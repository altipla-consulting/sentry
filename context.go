package sentry

import (
	"context"
)

type key int

var keySentry key = 1

type Sentry struct {
	breadcrumbs []*Breadcrumb
}

func FromContext(ctx context.Context) *Sentry {
	return ctx.Value(keySentry).(*Sentry)
}

func WithContext(ctx context.Context) context.Context {
	return context.WithValue(ctx, keySentry, new(Sentry))
}
