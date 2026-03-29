package handlers

import (
	"net/http"
	"strings"

	userstateservice "diet/internal/services/user_state"
	"github.com/gin-gonic/gin"
)

type MeHandler struct {
	service *userstateservice.Service
}

func NewMeHandler(service *userstateservice.Service) *MeHandler {
	return &MeHandler{service: service}
}

func (h *MeHandler) GetMe(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id is required"})
		return
	}

	state, err := h.service.GetByUserID(c.Request.Context(), userID)
	if err != nil {
		if strings.Contains(err.Error(), "user_id is required") {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch user state"})
		return
	}

	c.JSON(http.StatusOK, state)
}
