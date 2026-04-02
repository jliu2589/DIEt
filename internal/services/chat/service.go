package chat

import (
	"context"
	"fmt"
	"strings"
	"time"

	mealservice "diet/internal/services/meal"
)

type MealProcessor interface {
	ProcessTextMeal(ctx context.Context, input mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error)
}

type Service struct {
	mealService MealProcessor
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

func NewService(mealService MealProcessor) *Service {
	return &Service{mealService: mealService}
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
