package recommendations

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"diet/internal/repositories"
)

const (
	defaultCandidateLimit = 30
	defaultItemsLimit     = 10
)

type Service struct {
	userSettingsRepo *repositories.UserSettingsRepository
	summaryRepo      *repositories.DailyNutritionSummaryRepository
	mealMemoryRepo   *repositories.MealMemoryRepository
	phraser          Phraser
	aiPhrasingOn     bool
}

type Response struct {
	CaloriesRemaining float64              `json:"calories_remaining"`
	ProteinRemainingG float64              `json:"protein_remaining_g"`
	MessageToUser     string               `json:"message_to_user,omitempty"`
	Items             []RecommendationItem `json:"items"`
}

type Phraser interface {
	PhraseRecommendations(ctx context.Context, input PhraseInput) (string, error)
}

type PhraseInput struct {
	CaloriesRemaining float64
	ProteinRemainingG float64
	Items             []RecommendationItem
}

type RecommendationItem struct {
	MealID              int64   `json:"meal_id"`
	CanonicalName       string  `json:"canonical_name"`
	CaloriesKcal        float64 `json:"calories_kcal"`
	ProteinG            float64 `json:"protein_g"`
	CarbohydrateG       float64 `json:"carbohydrate_g"`
	FatG                float64 `json:"fat_g"`
	RecommendationScore float64 `json:"recommendation_score"`
}

func NewService(userSettingsRepo *repositories.UserSettingsRepository, summaryRepo *repositories.DailyNutritionSummaryRepository, mealMemoryRepo *repositories.MealMemoryRepository) *Service {
	return &Service{
		userSettingsRepo: userSettingsRepo,
		summaryRepo:      summaryRepo,
		mealMemoryRepo:   mealMemoryRepo,
		aiPhrasingOn:     true,
	}
}

func (s *Service) GetForUserToday(ctx context.Context, userID string, now time.Time) (*Response, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	settings, err := s.userSettingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user settings: %w", err)
	}

	today := time.Date(now.UTC().Year(), now.UTC().Month(), now.UTC().Day(), 0, 0, 0, 0, time.UTC)
	summary, err := s.summaryRepo.GetByUserIDAndDate(ctx, userID, today)
	if err != nil {
		return nil, fmt.Errorf("get daily summary: %w", err)
	}

	calorieGoal := valueOrZero(nil)
	proteinGoal := valueOrZero(nil)
	if settings != nil {
		calorieGoal = valueOrZero(settings.CalorieGoal)
		proteinGoal = valueOrZero(settings.ProteinGoalG)
	}

	caloriesToday := 0.0
	proteinToday := 0.0
	if summary != nil {
		caloriesToday = valueOrZero(summary.CaloriesKcal)
		proteinToday = valueOrZero(summary.ProteinG)
	}

	caloriesRemaining := math.Max(calorieGoal-caloriesToday, 0)
	proteinRemaining := math.Max(proteinGoal-proteinToday, 0)

	candidates, err := s.mealMemoryRepo.ListRecommendationCandidates(ctx, defaultCandidateLimit)
	if err != nil {
		return nil, fmt.Errorf("get recommendation candidates: %w", err)
	}

	items := make([]RecommendationItem, 0, len(candidates))
	for _, c := range candidates {
		cal := valueOrZero(c.CaloriesKcal)
		pro := valueOrZero(c.ProteinG)
		score := recommendationScore(caloriesRemaining, proteinRemaining, cal, pro)
		items = append(items, RecommendationItem{
			MealID:              c.MealID,
			CanonicalName:       c.CanonicalName,
			CaloriesKcal:        cal,
			ProteinG:            pro,
			CarbohydrateG:       valueOrZero(c.CarbohydrateG),
			FatG:                valueOrZero(c.FatG),
			RecommendationScore: score,
		})
	}

	sort.Slice(items, func(i, j int) bool {
		if items[i].RecommendationScore == items[j].RecommendationScore {
			return items[i].MealID < items[j].MealID
		}
		return items[i].RecommendationScore > items[j].RecommendationScore
	})

	if len(items) > defaultItemsLimit {
		items = items[:defaultItemsLimit]
	}

	response := &Response{
		CaloriesRemaining: caloriesRemaining,
		ProteinRemainingG: proteinRemaining,
		MessageToUser:     defaultExplanation(caloriesRemaining, proteinRemaining),
		Items:             items,
	}

	if s.aiPhrasingOn && s.phraser != nil {
		if message, err := s.phraser.PhraseRecommendations(ctx, PhraseInput{
			CaloriesRemaining: caloriesRemaining,
			ProteinRemainingG: proteinRemaining,
			Items:             items,
		}); err == nil && strings.TrimSpace(message) != "" {
			response.MessageToUser = strings.TrimSpace(message)
		}
	}

	return response, nil
}

func recommendationScore(calRemain, proRemain, cal, pro float64) float64 {
	dCal := math.Abs(calRemain - cal)
	dPro := math.Abs(proRemain - pro)
	distance := dCal + (dPro * 10)
	return 1 / (1 + distance)
}

func valueOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}

func defaultExplanation(caloriesRemaining, proteinRemaining float64) string {
	return fmt.Sprintf(
		"You still have about %.0f calories and %.0fg protein left. These meals would help you close the gap.",
		caloriesRemaining,
		proteinRemaining,
	)
}

func (s *Service) SetPhraser(phraser Phraser) {
	s.phraser = phraser
}

func (s *Service) SetAIPhrasingEnabled(enabled bool) {
	s.aiPhrasingOn = enabled
}
