package repositories

import "github.com/jackc/pgx/v5/pgxpool"

// MealRepository is the data-access layer contract for meal entries.
type MealRepository interface{}

type mealRepository struct {
	pool *pgxpool.Pool
}

func NewMealRepository(pool *pgxpool.Pool) MealRepository {
	return &mealRepository{pool: pool}
}
