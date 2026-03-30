package handlers

import (
	"net/http"
	"time"

	chatservice "diet/internal/services/chat"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	service *chatservice.Service
}

func NewChatHandler(service *chatservice.Service) *ChatHandler {
	return &ChatHandler{service: service}
}

type chatRequest struct {
	UserID   string     `json:"user_id"`
	Message  string     `json:"message" binding:"required"`
	LoggedAt *time.Time `json:"logged_at"`
}

func (h *ChatHandler) PostChat(c *gin.Context) {
	var req chatRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	userID, ok := resolveUserID(c, req.UserID, false)
	if !ok {
		return
	}

	resp, err := h.service.HandleMessage(c.Request.Context(), chatservice.Request{
		UserID:   userID,
		Message:  req.Message,
		LoggedAt: req.LoggedAt,
	})
	if err != nil {
		if isValidationError(err) || isWeightValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process chat message"})
		return
	}

	c.JSON(http.StatusOK, resp)
}
