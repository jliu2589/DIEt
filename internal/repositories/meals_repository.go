package repositories

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"diet/internal/models"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MealsRepository struct {
	pool *pgxpool.Pool
}

func NewMealsRepository(pool *pgxpool.Pool) *MealsRepository {
	return &MealsRepository{pool: pool}
}

func (r *MealsRepository) Create(ctx context.Context, in models.Meal) (*models.Meal, error) {
	name := strings.TrimSpace(in.CanonicalName)
	if name == "" {
		return nil, fmt.Errorf("canonical_name is required")
	}

	const q = `
		INSERT INTO meals (
			canonical_name,
			fingerprint_hash,
			source_type,
			confidence_score
		) VALUES (
			$1,
			$2,
			$3,
			$4
		)
		RETURNING
			id,
			canonical_name,
			fingerprint_hash,
			source_type,
			confidence_score,
			created_at,
			updated_at
	`

	var out models.Meal
	if err := r.pool.QueryRow(ctx, q, name, in.FingerprintHash, in.SourceType, in.ConfidenceScore).Scan(
		&out.ID,
		&out.CanonicalName,
		&out.FingerprintHash,
		&out.SourceType,
		&out.ConfidenceScore,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, fmt.Errorf("create meal: fingerprint_hash already exists")
		}
		return nil, fmt.Errorf("create meal: %w", err)
	}

	return &out, nil
}

func (r *MealsRepository) GetByID(ctx context.Context, id int64) (*models.Meal, error) {
	const q = `
		SELECT
			id,
			canonical_name,
			fingerprint_hash,
			source_type,
			confidence_score,
			created_at,
			updated_at
		FROM meals
		WHERE id = $1
	`

	var out models.Meal
	if err := r.pool.QueryRow(ctx, q, id).Scan(
		&out.ID,
		&out.CanonicalName,
		&out.FingerprintHash,
		&out.SourceType,
		&out.ConfidenceScore,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get meal by id: %w", err)
	}

	return &out, nil
}

func (r *MealsRepository) GetByFingerprintHash(ctx context.Context, fingerprintHash string) (*models.Meal, error) {
	hash := strings.TrimSpace(fingerprintHash)
	if hash == "" {
		return nil, fmt.Errorf("fingerprint_hash is required")
	}

	const q = `
		SELECT
			id,
			canonical_name,
			fingerprint_hash,
			source_type,
			confidence_score,
			created_at,
			updated_at
		FROM meals
		WHERE fingerprint_hash = $1
	`

	var out models.Meal
	if err := r.pool.QueryRow(ctx, q, hash).Scan(
		&out.ID,
		&out.CanonicalName,
		&out.FingerprintHash,
		&out.SourceType,
		&out.ConfidenceScore,
		&out.CreatedAt,
		&out.UpdatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("get meal by fingerprint_hash: %w", err)
	}

	return &out, nil
}
