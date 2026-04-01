package openai

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

const defaultBaseURL = "https://api.openai.com/v1"

// NutritionV1 contains only the V1 nutrition fields.
type NutritionV1 struct {
	CaloriesKcal  *float64 `json:"calories_kcal"`
	ProteinG      *float64 `json:"protein_g"`
	CarbohydrateG *float64 `json:"carbohydrate_g"`
	FatG          *float64 `json:"fat_g"`
	FiberG        *float64 `json:"fiber_g"`
	SugarsG       *float64 `json:"sugars_g"`
	SaturatedFatG *float64 `json:"saturated_fat_g"`
	SodiumMg      *float64 `json:"sodium_mg"`
	PotassiumMg   *float64 `json:"potassium_mg"`
	CalciumMg     *float64 `json:"calcium_mg"`
	MagnesiumMg   *float64 `json:"magnesium_mg"`
	IronMg        *float64 `json:"iron_mg"`
	ZincMg        *float64 `json:"zinc_mg"`
	VitaminDMcg   *float64 `json:"vitamin_d_mcg"`
	VitaminB12Mcg *float64 `json:"vitamin_b12_mcg"`
	FolateB9Mcg   *float64 `json:"folate_b9_mcg"`
	VitaminCMg    *float64 `json:"vitamin_c_mg"`
}

// MealItem is one item within a meal analysis result.
type MealItem struct {
	Name          string   `json:"name"`
	Quantity      *float64 `json:"quantity,omitempty"`
	Unit          string   `json:"unit,omitempty"`
	CaloriesKcal  *float64 `json:"calories_kcal,omitempty"`
	ProteinG      *float64 `json:"protein_g,omitempty"`
	CarbohydrateG *float64 `json:"carbohydrate_g,omitempty"`
	FatG          *float64 `json:"fat_g,omitempty"`
}

// MealTextAnalysis is the strongly-typed OpenAI response shape for meal analysis.
type MealTextAnalysis struct {
	CanonicalName   string      `json:"canonical_name"`
	IsMeal          *bool       `json:"is_meal,omitempty"`
	RejectionReason string      `json:"rejection_reason,omitempty"`
	ConfidenceScore *float64    `json:"confidence_score"`
	Assumptions     []string    `json:"assumptions"`
	Items           []MealItem  `json:"items"`
	Nutrition       NutritionV1 `json:"nutrition"`
}

type Client struct {
	apiKey     string
	model      string
	baseURL    string
	httpClient *http.Client
}

func NewClient(apiKey, model string) *Client {
	if model == "" {
		model = "gpt-4o-mini"
	}

	return &Client{
		apiKey:  apiKey,
		model:   model,
		baseURL: defaultBaseURL,
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (c *Client) AnalyzeMealText(ctx context.Context, rawText string) (MealTextAnalysis, error) {
	if strings.TrimSpace(rawText) == "" {
		return MealTextAnalysis{}, fmt.Errorf("rawText is required")
	}
	if strings.TrimSpace(c.apiKey) == "" {
		return MealTextAnalysis{}, fmt.Errorf("openai api key is required")
	}

	requestBody := map[string]any{
		"model": c.model,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt()},
			{"role": "user", "content": rawText},
		},
		"response_format": map[string]string{"type": "json_object"},
		"temperature":     0.2,
	}

	payload, err := json.Marshal(requestBody)
	if err != nil {
		return MealTextAnalysis{}, fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.baseURL+"/chat/completions", bytes.NewReader(payload))
	if err != nil {
		return MealTextAnalysis{}, fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Authorization", "Bearer "+c.apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return MealTextAnalysis{}, fmt.Errorf("send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return MealTextAnalysis{}, fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode >= http.StatusBadRequest {
		return MealTextAnalysis{}, fmt.Errorf("openai api error (%d): %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}

	var completion chatCompletionResponse
	if err := json.Unmarshal(body, &completion); err != nil {
		return MealTextAnalysis{}, fmt.Errorf("decode completion response: %w", err)
	}
	if len(completion.Choices) == 0 {
		return MealTextAnalysis{}, fmt.Errorf("openai api returned no choices")
	}

	content := strings.TrimSpace(completion.Choices[0].Message.Content)
	if content == "" {
		return MealTextAnalysis{}, fmt.Errorf("openai api returned empty content")
	}

	var analysis MealTextAnalysis
	if err := json.Unmarshal([]byte(content), &analysis); err != nil {
		return MealTextAnalysis{}, fmt.Errorf("decode analysis json: %w", err)
	}

	if err := validateAnalysis(analysis); err != nil {
		return MealTextAnalysis{}, err
	}

	return analysis, nil
}

func validateAnalysis(a MealTextAnalysis) error {
	isMeal := true
	if a.IsMeal != nil {
		isMeal = *a.IsMeal
	}
	if isMeal && strings.TrimSpace(a.CanonicalName) == "" {
		return fmt.Errorf("analysis validation: canonical_name is required")
	}

	if a.ConfidenceScore != nil && (*a.ConfidenceScore < 0 || *a.ConfidenceScore > 1) {
		return fmt.Errorf("analysis validation: confidence_score must be between 0 and 1")
	}

	return nil
}

func systemPrompt() string {
	return `You are a nutrition analysis assistant.
Return JSON only and do not include markdown, comments, or prose.
Output shape must be:
	{
	  "is_meal": boolean,
	  "rejection_reason": string,
	  "canonical_name": string,
	  "confidence_score": number|null,
  "assumptions": string[],
  "items": [
    {
      "name": string,
      "quantity": number|null,
      "unit": string,
      "calories_kcal": number|null,
      "protein_g": number|null,
      "carbohydrate_g": number|null,
      "fat_g": number|null
    }
  ],
  "nutrition": {
    "calories_kcal": number|null,
    "protein_g": number|null,
    "carbohydrate_g": number|null,
    "fat_g": number|null,
    "fiber_g": number|null,
    "sugars_g": number|null,
    "saturated_fat_g": number|null,
    "sodium_mg": number|null,
    "potassium_mg": number|null,
    "calcium_mg": number|null,
    "magnesium_mg": number|null,
    "iron_mg": number|null,
    "zinc_mg": number|null,
    "vitamin_d_mcg": number|null,
    "vitamin_b12_mcg": number|null,
    "folate_b9_mcg": number|null,
    "vitamin_c_mg": number|null
  }
}
If the text is nonsense or not describing food intake, set "is_meal": false and explain why in "rejection_reason".
When "is_meal" is false, set "canonical_name" to "not_a_meal".
If unsure, use null.
Normalize canonical_name to a common standardized meal name.
Include assumptions for inferred portions or ingredients.
Estimate common meals reasonably when details are incomplete.`
}

type chatCompletionResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
}
