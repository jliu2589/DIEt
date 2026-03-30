package repositories

import (
	"context"
	"errors"
	"fmt"

	"diet/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MealAnalysisRepository struct {
	db DBTX
}

func NewMealAnalysisRepository(pool *pgxpool.Pool) *MealAnalysisRepository {
	return &MealAnalysisRepository{db: pool}
}

func NewMealAnalysisRepositoryWithDB(db DBTX) *MealAnalysisRepository {
	return &MealAnalysisRepository{db: db}
}

func (r *MealAnalysisRepository) Insert(ctx context.Context, analysis models.MealAnalysis) error {
	const q = `
		INSERT INTO meal_analysis (
			meal_event_id, user_id, canonical_name, confidence_score,
			assumptions_json, items_json, raw_analysis_json,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg
		) VALUES (
			$1,$2,$3,$4,
			$5,$6,$7,
			$8,$9,$10,$11,$12,$13,$14,
			$15,$16,$17,$18,$19,$20,
			$21,$22,$23,$24
		)
	`

	if _, err := r.db.Exec(
		ctx,
		q,
		analysis.MealEventID,
		analysis.UserID,
		analysis.CanonicalName,
		analysis.ConfidenceScore,
		analysis.AssumptionsJSON,
		analysis.ItemsJSON,
		analysis.RawAnalysisJSON,
		analysis.CaloriesKcal,
		analysis.ProteinG,
		analysis.CarbohydrateG,
		analysis.FatG,
		analysis.FiberG,
		analysis.SugarsG,
		analysis.SaturatedFatG,
		analysis.SodiumMg,
		analysis.PotassiumMg,
		analysis.CalciumMg,
		analysis.MagnesiumMg,
		analysis.IronMg,
		analysis.ZincMg,
		analysis.VitaminDMcg,
		analysis.VitaminB12Mcg,
		analysis.FolateB9Mcg,
		analysis.VitaminCMg,
	); err != nil {
		return fmt.Errorf("insert meal_analysis: %w", err)
	}

	return nil
}

func (r *MealAnalysisRepository) GetByMealEventIDAndUserID(ctx context.Context, mealEventID int64, userID string) (*models.MealAnalysis, error) {
	const q = `
		SELECT
			id, meal_event_id, user_id, canonical_name, confidence_score,
			assumptions_json, items_json, raw_analysis_json,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg,
			created_at, updated_at
		FROM meal_analysis
		WHERE meal_event_id = $1 AND user_id = $2
		LIMIT 1
	`
	var out models.MealAnalysis
	if err := r.db.QueryRow(ctx, q, mealEventID, userID).Scan(
		&out.ID,
		&out.MealEventID,
		&out.UserID,
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
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get meal_analysis by meal_event_id and user_id: %w", err)
	}
	return &out, nil
}

func (r *MealAnalysisRepository) UpdateEditableByMealEventIDAndUserID(ctx context.Context, in models.MealAnalysis) error {
	const q = `
		UPDATE meal_analysis
		SET
			canonical_name = COALESCE(NULLIF($3, ''), canonical_name),
			calories_kcal = COALESCE($4, calories_kcal),
			protein_g = COALESCE($5, protein_g),
			carbohydrate_g = COALESCE($6, carbohydrate_g),
			fat_g = COALESCE($7, fat_g),
			updated_at = NOW()
		WHERE meal_event_id = $1 AND user_id = $2
	`
	if _, err := r.db.Exec(ctx, q, in.MealEventID, in.UserID, in.CanonicalName, in.CaloriesKcal, in.ProteinG, in.CarbohydrateG, in.FatG); err != nil {
		return fmt.Errorf("update meal_analysis editable fields: %w", err)
	}
	return nil
}
