package models

// HealthResponse is returned by health-check endpoints.
type HealthResponse struct {
	OK bool `json:"ok"`
}
