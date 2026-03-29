package user_state

import (
	"context"
	"fmt"
	"strings"
	"time"

	"diet/internal/models"
	"diet/internal/repositories"
)

type Service struct {
	userSettingsRepo  *repositories.UserSettingsRepository
	weightEntriesRepo *repositories.WeightEntriesRepository
}

type State struct {
	Settings     *SettingsState     `json:"settings"`
	LatestWeight *LatestWeightState `json:"latest_weight"`
}

type SettingsState struct {
	Name         *string  `json:"name,omitempty"`
	HeightCM     *float64 `json:"height_cm,omitempty"`
	WeightGoalKG *float64 `json:"weight_goal_kg,omitempty"`
	CalorieGoal  *float64 `json:"calorie_goal,omitempty"`
	ProteinGoalG *float64 `json:"protein_goal_g,omitempty"`
	WeightUnit   string   `json:"weight_unit"`
}

type LatestWeightState struct {
	Weight   float64 `json:"weight"`
	Unit     string  `json:"unit"`
	LoggedAt string  `json:"logged_at"`
}

func NewService(userSettingsRepo *repositories.UserSettingsRepository, weightEntriesRepo *repositories.WeightEntriesRepository) *Service {
	return &Service{userSettingsRepo: userSettingsRepo, weightEntriesRepo: weightEntriesRepo}
}

func (s *Service) GetByUserID(ctx context.Context, userID string) (*State, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	settings, err := s.userSettingsRepo.GetByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get user settings: %w", err)
	}

	latestWeight, err := s.weightEntriesRepo.GetLatestByUserID(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("get latest weight: %w", err)
	}

	return &State{
		Settings:     toSettingsState(settings),
		LatestWeight: toLatestWeightState(latestWeight),
	}, nil
}

func toSettingsState(settings *models.UserSettings) *SettingsState {
	if settings == nil {
		return nil
	}

	return &SettingsState{
		Name:         settings.Name,
		HeightCM:     settings.HeightCM,
		WeightGoalKG: settings.WeightGoalKG,
		CalorieGoal:  settings.CalorieGoal,
		ProteinGoalG: settings.ProteinGoalG,
		WeightUnit:   settings.WeightUnit,
	}
}

func toLatestWeightState(entry *models.WeightEntry) *LatestWeightState {
	if entry == nil {
		return nil
	}

	return &LatestWeightState{
		Weight:   entry.Weight,
		Unit:     entry.Unit,
		LoggedAt: entry.LoggedAt.UTC().Format(time.RFC3339),
	}
}
