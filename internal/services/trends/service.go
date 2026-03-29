package trends

import (
	"context"
	"fmt"
	"strings"
	"time"

	"diet/internal/repositories"
)

type Service struct {
	weightRepo  *repositories.WeightEntriesRepository
	summaryRepo *repositories.DailyNutritionSummaryRepository
}

type Point struct {
	Date          string   `json:"date"`
	Weight        *float64 `json:"weight"`
	CaloriesKcal  *float64 `json:"calories_kcal"`
	ProteinG      *float64 `json:"protein_g"`
	CarbohydrateG *float64 `json:"carbohydrate_g"`
	FatG          *float64 `json:"fat_g"`
}

type Result struct {
	UserID string  `json:"user_id"`
	Range  string  `json:"range"`
	Points []Point `json:"points"`
}

func NewService(weightRepo *repositories.WeightEntriesRepository, summaryRepo *repositories.DailyNutritionSummaryRepository) *Service {
	return &Service{weightRepo: weightRepo, summaryRepo: summaryRepo}
}

func (s *Service) GetTrends(ctx context.Context, userID, rangeKey string, now time.Time) (*Result, error) {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	days, err := rangeToDays(rangeKey)
	if err != nil {
		return nil, err
	}

	now = now.UTC()
	endDate := dateOnly(now)
	startDate := endDate.AddDate(0, 0, -(days - 1))
	endExclusive := endDate.AddDate(0, 0, 1)

	weights, err := s.weightRepo.ListDailyLatestByUserIDAndDateRange(ctx, userID, startDate, endExclusive)
	if err != nil {
		return nil, fmt.Errorf("get weight trends: %w", err)
	}

	summaries, err := s.summaryRepo.ListByUserIDAndDateRange(ctx, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("get nutrition trends: %w", err)
	}

	weightByDate := make(map[string]*float64, len(weights))
	for _, item := range weights {
		value := item.Weight
		weightByDate[item.Date.Format("2006-01-02")] = &value
	}

	summaryByDate := make(map[string]repositories.DailyNutritionSummaryRow, len(summaries))
	for _, item := range summaries {
		summaryByDate[item.SummaryDate.Format("2006-01-02")] = item
	}

	points := make([]Point, 0, days)
	for d := startDate; !d.After(endDate); d = d.AddDate(0, 0, 1) {
		key := d.Format("2006-01-02")
		point := Point{Date: key, Weight: weightByDate[key]}
		if summary, ok := summaryByDate[key]; ok {
			point.CaloriesKcal = summary.CaloriesKcal
			point.ProteinG = summary.ProteinG
			point.CarbohydrateG = summary.CarbohydrateG
			point.FatG = summary.FatG
		}
		points = append(points, point)
	}

	return &Result{UserID: userID, Range: rangeKey, Points: points}, nil
}

func rangeToDays(rangeKey string) (int, error) {
	switch strings.TrimSpace(rangeKey) {
	case "7d":
		return 7, nil
	case "30d":
		return 30, nil
	case "90d":
		return 90, nil
	case "1y":
		return 365, nil
	default:
		return 0, fmt.Errorf("range must be one of: 7d, 30d, 90d, 1y")
	}
}

func dateOnly(t time.Time) time.Time {
	u := t.UTC()
	return time.Date(u.Year(), u.Month(), u.Day(), 0, 0, 0, 0, time.UTC)
}
