package handlers

import (
	"net/http"
	"strings"
	"time"

	"diet/internal/models"
	"diet/internal/repositories"
	"github.com/gin-gonic/gin"
)

type SummaryHandler struct {
	repo *repositories.DailyNutritionSummaryRepository
}

func NewSummaryHandler(repo *repositories.DailyNutritionSummaryRepository) *SummaryHandler {
	return &SummaryHandler{repo: repo}
}

type dailySummaryResponse struct {
	UserID        string  `json:"user_id"`
	Date          string  `json:"date"`
	CaloriesKcal  float64 `json:"calories_kcal"`
	ProteinG      float64 `json:"protein_g"`
	CarbohydrateG float64 `json:"carbohydrate_g"`
	FatG          float64 `json:"fat_g"`
	FiberG        float64 `json:"fiber_g"`
	SugarsG       float64 `json:"sugars_g"`
	SaturatedFatG float64 `json:"saturated_fat_g"`
	SodiumMg      float64 `json:"sodium_mg"`
	PotassiumMg   float64 `json:"potassium_mg"`
	CalciumMg     float64 `json:"calcium_mg"`
	MagnesiumMg   float64 `json:"magnesium_mg"`
	IronMg        float64 `json:"iron_mg"`
	ZincMg        float64 `json:"zinc_mg"`
	VitaminDMcg   float64 `json:"vitamin_d_mcg"`
	VitaminB12Mcg float64 `json:"vitamin_b12_mcg"`
	FolateB9Mcg   float64 `json:"folate_b9_mcg"`
	VitaminCMg    float64 `json:"vitamin_c_mg"`
}

func (h *SummaryHandler) GetDailySummary(c *gin.Context) {
	userID := strings.TrimSpace(c.Query("user_id"))
	dateRaw := strings.TrimSpace(c.Query("date"))
	if userID == "" || dateRaw == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "user_id and date are required"})
		return
	}

	summaryDate, err := time.Parse("2006-01-02", dateRaw)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "date must be in YYYY-MM-DD format"})
		return
	}

	summary, err := h.repo.GetByUserIDAndDate(c.Request.Context(), userID, summaryDate)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch daily summary"})
		return
	}

	if summary == nil {
		c.JSON(http.StatusOK, newDailySummaryResponse(userID, dateRaw, models.NutritionFields{}))
		return
	}

	c.JSON(http.StatusOK, newDailySummaryResponse(userID, dateRaw, summary.NutritionFields))
}

func newDailySummaryResponse(userID, date string, n models.NutritionFields) dailySummaryResponse {
	return dailySummaryResponse{
		UserID:        userID,
		Date:          date,
		CaloriesKcal:  floatOrZero(n.CaloriesKcal),
		ProteinG:      floatOrZero(n.ProteinG),
		CarbohydrateG: floatOrZero(n.CarbohydrateG),
		FatG:          floatOrZero(n.FatG),
		FiberG:        floatOrZero(n.FiberG),
		SugarsG:       floatOrZero(n.SugarsG),
		SaturatedFatG: floatOrZero(n.SaturatedFatG),
		SodiumMg:      floatOrZero(n.SodiumMg),
		PotassiumMg:   floatOrZero(n.PotassiumMg),
		CalciumMg:     floatOrZero(n.CalciumMg),
		MagnesiumMg:   floatOrZero(n.MagnesiumMg),
		IronMg:        floatOrZero(n.IronMg),
		ZincMg:        floatOrZero(n.ZincMg),
		VitaminDMcg:   floatOrZero(n.VitaminDMcg),
		VitaminB12Mcg: floatOrZero(n.VitaminB12Mcg),
		FolateB9Mcg:   floatOrZero(n.FolateB9Mcg),
		VitaminCMg:    floatOrZero(n.VitaminCMg),
	}
}

func floatOrZero(v *float64) float64 {
	if v == nil {
		return 0
	}
	return *v
}
