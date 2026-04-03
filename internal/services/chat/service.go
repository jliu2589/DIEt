package chat

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	inputclassifier "diet/internal/services/input_classifier"
	mealservice "diet/internal/services/meal"
	weightservice "diet/internal/services/weight"
)

type MealProcessor interface {
	ProcessTextMeal(ctx context.Context, input mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error)
}

type InputClassifier interface {
	Classify(rawText string) string
}

type WeightProcessor interface {
	CreateEntry(ctx context.Context, input weightservice.CreateEntryInput) (*weightserviceEntry, error)
}

type weightserviceEntry struct {
	Weight   float64
	Unit     string
	LoggedAt time.Time
}

type weightProcessorAdapter struct {
	inner *weightservice.Service
}

func (w weightProcessorAdapter) CreateEntry(ctx context.Context, input weightservice.CreateEntryInput) (*weightserviceEntry, error) {
	entry, err := w.inner.CreateEntry(ctx, input)
	if err != nil || entry == nil {
		return nil, err
	}
	return &weightserviceEntry{Weight: entry.Weight, Unit: entry.Unit, LoggedAt: entry.LoggedAt}, nil
}

type Service struct {
	mealService mealProcessor
	classifier  InputClassifier
	weightSvc   WeightProcessor
}

type mealProcessor interface {
	ProcessTextMeal(ctx context.Context, input mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error)
}

type Request struct {
	UserID   string
	Message  string
	LoggedAt *time.Time
}

type Response struct {
	Intent               string                `json:"intent"`
	MessageToUser        string                `json:"message_to_user"`
	MealResult           *MealResult           `json:"meal_result,omitempty"`
	WeightResult         *WeightResult         `json:"weight_result,omitempty"`
	RecommendationResult *RecommendationResult `json:"recommendation_result,omitempty"`
}

type MealResult struct {
	MealEventID     int64     `json:"meal_event_id"`
	CanonicalName   string    `json:"canonical_name"`
	LoggedAt        time.Time `json:"logged_at"`
	EatenAt         time.Time `json:"eaten_at"`
	TimeSource      string    `json:"time_source"`
	Source          string    `json:"source"`
	ConfidenceScore *float64  `json:"confidence_score,omitempty"`
	CaloriesKcal    *float64  `json:"calories_kcal,omitempty"`
	ProteinG        *float64  `json:"protein_g,omitempty"`
	CarbohydrateG   *float64  `json:"carbohydrate_g,omitempty"`
	FatG            *float64  `json:"fat_g,omitempty"`
}

type WeightResult struct {
	Weight   float64   `json:"weight"`
	Unit     string    `json:"unit"`
	LoggedAt time.Time `json:"logged_at"`
}

type RecommendationResult struct {
	Text string `json:"text"`
}

func NewService(mealService MealProcessor, classifier InputClassifier, weightSvc *weightservice.Service) *Service {
	if classifier == nil {
		classifier = inputclassifier.NewService()
	}
	var adapted WeightProcessor
	if weightSvc != nil {
		adapted = weightProcessorAdapter{inner: weightSvc}
	}
	return &Service{mealService: mealService, classifier: classifier, weightSvc: adapted}
}

var weightPattern = regexp.MustCompile(`(?i)(\d{2,3}(?:\.\d+)?)\s*(kg|kgs|kilograms?|lb|lbs|pounds?)\b`)

func parseWeightFromMessage(message string) (float64, string, bool) {
	matches := weightPattern.FindStringSubmatch(message)
	if len(matches) != 3 {
		return 0, "", false
	}
	weight, err := strconv.ParseFloat(matches[1], 64)
	if err != nil || weight <= 0 {
		return 0, "", false
	}
	unitRaw := strings.ToLower(matches[2])
	if strings.HasPrefix(unitRaw, "kg") || strings.HasPrefix(unitRaw, "kilo") {
		return weight, "kg", true
	}
	return weight, "lb", true
}

func (s *Service) HandleMessage(ctx context.Context, req Request) (*Response, error) {
	userID := strings.TrimSpace(req.UserID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	message := strings.TrimSpace(req.Message)
	if message == "" {
		return nil, fmt.Errorf("message is required")
	}

	intent := inputclassifier.IntentMealLog
	if s.classifier != nil {
		intent = s.classifier.Classify(message)
	}

	switch intent {
	case inputclassifier.IntentWeightLog:
		if s.weightSvc == nil {
			return &Response{Intent: "weight_log", MessageToUser: "Weight logging is unavailable right now."}, nil
		}
		weight, unit, ok := parseWeightFromMessage(message)
		if !ok {
			return &Response{Intent: "weight_log", MessageToUser: "Please include your weight like '176 lb' or '80 kg'."}, nil
		}
		loggedAt := time.Now().UTC()
		if req.LoggedAt != nil && !req.LoggedAt.IsZero() {
			loggedAt = req.LoggedAt.UTC()
		}
		entry, err := s.weightSvc.CreateEntry(ctx, weightservice.CreateEntryInput{
			UserID:   userID,
			Weight:   weight,
			Unit:     unit,
			LoggedAt: loggedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("process weight message: %w", err)
		}
		return &Response{
			Intent:        "weight_log",
			MessageToUser: "Logged your weight.",
			WeightResult: &WeightResult{
				Weight:   entry.Weight,
				Unit:     entry.Unit,
				LoggedAt: entry.LoggedAt.UTC(),
			},
		}, nil

	case inputclassifier.IntentGeneralChat:
		return &Response{Intent: "general_chat", MessageToUser: "I can log meals and weight. Try: 'chicken bowl' or '176 lb'."}, nil
	case inputclassifier.IntentRecommendationRequest:
		return &Response{Intent: "recommendation_request", MessageToUser: "Recommendations are not wired in chat yet."}, nil
	case inputclassifier.IntentUnknown:
		return &Response{Intent: "unknown", MessageToUser: "I couldn’t classify that. Try a meal or a weight entry."}, nil
	}

	mealResult, err := s.mealService.ProcessTextMeal(ctx, mealservice.ProcessTextMealInput{
		UserID:  userID,
		Source:  "chat",
		RawText: message,
	})
	if err != nil {
		return nil, fmt.Errorf("process meal message: %w", err)
	}
	if mealResult == nil || !mealResult.Logged {
		return &Response{Intent: "unknown", MessageToUser: "I couldn’t log that as a meal."}, nil
	}
	return &Response{
		Intent:        "meal_log",
		MessageToUser: "Logged your meal.",
		MealResult: &MealResult{
			MealEventID:     mealResult.MealEventID,
			CanonicalName:   mealResult.CanonicalName,
			LoggedAt:        mealResult.LoggedAt.UTC(),
			EatenAt:         mealResult.EatenAt.UTC(),
			TimeSource:      mealResult.TimeSource,
			Source:          mealResult.Source,
			ConfidenceScore: mealResult.ConfidenceScore,
			CaloriesKcal:    mealResult.Nutrition.CaloriesKcal,
			ProteinG:        mealResult.Nutrition.ProteinG,
			CarbohydrateG:   mealResult.Nutrition.CarbohydrateG,
			FatG:            mealResult.Nutrition.FatG,
		},
	}, nil
}
