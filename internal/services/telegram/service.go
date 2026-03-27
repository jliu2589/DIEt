package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	mealservice "diet/internal/services/meal"
)

const telegramSource = "telegram"

type MealProcessor interface {
	ProcessTextMeal(ctx context.Context, input mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error)
}

type MessageSender interface {
	SendMessage(chatID int64, text string) error
}

// Service coordinates Telegram updates with the existing meal service.
// V1 only supports plain text messages.
type Service struct {
	mealService MealProcessor
	botClient   MessageSender
}

func NewService(mealService MealProcessor, botClient MessageSender) *Service {
	return &Service{mealService: mealService, botClient: botClient}
}

func (s *Service) ProcessUpdate(ctx context.Context, update Update) error {
	if s == nil || s.mealService == nil || s.botClient == nil {
		return fmt.Errorf("telegram service dependencies are not configured")
	}

	message := update.Message
	if message == nil || message.From == nil {
		return nil
	}

	messageText := strings.TrimSpace(message.Text)
	if messageText == "" {
		// V1: ignore unsupported/non-text message types.
		return nil
	}

	result, err := s.mealService.ProcessTextMeal(ctx, mealservice.ProcessTextMealInput{
		UserID:  fmt.Sprintf("telegram:%d", message.From.ID),
		Source:  telegramSource,
		RawText: messageText,
		EatenAt: time.Unix(message.Date, 0).UTC(),
	})
	if err != nil {
		return fmt.Errorf("process telegram text meal: %w", err)
	}

	if err := s.botClient.SendMessage(message.Chat.ID, buildNutritionReply(result)); err != nil {
		return fmt.Errorf("send telegram reply: %w", err)
	}

	return nil
}

func buildNutritionReply(result *mealservice.ProcessTextMealResult) string {
	if result == nil {
		return "Meal saved."
	}

	return fmt.Sprintf(
		"Meal: %s\nCalories: %.2f kcal\nProtein: %.2f g\nCarbs: %.2f g\nFat: %.2f g",
		result.CanonicalName,
		floatOrZero(result.Nutrition.CaloriesKcal),
		floatOrZero(result.Nutrition.ProteinG),
		floatOrZero(result.Nutrition.CarbohydrateG),
		floatOrZero(result.Nutrition.FatG),
	)
}

func floatOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
