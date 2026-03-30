package repositories

import (
	"context"
	"errors"
	"fmt"

	"diet/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MealMemoryRepository struct {
	db DBTX
}

type MealRecommendationCandidate struct {
	MealID        int64
	CanonicalName string
	CaloriesKcal  *float64
	ProteinG      *float64
	CarbohydrateG *float64
	FatG          *float64
}

type MealRecommendationCandidate struct {
	MealID        int64
	CanonicalName string
	CaloriesKcal  *float64
	ProteinG      *float64
	CarbohydrateG *float64
	FatG          *float64
}

func NewMealMemoryRepository(pool *pgxpool.Pool) *MealMemoryRepository {
	return &MealMemoryRepository{db: pool}
}

func NewMealMemoryRepositoryWithDB(db DBTX) *MealMemoryRepository {
	return &MealMemoryRepository{db: db}
}

func (r *MealMemoryRepository) ListRecommendationCandidates(ctx context.Context, limit int) ([]MealRecommendationCandidate, error) {
	const q = `
		SELECT
			id,
			canonical_name,
			calories_kcal,
			protein_g,
			carbohydrate_g,
			fat_g
		FROM meal_memory
		WHERE canonical_name <> ''
		ORDER BY usage_count DESC, last_used_at DESC NULLS LAST, id DESC
		LIMIT $1
	`

	rows, err := r.db.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list meal_memory recommendation candidates: %w", err)
	}
	defer rows.Close()

	out := make([]MealRecommendationCandidate, 0, limit)
	for rows.Next() {
		var item MealRecommendationCandidate
		if err := rows.Scan(
			&item.MealID,
			&item.CanonicalName,
			&item.CaloriesKcal,
			&item.ProteinG,
			&item.CarbohydrateG,
			&item.FatG,
		); err != nil {
			return nil, fmt.Errorf("scan meal_memory recommendation candidate: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meal_memory recommendation candidates: %w", err)
	}

	return out, nil
}

func (r *MealMemoryRepository) ListRecommendationCandidates(ctx context.Context, limit int) ([]MealRecommendationCandidate, error) {
	const q = `
		SELECT
			id,
			canonical_name,
			calories_kcal,
			protein_g,
			carbohydrate_g,
			fat_g
		FROM meal_memory
		WHERE canonical_name <> ''
		ORDER BY usage_count DESC, last_used_at DESC NULLS LAST, id DESC
		LIMIT $1
	`

	rows, err := r.pool.Query(ctx, q, limit)
	if err != nil {
		return nil, fmt.Errorf("list meal_memory recommendation candidates: %w", err)
	}
	defer rows.Close()

	out := make([]MealRecommendationCandidate, 0, limit)
	for rows.Next() {
		var item MealRecommendationCandidate
		if err := rows.Scan(
			&item.MealID,
			&item.CanonicalName,
			&item.CaloriesKcal,
			&item.ProteinG,
			&item.CarbohydrateG,
			&item.FatG,
		); err != nil {
			return nil, fmt.Errorf("scan meal_memory recommendation candidate: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meal_memory recommendation candidates: %w", err)
	}

	return out, nil
}

func (r *MealMemoryRepository) FindByFingerprintHash(ctx context.Context, fingerprintHash string) (*models.MealMemory, error) {
	const q = `
		SELECT
			id, fingerprint_hash, canonical_name, confidence_score,
			assumptions_json, items_json, raw_analysis_json,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg,
			usage_count, last_used_at, created_at, updated_at
		FROM meal_memory
		WHERE fingerprint_hash = $1
	`

	var out models.MealMemory
	if err := r.db.QueryRow(ctx, q, fingerprintHash).Scan(
		&out.ID,
		&out.FingerprintHash,
		&out.CanonicalName,
		&out.ConfidenceScore,
		&out.AssumptionsJSON,
		&out.ItemsJSON,
		&out.RawAnalysisJSON,
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
		&out.UsageCount,
		&out.LastUsedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("find meal_memory by fingerprint_hash: %w", err)
	}

	return &out, nil
}

func (r *MealMemoryRepository) Upsert(ctx context.Context, memory models.MealMemory) (*models.MealMemory, error) {
	const q = `
		INSERT INTO meal_memory (
			fingerprint_hash, canonical_name, confidence_score,
			assumptions_json, items_json, raw_analysis_json,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg,
			usage_count, last_used_at
		) VALUES (
			$1,$2,$3,
			$4,$5,$6,
			$7,$8,$9,$10,$11,$12,$13,
			$14,$15,$16,$17,$18,$19,
			$20,$21,$22,$23,
			$24, COALESCE($25, NOW())
		)
		ON CONFLICT (fingerprint_hash) DO UPDATE
		SET
			canonical_name = EXCLUDED.canonical_name,
			confidence_score = EXCLUDED.confidence_score,
			assumptions_json = EXCLUDED.assumptions_json,
			items_json = EXCLUDED.items_json,
			raw_analysis_json = EXCLUDED.raw_analysis_json,
			calories_kcal = EXCLUDED.calories_kcal,
			protein_g = EXCLUDED.protein_g,
			carbohydrate_g = EXCLUDED.carbohydrate_g,
			fat_g = EXCLUDED.fat_g,
			fiber_g = EXCLUDED.fiber_g,
			sugars_g = EXCLUDED.sugars_g,
			saturated_fat_g = EXCLUDED.saturated_fat_g,
			sodium_mg = EXCLUDED.sodium_mg,
			potassium_mg = EXCLUDED.potassium_mg,
			calcium_mg = EXCLUDED.calcium_mg,
			magnesium_mg = EXCLUDED.magnesium_mg,
			iron_mg = EXCLUDED.iron_mg,
			zinc_mg = EXCLUDED.zinc_mg,
			vitamin_d_mcg = EXCLUDED.vitamin_d_mcg,
			vitamin_b12_mcg = EXCLUDED.vitamin_b12_mcg,
			folate_b9_mcg = EXCLUDED.folate_b9_mcg,
			vitamin_c_mg = EXCLUDED.vitamin_c_mg,
			usage_count = meal_memory.usage_count + 1,
			last_used_at = NOW(),
			updated_at = NOW()
		RETURNING
			id, fingerprint_hash, canonical_name, confidence_score,
			assumptions_json, items_json, raw_analysis_json,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg,
			usage_count, last_used_at, created_at, updated_at
	`

	var out models.MealMemory
	if err := r.db.QueryRow(
		ctx,
		q,
		memory.FingerprintHash,
		memory.CanonicalName,
		memory.ConfidenceScore,
		memory.AssumptionsJSON,
		memory.ItemsJSON,
		memory.RawAnalysisJSON,
		memory.CaloriesKcal,
		memory.ProteinG,
		memory.CarbohydrateG,
		memory.FatG,
		memory.FiberG,
		memory.SugarsG,
		memory.SaturatedFatG,
		memory.SodiumMg,
		memory.PotassiumMg,
		memory.CalciumMg,
		memory.MagnesiumMg,
		memory.IronMg,
		memory.ZincMg,
		memory.VitaminDMcg,
		memory.VitaminB12Mcg,
		memory.FolateB9Mcg,
		memory.VitaminCMg,
		memory.UsageCount,
		memory.LastUsedAt,
	).Scan(
		&out.ID,
		&out.FingerprintHash,
		&out.CanonicalName,
		&out.ConfidenceScore,
		&out.AssumptionsJSON,
		&out.ItemsJSON,
		&out.RawAnalysisJSON,
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
		&out.UsageCount,
		&out.LastUsedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("upsert meal_memory: %w", err)
	}

	return &out, nil
}
