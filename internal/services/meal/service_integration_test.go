package meal

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"diet/internal/models"
	"diet/internal/repositories"
	"diet/internal/services/openai"
	"diet/internal/testutil"
	"github.com/jackc/pgx/v5/pgxpool"
)

type classifierStub struct{ intent string }

func (s classifierStub) Classify(string) string { return s.intent }

type analyzerStub struct {
	calls int
	resp  openai.MealTextAnalysis
}

func (s *analyzerStub) AnalyzeMealText(context.Context, string) (openai.MealTextAnalysis, error) {
	s.calls++
	return s.resp, nil
}

func float64Ptr(v float64) *float64 { return &v }

func newMealServiceForTest(pool *pgxpool.Pool, analyzer Analyzer, intent string) *Service {
	repos := repositories.NewWithDB(pool)
	return NewService(
		repos.MealEvents,
		repos.MealAnalysis,
		repos.MealMemory,
		repos.DailyNutritionSummary,
		repos.Meals,
		repos.MealItems,
		repos.CanonicalFoods,
		nil,
		analyzer,
		classifierStub{intent: intent},
	)
}

func TestProcessTextMeal_GeneralChatDoesNotLogMeal(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	svc := newMealServiceForTest(pool, &analyzerStub{}, "general_chat")

	res, err := svc.ProcessTextMeal(context.Background(), ProcessTextMealInput{UserID: "u1", Source: "web", RawText: "hello"})
	if err != nil {
		t.Fatalf("ProcessTextMeal: %v", err)
	}
	if res.Logged {
		t.Fatalf("expected logged=false")
	}

	var events int
	if err := pool.QueryRow(context.Background(), `SELECT COUNT(*) FROM meal_events`).Scan(&events); err != nil {
		t.Fatalf("count meal_events: %v", err)
	}
	if events != 0 {
		t.Fatalf("expected 0 meal_events, got %d", events)
	}
}

func TestProcessTextMeal_CacheHit(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	repos := repositories.New(pool)
	analyzer := &analyzerStub{}
	svc := newMealServiceForTest(pool, analyzer, "meal_log")

	raw := "salmon rice"
	itemsJSON, _ := json.Marshal([]models.MealItem{{Name: "salmon"}})
	_, err := repos.MealMemory.Upsert(context.Background(), models.MealMemory{
		FingerprintHash: FingerprintFromText(raw),
		CanonicalName:   "salmon rice",
		ItemsJSON:       itemsJSON,
		NutritionFields: models.NutritionFields{CaloriesKcal: float64Ptr(520), ProteinG: float64Ptr(35)},
	})
	if err != nil {
		t.Fatalf("seed meal_memory: %v", err)
	}

	res, err := svc.ProcessTextMeal(context.Background(), ProcessTextMealInput{UserID: "u1", Source: "web", RawText: raw, EatenAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("ProcessTextMeal: %v", err)
	}
	if res.ProcessedFrom != "cache" {
		t.Fatalf("expected processed_from cache, got %s", res.ProcessedFrom)
	}
	if analyzer.calls != 0 {
		t.Fatalf("expected analyzer not called, got %d", analyzer.calls)
	}
}

func TestProcessTextMeal_ReusableDBHit(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	repos := repositories.New(pool)
	analyzer := &analyzerStub{}
	svc := newMealServiceForTest(pool, analyzer, "meal_log")

	raw := "bowl salmon"
	fingerprint := FingerprintFromText(raw)
	confidence := 0.9
	mealRow, err := repos.Meals.Create(context.Background(), models.Meal{CanonicalName: "salmon bowl", FingerprintHash: &fingerprint, SourceType: strPtr("seed"), ConfidenceScore: &confidence})
	if err != nil {
		t.Fatalf("create reusable meal: %v", err)
	}

	var foodID int64
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO foods (canonical_name, default_amount, default_unit, source_type)
		VALUES ('salmon', 100, 'g', 'seed') RETURNING id`).Scan(&foodID); err != nil {
		t.Fatalf("insert food: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `INSERT INTO food_nutrition (food_id, calories_kcal, protein_g) VALUES ($1, 200, 20)`, foodID); err != nil {
		t.Fatalf("insert food_nutrition: %v", err)
	}
	if _, err := repos.MealItems.Create(context.Background(), models.StoredMealItem{MealID: mealRow.ID, FoodID: foodID, Quantity: 100, Unit: "g"}); err != nil {
		t.Fatalf("create meal_item: %v", err)
	}

	res, err := svc.ProcessTextMeal(context.Background(), ProcessTextMealInput{UserID: "u1", Source: "web", RawText: raw, EatenAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("ProcessTextMeal: %v", err)
	}
	if res.ProcessedFrom != "reusable_db" {
		t.Fatalf("expected reusable_db, got %s", res.ProcessedFrom)
	}
	if analyzer.calls != 0 {
		t.Fatalf("expected analyzer not called")
	}
}

func TestProcessTextMeal_OpenAIFallback(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	analyzer := &analyzerStub{resp: openai.MealTextAnalysis{
		CanonicalName: "oatmeal",
		Nutrition: openai.NutritionV1{
			CaloriesKcal:  float64Ptr(300),
			ProteinG:      float64Ptr(12),
			CarbohydrateG: float64Ptr(50),
			FatG:          float64Ptr(6),
		},
	}}
	svc := newMealServiceForTest(pool, analyzer, "meal_log")

	res, err := svc.ProcessTextMeal(context.Background(), ProcessTextMealInput{UserID: "u1", Source: "web", RawText: "oatmeal", EatenAt: time.Now().UTC()})
	if err != nil {
		t.Fatalf("ProcessTextMeal: %v", err)
	}
	if res.ProcessedFrom != "openai" {
		t.Fatalf("expected openai, got %s", res.ProcessedFrom)
	}
	if analyzer.calls != 1 {
		t.Fatalf("expected analyzer called once, got %d", analyzer.calls)
	}
}

func TestProcessTextMeal_OpenAIFallbackDisabledReturnsClearError(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	svc := newMealServiceForTest(pool, nil, "meal_log")

	_, err := svc.ProcessTextMeal(context.Background(), ProcessTextMealInput{UserID: "u1", Source: "web", RawText: "needs analysis", EatenAt: time.Now().UTC()})
	if err == nil {
		t.Fatalf("expected error")
	}
	if err != ErrOpenAIFallbackDisabled {
		t.Fatalf("expected ErrOpenAIFallbackDisabled, got %v", err)
	}
}

func strPtr(v string) *string { return &v }
