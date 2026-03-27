package handlers

import (
	telegramsvc "diet/internal/services/telegram"
	"github.com/gin-gonic/gin"
)

// TelegramHandler is a placeholder for Telegram webhook delivery.
type TelegramHandler struct {
	secretPath string
	botClient  *telegramsvc.BotClient
}

func NewTelegramHandler(secretPath string, botClient *telegramsvc.BotClient) *TelegramHandler {
	return &TelegramHandler{secretPath: secretPath, botClient: botClient}
}

func (h *TelegramHandler) RegisterRoutes(router gin.IRouter) {
	// Example: router.POST("/telegram/"+h.secretPath, h.Webhook)
}
