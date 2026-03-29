package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	mealservice "diet/internal/services/meal"
	"github.com/gin-gonic/gin"
)

type MealHandler struct {
	service *mealservice.Service
}

func NewMealHandler(service *mealservice.Service) *MealHandler {
	return &MealHandler{service: service}
}

type createMealRequest struct {
	UserID  string    `json:"user_id" binding:"required"`
	Source  string    `json:"source" binding:"required"`
	RawText string    `json:"raw_text" binding:"required"`
	EatenAt time.Time `json:"eaten_at" binding:"required"`
}

type createMealResponse struct {
	Item mealResponseItem `json:"item"`
}

type recentMealsResponse struct {
	Items []mealResponseItem `json:"items"`
}

type mealResponseItem struct {
	MealEventID     int64    `json:"meal_event_id"`
	CanonicalName   string   `json:"canonical_name"`
	EatenAt         string   `json:"eaten_at"`
	Source          string   `json:"source"`
	ConfidenceScore *float64 `json:"confidence_score,omitempty"`
	CaloriesKcal    *float64 `json:"calories_kcal"`
	ProteinG        *float64 `json:"protein_g"`
	CarbohydrateG   *float64 `json:"carbohydrate_g"`
	FatG            *float64 `json:"fat_g"`
}

func (h *MealHandler) CreateMeal(c *gin.Context) {
	var req createMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, ok := normalizeRequiredUserID(req.UserID)
	if !ok || strings.TrimSpace(req.RawText) == "" || strings.TrimSpace(req.Source) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, source, and raw_text are required"})
		return
	}

	result, err := h.service.ProcessTextMeal(c.Request.Context(), mealservice.ProcessTextMealInput{
		UserID:  userID,
		Source:  req.Source,
		RawText: req.RawText,
		EatenAt: req.EatenAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process meal"})
		return
	}

	c.JSON(http.StatusOK, createMealResponse{Item: toMealResponseItemFromCreate(result)})
}

func (h *MealHandler) GetRecentMeals(c *gin.Context) {
	userID, ok := requiredUserIDFromQuery(c)
	if !ok {
		return
	}

	limit, err := parseLimitQuery(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	items, err := h.service.GetRecentMeals(c.Request.Context(), userID, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recent meals"})
		return
	}

	respItems := make([]mealResponseItem, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, toMealResponseItemFromRecent(item))
	}

	c.JSON(http.StatusOK, recentMealsResponse{Items: respItems})
}

func parseLimitQuery(limitRaw string) (int, error) {
	trimmed := strings.TrimSpace(limitRaw)
	if trimmed == "" {
		return 0, nil
	}

	limit, err := strconv.Atoi(trimmed)
	if err != nil || limit < 0 {
		return 0, fmt.Errorf("limit must be a non-negative integer")
	}

	return limit, nil
}

func toMealResponseItemFromCreate(result *mealservice.ProcessTextMealResult) mealResponseItem {
	if result == nil {
		return mealResponseItem{}
	}

	return mealResponseItem{
		MealEventID:     result.MealEventID,
		CanonicalName:   result.CanonicalName,
		EatenAt:         result.EatenAt.UTC().Format(time.RFC3339),
		Source:          result.Source,
		ConfidenceScore: result.ConfidenceScore,
		CaloriesKcal:    result.Nutrition.CaloriesKcal,
		ProteinG:        result.Nutrition.ProteinG,
		CarbohydrateG:   result.Nutrition.CarbohydrateG,
		FatG:            result.Nutrition.FatG,
	}
}

func toMealResponseItemFromRecent(item mealservice.RecentMealResult) mealResponseItem {
	return mealResponseItem{
		MealEventID:   item.MealEventID,
		CanonicalName: item.CanonicalName,
		EatenAt:       item.EatenAt.UTC().Format(time.RFC3339),
		Source:        item.Source,
		CaloriesKcal:  item.CaloriesKcal,
		ProteinG:      item.ProteinG,
		CarbohydrateG: item.CarbohydrateG,
		FatG:          item.FatG,
	}
}
