package meal_storage

import (
	"context"
	"fmt"
	"strings"

	"diet/internal/models"
	"diet/internal/repositories"
)

// Service provides reusable meal template/interpreted-meal storage helpers.
type Service struct {
	mealsRepo     *repositories.MealsRepository
	mealItemsRepo *repositories.MealItemsRepository
}

func NewService(mealsRepo *repositories.MealsRepository, mealItemsRepo *repositories.MealItemsRepository) *Service {
	return &Service{mealsRepo: mealsRepo, mealItemsRepo: mealItemsRepo}
}

type StoreMealInput struct {
	CanonicalName   string
	FingerprintHash *string
	SourceType      *string
	ConfidenceScore *float64
	Items           []models.StoredMealItem
}

type MealWithItems struct {
	Meal  *models.Meal            `json:"meal"`
	Items []models.StoredMealItem `json:"items"`
}

// Store inserts a meal and its meal_items for reusable caching/recommendation use-cases.
func (s *Service) Store(ctx context.Context, in StoreMealInput) (*MealWithItems, error) {
	if strings.TrimSpace(in.CanonicalName) == "" {
		return nil, fmt.Errorf("canonical_name is required")
	}

	meal, err := s.mealsRepo.Create(ctx, models.Meal{
		CanonicalName:   strings.TrimSpace(in.CanonicalName),
		FingerprintHash: in.FingerprintHash,
		SourceType:      in.SourceType,
		ConfidenceScore: in.ConfidenceScore,
	})
	if err != nil {
		return nil, fmt.Errorf("create meal: %w", err)
	}

	createdItems := make([]models.StoredMealItem, 0, len(in.Items))
	for _, it := range in.Items {
		item, err := s.mealItemsRepo.Create(ctx, models.StoredMealItem{
			MealID:   meal.ID,
			FoodID:   it.FoodID,
			Quantity: it.Quantity,
			Unit:     strings.TrimSpace(it.Unit),
		})
		if err != nil {
			return nil, fmt.Errorf("create meal item for meal_id=%d: %w", meal.ID, err)
		}
		createdItems = append(createdItems, *item)
	}

	return &MealWithItems{Meal: meal, Items: createdItems}, nil
}

func (s *Service) GetByFingerprintHash(ctx context.Context, fingerprintHash string) (*MealWithItems, error) {
	meal, err := s.mealsRepo.GetByFingerprintHash(ctx, fingerprintHash)
	if err != nil {
		return nil, fmt.Errorf("get meal by fingerprint_hash: %w", err)
	}
	if meal == nil {
		return nil, nil
	}

	items, err := s.mealItemsRepo.ListByMealID(ctx, meal.ID)
	if err != nil {
		return nil, fmt.Errorf("list meal items: %w", err)
	}

	return &MealWithItems{Meal: meal, Items: items}, nil
}
