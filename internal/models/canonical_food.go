package models

import "time"

// CanonicalFoodWithNutrition represents a canonical food row and its default nutrition profile.
type CanonicalFoodWithNutrition struct {
	ID            int64   `json:"id" db:"id"`
	CanonicalName string  `json:"canonical_name" db:"canonical_name"`
	DefaultAmount float64 `json:"default_amount" db:"default_amount"`
	DefaultUnit   string  `json:"default_unit" db:"default_unit"`
	Category      *string `json:"category,omitempty" db:"category"`
	SourceType    *string `json:"source_type,omitempty" db:"source_type"`

	NutritionFields

	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
