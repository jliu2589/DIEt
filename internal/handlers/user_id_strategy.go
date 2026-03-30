package handlers

import (
	"strings"

	"diet/internal/auth"
	"github.com/gin-gonic/gin"
)

const errUserIDRequired = "authenticated subject or user_id is required"

// Temporary user strategy (pre-production):
// 1) Prefer authenticated subject from middleware context.
// 2) Optionally allow explicit user_id override for internal/admin tooling.
// 3) Fall back to explicit user_id when no auth subject exists (dev compatibility).
func resolveUserID(c *gin.Context, explicitRaw string, allowOverride bool) (string, bool) {
	explicit, explicitOK := normalizeRequiredUserID(explicitRaw)

	if subject, ok := auth.SubjectFromContext(c.Request.Context()); ok {
		if allowOverride && explicitOK {
			return explicit, true
		}
		return subject.UserID, true
	}

	if explicitOK {
		return explicit, true
	}

	c.JSON(400, gin.H{"error": errUserIDRequired})
	return "", false
}

func normalizeRequiredUserID(raw string) (string, bool) {
	userID := strings.TrimSpace(raw)
	if userID == "" {
		return "", false
	}
	return userID, true
}

func requiredUserIDFromQuery(c *gin.Context) (string, bool) {
	return resolveUserID(c, c.Query("user_id"), true)
}
