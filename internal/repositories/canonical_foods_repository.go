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
	db DBTX
}

type CanonicalFoodRecommendationCandidate struct {
	FoodID        int64
	CanonicalName string
	DefaultAmount float64
	DefaultUnit   string
	CaloriesKcal  *float64
	ProteinG      *float64
	CarbohydrateG *float64
	FatG          *float64
}

func NewCanonicalFoodsRepository(pool *pgxpool.Pool) *CanonicalFoodsRepository {
	return &CanonicalFoodsRepository{db: pool}
}

func NewCanonicalFoodsRepositoryWithDB(db DBTX) *CanonicalFoodsRepository {
	return &CanonicalFoodsRepository{db: db}
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
	if err := r.db.QueryRow(ctx, q, id).Scan(
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
	if err := r.db.QueryRow(ctx, q, name).Scan(
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

func (r *CanonicalFoodsRepository) ListRecommendationCandidates(ctx context.Context, limit int) ([]CanonicalFoodRecommendationCandidate, error) {
	if limit <= 0 {
		limit = 50
	}

	const q = `
		SELECT
			f.id,
			f.canonical_name,
			f.default_amount,
			f.default_unit,
			fn.calories_kcal,
			fn.protein_g,
			fn.carbohydrate_g,
			fn.fat_g
		FROM foods f
		LEFT JOIN food_nutrition fn ON fn.food_id = f.id
		ORDER BY f.canonical_name ASC, f.id ASC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list canonical food recommendation candidates: %w", err)
	}
	defer rows.Close()

	out := make([]CanonicalFoodRecommendationCandidate, 0, limit)
	for rows.Next() {
		var item CanonicalFoodRecommendationCandidate
		if err := rows.Scan(
			&item.FoodID,
			&item.CanonicalName,
			&item.DefaultAmount,
			&item.DefaultUnit,
			&item.CaloriesKcal,
			&item.ProteinG,
			&item.CarbohydrateG,
			&item.FatG,
		); err != nil {
			return nil, fmt.Errorf("scan canonical food recommendation candidate: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate canonical food recommendation candidates: %w", err)
	}

	return out, nil
}
