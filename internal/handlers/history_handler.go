package handlers

import (
	"net/http"
	"time"

	"diet/internal/repositories"
	"github.com/gin-gonic/gin"
)

type HistoryHandler struct {
	mealEventsRepo *repositories.MealEventsRepository
}

func NewHistoryHandler(mealEventsRepo *repositories.MealEventsRepository) *HistoryHandler {
	return &HistoryHandler{mealEventsRepo: mealEventsRepo}
}

type historyWeeklyResponse struct {
	UserID    string              `json:"user_id"`
	StartDate string              `json:"start_date"`
	EndDate   string              `json:"end_date"`
	Days      []historyDayPayload `json:"days"`
}

type historyDayPayload struct {
	Date   string             `json:"date"`
	Totals historyTotals      `json:"totals"`
	Meals  []historyMealEntry `json:"meals"`
}

type historyTotals struct {
	CaloriesKcal  float64 `json:"calories_kcal"`
	ProteinG      float64 `json:"protein_g"`
	CarbohydrateG float64 `json:"carbohydrate_g"`
	FatG          float64 `json:"fat_g"`
}

type historyMealEntry struct {
	MealEventID   int64    `json:"meal_event_id"`
	CanonicalName string   `json:"canonical_name"`
	EatenAt       string   `json:"eaten_at"`
	CaloriesKcal  *float64 `json:"calories_kcal"`
	ProteinG      *float64 `json:"protein_g"`
	CarbohydrateG *float64 `json:"carbohydrate_g"`
	FatG          *float64 `json:"fat_g"`
}

func (h *HistoryHandler) GetWeekly(c *gin.Context) {
	userID, ok := requiredUserIDFromQuery(c)
	if !ok {
		return
	}

	startDateRaw := c.Query("start_date")
	var startDate time.Time
	var err error
	if startDateRaw == "" {
		now := time.Now().UTC()
		startDate = time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC).AddDate(0, 0, -6)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateRaw)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "start_date must be YYYY-MM-DD"})
			return
		}
		startDate = startDate.UTC()
	}

	endExclusive := startDate.AddDate(0, 0, 7)
	rows, err := h.mealEventsRepo.ListByUserIDAndDateRange(c.Request.Context(), userID, startDate, endExclusive)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch weekly history"})
		return
	}

	days := make([]historyDayPayload, 0, 7)
	byDate := make(map[string]*historyDayPayload, 7)
	for i := 0; i < 7; i++ {
		date := startDate.AddDate(0, 0, i)
		key := date.Format("2006-01-02")
		entry := historyDayPayload{Date: key, Meals: []historyMealEntry{}, Totals: historyTotals{}}
		days = append(days, entry)
		byDate[key] = &days[len(days)-1]
	}

	for _, row := range rows {
		key := row.EatenAt.UTC().Format("2006-01-02")
		day := byDate[key]
		if day == nil {
			continue
		}
		day.Meals = append(day.Meals, historyMealEntry{
			MealEventID:   row.MealEventID,
			CanonicalName: row.CanonicalName,
			EatenAt:       row.EatenAt.UTC().Format(time.RFC3339),
			CaloriesKcal:  row.CaloriesKcal,
			ProteinG:      row.ProteinG,
			CarbohydrateG: row.CarbohydrateG,
			FatG:          row.FatG,
		})
		day.Totals.CaloriesKcal += valueOrZero(row.CaloriesKcal)
		day.Totals.ProteinG += valueOrZero(row.ProteinG)
		day.Totals.CarbohydrateG += valueOrZero(row.CarbohydrateG)
		day.Totals.FatG += valueOrZero(row.FatG)
	}

	c.JSON(http.StatusOK, historyWeeklyResponse{
		UserID:    userID,
		StartDate: startDate.Format("2006-01-02"),
		EndDate:   endExclusive.AddDate(0, 0, -1).Format("2006-01-02"),
		Days:      days,
	})
}

func valueOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
