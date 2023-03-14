package sentry

import (
	"context"
	"time"

	"github.com/getsentry/sentry-go"
)

type key int

var keySentry key = 1

// Sentry accumulates info through out the whole request to send them in case
// an error is reported.
type Sentry struct {
	breadcrumbs []*sentry.Breadcrumb
}

// FromContext returns the Sentry instance stored in the context. If no instance
// was created it will return nil.
func FromContext(ctx context.Context) *Sentry {
	value, _ := ctx.Value(keySentry).(*Sentry)
	return value
}

// WithContext stores a new instance of Sentry in the context and returns the
// new generated context that you should use everywhere.
func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, keySentry, new(Sentry))
}

// LogBreadcrumb logs a new breadcrumb in the Sentry instance of the context.
func LogBreadcrumb(ctx context.Context, level sentry.Level, category, message string) {
	info := FromContext(ctx)
	if info == nil {
		return
	}

	info.breadcrumbs = append(info.breadcrumbs, &sentry.Breadcrumb{
		Timestamp: time.Now(),
		Type:      "default",
		Message:   message,
		Category:  category,
		Level:     level,
	})
}
