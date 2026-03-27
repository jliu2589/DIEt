package telegram

import (
	"context"
	"fmt"
	"strings"
	"time"

	mealservice "diet/internal/services/meal"
)

type MealProcessor interface {
	ProcessTextMeal(ctx context.Context, input mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error)
}

type Service struct {
	mealService MealProcessor
	botClient   *BotClient
}

func NewService(mealService MealProcessor, botClient *BotClient) *Service {
	return &Service{mealService: mealService, botClient: botClient}
}

func (s *Service) ProcessUpdate(ctx context.Context, update Update) error {
	if update.Message == nil || update.Message.From == nil {
		return nil
	}

	messageText := strings.TrimSpace(update.Message.Text)
	if messageText == "" {
		// V1: only text messages are supported.
		return nil
	}

	userID := fmt.Sprintf("telegram:%d", update.Message.From.ID)
	eatenAt := time.Unix(update.Message.Date, 0).UTC()

	result, err := s.mealService.ProcessTextMeal(ctx, mealservice.ProcessTextMealInput{
		UserID:  userID,
		Source:  "telegram",
		RawText: messageText,
		EatenAt: eatenAt,
	})
	if err != nil {
		return fmt.Errorf("process telegram text meal: %w", err)
	}

	reply := fmt.Sprintf(
		"Meal: %s\nCalories: %.2f kcal\nProtein: %.2f g\nCarbs: %.2f g\nFat: %.2f g",
		result.CanonicalName,
		floatOrZero(result.Nutrition.CaloriesKcal),
		floatOrZero(result.Nutrition.ProteinG),
		floatOrZero(result.Nutrition.CarbohydrateG),
		floatOrZero(result.Nutrition.FatG),
	)

	if err := s.botClient.SendMessage(update.Message.Chat.ID, reply); err != nil {
		return fmt.Errorf("send telegram reply: %w", err)
	}

	return nil
}

func floatOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
