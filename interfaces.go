package sentry

import (
	"time"

	"github.com/getsentry/raven-go"
)

type Extra struct {
	RequestID  string `json:"Request ID"`
	InstanceID string `json:"Instance ID"`
}

func (extra *Extra) Class() string {
	return "extra"
}

type Exception struct {
	Value      string            `json:"value"`
	Module     string            `json:"module"`
	Stacktrace *raven.Stacktrace `json:"stacktrace"`
	Type       string            `json:"type"`
}

func (e *Exception) Class() string {
	return "exception"
}

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
