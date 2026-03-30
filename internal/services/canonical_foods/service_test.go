package canonical_foods

import (
	"testing"

	"diet/internal/models"
)

func fptr(v float64) *float64 { return &v }

func TestScaleNutrition_ScalesByFactor(t *testing.T) {
	food := models.CanonicalFoodWithNutrition{
		CanonicalName: "egg",
		DefaultAmount: 1,
		DefaultUnit:   "piece",
		NutritionFields: models.NutritionFields{
			CaloriesKcal: fptr(70),
			ProteinG:     fptr(6),
			FatG:         fptr(5),
		},
	}

	out, err := ScaleNutrition(food, 3, "piece")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.Factor != 3 {
		t.Fatalf("factor: expected 3, got %v", out.Factor)
	}
	if got := *out.Nutrition.CaloriesKcal; got != 210 {
		t.Fatalf("calories: expected 210, got %v", got)
	}
	if got := *out.Nutrition.ProteinG; got != 18 {
		t.Fatalf("protein: expected 18, got %v", got)
	}
	if got := *out.Nutrition.FatG; got != 15 {
		t.Fatalf("fat: expected 15, got %v", got)
	}
}

func TestScaleNutrition_UsesDefaultUnitWhenRequestedUnitEmpty(t *testing.T) {
	food := models.CanonicalFoodWithNutrition{
		DefaultAmount: 100,
		DefaultUnit:   "g",
		NutritionFields: models.NutritionFields{
			CaloriesKcal: fptr(52),
		},
	}

	out, err := ScaleNutrition(food, 250, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if out.RequestedUnit != "g" {
		t.Fatalf("requested unit: expected g, got %q", out.RequestedUnit)
	}
	if out.Factor != 2.5 {
		t.Fatalf("factor: expected 2.5, got %v", out.Factor)
	}
	if got := *out.Nutrition.CaloriesKcal; got != 130 {
		t.Fatalf("calories: expected 130, got %v", got)
	}
}

func TestScaleNutrition_RejectsUnitMismatch(t *testing.T) {
	food := models.CanonicalFoodWithNutrition{
		DefaultAmount: 100,
		DefaultUnit:   "g",
	}

	if _, err := ScaleNutrition(food, 100, "ml"); err == nil {
		t.Fatal("expected unit mismatch error")
	}
}
