package models

import "time"

// UserSettings maps to the user_settings table.
type UserSettings struct {
	UserID       string    `json:"user_id" db:"user_id"`
	Name         *string   `json:"name,omitempty" db:"name"`
	HeightCM     *float64  `json:"height_cm,omitempty" db:"height_cm"`
	WeightGoalKG *float64  `json:"weight_goal_kg,omitempty" db:"weight_goal_kg"`
	CalorieGoal  *float64  `json:"calorie_goal,omitempty" db:"calorie_goal"`
	ProteinGoalG *float64  `json:"protein_goal_g,omitempty" db:"protein_goal_g"`
	WeightUnit   string    `json:"weight_unit" db:"weight_unit"`
	CreatedAt    time.Time `json:"created_at" db:"created_at"`
	UpdatedAt    time.Time `json:"updated_at" db:"updated_at"`
}
