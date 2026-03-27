package server

import (
	"diet/internal/handlers"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	HealthHandler   *handlers.HealthHandler
	TelegramHandler *handlers.TelegramHandler
}

func NewRouter(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", deps.HealthHandler.GetHealth)

	api := r.Group("/api/v1")
	{
		deps.TelegramHandler.RegisterRoutes(api)
	}

	return r
}
