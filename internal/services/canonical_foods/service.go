package canonical_foods

import (
	"fmt"
	"strings"

	"diet/internal/models"
)

// ScaleResult is a deterministic nutrition scaling output for a requested food quantity.
type ScaleResult struct {
	Factor          float64                `json:"factor"`
	RequestedAmount float64                `json:"requested_amount"`
	RequestedUnit   string                 `json:"requested_unit"`
	Nutrition       models.NutritionFields `json:"nutrition"`
}

// ScaleNutrition scales a canonical food's default nutrition values to a requested amount.
//
// The scale factor is computed as:
//
//	factor = requestedAmount / canonicalFood.DefaultAmount
//
// If requestedUnit is empty, canonicalFood.DefaultUnit is assumed.
func ScaleNutrition(canonicalFood models.CanonicalFoodWithNutrition, requestedAmount float64, requestedUnit string) (*ScaleResult, error) {
	if canonicalFood.DefaultAmount <= 0 {
		return nil, fmt.Errorf("canonical food default_amount must be > 0")
	}
	if requestedAmount < 0 {
		return nil, fmt.Errorf("requested_amount must be >= 0")
	}

	defaultUnit := normalizeUnit(canonicalFood.DefaultUnit)
	if defaultUnit == "" {
		return nil, fmt.Errorf("canonical food default_unit is required")
	}

	reqUnit := normalizeUnit(requestedUnit)
	if reqUnit == "" {
		reqUnit = defaultUnit
	}
	if reqUnit != defaultUnit {
		return nil, fmt.Errorf("unit mismatch: canonical default_unit=%q requested_unit=%q", defaultUnit, reqUnit)
	}

	factor := requestedAmount / canonicalFood.DefaultAmount

	return &ScaleResult{
		Factor:          factor,
		RequestedAmount: requestedAmount,
		RequestedUnit:   reqUnit,
		Nutrition:       scaleNutritionFields(canonicalFood.NutritionFields, factor),
	}, nil
}

func scaleNutritionFields(in models.NutritionFields, factor float64) models.NutritionFields {
	return models.NutritionFields{
		CaloriesKcal:  scalePtr(in.CaloriesKcal, factor),
		ProteinG:      scalePtr(in.ProteinG, factor),
		CarbohydrateG: scalePtr(in.CarbohydrateG, factor),
		FatG:          scalePtr(in.FatG, factor),
		FiberG:        scalePtr(in.FiberG, factor),
		SugarsG:       scalePtr(in.SugarsG, factor),
		SaturatedFatG: scalePtr(in.SaturatedFatG, factor),
		SodiumMg:      scalePtr(in.SodiumMg, factor),
		PotassiumMg:   scalePtr(in.PotassiumMg, factor),
		CalciumMg:     scalePtr(in.CalciumMg, factor),
		MagnesiumMg:   scalePtr(in.MagnesiumMg, factor),
		IronMg:        scalePtr(in.IronMg, factor),
		ZincMg:        scalePtr(in.ZincMg, factor),
		VitaminDMcg:   scalePtr(in.VitaminDMcg, factor),
		VitaminB12Mcg: scalePtr(in.VitaminB12Mcg, factor),
		FolateB9Mcg:   scalePtr(in.FolateB9Mcg, factor),
		VitaminCMg:    scalePtr(in.VitaminCMg, factor),
	}
}

func scalePtr(v *float64, factor float64) *float64 {
	if v == nil {
		return nil
	}
	scaled := (*v) * factor
	return &scaled
}

func normalizeUnit(unit string) string {
	return strings.ToLower(strings.TrimSpace(unit))
}
