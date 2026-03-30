package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"diet/internal/repositories"
	mealservice "diet/internal/services/meal"
	"github.com/gin-gonic/gin"
)

type MealHandler struct {
	service         *mealservice.Service
	idempotencyRepo *repositories.IdempotencyKeysRepository
}

func NewMealHandler(service *mealservice.Service, idempotencyRepo *repositories.IdempotencyKeysRepository) *MealHandler {
	return &MealHandler{service: service, idempotencyRepo: idempotencyRepo}
}

type createMealRequest struct {
	UserID         string    `json:"user_id"`
	Source         string    `json:"source" binding:"required"`
	RawText        string    `json:"raw_text" binding:"required"`
	EatenAt        time.Time `json:"eaten_at"`
	IdempotencyKey string    `json:"idempotency_key,omitempty"`
}

type createMealResponse struct {
	Intent  string            `json:"intent"`
	Logged  bool              `json:"logged"`
	Message string            `json:"message"`
	Item    *mealResponseItem `json:"item,omitempty"`
}

type recentMealsResponse struct {
	Items []mealResponseItem `json:"items"`
}

type mealResponseItem struct {
	MealEventID     int64    `json:"meal_event_id"`
	CanonicalName   string   `json:"canonical_name"`
	LoggedAt        string   `json:"logged_at"`
	EatenAt         string   `json:"eaten_at"`
	TimeSource      string   `json:"time_source"`
	Source          string   `json:"source"`
	ConfidenceScore *float64 `json:"confidence_score,omitempty"`
	CaloriesKcal    *float64 `json:"calories_kcal"`
	ProteinG        *float64 `json:"protein_g"`
	CarbohydrateG   *float64 `json:"carbohydrate_g"`
	FatG            *float64 `json:"fat_g"`
}

type editMealTimeRequest struct {
	UserID  string    `json:"user_id"`
	EatenAt time.Time `json:"eaten_at" binding:"required"`
}

type editMealTimeResponse struct {
	MealEventID   int64  `json:"meal_event_id"`
	CanonicalName string `json:"canonical_name"`
	EatenAt       string `json:"eaten_at"`
	TimeSource    string `json:"time_source"`
}

func (h *MealHandler) CreateMeal(c *gin.Context) {
	var req createMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, ok := resolveUserID(c, req.UserID, false)
	if !ok || strings.TrimSpace(req.RawText) == "" || strings.TrimSpace(req.Source) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "source and raw_text are required"})
		return
	}

	idempotencyKey := resolveIdempotencyKey(c, req.IdempotencyKey)
	idempotencyRecordID, handled := beginIdempotency(c, h.idempotencyRepo, userID, "POST:/v1/meals", idempotencyKey, req)
	if handled {
		return
	}

	result, err := h.service.ProcessTextMeal(c.Request.Context(), mealservice.ProcessTextMealInput{
		UserID:  userID,
		Source:  req.Source,
		RawText: req.RawText,
		EatenAt: req.EatenAt,
	})
	if err != nil {
		cleanupIdempotencyOnError(c, h.idempotencyRepo, idempotencyRecordID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process meal"})
		return
	}

	resp := createMealResponse{
		Intent:  result.Intent,
		Logged:  result.Logged,
		Message: result.Message,
	}
	if result.Logged {
		item := toMealResponseItemFromCreate(result)
		resp.Item = &item
	}

	saveIdempotencySuccess(c, h.idempotencyRepo, idempotencyRecordID, http.StatusOK, resp)
	c.JSON(http.StatusOK, resp)
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

func (h *MealHandler) EditMealTime(c *gin.Context) {
	mealEventID, err := strconv.ParseInt(strings.TrimSpace(c.Param("mealEventID")), 10, 64)
	if err != nil || mealEventID <= 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "mealEventID must be a positive integer"})
		return
	}

	var req editMealTimeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, ok := resolveUserID(c, req.UserID, false)
	if !ok {
		return
	}

	updated, err := h.service.EditMealTime(c.Request.Context(), mealEventID, userID, req.EatenAt)
	if err != nil {
		if strings.Contains(err.Error(), "is required") || strings.Contains(err.Error(), "must be greater than 0") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to edit meal time"})
		return
	}
	if updated == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "meal event not found"})
		return
	}

	c.JSON(http.StatusOK, editMealTimeResponse{
		MealEventID:   updated.MealEventID,
		CanonicalName: updated.CanonicalName,
		EatenAt:       updated.EatenAt.UTC().Format(time.RFC3339),
		TimeSource:    updated.TimeSource,
	})
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
		LoggedAt:        result.LoggedAt.UTC().Format(time.RFC3339),
		EatenAt:         result.EatenAt.UTC().Format(time.RFC3339),
		TimeSource:      result.TimeSource,
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
		LoggedAt:      item.LoggedAt.UTC().Format(time.RFC3339),
		EatenAt:       item.EatenAt.UTC().Format(time.RFC3339),
		TimeSource:    item.TimeSource,
		Source:        item.Source,
		CaloriesKcal:  item.CaloriesKcal,
		ProteinG:      item.ProteinG,
		CarbohydrateG: item.CarbohydrateG,
		FatG:          item.FatG,
	}
}
