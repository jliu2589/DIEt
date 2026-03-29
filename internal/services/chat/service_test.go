package chat

import (
	"context"
	"testing"
	"time"

	"diet/internal/models"
	mealservice "diet/internal/services/meal"
	weightservice "diet/internal/services/weight"
)

type classifierStub struct{ intent string }

func (s classifierStub) Classify(_ string) string { return s.intent }

type mealStub struct{ calls int }

func (m *mealStub) ProcessTextMeal(_ context.Context, _ mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error) {
	m.calls++
	return &mealservice.ProcessTextMealResult{Logged: true}, nil
}

type weightStub struct{ calls int }

func (w *weightStub) CreateEntry(_ context.Context, _ weightservice.CreateEntryInput) (*models.WeightEntry, error) {
	w.calls++
	return &models.WeightEntry{Weight: 70, Unit: "kg", LoggedAt: time.Now().UTC()}, nil
}

func TestHandleMessage_GeneralChat_DoesNotCallMealOrWeight(t *testing.T) {
	meal := &mealStub{}
	weight := &weightStub{}
	svc := NewService(classifierStub{intent: "general_chat"}, meal, weight)

	resp, err := svc.HandleMessage(context.Background(), Request{UserID: "u1", Message: "hi"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
	if resp.Intent != "general_chat" {
		t.Fatalf("expected general_chat intent, got %s", resp.Intent)
	}
	if resp.MessageToUser != nonDietHelpMessage {
		t.Fatalf("unexpected message_to_user: %s", resp.MessageToUser)
	}
	if meal.calls != 0 {
		t.Fatalf("expected meal service not called, got %d", meal.calls)
	}
	if weight.calls != 0 {
		t.Fatalf("expected weight service not called, got %d", weight.calls)
	}
}

func TestHandleMessage_Unknown_DoesNotCallMealOrWeight(t *testing.T) {
	meal := &mealStub{}
	weight := &weightStub{}
	svc := NewService(classifierStub{intent: "unknown"}, meal, weight)

	resp, err := svc.HandleMessage(context.Background(), Request{UserID: "u1", Message: "random words"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil {
		t.Fatalf("expected response")
	}
	if resp.Intent != "unknown" {
		t.Fatalf("expected unknown intent, got %s", resp.Intent)
	}
	if resp.MessageToUser != nonDietHelpMessage {
		t.Fatalf("unexpected message_to_user: %s", resp.MessageToUser)
	}
	if meal.calls != 0 {
		t.Fatalf("expected meal service not called, got %d", meal.calls)
	}
	if weight.calls != 0 {
		t.Fatalf("expected weight service not called, got %d", weight.calls)
	}
}
