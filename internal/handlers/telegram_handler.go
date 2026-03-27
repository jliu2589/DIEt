package handlers

import (
	"net/http"
	"strings"

	telegramsvc "diet/internal/services/telegram"
	"github.com/gin-gonic/gin"
)

// TelegramHandler handles Telegram webhook delivery.
type TelegramHandler struct {
	secretPath  string
	secretToken string
	service     *telegramsvc.Service
}

func NewTelegramHandler(secretPath, secretToken string, service *telegramsvc.Service) *TelegramHandler {
	return &TelegramHandler{
		secretPath:  strings.TrimSpace(secretPath),
		secretToken: strings.TrimSpace(secretToken),
		service:     service,
	}
}

func (h *TelegramHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/v1/integrations/telegram/webhook/:secretPath", h.Webhook)
}

func (h *TelegramHandler) Webhook(c *gin.Context) {
	if c.Param("secretPath") != h.secretPath {
		c.JSON(http.StatusNotFound, gin.H{"error": "not found"})
		return
	}

	if c.GetHeader("X-Telegram-Bot-Api-Secret-Token") != h.secretToken {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	var update telegramsvc.Update
	if err := c.ShouldBindJSON(&update); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid update payload"})
		return
	}

	if err := h.service.ProcessUpdate(c.Request.Context(), update); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process update"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"ok": true})
}
