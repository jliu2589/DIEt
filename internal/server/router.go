package server

import (
	"diet/internal/handlers"

	"github.com/gin-gonic/gin"
)

type Dependencies struct {
	HealthHandler   *handlers.HealthHandler
	MealHandler     *handlers.MealHandler
	SummaryHandler  *handlers.SummaryHandler
	TelegramHandler *handlers.TelegramHandler
}

func NewRouter(deps Dependencies) *gin.Engine {
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())

	r.GET("/health", deps.HealthHandler.GetHealth)
	r.POST("/v1/meals", deps.MealHandler.CreateMeal)
	r.GET("/v1/daily-summary", deps.SummaryHandler.GetDailySummary)
	deps.TelegramHandler.RegisterRoutes(r)

	return r
}
