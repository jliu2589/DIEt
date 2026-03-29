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

type WeightEntriesRepository struct {
	pool *pgxpool.Pool
}

type DailyWeightEntry struct {
	Date   time.Time
	Weight float64
}

func NewWeightEntriesRepository(pool *pgxpool.Pool) *WeightEntriesRepository {
	return &WeightEntriesRepository{pool: pool}
}

func (r *WeightEntriesRepository) Insert(ctx context.Context, entry models.WeightEntry) (*models.WeightEntry, error) {
	const q = `
		INSERT INTO weight_entries (
			user_id,
			weight,
			unit,
			logged_at
		) VALUES (
			$1,
			$2,
			$3,
			$4
		)
		RETURNING
			id,
			user_id,
			weight,
			unit,
			logged_at,
			created_at,
			updated_at
	`

	var out models.WeightEntry
	if err := r.pool.QueryRow(ctx, q, entry.UserID, entry.Weight, entry.Unit, entry.LoggedAt).Scan(
		&out.ID,
		&out.UserID,
		&out.Weight,
		&out.Unit,
		&out.LoggedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		return nil, fmt.Errorf("insert weight_entry: %w", err)
	}

	return &out, nil
}

func (r *WeightEntriesRepository) GetLatestByUserID(ctx context.Context, userID string) (*models.WeightEntry, error) {
	const q = `
		SELECT
			id,
			user_id,
			weight,
			unit,
			logged_at,
			created_at,
			updated_at
		FROM weight_entries
		WHERE user_id = $1
		ORDER BY logged_at DESC, id DESC
		LIMIT 1
	`

	var out models.WeightEntry
	if err := r.pool.QueryRow(ctx, q, userID).Scan(
		&out.ID,
		&out.UserID,
		&out.Weight,
		&out.Unit,
		&out.LoggedAt,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get latest weight_entry by user_id: %w", err)
	}

	return &out, nil
}

func (r *WeightEntriesRepository) ListRecentByUserID(ctx context.Context, userID string, limit int) ([]models.WeightEntry, error) {
	const q = `
		SELECT
			id,
			user_id,
			weight,
			unit,
			logged_at,
			created_at,
			updated_at
		FROM weight_entries
		WHERE user_id = $1
		ORDER BY logged_at DESC, id DESC
		LIMIT $2
	`

	rows, err := r.pool.Query(ctx, q, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("list recent weight_entries by user_id: %w", err)
	}
	defer rows.Close()

	entries := make([]models.WeightEntry, 0, limit)
	for rows.Next() {
		var item models.WeightEntry
		if err := rows.Scan(
			&item.ID,
			&item.UserID,
			&item.Weight,
			&item.Unit,
			&item.LoggedAt,
			&item.CreatedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan recent weight_entry: %w", err)
		}
		entries = append(entries, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate recent weight_entries: %w", err)
	}

	return entries, nil
}

func (r *WeightEntriesRepository) ListDailyLatestByUserIDAndDateRange(ctx context.Context, userID string, startAt, endExclusive time.Time) ([]DailyWeightEntry, error) {
	const q = `
		SELECT DISTINCT ON (entry_date)
			entry_date,
			weight
		FROM (
			SELECT
				(logged_at AT TIME ZONE 'UTC')::date AS entry_date,
				weight,
				logged_at,
				id
			FROM weight_entries
			WHERE user_id = $1
				AND logged_at >= $2
				AND logged_at < $3
		) w
		ORDER BY entry_date ASC, logged_at DESC, id DESC
	`

	rows, err := r.pool.Query(ctx, q, userID, startAt, endExclusive)
	if err != nil {
		return nil, fmt.Errorf("list daily latest weight_entries by user and date range: %w", err)
	}
	defer rows.Close()

	out := make([]DailyWeightEntry, 0)
	for rows.Next() {
		var item DailyWeightEntry
		if err := rows.Scan(&item.Date, &item.Weight); err != nil {
			return nil, fmt.Errorf("scan daily latest weight_entry: %w", err)
		}
		out = append(out, item)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate daily latest weight_entries: %w", err)
	}

	return out, nil
}
