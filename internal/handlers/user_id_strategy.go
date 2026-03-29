package handlers

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const errUserIDRequired = "user_id is required"

// Temporary user strategy (pre-auth):
// - clients provide user_id explicitly (query/body)
// - handlers normalize and require it consistently
// This keeps the surface area small so it can be replaced later by auth middleware.
func normalizeRequiredUserID(raw string) (string, bool) {
	userID := strings.TrimSpace(raw)
	if userID == "" {
		return "", false
	}
	return userID, true
}

func requiredUserIDFromQuery(c *gin.Context) (string, bool) {
	userID, ok := normalizeRequiredUserID(c.Query("user_id"))
	if !ok {
		c.JSON(400, gin.H{"error": errUserIDRequired})
		return "", false
	}
	return userID, true
}
