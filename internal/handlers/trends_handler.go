package handlers

import (
	"net/http"
	"strings"
	"time"

	trendsservice "diet/internal/services/trends"
	"github.com/gin-gonic/gin"
)

type TrendsHandler struct {
	service *trendsservice.Service
}

func NewTrendsHandler(service *trendsservice.Service) *TrendsHandler {
	return &TrendsHandler{service: service}
}

func (h *TrendsHandler) GetTrends(c *gin.Context) {
	userID, ok := requiredUserIDFromQuery(c)
	rangeKey := strings.TrimSpace(c.Query("range"))
	if !ok {
		return
	}
	if rangeKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "range is required"})
		return
	}

	result, err := h.service.GetTrends(c.Request.Context(), userID, rangeKey, time.Now())
	if err != nil {
		if strings.Contains(err.Error(), "range must be one of") || strings.Contains(err.Error(), "is required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch trends"})
		return
	}

	c.JSON(http.StatusOK, result)
}
