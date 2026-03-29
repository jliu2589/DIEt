package models

import "time"

// Meal represents a reusable meal template or interpreted meal cached in storage.
type Meal struct {
	ID              int64     `json:"id" db:"id"`
	CanonicalName   string    `json:"canonical_name" db:"canonical_name"`
	FingerprintHash *string   `json:"fingerprint_hash,omitempty" db:"fingerprint_hash"`
	SourceType      *string   `json:"source_type,omitempty" db:"source_type"`
	ConfidenceScore *float64  `json:"confidence_score,omitempty" db:"confidence_score"`
	CreatedAt       time.Time `json:"created_at" db:"created_at"`
	UpdatedAt       time.Time `json:"updated_at" db:"updated_at"`
}

// StoredMealItem links a stored meal to a canonical food and quantity.
type StoredMealItem struct {
	ID        int64     `json:"id" db:"id"`
	MealID    int64     `json:"meal_id" db:"meal_id"`
	FoodID    int64     `json:"food_id" db:"food_id"`
	Quantity  float64   `json:"quantity" db:"quantity"`
	Unit      string    `json:"unit" db:"unit"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
}
