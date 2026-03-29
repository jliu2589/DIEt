package user_settings

import (
	"context"
	"fmt"
	"strings"

	"diet/internal/models"
	"diet/internal/repositories"
)

type Service struct {
	repo *repositories.UserSettingsRepository
}

type UpsertInput struct {
	UserID       string
	Name         *string
	HeightCM     *float64
	WeightGoalKG *float64
	CalorieGoal  *float64
	ProteinGoalG *float64
	WeightUnit   string
}

func NewService(repo *repositories.UserSettingsRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByUserID(ctx context.Context, userID string) (*models.UserSettings, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	return s.repo.GetByUserID(ctx, userID)
}

func (s *Service) Upsert(ctx context.Context, input UpsertInput) (*models.UserSettings, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	weightUnit := strings.TrimSpace(input.WeightUnit)
	if weightUnit == "" {
		weightUnit = "kg"
	}

	payload := models.UserSettings{
		UserID:       userID,
		Name:         trimStringPtr(input.Name),
		HeightCM:     input.HeightCM,
		WeightGoalKG: input.WeightGoalKG,
		CalorieGoal:  input.CalorieGoal,
		ProteinGoalG: input.ProteinGoalG,
		WeightUnit:   weightUnit,
	}

	return s.repo.Upsert(ctx, payload)
}

func trimStringPtr(value *string) *string {
	if value == nil {
		return nil
	}

	trimmed := strings.TrimSpace(*value)
	if trimmed == "" {
		return nil
	}

	return &trimmed
}
