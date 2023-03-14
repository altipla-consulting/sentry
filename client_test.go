package sentry_test

import (
	"context"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/altipla-consulting/errors"
	"github.com/altipla-consulting/sentry"
)

func TestReportError(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}
	defer time.Sleep(3 * time.Second)

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))

	client.Report(context.Background(), foo("foo error"))
	client.Report(context.Background(), foo("bar error"))
	client.Report(context.Background(), CustomErr{Message: "foo custom"})
}

func foo(msg string) error {
	return errors.Trace(bar(msg))
}

func bar(msg string) error {
	return errors.Trace(baz(msg))
}

func baz(msg string) error {
	return errors.Errorf(msg)
}

type CustomErr struct {
	Message string
}

func (e CustomErr) Error() string {
	return "custom error: " + e.Message
}

func TestReportPanic(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}
	defer time.Sleep(3 * time.Second)

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.ReportPanics(context.Background())

	panic("foo")
}

func TestIgnoreAbortError(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}
	defer time.Sleep(3 * time.Second)

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.ReportPanics(context.Background())

	panic(http.ErrAbortHandler)
}
