package handlers

import (
	"net/http"
	"strings"

	userSettingsService "diet/internal/services/user_settings"
	"github.com/gin-gonic/gin"
)

type UserSettingsHandler struct {
	service *userSettingsService.Service
}

func NewUserSettingsHandler(service *userSettingsService.Service) *UserSettingsHandler {
	return &UserSettingsHandler{service: service}
}

type upsertUserSettingsRequest struct {
	UserID       string   `json:"user_id" binding:"required"`
	Name         *string  `json:"name"`
	HeightCM     *float64 `json:"height_cm"`
	WeightGoalKG *float64 `json:"weight_goal_kg"`
	CalorieGoal  *float64 `json:"calorie_goal"`
	ProteinGoalG *float64 `json:"protein_goal_g"`
	WeightUnit   string   `json:"weight_unit" binding:"omitempty,oneof=kg lb"`
}

type userSettingsResponse struct {
	UserID       string   `json:"user_id"`
	Name         *string  `json:"name,omitempty"`
	HeightCM     *float64 `json:"height_cm,omitempty"`
	WeightGoalKG *float64 `json:"weight_goal_kg,omitempty"`
	CalorieGoal  *float64 `json:"calorie_goal,omitempty"`
	ProteinGoalG *float64 `json:"protein_goal_g,omitempty"`
	WeightUnit   string   `json:"weight_unit"`
}

func (h *UserSettingsHandler) GetUserSettings(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	settings, err := h.service.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		if isValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user settings"})
		return
	}

	if settings == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "user settings not found"})
		return
	}

	c.JSON(http.StatusOK, toUserSettingsResponse(settings.UserID, settings.Name, settings.HeightCM, settings.WeightGoalKG, settings.CalorieGoal, settings.ProteinGoalG, settings.WeightUnit))
}

func (h *UserSettingsHandler) UpsertUserSettings(c *gin.Context) {
	var req upsertUserSettingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if strings.TrimSpace(req.UserID) == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	settings, err := h.service.Upsert(c.Request.Context(), userSettingsService.UpsertInput{
		UserID:       req.UserID,
		Name:         req.Name,
		HeightCM:     req.HeightCM,
		WeightGoalKG: req.WeightGoalKG,
		CalorieGoal:  req.CalorieGoal,
		ProteinGoalG: req.ProteinGoalG,
		WeightUnit:   req.WeightUnit,
	})
	if err != nil {
		if isValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save user settings"})
		return
	}

	c.JSON(http.StatusOK, toUserSettingsResponse(settings.UserID, settings.Name, settings.HeightCM, settings.WeightGoalKG, settings.CalorieGoal, settings.ProteinGoalG, settings.WeightUnit))
}

func isValidationError(err error) bool {
	if err == nil {
		return false
	}

	return strings.Contains(err.Error(), "is required") ||
		strings.Contains(err.Error(), "weight_unit must be one of")
}

func toUserSettingsResponse(
	userID string,
	name *string,
	heightCM *float64,
	weightGoalKG *float64,
	calorieGoal *float64,
	proteinGoalG *float64,
	weightUnit string,
) userSettingsResponse {
	return userSettingsResponse{
		UserID:       userID,
		Name:         name,
		HeightCM:     heightCM,
		WeightGoalKG: weightGoalKG,
		CalorieGoal:  calorieGoal,
		ProteinGoalG: proteinGoalG,
		WeightUnit:   weightUnit,
	}
}
