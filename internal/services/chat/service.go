package chat

import (
	"context"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"

	"diet/internal/models"
	inputclassifier "diet/internal/services/input_classifier"
	mealservice "diet/internal/services/meal"
	weightservice "diet/internal/services/weight"
)

type Classifier interface {
	Classify(rawText string) string
}

type MealProcessor interface {
	ProcessTextMeal(ctx context.Context, input mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error)
}

type WeightLogger interface {
	CreateEntry(ctx context.Context, input weightservice.CreateEntryInput) (*models.WeightEntry, error)
}

type Service struct {
	classifier    Classifier
	mealService   MealProcessor
	weightService WeightLogger
}

const nonDietHelpMessage = "I can help log meals, log weight, and suggest meals to hit your goals."

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

var weightPattern = regexp.MustCompile(`(?i)\b(\d{2,3}(?:\.\d+)?)\s*(kg|kgs|kilograms?|lb|lbs|pounds?)\b`)

func NewService(classifier Classifier, mealService MealProcessor, weightService WeightLogger) *Service {
	return &Service{classifier: classifier, mealService: mealService, weightService: weightService}
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

	intent := inputclassifier.IntentUnknown
	if s.classifier != nil {
		intent = s.classifier.Classify(message)
	}

	switch intent {
	case inputclassifier.IntentMealLog:
		mealResult, err := s.mealService.ProcessTextMeal(ctx, mealservice.ProcessTextMealInput{
			UserID:  userID,
			Source:  "chat",
			RawText: message,
		})
		if err != nil {
			return nil, fmt.Errorf("process meal message: %w", err)
		}
		if mealResult == nil || !mealResult.Logged {
			return &Response{Intent: intent, MessageToUser: "I couldn’t log that as a meal."}, nil
		}
		return &Response{
			Intent:        intent,
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
	case inputclassifier.IntentWeightLog:
		weightValue, unit, ok := parseWeightFromText(message)
		if !ok {
			return &Response{Intent: intent, MessageToUser: "I detected a weight log, but I couldn’t parse the value."}, nil
		}
		loggedAt := time.Now().UTC()
		if req.LoggedAt != nil && !req.LoggedAt.IsZero() {
			loggedAt = req.LoggedAt.UTC()
		}
		entry, err := s.weightService.CreateEntry(ctx, weightservice.CreateEntryInput{
			UserID:   userID,
			Weight:   weightValue,
			Unit:     unit,
			LoggedAt: loggedAt,
		})
		if err != nil {
			return nil, fmt.Errorf("log weight message: %w", err)
		}
		return &Response{
			Intent:        intent,
			MessageToUser: "Logged your weight entry.",
			WeightResult:  &WeightResult{Weight: entry.Weight, Unit: entry.Unit, LoggedAt: entry.LoggedAt.UTC()},
		}, nil
	case inputclassifier.IntentRecommendationRequest:
		return &Response{
			Intent:        intent,
			MessageToUser: "Got it — I can help with recommendations.",
			RecommendationResult: &RecommendationResult{
				Text: "Try a high-protein meal with lean protein, vegetables, and whole grains.",
			},
		}, nil
	case inputclassifier.IntentGeneralChat:
		return &Response{Intent: intent, MessageToUser: nonDietHelpMessage}, nil
	default:
		return &Response{Intent: inputclassifier.IntentUnknown, MessageToUser: nonDietHelpMessage}, nil
	}
}

func parseWeightFromText(text string) (float64, string, bool) {
	m := weightPattern.FindStringSubmatch(strings.ToLower(text))
	if len(m) != 3 {
		return 0, "", false
	}
	value, err := strconv.ParseFloat(m[1], 64)
	if err != nil {
		return 0, "", false
	}
	unit := normalizeWeightUnit(m[2])
	if unit == "" {
		return 0, "", false
	}
	return value, unit, true
}

func normalizeWeightUnit(raw string) string {
	raw = strings.ToLower(strings.TrimSpace(raw))
	switch raw {
	case "kg", "kgs", "kilogram", "kilograms":
		return "kg"
	case "lb", "lbs", "pound", "pounds":
		return "lb"
	default:
		return ""
	}
}
