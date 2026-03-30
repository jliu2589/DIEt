package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diet/internal/models"
	chatservice "diet/internal/services/chat"
	mealservice "diet/internal/services/meal"
	weightservice "diet/internal/services/weight"
	"github.com/gin-gonic/gin"
)

type chatClassifierStub struct{ intent string }

func (s chatClassifierStub) Classify(string) string { return s.intent }

type chatMealStub struct{}

func (chatMealStub) ProcessTextMeal(context.Context, mealservice.ProcessTextMealInput) (*mealservice.ProcessTextMealResult, error) {
	return &mealservice.ProcessTextMealResult{
		Logged:        true,
		MealEventID:   12,
		CanonicalName: "chat meal",
		LoggedAt:      time.Now().UTC(),
		EatenAt:       time.Now().UTC(),
		TimeSource:    "explicit",
		Source:        "chat",
	}, nil
}

type chatWeightStub struct{}

func (chatWeightStub) CreateEntry(context.Context, weightservice.CreateEntryInput) (*models.WeightEntry, error) {
	return &models.WeightEntry{Weight: 70, Unit: "kg", LoggedAt: time.Now().UTC()}, nil
}

func TestChatHandler_PostChat(t *testing.T) {
	gin.SetMode(gin.TestMode)
	svc := chatservice.NewService(chatClassifierStub{intent: "meal_log"}, chatMealStub{}, chatWeightStub{})
	h := NewChatHandler(svc, nil)
	r := gin.New()
	r.POST("/v1/chat", h.PostChat)

	req := httptest.NewRequest(http.MethodPost, "/v1/chat", bytes.NewBufferString(`{"user_id":"u1","message":"i ate salmon"}`))
	req.Header.Set("Content-Type", "application/json")
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp map[string]any
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp["meal_result"] == nil {
		t.Fatalf("expected meal_result in response")
	}
}
