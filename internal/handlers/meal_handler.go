package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"diet/internal/models"
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
	MealEventID     int64                  `json:"meal_event_id"`
	ProcessedFrom   string                 `json:"processed_from"`
	CanonicalName   string                 `json:"canonical_name"`
	Nutrition       models.NutritionFields `json:"nutrition"`
	ConfidenceScore *float64               `json:"confidence_score,omitempty"`
}

type recentMealsResponse struct {
	Items []recentMealItem `json:"items"`
}

type recentMealItem struct {
	MealEventID   int64   `json:"meal_event_id"`
	CanonicalName string  `json:"canonical_name"`
	EatenAt       string  `json:"eaten_at"`
	CaloriesKcal  float64 `json:"calories_kcal"`
	ProteinG      float64 `json:"protein_g"`
	CarbohydrateG float64 `json:"carbohydrate_g"`
	FatG          float64 `json:"fat_g"`
	Source        string  `json:"source"`
}

func (h *MealHandler) CreateMeal(c *gin.Context) {
	var req createMealRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.UserID) == "" || strings.TrimSpace(req.RawText) == "" || strings.TrimSpace(req.Source) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id, source, and raw_text are required"})
		return
	}

	result, err := h.service.ProcessTextMeal(c.Request.Context(), mealservice.ProcessTextMealInput{
		UserID:  req.UserID,
		Source:  req.Source,
		RawText: req.RawText,
		EatenAt: req.EatenAt,
	})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process meal"})
		return
	}

	c.JSON(http.StatusOK, createMealResponse{
		MealEventID:     result.MealEventID,
		ProcessedFrom:   result.Source,
		CanonicalName:   result.CanonicalName,
		Nutrition:       result.Nutrition,
		ConfidenceScore: result.ConfidenceScore,
	})
}

func (h *MealHandler) GetRecentMeals(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
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

	respItems := make([]recentMealItem, 0, len(items))
	for _, item := range items {
		respItems = append(respItems, recentMealItem{
			MealEventID:   item.MealEventID,
			CanonicalName: item.CanonicalName,
			EatenAt:       item.EatenAt.UTC().Format(time.RFC3339),
			CaloriesKcal:  floatPtrOrZero(item.CaloriesKcal),
			ProteinG:      floatPtrOrZero(item.ProteinG),
			CarbohydrateG: floatPtrOrZero(item.CarbohydrateG),
			FatG:          floatPtrOrZero(item.FatG),
			Source:        item.Source,
		})
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

func floatPtrOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
