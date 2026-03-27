package server

import (
	"time"

	"diet/internal/handlers"

	"github.com/gin-contrib/cors"
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

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"http://localhost:3000"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	r.GET("/health", deps.HealthHandler.GetHealth)
	r.POST("/v1/meals", deps.MealHandler.CreateMeal)
	r.GET("/v1/daily-summary", deps.SummaryHandler.GetDailySummary)
	if deps.TelegramHandler != nil {
		deps.TelegramHandler.RegisterRoutes(r)
	}

	return r
}
