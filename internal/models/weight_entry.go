package models

import "time"

// WeightEntry maps to the weight_entries table.
type WeightEntry struct {
	ID        int64     `json:"id" db:"id"`
	UserID    string    `json:"user_id" db:"user_id"`
	Weight    float64   `json:"weight" db:"weight"`
	Unit      string    `json:"unit" db:"unit"`
	LoggedAt  time.Time `json:"logged_at" db:"logged_at"`
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}
