package config

import (
	"fmt"
	"os"
	"strings"
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
		Port:                       os.Getenv("PORT"),
		DatabaseURL:                os.Getenv("DATABASE_URL"),
		OpenAIAPIKey:               os.Getenv("OPENAI_API_KEY"),
		TelegramBotToken:           os.Getenv("TELEGRAM_BOT_TOKEN"),
		TelegramWebhookSecretPath:  os.Getenv("TELEGRAM_WEBHOOK_SECRET_PATH"),
		TelegramWebhookSecretToken: os.Getenv("TELEGRAM_WEBHOOK_SECRET_TOKEN"),
	}

	missing := make([]string, 0, 6)
	if cfg.Port == "" {
		missing = append(missing, "PORT")
	}
	if cfg.DatabaseURL == "" {
		missing = append(missing, "DATABASE_URL")
	}
	if cfg.OpenAIAPIKey == "" {
		missing = append(missing, "OPENAI_API_KEY")
	}
	if cfg.TelegramBotToken == "" {
		missing = append(missing, "TELEGRAM_BOT_TOKEN")
	}
	if cfg.TelegramWebhookSecretPath == "" {
		missing = append(missing, "TELEGRAM_WEBHOOK_SECRET_PATH")
	}
	if cfg.TelegramWebhookSecretToken == "" {
		missing = append(missing, "TELEGRAM_WEBHOOK_SECRET_TOKEN")
	}

	if len(missing) > 0 {
		return Config{}, fmt.Errorf("missing required environment variables: %s", strings.Join(missing, ", "))
	}

	return cfg, nil
}
