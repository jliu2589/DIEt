package config

import (
	"fmt"
	"os"
)

// Config stores runtime configuration loaded from environment variables.
type Config struct {
	Port                       string
	DatabaseURL                string
	OpenAIAPIKey               string
	TelegramBotToken           string
	TelegramWebhookSecretPath  string
	TelegramWebhookSecretToken string
}

func Load() (Config, error) {
	cfg := Config{
		Port:                       getEnv("PORT", "8080"),
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:               os.Getenv("OPENAI_API_KEY"),
		TelegramBotToken:           os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecretPath:  os.Getenv("TELEGRAM_WEBHOOK_SECRET_PATH"),
		TelegramWebhookSecretToken: os.Getenv("TELEGRAM_WEBHOOK_SECRET_TOKEN"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}

	return cfg, nil
}

func getEnv(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
