package chat

import (
	"context"
	"testing"
	"time"

	inputclassifier "diet/internal/services/input_classifier"
	mealservice "diet/internal/services/meal"
	weightservice "diet/internal/services/weight"
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

type classifierStub struct{ intent string }

func (c classifierStub) Classify(string) string { return c.intent }

type weightStub struct{ calls int }

func (w *weightStub) CreateEntry(_ context.Context, input weightservice.CreateEntryInput) (*weightserviceEntry, error) {
	w.calls++
	return &weightserviceEntry{Weight: input.Weight, Unit: input.Unit, LoggedAt: input.LoggedAt}, nil
}

func TestHandleMessage_MealIntentCallsMealService(t *testing.T) {
	meal := &mealStub{}
	svc := &Service{mealService: meal, classifier: classifierStub{intent: inputclassifier.IntentMealLog}}

	resp, err := svc.HandleMessage(context.Background(), Request{UserID: "u1", Message: "i ate eggs"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp == nil || resp.Intent != "meal_log" {
		t.Fatalf("expected meal_log response")
	}
	if meal.calls != 1 {
		t.Fatalf("expected meal service called once, got %d", meal.calls)
	}
}

func TestHandleMessage_WeightIntentLogsWeight(t *testing.T) {
	meal := &mealStub{}
	weight := &weightStub{}
	svc := &Service{mealService: meal, classifier: classifierStub{intent: inputclassifier.IntentWeightLog}, weightSvc: weight}

	resp, err := svc.HandleMessage(context.Background(), Request{UserID: "u1", Message: "I weigh 176.2 lb"})
	if err != nil {
		t.Fatalf("expected nil error, got %v", err)
	}
	if resp.Intent != "weight_log" {
		t.Fatalf("expected weight_log intent, got %s", resp.Intent)
	}
	if resp.WeightResult == nil {
		t.Fatalf("expected weight_result")
	}
	if meal.calls != 0 {
		t.Fatalf("expected meal service not called")
	}
	if weight.calls != 1 {
		t.Fatalf("expected weight service called once")
	}
}

func TestHandleMessage_WeightIntentInvalidFormat(t *testing.T) {
	svc := &Service{mealService: &mealStub{}, classifier: classifierStub{intent: inputclassifier.IntentWeightLog}, weightSvc: &weightStub{}}
	resp, err := svc.HandleMessage(context.Background(), Request{UserID: "u1", Message: "i weigh a lot"})
	if err != nil {
		t.Fatalf("expected nil err, got %v", err)
	}
	if resp.Intent != "weight_log" {
		t.Fatalf("expected weight_log intent")
	}
	if resp.WeightResult != nil {
		t.Fatalf("expected no weight_result")
	}
}

func TestParseWeightFromMessage(t *testing.T) {
	w, unit, ok := parseWeightFromMessage("I weigh 80 kg")
	if !ok || unit != "kg" || w != 80 {
		t.Fatalf("unexpected parse result: %v %s %v", w, unit, ok)
	}
}
