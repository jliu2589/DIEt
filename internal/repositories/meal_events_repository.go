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

type MealEventsRepository struct {
	db DBTX
}

type RecentMeal struct {
	MealEventID   int64
	CanonicalName string
	LoggedAt      time.Time
	EatenAt       time.Time
	TimeSource    string
	CaloriesKcal  *float64
	ProteinG      *float64
	CarbohydrateG *float64
	FatG          *float64
	Source        string
}

type MealTimeUpdate struct {
	MealEventID   int64
	CanonicalName string
	EatenAt       time.Time
	TimeSource    string
}

func NewMealEventsRepository(pool *pgxpool.Pool) *MealEventsRepository {
	return &MealEventsRepository{db: pool}
}

func NewMealEventsRepositoryWithDB(db DBTX) *MealEventsRepository {
	return &MealEventsRepository{db: db}
}

func (r *MealEventsRepository) Insert(ctx context.Context, event models.MealEvent) (*models.MealEvent, error) {
	const q = `
		INSERT INTO meal_events (
			user_id, source, source_message_id, event_type, raw_text, image_url, logged_at, eaten_at, time_source, processing_status, fingerprint_hash
		) VALUES ($1,$2,$3,$4,$5,$6,$7,$8,$9,$10,$11)
		RETURNING id, user_id, source, source_message_id, event_type, raw_text, image_url, logged_at, eaten_at, time_source, processing_status, fingerprint_hash, created_at, updated_at
	`

	var out models.MealEvent
	if err := r.db.QueryRow(
		ctx,
		q,
		event.UserID,
		event.Source,
		event.SourceMessageID,
		event.EventType,
		event.RawText,
		event.ImageURL,
		event.LoggedAt,
		event.EatenAt,
		event.TimeSource,
		event.ProcessingStatus,
		event.FingerprintHash,
	).Scan(
		&out.ID,
		&out.UserID,
		&out.Source,
		&out.SourceMessageID,
		&out.EventType,
		&out.RawText,
		&out.ImageURL,
		&out.LoggedAt,
		&out.EatenAt,
		&out.TimeSource,
		&out.ProcessingStatus,
		&out.FingerprintHash,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert meal_event: %w", err)
	}

	return &out, nil
}

func (r *MealEventsRepository) GetByID(ctx context.Context, id int64) (*models.MealEvent, error) {
	const q = `
		SELECT id, user_id, source, source_message_id, event_type, raw_text, image_url, logged_at, eaten_at, time_source, processing_status, fingerprint_hash, created_at, updated_at
		FROM meal_events
		WHERE id = $1
	`

	var out models.MealEvent
	if err := r.db.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.UserID,
		&out.Source,
		&out.SourceMessageID,
		&out.EventType,
		&out.RawText,
		&out.ImageURL,
		&out.LoggedAt,
		&out.EatenAt,
		&out.TimeSource,
		&out.ProcessingStatus,
		&out.FingerprintHash,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("get meal_event by id: %w", err)
	}

	return &out, nil
}

func (r *MealEventsRepository) UpdateProcessingStatus(ctx context.Context, id int64, status string) error {
	const q = `
		UPDATE meal_events
		SET processing_status = $2, updated_at = NOW()
		WHERE id = $1
	`

	if _, err := r.db.Exec(ctx, q, id, status); err != nil {
		return fmt.Errorf("update meal_event processing status: %w", err)
	}

	return nil
}

func (r *MealEventsRepository) ListRecentByUserID(ctx context.Context, userID string, limit int) ([]RecentMeal, error) {
	const q = `
		SELECT
			me.id,
			ma.canonical_name,
			me.logged_at,
			me.eaten_at,
			me.time_source,
			ma.calories_kcal,
			ma.protein_g,
			ma.carbohydrate_g,
			ma.fat_g,
			me.source
		FROM meal_events me
		INNER JOIN meal_analysis ma ON ma.meal_event_id = me.id
		WHERE me.user_id = $1
		ORDER BY me.eaten_at DESC, me.id DESC
		LIMIT $2
	`

	rows, err := r.db.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent meals by user id: %w", err)
	}
	defer rows.Close()

	out := make([]RecentMeal, 0, limit)
	for rows.Next() {
		var item RecentMeal
		if err := rows.Scan(
			&item.MealEventID,
			&item.CanonicalName,
			&item.LoggedAt,
			&item.EatenAt,
			&item.TimeSource,
			&item.CaloriesKcal,
			&item.ProteinG,
			&item.CarbohydrateG,
			&item.FatG,
			&item.Source,
		); err != nil {
			return nil, fmt.Errorf("scan recent meal row: %w", err)
		}
		out = append(out, item)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent meal rows: %w", err)
	}

	return out, nil
}

func (r *MealEventsRepository) UpdateEatenAtByIDAndUserID(ctx context.Context, mealEventID int64, userID string, eatenAt time.Time) (*MealTimeUpdate, error) {
	const q = `
		UPDATE meal_events me
		SET eaten_at = $3, time_source = 'edited', updated_at = NOW()
		WHERE me.id = $1 AND me.user_id = $2
		RETURNING
			me.id,
			COALESCE((
				SELECT ma.canonical_name
				FROM meal_analysis ma
				WHERE ma.meal_event_id = me.id
				LIMIT 1
			), ''),
			me.eaten_at,
			me.time_source
	`

	var out MealTimeUpdate
	if err := r.db.QueryRow(ctx, q, mealEventID, userID, eatenAt).Scan(
		&out.MealEventID,
		&out.CanonicalName,
		&out.EatenAt,
		&out.TimeSource,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("update meal_event eaten_at by id and user_id: %w", err)
	}

	return &out, nil
}
