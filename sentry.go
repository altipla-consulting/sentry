package sentry

import (
	"context"
	"time"
)

type key int

var keySentry key = 1

type Sentry struct {
	breadcrumbs []*Breadcrumb
}

func FromContext(ctx context.Context) *Sentry {
	return ctx.Value(keySentry).(*Sentry)
}

func WithContext(ctx context.Context, sentry *Sentry) context.Context {
	return context.WithValue(ctx, keySentry, sentry)
}

func NewContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, keySentry, new(Sentry))
}

func LogBreadcrumb(ctx context.Context, level Level, category, message string) {
	info := FromContext(ctx)
	info.breadcrumbs = append(info.breadcrumbs, &Breadcrumb{
		Timestamp: time.Now(),
		Type:      "default",
		Message:   message,
		Category:  category,
		Level:     level,
	})
}
