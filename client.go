package sentry

import (
	"context"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/getsentry/raven-go"
	"github.com/juju/errors"
	log "github.com/sirupsen/logrus"
)

type Client struct {
	dsn string
}

func NewClient(dsn string) *Client {
	return &Client{
		dsn: dsn,
	}
}

func (client *Client) ReportInternal(ctx context.Context, appErr error) {
	client.report(ctx, appErr, nil)
}

func (client *Client) ReportRequest(appErr error, r *http.Request) {
	client.report(r.Context(), appErr, r)
}

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

type StackTracer interface {
	StackTrace() []string
}

func (client *Client) report(ctx context.Context, appErr error, r *http.Request) {
	event := make(chan string, 1)
	go func() {
		jujuErr, ok := appErr.(StackTracer)
		if !ok {
			jujuErr = errors.Errorf("unknown error type: %s", appErr.Error()).(StackTracer)
		}
		stacktrace := new(raven.Stacktrace)
		for _, entry := range jujuErr.StackTrace() {
			parts := strings.Split(entry, ":")
			if len(parts) > 2 {
				n, err := strconv.ParseInt(parts[1], 10, 64)
				if err == nil {
					stacktrace.Frames = append(stacktrace.Frames, &raven.StacktraceFrame{
						Filename:    parts[0],
						Lineno:      int(n),
						ContextLine: entry,
					})
					continue
				}
			}

			// Fallback to avoid erroring out here if no location is found
			stacktrace.Frames = append(stacktrace.Frames, &raven.StacktraceFrame{
				Filename:    entry,
				ContextLine: entry,
			})
		}

		// Invert frames to show them in the correct order in the Sentry UI
		for i, j := 0, len(stacktrace.Frames)-1; i < j; i, j = i+1, j-1 {
			stacktrace.Frames[i], stacktrace.Frames[j] = stacktrace.Frames[j], stacktrace.Frames[i]
		}

		client, err := raven.New(client.dsn)
		if err != nil {
			log.WithField("error", err).Error("Cannot create client")
			return
		}
		client.SetRelease(os.Getenv("VERSION"))

		info := FromContext(ctx)
		interfaces := []raven.Interface{
			&Exception{
				Stacktrace: stacktrace,
				Module:     "backend",
				Value:      appErr.Error(),
				Type:       appErr.Error(),
			},
			&Breadcrumbs{
				Values: info.breadcrumbs,
			},
		}
		if r != nil {
			interfaces = append(interfaces, raven.NewHttp(r), &raven.User{IP: r.RemoteAddr})
		}
		packet := raven.NewPacket(appErr.Error(), interfaces...)
		eventID, ch := client.Capture(packet, nil)
		<-ch
		event <- eventID
	}()

	select {
	case eventID := <-event:
		log.WithField("eventID", eventID).Info("Error logged to sentry")

	case <-time.After(5 * time.Second):
		log.Error("Timeout trying to reach sentry")
	}
}
