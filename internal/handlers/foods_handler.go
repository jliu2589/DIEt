package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"diet/internal/repositories"
	"github.com/gin-gonic/gin"
)

type FoodsHandler struct {
	repo *repositories.CanonicalFoodsRepository
}

func NewFoodsHandler(repo *repositories.CanonicalFoodsRepository) *FoodsHandler {
	return &FoodsHandler{repo: repo}
}

func (h *FoodsHandler) Search(c *gin.Context) {
	q := strings.TrimSpace(c.Query("q"))
	if q == "" {
		c.JSON(http.StatusOK, gin.H{"items": []any{}})
		return
	}
	limit := 20
	if raw := strings.TrimSpace(c.Query("limit")); raw != "" {
		if parsed, err := strconv.Atoi(raw); err == nil && parsed > 0 && parsed <= 100 {
			limit = parsed
		}
	}

	items, err := h.repo.SearchByName(c.Request.Context(), q, limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to search foods"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"items": items})
}
