package handlers

import (
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"diet/internal/models"
	weightservice "diet/internal/services/weight"
	"github.com/gin-gonic/gin"
)

type WeightHandler struct {
	service *weightservice.Service
}

func NewWeightHandler(service *weightservice.Service) *WeightHandler {
	return &WeightHandler{service: service}
}

type createWeightEntryRequest struct {
	UserID   string    `json:"user_id" binding:"required"`
	Weight   float64   `json:"weight" binding:"required"`
	Unit     string    `json:"unit" binding:"required,oneof=kg lb"`
	LoggedAt time.Time `json:"logged_at" binding:"required"`
}

type weightEntryResponse struct {
	ID       int64   `json:"id"`
	UserID   string  `json:"user_id"`
	Weight   float64 `json:"weight"`
	Unit     string  `json:"unit"`
	LoggedAt string  `json:"logged_at"`
}

type recentWeightEntriesResponse struct {
	Items []weightEntryResponse `json:"items"`
}

func (h *WeightHandler) CreateWeightEntry(c *gin.Context) {
	var req createWeightEntryRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, ok := normalizeRequiredUserID(req.UserID)
	if !ok {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	entry, err := h.service.CreateEntry(c.Request.Context(), weightservice.CreateEntryInput{
		UserID:   userID,
		Weight:   req.Weight,
		Unit:     req.Unit,
		LoggedAt: req.LoggedAt,
	})
	if err != nil {
		if isWeightValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create weight entry"})
		return
	}

	c.JSON(http.StatusOK, toWeightEntryResponse(*entry))
}

func (h *WeightHandler) GetLatestWeightEntry(c *gin.Context) {
	userID, ok := requiredUserIDFromQuery(c)
	if !ok {
		return
	}

	entry, err := h.service.GetLatestEntry(c.Request.Context(), userID)
	if err != nil {
		if isWeightValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch latest weight entry"})
		return
	}

	if entry == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "weight entry not found"})
		return
	}

	c.JSON(http.StatusOK, toWeightEntryResponse(*entry))
}

func (h *WeightHandler) GetRecentWeightEntries(c *gin.Context) {
	userID, ok := requiredUserIDFromQuery(c)
	if !ok {
		return
	}

	limit, err := parseWeightLimitQuery(c.Query("limit"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	entries, err := h.service.GetRecentEntries(c.Request.Context(), userID, limit)
	if err != nil {
		if isWeightValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recent weight entries"})
		return
	}

	items := make([]weightEntryResponse, 0, len(entries))
	for _, entry := range entries {
		items = append(items, toWeightEntryResponse(entry))
	}

	c.JSON(http.StatusOK, recentWeightEntriesResponse{Items: items})
}

func parseWeightLimitQuery(limitRaw string) (int, error) {
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

func isWeightValidationError(err error) bool {
	if err == nil {
		return false
	}

	message := err.Error()
	return strings.Contains(message, "is required") ||
		strings.Contains(message, "must be greater than 0") ||
		strings.Contains(message, "must be one of")
}

func toWeightEntryResponse(entry models.WeightEntry) weightEntryResponse {
	return weightEntryResponse{
		ID:       entry.ID,
		UserID:   entry.UserID,
		Weight:   entry.Weight,
		Unit:     entry.Unit,
		LoggedAt: entry.LoggedAt.UTC().Format(time.RFC3339),
	}
}
