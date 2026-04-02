package handlers

import (
	"net/http"
	"time"

	"diet/internal/models"
	"diet/internal/repositories"
	mealservice "diet/internal/services/meal"
	recommendationsservice "diet/internal/services/recommendations"
	"github.com/gin-gonic/gin"
)

type DashboardHandler struct {
	summaryRepo        *repositories.DailyNutritionSummaryRepository
	mealService        *mealservice.Service
	recommendationsSvc *recommendationsservice.Service
}

func NewDashboardHandler(summaryRepo *repositories.DailyNutritionSummaryRepository, mealService *mealservice.Service, recommendationsSvc *recommendationsservice.Service) *DashboardHandler {
	return &DashboardHandler{summaryRepo: summaryRepo, mealService: mealService, recommendationsSvc: recommendationsSvc}
}

type dashboardTodayResponse struct {
	UserID          string                           `json:"user_id"`
	Date            string                           `json:"date"`
	DailySummary    dailySummaryResponse             `json:"daily_summary"`
	RecentMeals     []mealResponseItem               `json:"recent_meals"`
	Recommendations *recommendationsservice.Response `json:"recommendations,omitempty"`
}

func (h *DashboardHandler) GetToday(c *gin.Context) {
	userID, ok := requiredUserIDFromQuery(c)
	if !ok {
		return
	}

	now := time.Now().UTC()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, time.UTC)
	if err := h.summaryRepo.ReconcileForUserDate(c.Request.Context(), userID, today); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to recalculate today summary"})
		return
	}

	summaryModel, err := h.summaryRepo.GetByUserIDAndDate(c.Request.Context(), userID, today)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch today summary"})
		return
	}
	nutrition := models.NutritionFields{}
	if summaryModel != nil {
		nutrition = summaryModel.NutritionFields
	}
	summary := newDailySummaryResponse(userID, today.Format("2006-01-02"), nutrition)

	recent, err := h.mealService.GetRecentMeals(c.Request.Context(), userID, 10)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch today meals"})
		return
	}
	recentItems := make([]mealResponseItem, 0, len(recent))
	for _, item := range recent {
		itemDay := item.EatenAt.UTC()
		if itemDay.Year() != today.Year() || itemDay.Month() != today.Month() || itemDay.Day() != today.Day() {
			continue
		}
		recentItems = append(recentItems, toMealResponseItemFromRecent(item))
	}

	recommendations, err := h.recommendationsSvc.GetForUserToday(c.Request.Context(), userID, now)
	if err != nil {
		recommendations = nil
	}

	c.JSON(http.StatusOK, dashboardTodayResponse{
		UserID:          userID,
		Date:            today.Format("2006-01-02"),
		DailySummary:    summary,
		RecentMeals:     recentItems,
		Recommendations: recommendations,
	})
}
