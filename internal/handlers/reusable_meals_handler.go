package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"diet/internal/repositories"
	"github.com/gin-gonic/gin"
)

type ReusableMealsHandler struct {
	repo *repositories.MealsRepository
}

func NewReusableMealsHandler(repo *repositories.MealsRepository) *ReusableMealsHandler {
	return &ReusableMealsHandler{repo: repo}
}

func (h *ReusableMealsHandler) List(c *gin.Context) {
	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	items, err := h.repo.List(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to list reusable meals"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
