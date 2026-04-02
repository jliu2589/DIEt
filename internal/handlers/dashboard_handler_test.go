package handlers

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"diet/internal/repositories"
	mealservice "diet/internal/services/meal"
	"diet/internal/services/openai"
	recommendationsservice "diet/internal/services/recommendations"
	"diet/internal/testutil"
	"github.com/gin-gonic/gin"
)

type dashboardAnalyzerStub struct{}

func (dashboardAnalyzerStub) AnalyzeMealText(context.Context, string) (openai.MealTextAnalysis, error) {
	return openai.MealTextAnalysis{
		CanonicalName: "dashboard meal",
		Nutrition: openai.NutritionV1{
			CaloriesKcal: numPtr(400),
			ProteinG:     numPtr(25),
		},
	}, nil
}

func TestDashboardToday_RecalculatesSummaryFromTodayMeals(t *testing.T) {
	gin.SetMode(gin.TestMode)
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	repos := repositories.New(pool)

	mealSvc := mealservice.NewService(
		repos.MealEvents,
		repos.MealAnalysis,
		repos.MealMemory,
		repos.DailyNutritionSummary,
		repos.Meals,
		repos.MealItems,
		repos.CanonicalFoods,
		nil,
		dashboardAnalyzerStub{},
	)

	today := time.Now().UTC()
	_, err := mealSvc.ProcessTextMeal(context.Background(), mealservice.ProcessTextMealInput{
		UserID:  "u1",
		Source:  "web",
		RawText: "meal for dashboard",
		EatenAt: today,
	})
	if err != nil {
		t.Fatalf("seed meal: %v", err)
	}

	if _, err := pool.Exec(context.Background(), `DELETE FROM daily_nutrition_summary WHERE user_id = $1`, "u1"); err != nil {
		t.Fatalf("clear daily summary: %v", err)
	}

	recommendationsSvc := recommendationsservice.NewService(
		repos.UserSettings,
		repos.DailyNutritionSummary,
		repos.MealMemory,
		repos.Meals,
		repos.MealItems,
		repos.CanonicalFoods,
	)
	h := NewDashboardHandler(repos.DailyNutritionSummary, mealSvc, recommendationsSvc)

	r := gin.New()
	r.GET("/v1/dashboard/today", h.GetToday)

	req := httptest.NewRequest(http.MethodGet, "/v1/dashboard/today?user_id=u1", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("status=%d body=%s", w.Code, w.Body.String())
	}

	var resp struct {
		DailySummary struct {
			Totals struct {
				CaloriesKcal float64 `json:"calories_kcal"`
			} `json:"totals"`
		} `json:"daily_summary"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if resp.DailySummary.Totals.CaloriesKcal != 400 {
		t.Fatalf("expected recalculated calories 400, got %v", resp.DailySummary.Totals.CaloriesKcal)
	}
}
