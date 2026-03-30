package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// DemoSubjectMiddleware is a temporary pre-production subject extraction layer.
//
// It reads X-Demo-User-ID and attaches a stable subject to request context.
// Replace this middleware with Supabase/JWT verification in production.
func DemoSubjectMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		demoUserID := strings.TrimSpace(c.GetHeader(DemoUserHeader))
		if demoUserID != "" {
			subject := Subject{UserID: demoUserID, Source: "demo_header"}
			c.Request = c.Request.WithContext(WithSubject(c.Request.Context(), subject))
		}
		c.Next()
	}
}
