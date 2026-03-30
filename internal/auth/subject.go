package auth

import (
	"context"
	"strings"
)

// DemoUserHeader is a temporary pre-production header for subject identity in dev.
const DemoUserHeader = "X-Demo-User-ID"

type Subject struct {
	UserID string
	Source string
}

type contextKey string

const subjectContextKey contextKey = "auth.subject"

func WithSubject(ctx context.Context, subject Subject) context.Context {
	return context.WithValue(ctx, subjectContextKey, subject)
}

func SubjectFromContext(ctx context.Context) (*Subject, bool) {
	raw := ctx.Value(subjectContextKey)
	subject, ok := raw.(Subject)
	if !ok {
		return nil, false
	}
	subject.UserID = strings.TrimSpace(subject.UserID)
	if subject.UserID == "" {
		return nil, false
	}
	return &subject, true
}
