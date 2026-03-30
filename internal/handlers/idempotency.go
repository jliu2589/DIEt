package handlers

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"diet/internal/repositories"
	"github.com/gin-gonic/gin"
)

func requestHash(v any) (string, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:]), nil
}

func resolveIdempotencyKey(c *gin.Context, bodyKey string) string {
	headerKey := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
	if headerKey != "" {
		return headerKey
	}
	return strings.TrimSpace(bodyKey)
}

func handleExistingIdempotencyRecord(c *gin.Context, rec *repositories.IdempotencyRecord, reqHash string) bool {
	if rec == nil {
		return false
	}
	if rec.RequestHash != reqHash {
		c.JSON(http.StatusConflict, gin.H{"error": "idempotency key already used with different request payload"})
		return true
	}
	if rec.Status == repositories.IdempotencyStatusSucceeded {
		status := http.StatusOK
		if rec.HTTPStatus != nil {
			status = *rec.HTTPStatus
		}
		var payload any
		if len(rec.ResponseJSON) > 0 {
			_ = json.Unmarshal(rec.ResponseJSON, &payload)
		}
		if payload == nil {
			payload = gin.H{"message": "already processed"}
		}
		c.JSON(status, payload)
		return true
	}

	c.JSON(http.StatusConflict, gin.H{"error": "request with same idempotency key is already processing"})
	return true
}

func saveIdempotencySuccess(c *gin.Context, repo *repositories.IdempotencyKeysRepository, recID int64, httpStatus int, payload any) {
	if repo == nil || recID == 0 {
		return
	}
	data, err := json.Marshal(payload)
	if err != nil {
		return
	}
	_ = repo.MarkSucceeded(c.Request.Context(), recID, httpStatus, data)
}

func cleanupIdempotencyOnError(c *gin.Context, repo *repositories.IdempotencyKeysRepository, recID int64) {
	if repo == nil || recID == 0 {
		return
	}
	_ = repo.Delete(c.Request.Context(), recID)
}

func beginIdempotency(c *gin.Context, repo *repositories.IdempotencyKeysRepository, userID, endpoint, key string, payload any) (int64, bool) {
	if repo == nil || strings.TrimSpace(key) == "" {
		return 0, false
	}
	hash, err := requestHash(payload)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("invalid idempotency payload: %v", err)})
		return 0, true
	}
	rec, created, err := repo.BeginOrGet(c.Request.Context(), userID, endpoint, key, hash)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to process idempotency key"})
		return 0, true
	}
	if !created {
		if handleExistingIdempotencyRecord(c, rec, hash) {
			return 0, true
		}
	}
	return rec.ID, false
}
