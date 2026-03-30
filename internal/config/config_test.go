package config

import "testing"

func TestLoad_OnlyCoreVarsRequiredByDefault(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("OPENAI_API_KEY", "")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET_PATH", "")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET_TOKEN", "")
	t.Setenv("ENABLE_OPENAI_FALLBACK", "")
	t.Setenv("ENABLE_TELEGRAM_INTEGRATION", "")

	cfg, err := Load()
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if cfg.EnableOpenAIFallback {
		t.Fatalf("expected openai fallback disabled")
	}
	if cfg.EnableTelegramIntegration {
		t.Fatalf("expected telegram integration disabled")
	}
}

func TestLoad_OpenAIEnabledRequiresAPIKey(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("ENABLE_OPENAI_FALLBACK", "true")
	t.Setenv("OPENAI_API_KEY", "")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error")
	}
}

func TestLoad_TelegramEnabledRequiresTelegramVars(t *testing.T) {
	t.Setenv("PORT", "8080")
	t.Setenv("DATABASE_URL", "postgres://example")
	t.Setenv("ENABLE_TELEGRAM_INTEGRATION", "true")
	t.Setenv("TELEGRAM_BOT_TOKEN", "")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET_PATH", "")
	t.Setenv("TELEGRAM_WEBHOOK_SECRET_TOKEN", "")

	_, err := Load()
	if err == nil {
		t.Fatalf("expected error")
	}
}
