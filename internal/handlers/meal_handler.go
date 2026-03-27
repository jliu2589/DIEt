package handlers

import (
	"net/http"
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
