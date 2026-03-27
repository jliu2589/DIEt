package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"diet/internal/config"
	"diet/internal/db"
	"diet/internal/handlers"
	"diet/internal/repositories"
	"diet/internal/server"
	mealservice "diet/internal/services/meal"
	openaiservice "diet/internal/services/openai"
	telegramservice "diet/internal/services/telegram"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("load config: %v", err)
	}

	pool, err := db.NewPool(ctx, cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}

	var closeDB sync.Once
	closePool := func() {
		closeDB.Do(func() {
			pool.Close()
		})
	}
	defer closePool()

	repos := repositories.New(pool)
	openAIClient := openaiservice.NewClient(cfg.OpenAIAPIKey, "")
	telegramBotClient := telegramservice.NewBotClient(cfg.TelegramBotToken)
	mealSvc := mealservice.NewService(
		repos.MealEvents,
		repos.MealAnalysis,
		repos.MealMemory,
		repos.DailyNutritionSummary,
		openAIClient,
	)
	telegramSvc := telegramservice.NewService(mealSvc, telegramBotClient)

	router := server.NewRouter(server.Dependencies{
		HealthHandler:  handlers.NewHealthHandler(),
		MealHandler:    handlers.NewMealHandler(mealSvc),
		SummaryHandler: handlers.NewSummaryHandler(repos.DailyNutritionSummary),
		TelegramHandler: handlers.NewTelegramHandler(
			cfg.TelegramWebhookSecretPath,
			cfg.TelegramWebhookSecretToken,
			telegramSvc,
		),
	})

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Printf("http server shutdown: %v", err)
		}
		closePool()
	}()

	log.Printf("api listening on :%s", cfg.Port)
	if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		log.Fatalf("http server: %v", err)
	}
}
