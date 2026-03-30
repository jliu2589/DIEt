package repositories

import (
	"context"
	"fmt"

	"diet/internal/models"
	"github.com/jackc/pgx/v5/pgxpool"
)

type MealItemsRepository struct {
	db DBTX
}

func NewMealItemsRepository(pool *pgxpool.Pool) *MealItemsRepository {
	return &MealItemsRepository{db: pool}
}

func NewMealItemsRepositoryWithDB(db DBTX) *MealItemsRepository {
	return &MealItemsRepository{db: db}
}

func (r *MealItemsRepository) Create(ctx context.Context, in models.StoredMealItem) (*models.StoredMealItem, error) {
	if in.MealID <= 0 {
		return nil, fmt.Errorf("meal_id must be > 0")
	}
	if in.FoodID <= 0 {
		return nil, fmt.Errorf("food_id must be > 0")
	}
	if in.Quantity < 0 {
		return nil, fmt.Errorf("quantity must be >= 0")
	}
	if in.Unit == "" {
		return nil, fmt.Errorf("unit is required")
	}

	const q = `
		INSERT INTO meal_items (
			meal_id,
			food_id,
			quantity,
			unit
		) VALUES (
			$1,
			$2,
			$3,
			$4
		)
		RETURNING
			id,
			meal_id,
			food_id,
			quantity,
			unit,
			created_at
	`

	var out models.StoredMealItem
	if err := r.db.QueryRow(ctx, q, in.MealID, in.FoodID, in.Quantity, in.Unit).Scan(
		&out.ID,
		&out.MealID,
		&out.FoodID,
		&out.Quantity,
		&out.Unit,
		&out.CreatedAt,
	); err != nil {
		return nil, fmt.Errorf("create meal_item: %w", err)
	}

	return &out, nil
}

func (r *MealItemsRepository) ListByMealID(ctx context.Context, mealID int64) ([]models.StoredMealItem, error) {
	const q = `
		SELECT
			id,
			meal_id,
			food_id,
			quantity,
			unit,
			created_at
		FROM meal_items
		WHERE meal_id = $1
		ORDER BY id ASC
	`

	rows, err := r.db.Query(ctx, q, mealID)
	if err != nil {
		return nil, fmt.Errorf("list meal_items by meal_id: %w", err)
	}
	defer rows.Close()

	items := make([]models.StoredMealItem, 0)
	for rows.Next() {
		var out models.StoredMealItem
		if err := rows.Scan(
			&out.ID,
			&out.MealID,
			&out.FoodID,
			&out.Quantity,
			&out.Unit,
			&out.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("scan meal_item row: %w", err)
		}
		items = append(items, out)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate meal_item rows: %w", err)
	}

	return items, nil
}
