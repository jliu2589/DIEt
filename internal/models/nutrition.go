package models

import (
	"encoding/json"
	"time"
)

// NutritionFields is a reusable set of nutrition metrics.
type NutritionFields struct {
	CaloriesKcal  *float64 `json:"calories_kcal,omitempty" db:"calories_kcal"`
	ProteinG      *float64 `json:"protein_g,omitempty" db:"protein_g"`
	CarbohydrateG *float64 `json:"carbohydrate_g,omitempty" db:"carbohydrate_g"`
	FatG          *float64 `json:"fat_g,omitempty" db:"fat_g"`
	FiberG        *float64 `json:"fiber_g,omitempty" db:"fiber_g"`
	SugarsG       *float64 `json:"sugars_g,omitempty" db:"sugars_g"`
	SaturatedFatG *float64 `json:"saturated_fat_g,omitempty" db:"saturated_fat_g"`
	SodiumMg      *float64 `json:"sodium_mg,omitempty" db:"sodium_mg"`
	PotassiumMg   *float64 `json:"potassium_mg,omitempty" db:"potassium_mg"`
	CalciumMg     *float64 `json:"calcium_mg,omitempty" db:"calcium_mg"`
	MagnesiumMg   *float64 `json:"magnesium_mg,omitempty" db:"magnesium_mg"`
	IronMg        *float64 `json:"iron_mg,omitempty" db:"iron_mg"`
	ZincMg        *float64 `json:"zinc_mg,omitempty" db:"zinc_mg"`
	VitaminDMcg   *float64 `json:"vitamin_d_mcg,omitempty" db:"vitamin_d_mcg"`
	VitaminB12Mcg *float64 `json:"vitamin_b12_mcg,omitempty" db:"vitamin_b12_mcg"`
	FolateB9Mcg   *float64 `json:"folate_b9_mcg,omitempty" db:"folate_b9_mcg"`
	VitaminCMg    *float64 `json:"vitamin_c_mg,omitempty" db:"vitamin_c_mg"`
}

// MealEvent maps to the meal_events table.
type MealEvent struct {
	ID               int64     `json:"id" db:"id"`
	UserID           string    `json:"user_id" db:"user_id"`
	Source           string    `json:"source" db:"source"`
	SourceMessageID  *string   `json:"source_message_id,omitempty" db:"source_message_id"`
	EventType        string    `json:"event_type" db:"event_type"`
	RawText          *string   `json:"raw_text,omitempty" db:"raw_text"`
	ImageURL         *string   `json:"image_url,omitempty" db:"image_url"`
	LoggedAt         time.Time `json:"logged_at" db:"logged_at"`
	EatenAt          time.Time `json:"eaten_at" db:"eaten_at"`
	TimeSource       string    `json:"time_source" db:"time_source"`
	ProcessingStatus string    `json:"processing_status" db:"processing_status"`
	FingerprintHash  *string   `json:"fingerprint_hash,omitempty" db:"fingerprint_hash"`
	CreatedAt        time.Time `json:"created_at" db:"created_at"`
	UpdatedAt        time.Time `json:"updated_at" db:"updated_at"`
}

// MealAnalysis maps to the meal_analysis table.
type MealAnalysis struct {
	ID              int64           `json:"id" db:"id"`
	MealEventID     int64           `json:"meal_event_id" db:"meal_event_id"`
	UserID          string          `json:"user_id" db:"user_id"`
	CanonicalName   string          `json:"canonical_name" db:"canonical_name"`
	ConfidenceScore *float64        `json:"confidence_score,omitempty" db:"confidence_score"`
	AssumptionsJSON json.RawMessage `json:"assumptions_json,omitempty" db:"assumptions_json"`
	ItemsJSON       json.RawMessage `json:"items_json,omitempty" db:"items_json"`
	RawAnalysisJSON json.RawMessage `json:"raw_analysis_json,omitempty" db:"raw_analysis_json"`
	NutritionFields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MealMemory maps to the meal_memory table.
type MealMemory struct {
	ID              int64           `json:"id" db:"id"`
	FingerprintHash string          `json:"fingerprint_hash" db:"fingerprint_hash"`
	CanonicalName   string          `json:"canonical_name" db:"canonical_name"`
	ConfidenceScore *float64        `json:"confidence_score,omitempty" db:"confidence_score"`
	AssumptionsJSON json.RawMessage `json:"assumptions_json,omitempty" db:"assumptions_json"`
	ItemsJSON       json.RawMessage `json:"items_json,omitempty" db:"items_json"`
	RawAnalysisJSON json.RawMessage `json:"raw_analysis_json,omitempty" db:"raw_analysis_json"`
	NutritionFields
	UsageCount int32      `json:"usage_count" db:"usage_count"`
	LastUsedAt *time.Time `json:"last_used_at,omitempty" db:"last_used_at"`
	CreatedAt  time.Time  `json:"created_at" db:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at" db:"updated_at"`
}

// DailyNutritionSummary maps to the daily_nutrition_summary table.
type DailyNutritionSummary struct {
	ID          int64     `json:"id" db:"id"`
	UserID      string    `json:"user_id" db:"user_id"`
	SummaryDate time.Time `json:"summary_date" db:"summary_date"`
	NutritionFields
	CreatedAt time.Time `json:"created_at" db:"created_at"`
	UpdatedAt time.Time `json:"updated_at" db:"updated_at"`
}

// MealItem is a normalized item-level view used in analysis payloads.
type MealItem struct {
	Name          string   `json:"name"`
	Quantity      *float64 `json:"quantity,omitempty"`
	Unit          string   `json:"unit,omitempty"`
	CaloriesKcal  *float64 `json:"calories_kcal,omitempty"`
	ProteinG      *float64 `json:"protein_g,omitempty"`
	CarbohydrateG *float64 `json:"carbohydrate_g,omitempty"`
	FatG          *float64 `json:"fat_g,omitempty"`
}

// OpenAIMealAnalysisResponse represents a parsed OpenAI meal-analysis payload.
type OpenAIMealAnalysisResponse struct {
	CanonicalName   string     `json:"canonical_name"`
	ConfidenceScore *float64   `json:"confidence_score,omitempty"`
	Assumptions     []string   `json:"assumptions,omitempty"`
	Items           []MealItem `json:"items,omitempty"`
	NutritionFields
}
