package repositories

import (
	"context"
	"errors"
	"fmt"

	"diet/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserSettingsRepository struct {
	pool *pgxpool.Pool
}

func NewUserSettingsRepository(pool *pgxpool.Pool) *UserSettingsRepository {
	return &UserSettingsRepository{pool: pool}
}

func (r *UserSettingsRepository) GetByUserID(ctx context.Context, userID string) (*models.UserSettings, error) {
	const q = `
		SELECT
			user_id,
			name,
			height_cm,
			weight_goal_kg,
			calorie_goal,
			protein_goal_g,
			weight_unit,
			created_at,
			updated_at
		FROM user_settings
		WHERE user_id = $1
	`

	var out models.UserSettings
	if err := r.pool.QueryRow(ctx, q, userID).Scan(
		&out.UserID,
		&out.Name,
		&out.HeightCM,
		&out.WeightGoalKG,
		&out.CalorieGoal,
		&out.ProteinGoalG,
		&out.WeightUnit,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get user_settings by user_id: %w", err)
	}

	return &out, nil
}

func (r *UserSettingsRepository) Upsert(ctx context.Context, settings models.UserSettings) (*models.UserSettings, error) {
	const q = `
		INSERT INTO user_settings (
			user_id,
			name,
			height_cm,
			weight_goal_kg,
			calorie_goal,
			protein_goal_g,
			weight_unit
		) VALUES (
			$1,
			$2,
			$3,
			$4,
			$5,
			$6,
			$7
		)
		ON CONFLICT (user_id) DO UPDATE SET
			name = EXCLUDED.name,
			height_cm = EXCLUDED.height_cm,
			weight_goal_kg = EXCLUDED.weight_goal_kg,
			calorie_goal = EXCLUDED.calorie_goal,
			protein_goal_g = EXCLUDED.protein_goal_g,
			weight_unit = EXCLUDED.weight_unit,
			updated_at = NOW()
		RETURNING
			user_id,
			name,
			height_cm,
			weight_goal_kg,
			calorie_goal,
			protein_goal_g,
			weight_unit,
			created_at,
			updated_at
	`

	var out models.UserSettings
	if err := r.pool.QueryRow(
		ctx,
		q,
		settings.UserID,
		settings.Name,
		settings.HeightCM,
		settings.WeightGoalKG,
		settings.CalorieGoal,
		settings.ProteinGoalG,
		settings.WeightUnit,
	).Scan(
		&out.UserID,
		&out.Name,
		&out.HeightCM,
		&out.WeightGoalKG,
		&out.CalorieGoal,
		&out.ProteinGoalG,
		&out.WeightUnit,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("upsert user_settings: %w", err)
	}

	return &out, nil
}
