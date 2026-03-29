package weight

import (
	"context"
	"fmt"
	"strings"
	"time"

	"diet/internal/models"
	"diet/internal/repositories"
)

const (
	defaultRecentLimit = 30
	maxRecentLimit     = 100
)

type Service struct {
	repo *repositories.WeightEntriesRepository
}

type CreateEntryInput struct {
	UserID   string
	Weight   float64
	Unit     string
	LoggedAt time.Time
}

func NewService(repo *repositories.WeightEntriesRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) CreateEntry(ctx context.Context, input CreateEntryInput) (*models.WeightEntry, error) {
	userID := strings.TrimSpace(input.UserID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if input.Weight <= 0 {
		return nil, fmt.Errorf("weight must be greater than 0")
	}

	unit := strings.ToLower(strings.TrimSpace(input.Unit))
	if unit == "" {
		unit = "kg"
	}
	if unit != "kg" && unit != "lb" {
		return nil, fmt.Errorf("unit must be one of: kg, lb")
	}
	if input.LoggedAt.IsZero() {
		return nil, fmt.Errorf("logged_at is required")
	}

	return s.repo.Insert(ctx, models.WeightEntry{
		UserID:   userID,
		Weight:   input.Weight,
		Unit:     unit,
		LoggedAt: input.LoggedAt.UTC(),
	})
}

func (s *Service) GetLatestEntry(ctx context.Context, userID string) (*models.WeightEntry, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	return s.repo.GetLatestByUserID(ctx, userID)
}

func (s *Service) GetRecentEntries(ctx context.Context, userID string, limit int) ([]models.WeightEntry, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	switch {
	case limit <= 0:
		limit = defaultRecentLimit
	case limit > maxRecentLimit:
		limit = maxRecentLimit
	}

	return s.repo.ListRecentByUserID(ctx, userID, limit)
}
