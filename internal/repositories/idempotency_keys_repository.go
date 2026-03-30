package repositories

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type IdempotencyStatus string

const (
	IdempotencyStatusProcessing IdempotencyStatus = "processing"
	IdempotencyStatusSucceeded  IdempotencyStatus = "succeeded"
)

type IdempotencyRecord struct {
	ID             int64
	UserID         string
	Endpoint       string
	IdempotencyKey string
	RequestHash    string
	Status         IdempotencyStatus
	HTTPStatus     *int
	ResponseJSON   json.RawMessage
	CreatedAt      time.Time
	UpdatedAt      time.Time
}

type IdempotencyKeysRepository struct {
	db DBTX
}

func NewIdempotencyKeysRepository(pool *pgxpool.Pool) *IdempotencyKeysRepository {
	return &IdempotencyKeysRepository{db: pool}
}

func (r *IdempotencyKeysRepository) BeginOrGet(ctx context.Context, userID, endpoint, key, requestHash string) (*IdempotencyRecord, bool, error) {
	const insertQ = `
		INSERT INTO idempotency_keys (
			user_id,
			endpoint,
			idempotency_key,
			request_hash,
			status
		) VALUES ($1,$2,$3,$4,'processing')
		ON CONFLICT (user_id, endpoint, idempotency_key) DO NOTHING
		RETURNING id, user_id, endpoint, idempotency_key, request_hash, status, http_status, response_json, created_at, updated_at
	`

	var rec IdempotencyRecord
	if err := r.db.QueryRow(ctx, insertQ, userID, endpoint, key, requestHash).Scan(
		&rec.ID,
		&rec.UserID,
		&rec.Endpoint,
		&rec.IdempotencyKey,
		&rec.RequestHash,
		&rec.Status,
		&rec.HTTPStatus,
		&rec.ResponseJSON,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err == nil {
		return &rec, true, nil
	} else if !errors.Is(err, pgx.ErrNoRows) {
		return nil, false, fmt.Errorf("begin idempotency key: %w", err)
	}

	const getQ = `
		SELECT id, user_id, endpoint, idempotency_key, request_hash, status, http_status, response_json, created_at, updated_at
		FROM idempotency_keys
		WHERE user_id = $1 AND endpoint = $2 AND idempotency_key = $3
	`
	if err := r.db.QueryRow(ctx, getQ, userID, endpoint, key).Scan(
		&rec.ID,
		&rec.UserID,
		&rec.Endpoint,
		&rec.IdempotencyKey,
		&rec.RequestHash,
		&rec.Status,
		&rec.HTTPStatus,
		&rec.ResponseJSON,
		&rec.CreatedAt,
		&rec.UpdatedAt,
	); err != nil {
		return nil, false, fmt.Errorf("get existing idempotency key: %w", err)
	}

	return &rec, false, nil
}

func (r *IdempotencyKeysRepository) MarkSucceeded(ctx context.Context, id int64, httpStatus int, responseJSON []byte) error {
	const q = `
		UPDATE idempotency_keys
		SET status = 'succeeded', http_status = $2, response_json = $3, updated_at = NOW()
		WHERE id = $1
	`
	if _, err := r.db.Exec(ctx, q, id, httpStatus, responseJSON); err != nil {
		return fmt.Errorf("mark idempotency key succeeded: %w", err)
	}
	return nil
}

func (r *IdempotencyKeysRepository) Delete(ctx context.Context, id int64) error {
	const q = `DELETE FROM idempotency_keys WHERE id = $1`
	if _, err := r.db.Exec(ctx, q, id); err != nil {
		return fmt.Errorf("delete idempotency key: %w", err)
	}
	return nil
}
