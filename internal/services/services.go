package services

// OpenAIService defines behavior expected from OpenAI integrations.
type OpenAIService interface {
	// SummarizeMeal can be used to normalize a raw meal description.
	SummarizeMeal(input string) (string, error)
}

// TelegramService defines behavior expected from Telegram integrations.
type TelegramService interface {
	// HandleWebhook processes raw Telegram webhook payloads.
	HandleWebhook(payload []byte, secretToken string) error
}
