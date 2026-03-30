package handlers

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"diet/internal/repositories"
	mealservice "diet/internal/services/meal"
	"diet/internal/services/openai"
	"diet/internal/testutil"
	"github.com/gin-gonic/gin"
)

type mealClassifierStub struct{}

func (mealClassifierStub) Classify(string) string { return "meal_log" }

type mealAnalyzerStub struct{}

func (mealAnalyzerStub) AnalyzeMealText(context.Context, string) (openai.MealTextAnalysis, error) {
	return openai.MealTextAnalysis{
		CanonicalName: "test meal",
		Nutrition: openai.NutritionV1{
			CaloriesKcal:  numPtr(100),
			ProteinG:      numPtr(10),
			CarbohydrateG: numPtr(12),
			FatG:          numPtr(3),
		},
	}, nil
}

func setupMealHandler(t *testing.T) *MealHandler {
	t.Helper()
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	repos := repositories.New(pool)
	svc := mealservice.NewService(repos.MealEvents, repos.MealAnalysis, repos.MealMemory, repos.DailyNutritionSummary, repos.Meals, repos.MealItems, repos.CanonicalFoods, nil, mealAnalyzerStub{}, mealClassifierStub{})
	return NewMealHandler(svc, nil)
}

func TestMealHandler_CreateMeal(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupMealHandler(t)
	r := gin.New()
	r.POST("/v1/meals", h.CreateMeal)

	body := `{"user_id":"u1","source":"web","raw_text":"salad"}`
	req := httptest.NewRequest(http.MethodPost, "/v1/meals", bytes.NewBufferString(body))
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
	if resp["logged"] != true {
		t.Fatalf("expected logged=true, got %v", resp["logged"])
	}
	if resp["item"] == nil {
		t.Fatalf("expected item in response")
	}
}

func TestMealHandler_GetRecentMeals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupMealHandler(t)
	r := gin.New()
	r.POST("/v1/meals", h.CreateMeal)
	r.GET("/v1/meals/recent", h.GetRecentMeals)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/meals", bytes.NewBufferString(`{"user_id":"u1","source":"web","raw_text":"yogurt"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusOK {
		t.Fatalf("seed create status=%d body=%s", createW.Code, createW.Body.String())
	}

	req := httptest.NewRequest(http.MethodGet, "/v1/meals/recent?user_id=u1&limit=5", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}
	var resp struct {
		Items []map[string]any `json:"items"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if len(resp.Items) == 0 {
		t.Fatalf("expected at least one recent meal")
	}
	if _, ok := resp.Items[0]["time_source"]; !ok {
		t.Fatalf("expected time_source in recent meal payload")
	}
}

func TestMealHandler_EditMealTime(t *testing.T) {
	gin.SetMode(gin.TestMode)
	h := setupMealHandler(t)
	r := gin.New()
	r.POST("/v1/meals", h.CreateMeal)
	r.PATCH("/v1/meals/:mealEventID/time", h.EditMealTime)

	createReq := httptest.NewRequest(http.MethodPost, "/v1/meals", bytes.NewBufferString(`{"user_id":"u1","source":"web","raw_text":"omelette"}`))
	createReq.Header.Set("Content-Type", "application/json")
	createW := httptest.NewRecorder()
	r.ServeHTTP(createW, createReq)
	if createW.Code != http.StatusOK {
		t.Fatalf("seed create status=%d body=%s", createW.Code, createW.Body.String())
	}
	var created struct {
		Item struct {
			MealEventID float64 `json:"meal_event_id"`
		} `json:"item"`
	}
	if err := json.Unmarshal(createW.Body.Bytes(), &created); err != nil {
		t.Fatalf("decode create response: %v", err)
	}

	newTime := time.Now().UTC().Add(-2 * time.Hour).Format(time.RFC3339)
	patchBody := `{"user_id":"u1","eaten_at":"` + newTime + `"}`
	patchReq := httptest.NewRequest(http.MethodPatch, "/v1/meals/"+jsonNumberToID(created.Item.MealEventID)+"/time", bytes.NewBufferString(patchBody))
	patchReq.Header.Set("Content-Type", "application/json")
	patchW := httptest.NewRecorder()
	r.ServeHTTP(patchW, patchReq)
	if patchW.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", patchW.Code, patchW.Body.String())
	}
	if !bytes.Contains(patchW.Body.Bytes(), []byte(`"time_source":"edited"`)) {
		t.Fatalf("expected edited time_source, body=%s", patchW.Body.String())
	}
}

func jsonNumberToID(v float64) string {
	return strconv.FormatInt(int64(v), 10)
}

func numPtr(v float64) *float64 { return &v }
