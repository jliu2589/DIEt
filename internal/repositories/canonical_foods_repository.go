package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"diet/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CanonicalFoodsRepository struct {
	pool *pgxpool.Pool
}

func NewCanonicalFoodsRepository(pool *pgxpool.Pool) *CanonicalFoodsRepository {
	return &CanonicalFoodsRepository{pool: pool}
}

func (r *CanonicalFoodsRepository) GetByID(ctx context.Context, id int64) (*models.CanonicalFoodWithNutrition, error) {
	const q = `
		SELECT
			f.id,
			f.canonical_name,
			f.default_amount,
			f.default_unit,
			f.category,
			f.source_type,
			fn.calories_kcal,
			fn.protein_g,
			fn.carbohydrate_g,
			fn.fat_g,
			fn.fiber_g,
			fn.sugars_g,
			fn.saturated_fat_g,
			fn.sodium_mg,
			fn.potassium_mg,
			fn.calcium_mg,
			fn.magnesium_mg,
			fn.iron_mg,
			fn.zinc_mg,
			fn.vitamin_d_mcg,
			fn.vitamin_b12_mcg,
			fn.folate_b9_mcg,
			fn.vitamin_c_mg,
			f.created_at,
			f.updated_at
		FROM foods f
		LEFT JOIN food_nutrition fn ON fn.food_id = f.id
		WHERE f.id = $1
	`

	var out models.CanonicalFoodWithNutrition
	if err := r.pool.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.CanonicalName,
		&out.DefaultAmount,
		&out.DefaultUnit,
		&out.Category,
		&out.SourceType,
		&out.CaloriesKcal,
		&out.ProteinG,
		&out.CarbohydrateG,
		&out.FatG,
		&out.FiberG,
		&out.SugarsG,
		&out.SaturatedFatG,
		&out.SodiumMg,
		&out.PotassiumMg,
		&out.CalciumMg,
		&out.MagnesiumMg,
		&out.IronMg,
		&out.ZincMg,
		&out.VitaminDMcg,
		&out.VitaminB12Mcg,
		&out.FolateB9Mcg,
		&out.VitaminCMg,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get canonical food by id: %w", err)
	}

	return &out, nil
}

func (r *CanonicalFoodsRepository) GetByCanonicalName(ctx context.Context, canonicalName string) (*models.CanonicalFoodWithNutrition, error) {
	name := strings.TrimSpace(canonicalName)
	if name == "" {
		return nil, fmt.Errorf("canonical_name is required")
	}

	const q = `
		SELECT
			f.id,
			f.canonical_name,
			f.default_amount,
			f.default_unit,
			f.category,
			f.source_type,
			fn.calories_kcal,
			fn.protein_g,
			fn.carbohydrate_g,
			fn.fat_g,
			fn.fiber_g,
			fn.sugars_g,
			fn.saturated_fat_g,
			fn.sodium_mg,
			fn.potassium_mg,
			fn.calcium_mg,
			fn.magnesium_mg,
			fn.iron_mg,
			fn.zinc_mg,
			fn.vitamin_d_mcg,
			fn.vitamin_b12_mcg,
			fn.folate_b9_mcg,
			fn.vitamin_c_mg,
			f.created_at,
			f.updated_at
		FROM foods f
		LEFT JOIN food_nutrition fn ON fn.food_id = f.id
		WHERE lower(f.canonical_name) = lower($1)
	`

	var out models.CanonicalFoodWithNutrition
	if err := r.pool.QueryRow(ctx, q, name).Scan(
		&out.ID,
		&out.CanonicalName,
		&out.DefaultAmount,
		&out.DefaultUnit,
		&out.Category,
		&out.SourceType,
		&out.CaloriesKcal,
		&out.ProteinG,
		&out.CarbohydrateG,
		&out.FatG,
		&out.FiberG,
		&out.SugarsG,
		&out.SaturatedFatG,
		&out.SodiumMg,
		&out.PotassiumMg,
		&out.CalciumMg,
		&out.MagnesiumMg,
		&out.IronMg,
		&out.ZincMg,
		&out.VitaminDMcg,
		&out.VitaminB12Mcg,
		&out.FolateB9Mcg,
		&out.VitaminCMg,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get canonical food by canonical_name: %w", err)
	}

	return &out, nil
}
