package recommendations

import (
	"context"
	"fmt"
	"log/slog"
	"math"
	"sort"
	"strings"
	"time"

	"diet/internal/repositories"
	canonicalfoodssvc "diet/internal/services/canonical_foods"
)

const (
	defaultCandidateLimit = 30
	defaultItemsLimit     = 10

	sourceMealMemory    = "meal_memory"
	sourceReusableMeal  = "reusable_meal"
	sourceCanonicalFood = "canonical_food"
)

type Service struct {
	userSettingsRepo *repositories.UserSettingsRepository
	summaryRepo      *repositories.DailyNutritionSummaryRepository
	mealMemoryRepo   *repositories.MealMemoryRepository
	mealsRepo        *repositories.MealsRepository
	mealItemsRepo    *repositories.MealItemsRepository
	canonicalFoods   *repositories.CanonicalFoodsRepository
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
	Source              string  `json:"source"`
	SourceID            int64   `json:"source_id"`
	RecommendationScore float64 `json:"recommendation_score"`
}

func NewService(
	userSettingsRepo *repositories.UserSettingsRepository,
	summaryRepo *repositories.DailyNutritionSummaryRepository,
	mealMemoryRepo *repositories.MealMemoryRepository,
	mealsRepo *repositories.MealsRepository,
	mealItemsRepo *repositories.MealItemsRepository,
	canonicalFoodsRepo *repositories.CanonicalFoodsRepository,
) *Service {
	return &Service{
		userSettingsRepo: userSettingsRepo,
		summaryRepo:      summaryRepo,
		mealMemoryRepo:   mealMemoryRepo,
		mealsRepo:        mealsRepo,
		mealItemsRepo:    mealItemsRepo,
		canonicalFoods:   canonicalFoodsRepo,
		aiPhrasingOn:     true,
	}
}

func (s *Service) GetForUserToday(ctx context.Context, userID string, now time.Time) (*Response, error) {
	started := time.Now()
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

	candidates, err := s.buildCandidates(ctx, caloriesRemaining, proteinRemaining)
	if err != nil {
		slog.Error("recommendations.generate_failed", "user_id", userID, "error", err)
		return nil, fmt.Errorf("get recommendation candidates: %w", err)
	}

	response := &Response{
		CaloriesRemaining: caloriesRemaining,
		ProteinRemainingG: proteinRemaining,
		MessageToUser:     defaultExplanation(caloriesRemaining, proteinRemaining),
		Items:             candidates,
	}

	if s.aiPhrasingOn && s.phraser != nil {
		if message, err := s.phraser.PhraseRecommendations(ctx, PhraseInput{
			CaloriesRemaining: caloriesRemaining,
			ProteinRemainingG: proteinRemaining,
			Items:             candidates,
		}); err == nil && strings.TrimSpace(message) != "" {
			response.MessageToUser = strings.TrimSpace(message)
		}
	}
	slog.Info("recommendations.generated",
		"user_id", userID,
		"calories_remaining", caloriesRemaining,
		"protein_remaining_g", proteinRemaining,
		"items_count", len(response.Items),
		"latency_ms", time.Since(started).Milliseconds(),
	)

	return response, nil
}

func (s *Service) buildCandidates(ctx context.Context, caloriesRemaining, proteinRemaining float64) ([]RecommendationItem, error) {
	items := make([]RecommendationItem, 0, defaultCandidateLimit*3)
	mealMemoryCount := 0
	reusableCount := 0
	canonicalCount := 0

	if s.mealMemoryRepo != nil {
		memoryCandidates, err := s.mealMemoryRepo.ListRecommendationCandidates(ctx, defaultCandidateLimit)
		if err != nil {
			return nil, err
		}
		for _, c := range memoryCandidates {
			cal := valueOrZero(c.CaloriesKcal)
			pro := valueOrZero(c.ProteinG)
			items = append(items, RecommendationItem{
				MealID:              c.MealID,
				CanonicalName:       c.CanonicalName,
				CaloriesKcal:        cal,
				ProteinG:            pro,
				CarbohydrateG:       valueOrZero(c.CarbohydrateG),
				FatG:                valueOrZero(c.FatG),
				Source:              sourceMealMemory,
				SourceID:            c.MealID,
				RecommendationScore: recommendationScore(caloriesRemaining, proteinRemaining, cal, pro),
			})
			mealMemoryCount++
		}
	}

	if s.mealsRepo != nil && s.mealItemsRepo != nil && s.canonicalFoods != nil {
		reusable, err := s.buildReusableMealCandidates(ctx, caloriesRemaining, proteinRemaining)
		if err != nil {
			return nil, err
		}
		items = append(items, reusable...)
		reusableCount = len(reusable)
	}

	if s.canonicalFoods != nil {
		candidates, err := s.canonicalFoods.ListRecommendationCandidates(ctx, defaultCandidateLimit)
		if err != nil {
			return nil, err
		}
		for _, c := range candidates {
			cal := valueOrZero(c.CaloriesKcal)
			pro := valueOrZero(c.ProteinG)
			items = append(items, RecommendationItem{
				MealID:              c.FoodID,
				CanonicalName:       c.CanonicalName,
				CaloriesKcal:        cal,
				ProteinG:            pro,
				CarbohydrateG:       valueOrZero(c.CarbohydrateG),
				FatG:                valueOrZero(c.FatG),
				Source:              sourceCanonicalFood,
				SourceID:            c.FoodID,
				RecommendationScore: recommendationScore(caloriesRemaining, proteinRemaining, cal, pro),
			})
			canonicalCount++
		}
	}

	items = dedupeByName(items)
	sort.Slice(items, func(i, j int) bool {
		if items[i].RecommendationScore == items[j].RecommendationScore {
			if sourcePriority(items[i].Source) == sourcePriority(items[j].Source) {
				return items[i].SourceID < items[j].SourceID
			}
			return sourcePriority(items[i].Source) < sourcePriority(items[j].Source)
		}
		return items[i].RecommendationScore > items[j].RecommendationScore
	})

	if len(items) > defaultItemsLimit {
		items = items[:defaultItemsLimit]
	}
	slog.Info("recommendations.candidate_path",
		"meal_memory_candidates", mealMemoryCount,
		"reusable_meal_candidates", reusableCount,
		"canonical_food_candidates", canonicalCount,
		"final_candidates", len(items),
	)
	return items, nil
}

func (s *Service) buildReusableMealCandidates(ctx context.Context, caloriesRemaining, proteinRemaining float64) ([]RecommendationItem, error) {
	mealRows, err := s.mealsRepo.ListCandidates(ctx, defaultCandidateLimit)
	if err != nil {
		return nil, fmt.Errorf("list reusable meal candidates: %w", err)
	}

	out := make([]RecommendationItem, 0, len(mealRows))
	for _, mealRow := range mealRows {
		storedItems, err := s.mealItemsRepo.ListByMealID(ctx, mealRow.ID)
		if err != nil {
			return nil, fmt.Errorf("list meal_items for reusable candidate: %w", err)
		}

		totalCal := 0.0
		totalProtein := 0.0
		totalCarb := 0.0
		totalFat := 0.0
		for _, item := range storedItems {
			food, err := s.canonicalFoods.GetByID(ctx, item.FoodID)
			if err != nil {
				return nil, fmt.Errorf("get canonical food for reusable candidate: %w", err)
			}
			if food == nil {
				continue
			}

			scaled, err := canonicalfoodssvc.ScaleNutrition(*food, item.Quantity, item.Unit)
			if err != nil {
				continue
			}
			totalCal += valueOrZero(scaled.Nutrition.CaloriesKcal)
			totalProtein += valueOrZero(scaled.Nutrition.ProteinG)
			totalCarb += valueOrZero(scaled.Nutrition.CarbohydrateG)
			totalFat += valueOrZero(scaled.Nutrition.FatG)
		}

		if totalCal == 0 && totalProtein == 0 && totalCarb == 0 && totalFat == 0 {
			continue
		}
		out = append(out, RecommendationItem{
			MealID:              mealRow.ID,
			CanonicalName:       mealRow.CanonicalName,
			CaloriesKcal:        totalCal,
			ProteinG:            totalProtein,
			CarbohydrateG:       totalCarb,
			FatG:                totalFat,
			Source:              sourceReusableMeal,
			SourceID:            mealRow.ID,
			RecommendationScore: recommendationScore(caloriesRemaining, proteinRemaining, totalCal, totalProtein),
		})
	}

	return out, nil
}

func sourcePriority(source string) int {
	switch source {
	case sourceReusableMeal:
		return 0
	case sourceMealMemory:
		return 1
	case sourceCanonicalFood:
		return 2
	default:
		return 3
	}
}

func dedupeByName(items []RecommendationItem) []RecommendationItem {
	byName := make(map[string]RecommendationItem, len(items))
	for _, item := range items {
		key := strings.ToLower(strings.TrimSpace(item.CanonicalName))
		if key == "" {
			continue
		}
		existing, ok := byName[key]
		if !ok || item.RecommendationScore > existing.RecommendationScore ||
			(item.RecommendationScore == existing.RecommendationScore && sourcePriority(item.Source) < sourcePriority(existing.Source)) {
			byName[key] = item
		}
	}

	out := make([]RecommendationItem, 0, len(byName))
	for _, item := range byName {
		out = append(out, item)
	}
	return out
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
