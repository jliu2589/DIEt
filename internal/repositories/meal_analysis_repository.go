package repositories

import (
	"context"
	"fmt"

	"diet/internal/models"
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
