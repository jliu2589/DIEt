package repositories

import (
	"context"
	"errors"
	"fmt"
	"time"

	"diet/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DailyNutritionSummaryRepository struct {
	pool *pgxpool.Pool
}

func NewDailyNutritionSummaryRepository(pool *pgxpool.Pool) *DailyNutritionSummaryRepository {
	return &DailyNutritionSummaryRepository{pool: pool}
}

func (r *DailyNutritionSummaryRepository) GetByUserIDAndDate(ctx context.Context, userID string, summaryDate time.Time) (*models.DailyNutritionSummary, error) {
	const q = `
		SELECT
			id, user_id, summary_date,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg,
			created_at, updated_at
		FROM daily_nutrition_summary
		WHERE user_id = $1 AND summary_date = $2
	`

	var out models.DailyNutritionSummary
	if err := r.pool.QueryRow(ctx, q, userID, summaryDate).Scan(
		&out.ID,
		&out.UserID,
		&out.SummaryDate,
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
		return nil, fmt.Errorf("get daily_nutrition_summary by user and date: %w", err)
	}

	return &out, nil
}

func (r *DailyNutritionSummaryRepository) UpsertTotals(ctx context.Context, summary models.DailyNutritionSummary) (*models.DailyNutritionSummary, error) {
	const q = `
		INSERT INTO daily_nutrition_summary (
			user_id, summary_date,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg
		) VALUES (
			$1,$2,
			$3,$4,$5,$6,$7,$8,$9,
			$10,$11,$12,$13,$14,$15,
			$16,$17,$18,$19
		)
		ON CONFLICT (user_id, summary_date) DO UPDATE SET
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
			updated_at = NOW()
		RETURNING
			id, user_id, summary_date,
			calories_kcal, protein_g, carbohydrate_g, fat_g, fiber_g, sugars_g, saturated_fat_g,
			sodium_mg, potassium_mg, calcium_mg, magnesium_mg, iron_mg, zinc_mg,
			vitamin_d_mcg, vitamin_b12_mcg, folate_b9_mcg, vitamin_c_mg,
			created_at, updated_at
	`

	var out models.DailyNutritionSummary
	if err := r.pool.QueryRow(
		ctx,
		q,
		summary.UserID,
		summary.SummaryDate,
		summary.CaloriesKcal,
		summary.ProteinG,
		summary.CarbohydrateG,
		summary.FatG,
		summary.FiberG,
		summary.SugarsG,
		summary.SaturatedFatG,
		summary.SodiumMg,
		summary.PotassiumMg,
		summary.CalciumMg,
		summary.MagnesiumMg,
		summary.IronMg,
		summary.ZincMg,
		summary.VitaminDMcg,
		summary.VitaminB12Mcg,
		summary.FolateB9Mcg,
		summary.VitaminCMg,
	).Scan(
		&out.ID,
		&out.UserID,
		&out.SummaryDate,
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
		return nil, fmt.Errorf("upsert daily_nutrition_summary totals: %w", err)
	}

	return &out, nil
}
