package recommendations

import (
	"context"
	"testing"
	"time"

	"diet/internal/models"
	"diet/internal/repositories"
	"diet/internal/testutil"
)

func TestGetForUserToday_UsesMultipleCandidateSources(t *testing.T) {
	pool := testutil.OpenTestDB(t)
	t.Cleanup(pool.Close)
	repos := repositories.New(pool)

	calGoal := 2000.0
	proteinGoal := 150.0
	_, err := repos.UserSettings.Upsert(context.Background(), models.UserSettings{
		UserID:       "u1",
		CalorieGoal:  &calGoal,
		ProteinGoalG: &proteinGoal,
		WeightUnit:   "kg",
	})
	if err != nil {
		t.Fatalf("upsert user settings: %v", err)
	}

	// meal_memory candidate
	_, err = repos.MealMemory.Upsert(context.Background(), models.MealMemory{
		FingerprintHash: "fp-memory",
		CanonicalName:   "memory bowl",
		NutritionFields: models.NutritionFields{
			CaloriesKcal:  floatPtr(600),
			ProteinG:      floatPtr(40),
			CarbohydrateG: floatPtr(50),
			FatG:          floatPtr(20),
		},
	})
	if err != nil {
		t.Fatalf("seed meal_memory: %v", err)
	}

	// reusable meal candidate
	fingerprint := "fp-reusable"
	reusableRow, err := repos.Meals.Create(context.Background(), models.Meal{CanonicalName: "reusable plate", FingerprintHash: &fingerprint})
	if err != nil {
		t.Fatalf("seed reusable meal: %v", err)
	}

	var reusableFoodID int64
	if err := pool.QueryRow(context.Background(), `
		INSERT INTO foods (canonical_name, default_amount, default_unit, source_type)
		VALUES ('reusable chicken', 100, 'g', 'seed') RETURNING id`).Scan(&reusableFoodID); err != nil {
		t.Fatalf("insert reusable food: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `INSERT INTO food_nutrition (food_id, calories_kcal, protein_g, carbohydrate_g, fat_g) VALUES ($1, 400, 55, 10, 15)`, reusableFoodID); err != nil {
		t.Fatalf("insert reusable food nutrition: %v", err)
	}
	if _, err := repos.MealItems.Create(context.Background(), models.StoredMealItem{MealID: reusableRow.ID, FoodID: reusableFoodID, Quantity: 100, Unit: "g"}); err != nil {
		t.Fatalf("insert reusable meal_item: %v", err)
	}

	// canonical food candidate
	if _, err := pool.Exec(context.Background(), `
		INSERT INTO foods (canonical_name, default_amount, default_unit, source_type)
		VALUES ('canonical snack', 50, 'g', 'seed')`); err != nil {
		t.Fatalf("insert canonical food: %v", err)
	}
	var canonicalID int64
	if err := pool.QueryRow(context.Background(), `SELECT id FROM foods WHERE canonical_name='canonical snack'`).Scan(&canonicalID); err != nil {
		t.Fatalf("read canonical food id: %v", err)
	}
	if _, err := pool.Exec(context.Background(), `INSERT INTO food_nutrition (food_id, calories_kcal, protein_g, carbohydrate_g, fat_g) VALUES ($1, 250, 20, 18, 9)`, canonicalID); err != nil {
		t.Fatalf("insert canonical nutrition: %v", err)
	}

	svc := NewService(repos.UserSettings, repos.DailyNutritionSummary, repos.MealMemory, repos.Meals, repos.MealItems, repos.CanonicalFoods)
	resp, err := svc.GetForUserToday(context.Background(), "u1", time.Now().UTC())
	if err != nil {
		t.Fatalf("GetForUserToday: %v", err)
	}

	if len(resp.Items) == 0 {
		t.Fatalf("expected recommendations")
	}

	sources := map[string]bool{}
	for _, item := range resp.Items {
		sources[item.Source] = true
	}
	if !sources[sourceMealMemory] {
		t.Fatalf("expected meal_memory source in recommendations")
	}
	if !sources[sourceReusableMeal] {
		t.Fatalf("expected reusable_meal source in recommendations")
	}
	if !sources[sourceCanonicalFood] {
		t.Fatalf("expected canonical_food source in recommendations")
	}
}

func floatPtr(v float64) *float64 { return &v }
