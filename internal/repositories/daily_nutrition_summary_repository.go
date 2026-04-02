package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"diet/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type DailyNutritionSummaryRepository struct {
	db DBTX
}

type DailyNutritionSummaryRow struct {
	SummaryDate   time.Time
	CaloriesKcal  *float64
	ProteinG      *float64
	CarbohydrateG *float64
	FatG          *float64
}

func NewDailyNutritionSummaryRepository(pool *pgxpool.Pool) *DailyNutritionSummaryRepository {
	return &DailyNutritionSummaryRepository{db: pool}
}

func NewDailyNutritionSummaryRepositoryWithDB(db DBTX) *DailyNutritionSummaryRepository {
	return &DailyNutritionSummaryRepository{db: db}
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
	if err := r.db.QueryRow(ctx, q, userID, summaryDate).Scan(
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
	if err := r.db.QueryRow(
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

func (r *DailyNutritionSummaryRepository) ListByUserIDAndDateRange(ctx context.Context, userID string, startDate, endDate time.Time) ([]DailyNutritionSummaryRow, error) {
	const q = `
		SELECT
			summary_date,
			calories_kcal,
			protein_g,
			carbohydrate_g,
			fat_g
		FROM daily_nutrition_summary
		WHERE user_id = $1
			AND summary_date >= $2
			AND summary_date <= $3
		ORDER BY summary_date ASC
	`

	rows, err := r.db.Query(ctx, q, userID, startDate, endDate)
	if err != nil {
		return nil, fmt.Errorf("list daily_nutrition_summary by user and date range: %w", err)
	}
	defer rows.Close()

	out := make([]DailyNutritionSummaryRow, 0)
	for rows.Next() {
		var item DailyNutritionSummaryRow
		if err := rows.Scan(
			&item.SummaryDate,
			&item.CaloriesKcal,
			&item.ProteinG,
			&item.CarbohydrateG,
			&item.FatG,
		); err != nil {
			return nil, fmt.Errorf("scan daily_nutrition_summary row: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily_nutrition_summary rows: %w", err)
	}

	return out, nil
}

// ReconcileForUserDate deterministically recomputes and persists the daily summary
// from meal_events + meal_analysis for a user/date.
func (r *DailyNutritionSummaryRepository) ReconcileForUserDate(ctx context.Context, userID string, summaryDate time.Time) error {
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return fmt.Errorf("user_id is required")
	}
	summaryDate = time.Date(summaryDate.UTC().Year(), summaryDate.UTC().Month(), summaryDate.UTC().Day(), 0, 0, 0, 0, time.UTC)

	summary, err := r.AggregateForUserDate(ctx, userID, summaryDate)
	if err != nil {
		return err
	}

	if !hasAnyNutrition(*summary) {
		const deleteQ = `
			DELETE FROM daily_nutrition_summary
			WHERE user_id = $1 AND summary_date = $2
		`
		if _, err := r.db.Exec(ctx, deleteQ, userID, summaryDate); err != nil {
			return fmt.Errorf("delete empty daily summary during reconciliation: %w", err)
		}
		return nil
	}

	if _, err := r.UpsertTotals(ctx, *summary); err != nil {
		return fmt.Errorf("upsert reconciled daily summary: %w", err)
	}
	return nil
}

func (r *DailyNutritionSummaryRepository) AggregateForUserDate(ctx context.Context, userID string, summaryDate time.Time) (*models.DailyNutritionSummary, error) {
	const aggregateQ = `
			SELECT
				SUM(ma.calories_kcal),
			SUM(ma.protein_g),
			SUM(ma.carbohydrate_g),
			SUM(ma.fat_g),
			SUM(ma.fiber_g),
			SUM(ma.sugars_g),
			SUM(ma.saturated_fat_g),
			SUM(ma.sodium_mg),
			SUM(ma.potassium_mg),
			SUM(ma.calcium_mg),
			SUM(ma.magnesium_mg),
			SUM(ma.iron_mg),
			SUM(ma.zinc_mg),
			SUM(ma.vitamin_d_mcg),
			SUM(ma.vitamin_b12_mcg),
			SUM(ma.folate_b9_mcg),
			SUM(ma.vitamin_c_mg)
		FROM meal_events me
		INNER JOIN meal_analysis ma ON ma.meal_event_id = me.id
			WHERE me.user_id = $1
				AND me.eaten_at::date = $2::date
				AND me.processing_status = 'processed'
		`

	summary := models.DailyNutritionSummary{
		UserID:      userID,
		SummaryDate: summaryDate,
	}
	if err := r.db.QueryRow(ctx, aggregateQ, userID, summaryDate).Scan(
		&summary.CaloriesKcal,
		&summary.ProteinG,
		&summary.CarbohydrateG,
		&summary.FatG,
		&summary.FiberG,
		&summary.SugarsG,
		&summary.SaturatedFatG,
		&summary.SodiumMg,
		&summary.PotassiumMg,
		&summary.CalciumMg,
		&summary.MagnesiumMg,
		&summary.IronMg,
		&summary.ZincMg,
		&summary.VitaminDMcg,
		&summary.VitaminB12Mcg,
		&summary.FolateB9Mcg,
		&summary.VitaminCMg,
	); err != nil {
		return nil, fmt.Errorf("aggregate daily summary: %w", err)
	}
	return &summary, nil
}

func hasAnyNutrition(s models.DailyNutritionSummary) bool {
	return s.CaloriesKcal != nil ||
		s.ProteinG != nil ||
		s.CarbohydrateG != nil ||
		s.FatG != nil ||
		s.FiberG != nil ||
		s.SugarsG != nil ||
		s.SaturatedFatG != nil ||
		s.SodiumMg != nil ||
		s.PotassiumMg != nil ||
		s.CalciumMg != nil ||
		s.MagnesiumMg != nil ||
		s.IronMg != nil ||
		s.ZincMg != nil ||
		s.VitaminDMcg != nil ||
		s.VitaminB12Mcg != nil ||
		s.FolateB9Mcg != nil ||
		s.VitaminCMg != nil
}
