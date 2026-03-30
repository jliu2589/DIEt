package meal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"diet/internal/models"
	"diet/internal/repositories"
	canonicalfoodssvc "diet/internal/services/canonical_foods"
	inputclassifier "diet/internal/services/input_classifier"
	openaisvc "diet/internal/services/openai"
)

type Analyzer interface {
	AnalyzeMealText(ctx context.Context, rawText string) (openaisvc.MealTextAnalysis, error)
}

type InputClassifier interface {
	Classify(rawText string) string
}

type Service struct {
	mealEventsRepo     *repositories.MealEventsRepository
	mealAnalysisRepo   *repositories.MealAnalysisRepository
	mealMemoryRepo     *repositories.MealMemoryRepository
	dailySummaryRepo   *repositories.DailyNutritionSummaryRepository
	mealsRepo          *repositories.MealsRepository
	mealItemsRepo      *repositories.MealItemsRepository
	canonicalFoodsRepo *repositories.CanonicalFoodsRepository
	txManager          *repositories.TxManager
	mealTextAnalyzer   Analyzer
	classifier         InputClassifier
}

const (
	defaultRecentMealsLimit   = 20
	maxRecentMealsLimit       = 100
	defaultReusableMealSearch = 50
	reusableMatchThreshold    = 0.80
)

var ErrOpenAIFallbackDisabled = errors.New("openai fallback is disabled")

type ProcessTextMealInput struct {
	UserID          string
	RawText         string
	Source          string
	SourceMessageID *string
	EatenAt         time.Time
}

type ProcessTextMealResult struct {
	Intent            string                 `json:"intent"`
	Logged            bool                   `json:"logged"`
	Message           string                 `json:"message"`
	MealEventID       int64                  `json:"meal_event_id"`
	Source            string                 `json:"source"`
	ProcessedFrom     string                 `json:"processed_from"`
	LoggedAt          time.Time              `json:"logged_at"`
	EatenAt           time.Time              `json:"eaten_at"`
	TimeSource        string                 `json:"time_source"`
	CanonicalName     string                 `json:"canonical_name"`
	ConfidenceScore   *float64               `json:"confidence_score,omitempty"`
	Assumptions       []string               `json:"assumptions,omitempty"`
	Items             []models.MealItem      `json:"items,omitempty"`
	Nutrition         models.NutritionFields `json:"nutrition"`
	DailySummaryDate  string                 `json:"daily_summary_date"`
	MatchReason       *string                `json:"match_reason,omitempty"`
	TokenOverlapScore *float64               `json:"token_overlap_score,omitempty"`
}

type RecentMealResult struct {
	MealEventID   int64     `json:"meal_event_id"`
	CanonicalName string    `json:"canonical_name"`
	LoggedAt      time.Time `json:"logged_at"`
	EatenAt       time.Time `json:"eaten_at"`
	TimeSource    string    `json:"time_source"`
	CaloriesKcal  *float64  `json:"calories_kcal,omitempty"`
	ProteinG      *float64  `json:"protein_g,omitempty"`
	CarbohydrateG *float64  `json:"carbohydrate_g,omitempty"`
	FatG          *float64  `json:"fat_g,omitempty"`
	Source        string    `json:"source"`
}

type EditMealTimeResult struct {
	MealEventID   int64     `json:"meal_event_id"`
	CanonicalName string    `json:"canonical_name"`
	EatenAt       time.Time `json:"eaten_at"`
	TimeSource    string    `json:"time_source"`
}

func NewService(
	mealEventsRepo *repositories.MealEventsRepository,
	mealAnalysisRepo *repositories.MealAnalysisRepository,
	mealMemoryRepo *repositories.MealMemoryRepository,
	dailySummaryRepo *repositories.DailyNutritionSummaryRepository,
	mealsRepo *repositories.MealsRepository,
	mealItemsRepo *repositories.MealItemsRepository,
	canonicalFoodsRepo *repositories.CanonicalFoodsRepository,
	txManager *repositories.TxManager,
	mealTextAnalyzer Analyzer,
	classifier InputClassifier,
) *Service {
	return &Service{
		mealEventsRepo:     mealEventsRepo,
		mealAnalysisRepo:   mealAnalysisRepo,
		mealMemoryRepo:     mealMemoryRepo,
		dailySummaryRepo:   dailySummaryRepo,
		mealsRepo:          mealsRepo,
		mealItemsRepo:      mealItemsRepo,
		canonicalFoodsRepo: canonicalFoodsRepo,
		txManager:          txManager,
		mealTextAnalyzer:   mealTextAnalyzer,
		classifier:         classifier,
	}
}

func (s *Service) ProcessTextMeal(ctx context.Context, input ProcessTextMealInput) (*ProcessTextMealResult, error) {
	if strings.TrimSpace(input.UserID) == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if strings.TrimSpace(input.RawText) == "" {
		return nil, fmt.Errorf("raw_text is required")
	}
	intent := inputclassifier.IntentMealLog
	if s.classifier != nil {
		intent = s.classifier.Classify(input.RawText)
	}
	if intent != inputclassifier.IntentMealLog {
		return &ProcessTextMealResult{
			Intent:  intent,
			Logged:  false,
			Message: nonMealIntentMessage(intent),
		}, nil
	}

	eatenAt := input.EatenAt
	timeSource := "explicit"
	if eatenAt.IsZero() {
		if parsedEatenAt, ok := parseEatenAtFromText(input.RawText, time.Now().UTC()); ok {
			eatenAt = parsedEatenAt
			timeSource = "explicit"
		} else {
			eatenAt = time.Now().UTC()
			timeSource = "default_now"
		}
	}
	eatenAt = eatenAt.UTC()
	if strings.TrimSpace(input.Source) == "" {
		input.Source = "text"
	}

	rawText := input.RawText
	fingerprint := FingerprintFromText(input.RawText)

	cached, err := s.mealMemoryRepo.FindByFingerprintHash(ctx, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("lookup meal memory: %w", err)
	}

	if s.txManager == nil {
		event, err := s.mealEventsRepo.Insert(ctx, models.MealEvent{
			UserID:           input.UserID,
			Source:           input.Source,
			SourceMessageID:  input.SourceMessageID,
			EventType:        "text",
			RawText:          &rawText,
			LoggedAt:         time.Now().UTC(),
			EatenAt:          eatenAt,
			TimeSource:       timeSource,
			ProcessingStatus: "pending",
		})
		if err != nil {
			return nil, fmt.Errorf("create meal_event: %w", err)
		}
		if cached != nil {
			return s.processFromCache(ctx, event, cached)
		}
		if reusableResult, err := s.processFromReusableDatabase(ctx, event, fingerprint, input.RawText); err != nil {
			return nil, err
		} else if reusableResult != nil {
			return reusableResult, nil
		}
		return s.processWithOpenAI(ctx, event, fingerprint, input.RawText)
	}

	var out *ProcessTextMealResult
	if err := s.txManager.WithinTx(ctx, func(txRepos repositories.Repositories) error {
		txSvc := s.withRepos(txRepos)
		event, err := txSvc.mealEventsRepo.Insert(ctx, models.MealEvent{
			UserID:           input.UserID,
			Source:           input.Source,
			SourceMessageID:  input.SourceMessageID,
			EventType:        "text",
			RawText:          &rawText,
			LoggedAt:         time.Now().UTC(),
			EatenAt:          eatenAt,
			TimeSource:       timeSource,
			ProcessingStatus: "pending",
		})
		if err != nil {
			return fmt.Errorf("create meal_event: %w", err)
		}

		if cached != nil {
			out, err = txSvc.processFromCache(ctx, event, cached)
			return err
		}

		if reusableResult, err := txSvc.processFromReusableDatabase(ctx, event, fingerprint, input.RawText); err != nil {
			return err
		} else if reusableResult != nil {
			out = reusableResult
			return nil
		}

		out, err = txSvc.processWithOpenAI(ctx, event, fingerprint, input.RawText)
		return err
	}); err != nil {
		return nil, err
	}

	return out, nil
}

func (s *Service) GetRecentMeals(ctx context.Context, userID string, limit int) ([]RecentMealResult, error) {
	if strings.TrimSpace(userID) == "" {
		return nil, fmt.Errorf("user_id is required")
	}

	switch {
	case limit <= 0:
		limit = defaultRecentMealsLimit
	case limit > maxRecentMealsLimit:
		limit = maxRecentMealsLimit
	}

	rows, err := s.mealEventsRepo.ListRecentByUserID(ctx, userID, limit)
	if err != nil {
		return nil, fmt.Errorf("get recent meals: %w", err)
	}

	out := make([]RecentMealResult, 0, len(rows))
	for _, row := range rows {
		out = append(out, RecentMealResult{
			MealEventID:   row.MealEventID,
			CanonicalName: row.CanonicalName,
			LoggedAt:      row.LoggedAt,
			EatenAt:       row.EatenAt,
			TimeSource:    row.TimeSource,
			CaloriesKcal:  row.CaloriesKcal,
			ProteinG:      row.ProteinG,
			CarbohydrateG: row.CarbohydrateG,
			FatG:          row.FatG,
			Source:        row.Source,
		})
	}

	return out, nil
}

func (s *Service) EditMealTime(ctx context.Context, mealEventID int64, userID string, eatenAt time.Time) (*EditMealTimeResult, error) {
	if mealEventID <= 0 {
		return nil, fmt.Errorf("meal_event_id must be greater than 0")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return nil, fmt.Errorf("user_id is required")
	}
	if eatenAt.IsZero() {
		return nil, fmt.Errorf("eaten_at is required")
	}

	updated, err := s.mealEventsRepo.UpdateEatenAtByIDAndUserID(ctx, mealEventID, userID, eatenAt.UTC())
	if err != nil {
		return nil, fmt.Errorf("edit meal time: %w", err)
	}
	if updated == nil {
		return nil, nil
	}

	oldDate := dateOnly(updated.PreviousEatenAt)
	newDate := dateOnly(updated.UpdatedEatenAt)
	if err := s.dailySummaryRepo.ReconcileForUserDate(ctx, userID, oldDate); err != nil {
		return nil, fmt.Errorf("reconcile old summary date: %w", err)
	}
	if !oldDate.Equal(newDate) {
		if err := s.dailySummaryRepo.ReconcileForUserDate(ctx, userID, newDate); err != nil {
			return nil, fmt.Errorf("reconcile new summary date: %w", err)
		}
	}

	return &EditMealTimeResult{
		MealEventID:   updated.MealEventID,
		CanonicalName: updated.CanonicalName,
		EatenAt:       updated.UpdatedEatenAt,
		TimeSource:    updated.TimeSource,
	}, nil
}

// ReconcileDailySummaryForDate recomputes a user's daily summary from meal history.
// This is intended for edit/delete maintenance flows.
func (s *Service) ReconcileDailySummaryForDate(ctx context.Context, userID string, date time.Time) error {
	if strings.TrimSpace(userID) == "" {
		return fmt.Errorf("user_id is required")
	}
	return s.dailySummaryRepo.ReconcileForUserDate(ctx, userID, date)
}

func (s *Service) processFromCache(ctx context.Context, event *models.MealEvent, cached *models.MealMemory) (*ProcessTextMealResult, error) {
	analysis := models.MealAnalysis{
		MealEventID:     event.ID,
		UserID:          event.UserID,
		CanonicalName:   cached.CanonicalName,
		ConfidenceScore: cached.ConfidenceScore,
		AssumptionsJSON: cached.AssumptionsJSON,
		ItemsJSON:       cached.ItemsJSON,
		RawAnalysisJSON: cached.RawAnalysisJSON,
		NutritionFields: cached.NutritionFields,
	}

	if err := s.mealAnalysisRepo.Insert(ctx, analysis); err != nil {
		return nil, fmt.Errorf("insert meal_analysis from cache: %w", err)
	}
	if err := s.mealEventsRepo.UpdateProcessingStatus(ctx, event.ID, "processed"); err != nil {
		return nil, fmt.Errorf("mark meal_event processed: %w", err)
	}

	summaryDate := dateOnly(event.EatenAt)
	if _, err := s.updateDailySummary(ctx, event.UserID, summaryDate, cached.NutritionFields); err != nil {
		return nil, err
	}

	result := &ProcessTextMealResult{
		Intent:           inputclassifier.IntentMealLog,
		Logged:           true,
		Message:          "Logged your meal.",
		MealEventID:      event.ID,
		Source:           event.Source,
		ProcessedFrom:    "cache",
		LoggedAt:         event.LoggedAt,
		EatenAt:          event.EatenAt,
		TimeSource:       event.TimeSource,
		CanonicalName:    cached.CanonicalName,
		ConfidenceScore:  cached.ConfidenceScore,
		Nutrition:        cached.NutritionFields,
		DailySummaryDate: summaryDate.Format("2006-01-02"),
	}

	_ = json.Unmarshal(cached.ItemsJSON, &result.Items)
	_ = json.Unmarshal(cached.AssumptionsJSON, &result.Assumptions)

	return result, nil
}

func (s *Service) processWithOpenAI(ctx context.Context, event *models.MealEvent, fingerprint, rawText string) (*ProcessTextMealResult, error) {
	if s.mealTextAnalyzer == nil {
		return nil, ErrOpenAIFallbackDisabled
	}

	openAIResult, err := s.mealTextAnalyzer.AnalyzeMealText(ctx, rawText)
	if err != nil {
		return nil, fmt.Errorf("analyze meal text: %w", err)
	}

	assumptionsJSON, err := json.Marshal(openAIResult.Assumptions)
	if err != nil {
		return nil, fmt.Errorf("marshal assumptions: %w", err)
	}

	items := toModelItems(openAIResult.Items)
	itemsJSON, err := json.Marshal(items)
	if err != nil {
		return nil, fmt.Errorf("marshal items: %w", err)
	}

	rawAnalysisJSON, err := json.Marshal(openAIResult)
	if err != nil {
		return nil, fmt.Errorf("marshal raw analysis: %w", err)
	}

	nutrition := toNutritionFields(openAIResult.Nutrition)
	analysis := models.MealAnalysis{
		MealEventID:     event.ID,
		UserID:          event.UserID,
		CanonicalName:   openAIResult.CanonicalName,
		ConfidenceScore: openAIResult.ConfidenceScore,
		AssumptionsJSON: assumptionsJSON,
		ItemsJSON:       itemsJSON,
		RawAnalysisJSON: rawAnalysisJSON,
		NutritionFields: nutrition,
	}

	if err := s.mealAnalysisRepo.Insert(ctx, analysis); err != nil {
		return nil, fmt.Errorf("insert meal_analysis from openai: %w", err)
	}

	if _, err := s.mealMemoryRepo.Upsert(ctx, models.MealMemory{
		FingerprintHash: fingerprint,
		CanonicalName:   openAIResult.CanonicalName,
		ConfidenceScore: openAIResult.ConfidenceScore,
		AssumptionsJSON: assumptionsJSON,
		ItemsJSON:       itemsJSON,
		RawAnalysisJSON: rawAnalysisJSON,
		NutritionFields: nutrition,
	}); err != nil {
		return nil, fmt.Errorf("upsert meal_memory: %w", err)
	}

	if err := s.maybeSaveReusableMealTemplate(ctx, fingerprint, openAIResult.CanonicalName, openAIResult.ConfidenceScore, items); err != nil {
		return nil, fmt.Errorf("save reusable meal template: %w", err)
	}

	if err := s.mealEventsRepo.UpdateProcessingStatus(ctx, event.ID, "processed"); err != nil {
		return nil, fmt.Errorf("mark meal_event processed: %w", err)
	}

	summaryDate := dateOnly(event.EatenAt)
	if _, err := s.updateDailySummary(ctx, event.UserID, summaryDate, nutrition); err != nil {
		return nil, err
	}

	return &ProcessTextMealResult{
		Intent:           inputclassifier.IntentMealLog,
		Logged:           true,
		Message:          "Logged your meal.",
		MealEventID:      event.ID,
		Source:           event.Source,
		ProcessedFrom:    "openai",
		LoggedAt:         event.LoggedAt,
		EatenAt:          event.EatenAt,
		TimeSource:       event.TimeSource,
		CanonicalName:    openAIResult.CanonicalName,
		ConfidenceScore:  openAIResult.ConfidenceScore,
		Assumptions:      openAIResult.Assumptions,
		Items:            items,
		Nutrition:        nutrition,
		DailySummaryDate: summaryDate.Format("2006-01-02"),
	}, nil
}

func (s *Service) processFromReusableDatabase(ctx context.Context, event *models.MealEvent, fingerprint string, rawText string) (*ProcessTextMealResult, error) {
	if s.mealsRepo == nil || s.mealItemsRepo == nil || s.canonicalFoodsRepo == nil {
		return nil, nil
	}

	matchReason := "exact_fingerprint_match"
	tokenOverlap := 1.0

	reusableMeal, err := s.mealsRepo.GetByFingerprintHash(ctx, fingerprint)
	if err != nil {
		return nil, fmt.Errorf("lookup reusable meal by fingerprint: %w", err)
	}
	if reusableMeal == nil {
		matched, reason, overlap, err := s.findReusableMealMatch(ctx, rawText)
		if err != nil {
			return nil, err
		}
		if matched == nil || overlap < reusableMatchThreshold {
			return nil, nil
		}
		matchReason = reason
		tokenOverlap = overlap
		reusableMeal, err = s.mealsRepo.GetByID(ctx, matched.ID)
		if err != nil {
			return nil, fmt.Errorf("get matched reusable meal by id: %w", err)
		}
		if reusableMeal == nil {
			return nil, nil
		}
	}

	storedItems, err := s.mealItemsRepo.ListByMealID(ctx, reusableMeal.ID)
	if err != nil {
		return nil, fmt.Errorf("list reusable meal items: %w", err)
	}

	var total models.NutritionFields
	analysisItems := make([]models.MealItem, 0, len(storedItems))
	for _, item := range storedItems {
		food, err := s.canonicalFoodsRepo.GetByID(ctx, item.FoodID)
		if err != nil {
			return nil, fmt.Errorf("get canonical food for reusable meal item: %w", err)
		}
		if food == nil {
			continue
		}

		scaled, err := canonicalfoodssvc.ScaleNutrition(*food, item.Quantity, item.Unit)
		if err != nil {
			return nil, fmt.Errorf("scale nutrition for reusable meal item: %w", err)
		}

		total = mergeNutrition(total, scaled.Nutrition)
		analysisItems = append(analysisItems, models.MealItem{
			Name:          food.CanonicalName,
			Quantity:      &item.Quantity,
			Unit:          item.Unit,
			CaloriesKcal:  scaled.Nutrition.CaloriesKcal,
			ProteinG:      scaled.Nutrition.ProteinG,
			CarbohydrateG: scaled.Nutrition.CarbohydrateG,
			FatG:          scaled.Nutrition.FatG,
		})
	}

	itemsJSON, _ := json.Marshal(analysisItems)
	analysis := models.MealAnalysis{
		MealEventID:     event.ID,
		UserID:          event.UserID,
		CanonicalName:   reusableMeal.CanonicalName,
		ConfidenceScore: reusableMeal.ConfidenceScore,
		ItemsJSON:       itemsJSON,
		NutritionFields: total,
	}

	if err := s.mealAnalysisRepo.Insert(ctx, analysis); err != nil {
		return nil, fmt.Errorf("insert meal_analysis from reusable meal: %w", err)
	}

	if _, err := s.mealMemoryRepo.Upsert(ctx, models.MealMemory{
		FingerprintHash: fingerprint,
		CanonicalName:   reusableMeal.CanonicalName,
		ConfidenceScore: reusableMeal.ConfidenceScore,
		ItemsJSON:       itemsJSON,
		NutritionFields: total,
	}); err != nil {
		return nil, fmt.Errorf("upsert meal_memory from reusable meal: %w", err)
	}

	if err := s.mealEventsRepo.UpdateProcessingStatus(ctx, event.ID, "processed"); err != nil {
		return nil, fmt.Errorf("mark meal_event processed: %w", err)
	}

	summaryDate := dateOnly(event.EatenAt)
	if _, err := s.updateDailySummary(ctx, event.UserID, summaryDate, total); err != nil {
		return nil, err
	}

	return &ProcessTextMealResult{
		Intent:            inputclassifier.IntentMealLog,
		Logged:            true,
		Message:           "Logged your meal.",
		MealEventID:       event.ID,
		Source:            event.Source,
		ProcessedFrom:     "reusable_db",
		LoggedAt:          event.LoggedAt,
		EatenAt:           event.EatenAt,
		TimeSource:        event.TimeSource,
		CanonicalName:     reusableMeal.CanonicalName,
		ConfidenceScore:   reusableMeal.ConfidenceScore,
		Items:             analysisItems,
		Nutrition:         total,
		DailySummaryDate:  summaryDate.Format("2006-01-02"),
		MatchReason:       &matchReason,
		TokenOverlapScore: &tokenOverlap,
	}, nil
}

func (s *Service) findReusableMealMatch(ctx context.Context, rawText string) (*repositories.MealCandidate, string, float64, error) {
	if s.mealsRepo == nil {
		return nil, "no_repository", 0, nil
	}

	candidates, err := s.mealsRepo.ListCandidates(ctx, defaultReusableMealSearch)
	if err != nil {
		return nil, "candidate_query_error", 0, fmt.Errorf("list reusable meal candidates: %w", err)
	}
	if len(candidates) == 0 {
		return nil, "no_candidates", 0, nil
	}

	normalizedText := NormalizeText(rawText)
	inputTokens := CanonicalTokens(rawText)
	if normalizedText == "" || len(inputTokens) == 0 {
		return nil, "empty_input", 0, nil
	}

	bestScore := 0.0
	bestReason := "no_match"
	var best *repositories.MealCandidate
	for i := range candidates {
		name := NormalizeText(candidates[i].CanonicalName)
		if name == "" {
			continue
		}
		if normalizedText == name {
			score := 1.0
			best = &candidates[i]
			bestScore = score
			bestReason = "canonical_name_match"
			break
		}

		candidateTokens := CanonicalTokens(candidates[i].CanonicalName)
		overlap := TokenOverlapScore(inputTokens, candidateTokens)
		if overlap > bestScore {
			best = &candidates[i]
			bestScore = overlap
			bestReason = "token_overlap_score"
		}
	}

	if best == nil {
		return nil, bestReason, 0, nil
	}
	return best, bestReason, bestScore, nil
}

func (s *Service) maybeSaveReusableMealTemplate(ctx context.Context, fingerprint string, canonicalName string, confidenceScore *float64, items []models.MealItem) error {
	if s.mealsRepo == nil || s.mealItemsRepo == nil || s.canonicalFoodsRepo == nil {
		return nil
	}

	if existing, err := s.mealsRepo.GetByFingerprintHash(ctx, fingerprint); err != nil {
		return fmt.Errorf("lookup reusable meal by fingerprint before save: %w", err)
	} else if existing != nil {
		return nil
	}

	resolvedItems, err := s.resolveStoredMealItems(ctx, items)
	if err != nil {
		return fmt.Errorf("resolve stored meal items: %w", err)
	}
	if len(resolvedItems) == 0 {
		return nil
	}

	structureHash := FingerprintFromCanonicalStructure(canonicalName, items)
	if structureHash == "" {
		return nil
	}

	if existing, err := s.mealsRepo.GetByStructureSignature(ctx, structureHash); err != nil {
		return fmt.Errorf("lookup reusable meal by structure signature: %w", err)
	} else if existing != nil {
		return nil
	}

	sourceType := "canonical_structure"
	mealRow, err := s.mealsRepo.Create(ctx, models.Meal{
		CanonicalName:      canonicalName,
		FingerprintHash:    &fingerprint,
		StructureSignature: &structureHash,
		SourceType:         &sourceType,
		ConfidenceScore:    confidenceScore,
	})
	if err != nil {
		return fmt.Errorf("create reusable meal template: %w", err)
	}

	for _, it := range resolvedItems {
		if _, err := s.mealItemsRepo.Create(ctx, models.StoredMealItem{
			MealID:   mealRow.ID,
			FoodID:   it.FoodID,
			Quantity: it.Quantity,
			Unit:     it.Unit,
		}); err != nil {
			return fmt.Errorf("create reusable meal item: %w", err)
		}
	}

	return nil
}

func (s *Service) resolveStoredMealItems(ctx context.Context, items []models.MealItem) ([]models.StoredMealItem, error) {
	if len(items) == 0 {
		return nil, nil
	}
	out := make([]models.StoredMealItem, 0, len(items))
	for _, it := range items {
		if strings.TrimSpace(it.Name) == "" || it.Quantity == nil || strings.TrimSpace(it.Unit) == "" {
			continue
		}
		food, err := s.canonicalFoodsRepo.GetByCanonicalName(ctx, it.Name)
		if err != nil {
			return nil, fmt.Errorf("resolve food by canonical name: %w", err)
		}
		if food == nil {
			continue
		}
		out = append(out, models.StoredMealItem{
			FoodID:   food.ID,
			Quantity: *it.Quantity,
			Unit:     strings.TrimSpace(it.Unit),
		})
	}
	return out, nil
}

func (s *Service) withRepos(repos repositories.Repositories) *Service {
	return &Service{
		mealEventsRepo:     repos.MealEvents,
		mealAnalysisRepo:   repos.MealAnalysis,
		mealMemoryRepo:     repos.MealMemory,
		dailySummaryRepo:   repos.DailyNutritionSummary,
		mealsRepo:          repos.Meals,
		mealItemsRepo:      repos.MealItems,
		canonicalFoodsRepo: repos.CanonicalFoods,
		txManager:          nil,
		mealTextAnalyzer:   s.mealTextAnalyzer,
		classifier:         s.classifier,
	}
}

func nonMealIntentMessage(intent string) string {
	switch intent {
	case inputclassifier.IntentWeightLog:
		return "I detected a weight log, so I didn’t save this as a meal."
	case inputclassifier.IntentRecommendationRequest:
		return "I detected a recommendation request, so I didn’t save this as a meal."
	case inputclassifier.IntentGeneralChat:
		return "I detected general chat, so I didn’t save this as a meal."
	default:
		return "I couldn’t confidently detect a meal log, so nothing was saved."
	}
}

func (s *Service) updateDailySummary(ctx context.Context, userID string, summaryDate time.Time, delta models.NutritionFields) (*models.DailyNutritionSummary, error) {
	existing, err := s.dailySummaryRepo.GetByUserIDAndDate(ctx, userID, summaryDate)
	if err != nil {
		return nil, fmt.Errorf("get daily summary: %w", err)
	}

	totals := delta
	if existing != nil {
		totals = mergeNutrition(existing.NutritionFields, delta)
	}

	summary, err := s.dailySummaryRepo.UpsertTotals(ctx, models.DailyNutritionSummary{
		UserID:          userID,
		SummaryDate:     summaryDate,
		NutritionFields: totals,
	})
	if err != nil {
		return nil, fmt.Errorf("upsert daily summary: %w", err)
	}

	return summary, nil
}

func toModelItems(items []openaisvc.MealItem) []models.MealItem {
	if len(items) == 0 {
		return nil
	}
	out := make([]models.MealItem, 0, len(items))
	for _, item := range items {
		out = append(out, models.MealItem{
			Name:          item.Name,
			Quantity:      item.Quantity,
			Unit:          item.Unit,
			CaloriesKcal:  item.CaloriesKcal,
			ProteinG:      item.ProteinG,
			CarbohydrateG: item.CarbohydrateG,
			FatG:          item.FatG,
		})
	}
	return out
}

func toNutritionFields(n openaisvc.NutritionV1) models.NutritionFields {
	return models.NutritionFields{
		CaloriesKcal:  n.CaloriesKcal,
		ProteinG:      n.ProteinG,
		CarbohydrateG: n.CarbohydrateG,
		FatG:          n.FatG,
		FiberG:        n.FiberG,
		SugarsG:       n.SugarsG,
		SaturatedFatG: n.SaturatedFatG,
		SodiumMg:      n.SodiumMg,
		PotassiumMg:   n.PotassiumMg,
		CalciumMg:     n.CalciumMg,
		MagnesiumMg:   n.MagnesiumMg,
		IronMg:        n.IronMg,
		ZincMg:        n.ZincMg,
		VitaminDMcg:   n.VitaminDMcg,
		VitaminB12Mcg: n.VitaminB12Mcg,
		FolateB9Mcg:   n.FolateB9Mcg,
		VitaminCMg:    n.VitaminCMg,
	}
}

func mergeNutrition(base, delta models.NutritionFields) models.NutritionFields {
	return models.NutritionFields{
		CaloriesKcal:  addPtr(base.CaloriesKcal, delta.CaloriesKcal),
		ProteinG:      addPtr(base.ProteinG, delta.ProteinG),
		CarbohydrateG: addPtr(base.CarbohydrateG, delta.CarbohydrateG),
		FatG:          addPtr(base.FatG, delta.FatG),
		FiberG:        addPtr(base.FiberG, delta.FiberG),
		SugarsG:       addPtr(base.SugarsG, delta.SugarsG),
		SaturatedFatG: addPtr(base.SaturatedFatG, delta.SaturatedFatG),
		SodiumMg:      addPtr(base.SodiumMg, delta.SodiumMg),
		PotassiumMg:   addPtr(base.PotassiumMg, delta.PotassiumMg),
		CalciumMg:     addPtr(base.CalciumMg, delta.CalciumMg),
		MagnesiumMg:   addPtr(base.MagnesiumMg, delta.MagnesiumMg),
		IronMg:        addPtr(base.IronMg, delta.IronMg),
		ZincMg:        addPtr(base.ZincMg, delta.ZincMg),
		VitaminDMcg:   addPtr(base.VitaminDMcg, delta.VitaminDMcg),
		VitaminB12Mcg: addPtr(base.VitaminB12Mcg, delta.VitaminB12Mcg),
		FolateB9Mcg:   addPtr(base.FolateB9Mcg, delta.FolateB9Mcg),
		VitaminCMg:    addPtr(base.VitaminCMg, delta.VitaminCMg),
	}
}

func addPtr(a, b *float64) *float64 {
	if a == nil && b == nil {
		return nil
	}
	var av, bv float64
	if a != nil {
		av = *a
	}
	if b != nil {
		bv = *b
	}
	total := av + bv
	return &total
}

func dateOnly(t time.Time) time.Time {
	ut := t.UTC()
	return time.Date(ut.Year(), ut.Month(), ut.Day(), 0, 0, 0, 0, time.UTC)
}
