package sentry

import (
	"time"

	"golang.org/x/net/context"
)

type Breadcrumbs struct {
	Values []*Breadcrumb `json:"values,omitempty"`
}

func (b *Breadcrumbs) Class() string {
	return "breadcrumbs"
}

type Breadcrumb struct {
	Timestamp time.Time         `json:"timestamp"`
	Type      string            `json:"type"`
	Message   string            `json:"message"`
	Data      map[string]string `json:"data,omitempty"`
	Category  string            `json:"category"`
	Level     Level             `json:"level"`
}

type Level string

const (
	LevelCritical = Level("critical")
	LevelWarning  = Level("warning")
	LevelError    = Level("error")
	LevelInfo     = Level("info")
	LevelDebug    = Level("debug")
)

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
