package handlers

import "github.com/gin-gonic/gin"

// TelegramHandler is a placeholder for Telegram webhook delivery.
type TelegramHandler struct {
	secretPath string
}

func NewTelegramHandler(secretPath string) *TelegramHandler {
	return &TelegramHandler{secretPath: secretPath}
}

func (h *TelegramHandler) RegisterRoutes(router gin.IRouter) {
	// Example: router.POST("/telegram/"+h.secretPath, h.Webhook)
}
