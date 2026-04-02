package chat

import (
	"context"
	"testing"
	"time"

	mealservice "diet/internal/services/meal"
)

type mealStub struct {
	calls  int
	result *mealservice.ProcessTextMealResult
}

func (m *mealStub) ProcessTextMeal(_ context.Context, _ mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error) {
	m.calls++
	if m.result != nil {
		return m.result, nil
	}
	return &mealservice.ProcessTextMealResult{Logged: true, MealEventID: 1, LoggedAt: time.Now().UTC(), EatenAt: time.Now().UTC(), Source: "chat"}, nil
}

func TestHandleMessage_AlwaysCallsMealService(t *testing.T) {
	meal := &mealStub{}
	svc := NewService(meal)

	resp, err := svc.HandleMessage(context.Background(), Request{UserID: "u1", Message: "hi"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
	if meal.calls != 1 {
		t.Fatalf("expected meal service called once, got %d", meal.calls)
	}
	if resp.Intent != "meal_log" {
		t.Fatalf("expected meal_log intent, got %s", resp.Intent)
	}
}

func TestHandleMessage_NonMealResultReturnsUnknown(t *testing.T) {
	meal := &mealStub{result: &mealservice.ProcessTextMealResult{Logged: false}}
	svc := NewService(meal)

	resp, err := svc.HandleMessage(context.Background(), Request{UserID: "u1", Message: "random words"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Intent != "unknown" {
		t.Fatalf("expected unknown intent, got %s", resp.Intent)
	}
	if resp.MessageToUser != "I couldn’t log that as a meal." {
		t.Fatalf("unexpected message: %s", resp.MessageToUser)
	}
}
