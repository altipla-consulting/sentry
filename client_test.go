package sentry_test

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/altipla-consulting/sentry"
)

func TestReportError(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.Flush(5 * time.Second)

	client.Report(context.Background(), foo("foo error"))
	client.Report(context.Background(), foo("bar error"))
	client.Report(context.Background(), CustomErr{Message: "foo custom"})
}

func foo(msg string) error {
	return fmt.Errorf("foo function: %w", bar(msg))
}

func bar(msg string) error {
	return fmt.Errorf("bar function: %w", baz(msg))
}

func baz(msg string) error {
	return fmt.Errorf("baz function: %s", msg)
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

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.Flush(5 * time.Second)
	defer client.ReportPanics(context.Background())

	panic("foo")
}

func TestIgnoreAbortError(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.Flush(5 * time.Second)
	defer client.ReportPanics(context.Background())

	panic(http.ErrAbortHandler)
}

func TestReportStackTrace(t *testing.T) {
	if os.Getenv("SENTRY_DSN") == "" {
		t.Skip("Skipping sentry real tests without SENTRY_DSN=foo env variable")
	}

	client := sentry.NewClient(os.Getenv("SENTRY_DSN"))
	defer client.Flush(5 * time.Second)

	client.Report(context.Background(), fmt.Errorf("wrapper1: %w", wrapper1()))
}

func wrapper1() error {
	return fmt.Errorf("wrapper1: %w", wrapper2())
}

func wrapper2() error {
	if err := wrapper3(); err != nil {
		return fmt.Errorf("wrapper2: %w", err)
	}
	return nil
}

func wrapper3() error {
	return fmt.Errorf("wrapper3: new formatted error")
}
