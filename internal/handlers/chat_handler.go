package handlers

import (
	"net/http"
	"time"

	"diet/internal/repositories"
	chatservice "diet/internal/services/chat"
	"github.com/gin-gonic/gin"
)

type ChatHandler struct {
	service         *chatservice.Service
	idempotencyRepo *repositories.IdempotencyKeysRepository
}

func NewChatHandler(service *chatservice.Service, idempotencyRepo *repositories.IdempotencyKeysRepository) *ChatHandler {
	return &ChatHandler{service: service, idempotencyRepo: idempotencyRepo}
}

type chatRequest struct {
	UserID         string     `json:"user_id"`
	Message        string     `json:"message" binding:"required"`
	LoggedAt       *time.Time `json:"logged_at"`
	IdempotencyKey string     `json:"idempotency_key,omitempty"`
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

	idempotencyKey := resolveIdempotencyKey(c, req.IdempotencyKey)
	idempotencyRecordID, handled := beginIdempotency(c, h.idempotencyRepo, userID, "POST:/v1/chat", idempotencyKey, req)
	if handled {
		return
	}

	resp, err := h.service.HandleMessage(c.Request.Context(), chatservice.Request{
		UserID:   userID,
		Message:  req.Message,
		LoggedAt: req.LoggedAt,
	})
	if err != nil {
		cleanupIdempotencyOnError(c, h.idempotencyRepo, idempotencyRecordID)
		if isValidationError(err) || isWeightValidationError(err) {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process chat message"})
		return
	}

	saveIdempotencySuccess(c, h.idempotencyRepo, idempotencyRecordID, http.StatusOK, resp)
	c.JSON(http.StatusOK, resp)
}
