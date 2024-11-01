package sentry

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"reflect"
	"runtime/debug"

	"github.com/getsentry/sentry-go"
)

const (
	LevelFatal   = sentry.LevelFatal
	LevelWarning = sentry.LevelWarning
	LevelError   = sentry.LevelError
	LevelInfo    = sentry.LevelInfo
	LevelDebug   = sentry.LevelDebug
)

// Client wraps a Sentry connection.
type Client struct {
	client *sentry.Client
}

// NewClient opens a new connection to the Sentry report API. If dsn is empty
// it will return nil. A nil client can be used safely but it won't report anything.
func NewClient(dsn string) *Client {
	if dsn == "" {
		return nil
	}

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:            dsn,
		Release:        os.Getenv("VERSION"),
		SendDefaultPII: true,
	})
	if err != nil {
		panic(err)
	}

	return &Client{
		client: client,
	}
}

// Report reports an error to Sentry.
func (client *Client) Report(ctx context.Context, appErr error) {
	if client == nil {
		return
	}
	client.sendReport(ctx, appErr)
}

// Deprecated: Use Report() after a previous WithRequest() call.
func (client *Client) ReportRequest(r *http.Request, appErr error) {
	if client == nil {
		return
	}
	client.sendReport(r.Context(), appErr)
}

// ReportPanics detects panics in the rest of the body of the function and
// reports it if one occurs.
func (client *Client) ReportPanics(ctx context.Context) {
	if client == nil {
		return
	}
	client.ReportPanic(ctx, recover()) // revive:disable-line:defer
}

// ReportPanic sends a panic correctly formated to the server if the argument
// is not nil.
func (client *Client) ReportPanic(ctx context.Context, panicErr interface{}) {
	if client == nil || panicErr == nil {
		return
	}
	client.sendReportPanic(ctx, fmt.Errorf("panic: %v", panicErr), string(debug.Stack()))
}

// Deprecated: Use ReportPanics() after a previous WithRequest() call.
func (client *Client) ReportPanicsRequest(r *http.Request) {
	if client == nil {
		return
	}
	if rec := recover(); rec != nil { // revive:disable-line:defer
		client.sendReportPanic(r.Context(), fmt.Errorf("panic: %v", rec), string(debug.Stack()))
	}
}

func (client *Client) sendReport(ctx context.Context, appErr error) {
	go func() {
		event := sentry.NewEvent()
		event.Level = sentry.LevelError
		event.Message = appErr.Error()

		cause := appErr
		for {
			child := errors.Unwrap(cause)
			if child == nil {
				break
			}
			cause = child
		}
		event.Exception = []sentry.Exception{
			{
				Type:       reflect.TypeOf(cause).String(),
				Value:      appErr.Error(),
				Stacktrace: sentry.ExtractStacktrace(appErr),
			},
		}

		scope := scopeFromContext(ctx)
		if scope == nil {
			scope = sentry.NewScope()
		}

		eventID := sentry.NewHub(client.client, scope).CaptureEvent(event)
		slog.Info("Error sent to Sentry", slog.String("event-id", string(*eventID)), slog.String("error", appErr.Error()))
	}()
}

func (client *Client) sendReportPanic(ctx context.Context, appErr error, message string) {
	go func() {
		event := sentry.NewEvent()
		event.Level = sentry.LevelFatal
		event.Message = message
		event.Exception = []sentry.Exception{
			{
				Type:       appErr.Error(),
				Value:      appErr.Error(),
				Stacktrace: sentry.ExtractStacktrace(appErr),
			},
		}

		scope := scopeFromContext(ctx)
		if scope == nil {
			scope = sentry.NewScope()
		}

		eventID := sentry.NewHub(client.client, scope).CaptureEvent(event)
		slog.Info("Error sent to Sentry", slog.String("event-id", string(*eventID)), slog.String("error", appErr.Error()))
	}()
}
