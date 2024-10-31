package sentry

import (
	"context"
	"net/http"
	"time"

	"github.com/getsentry/sentry-go"
)

type key int

var (
	keySentry key = 1
)

func scopeFromContext(ctx context.Context) *sentry.Scope {
	value, _ := ctx.Value(keySentry).(*sentry.Scope)
	return value
}

// WithContext stores a new instance of Sentry in the context and returns the
// new generated context that you should use everywhere.
func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, keySentry, sentry.NewScope())
}

// WithRequest stores a new instance of Sentry in the context and returns the
// new generated request that you should use everywhere.
func WithRequest(r *http.Request) *http.Request {
	ctx := WithContext(r.Context())
	scope := scopeFromContext(ctx)
	scope.SetRequest(r)
	return r.WithContext(ctx)
}

// LogBreadcrumb logs a new breadcrumb in the Sentry instance of the context.
func LogBreadcrumb(ctx context.Context, level sentry.Level, category, message string) {
	scope := scopeFromContext(ctx)
	if scope == nil {
		return
	}
	scope.AddBreadcrumb(&sentry.Breadcrumb{
		Timestamp: time.Now(),
		Type:      "default",
		Message:   message,
		Category:  category,
		Level:     level,
	}, 500)
}

// Tag adds a new tag to the Sentry instance of the context.
func Tag(ctx context.Context, key string, value string) {
	scope := scopeFromContext(ctx)
	if scope == nil {
		return
	}
	scope.SetTag(key, value)
}
