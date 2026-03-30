package handlers

import (
	"net/http"
	"time"

	recommendationsservice "diet/internal/services/recommendations"
	"github.com/gin-gonic/gin"
)

type RecommendationsHandler struct {
	service *recommendationsservice.Service
}

func NewRecommendationsHandler(service *recommendationsservice.Service) *RecommendationsHandler {
	return &RecommendationsHandler{service: service}
}

func (h *RecommendationsHandler) GetRecommendations(c *gin.Context) {
	userID, ok := requiredUserIDFromQuery(c)
	if !ok {
		return
	}

	resp, err := h.service.GetForUserToday(c.Request.Context(), userID, time.Now())
	if err != nil {
		if isValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch recommendations"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
