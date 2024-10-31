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
	hub *sentry.Hub
}

// NewClient opens a new connection to the Sentry report API. If dsn is empty
// it will return nil. A nil client can be used safely but it won't report anything.
func NewClient(dsn string) *Client {
	if dsn == "" {
		return nil
	}

	client, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:     dsn,
		Release: os.Getenv("VERSION"),
	})
	if err != nil {
		panic(err)
	}

	return &Client{
		hub: sentry.NewHub(client, sentry.NewScope()),
	}
}

// Report reports an error to Sentry.
func (client *Client) Report(ctx context.Context, appErr error) {
	if client == nil {
		return
	}
	client.sendReport(ctx, appErr, nil)
}

// ReportRequest reports an error linked to a HTTP request.
func (client *Client) ReportRequest(r *http.Request, appErr error) {
	if client == nil {
		return
	}
	client.sendReport(r.Context(), appErr, r)
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
	client.sendReportPanic(ctx, fmt.Errorf("panic: %v", panicErr), string(debug.Stack()), nil)
}

// ReportPanicsRequest detects panics in the body of the function and reports them
// linked to a HTTP request.
func (client *Client) ReportPanicsRequest(r *http.Request) {
	if client == nil {
		return
	}
	if rec := recover(); rec != nil { // revive:disable-line:defer
		client.sendReportPanic(r.Context(), fmt.Errorf("panic: %v", rec), string(debug.Stack()), r)
	}
}

func (client *Client) sendReport(ctx context.Context, appErr error, r *http.Request) {
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

		info := FromContext(ctx)
		if info != nil {
			event.Breadcrumbs = info.breadcrumbs
			event.Tags = info.tags
		}

		if r != nil {
			event.Request = sentry.NewRequest(r)
		}

		eventID := client.hub.CaptureEvent(event)
		slog.Info("Error sent to Sentry", slog.String("event-id", string(*eventID)))
	}()
}

func (client *Client) sendReportPanic(ctx context.Context, appErr error, message string, r *http.Request) {
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

		info := FromContext(ctx)
		if info != nil {
			event.Breadcrumbs = info.breadcrumbs
		}

		if r != nil {
			event.Request = sentry.NewRequest(r)
		}

		eventID := client.hub.CaptureEvent(event)
		slog.Info("Error sent to Sentry", slog.String("event-id", string(*eventID)))
	}()
}
