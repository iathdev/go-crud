package sentry

import (
	"fmt"
	"learning-go/internal/infrastructure/config"
	"time"

	"github.com/getsentry/sentry-go"
)

// Init initializes the Sentry SDK. Returns a cleanup function.
// If SENTRY_DSN is empty, returns a no-op cleanup.
func Init(cfg *config.Config) (func(), error) {
	if cfg.SentryDSN == "" {
		return func() {}, nil
	}

	err := sentry.Init(sentry.ClientOptions{
		Dsn:                cfg.SentryDSN,
		Environment:        cfg.SentryEnvironment,
		TracesSampleRate:   cfg.GetSentrySampleRate(),
		Release:            cfg.GetServiceName(),
		AttachStacktrace:   true,
		EnableTracing:      true,
		SampleRate:         1.0,
		BeforeSend: func(event *sentry.Event, hint *sentry.EventHint) *sentry.Event {
			return event
		},
	})
	if err != nil {
		return nil, err
	}

	return func() {
		sentry.Flush(2 * time.Second)
	}, nil
}

// CaptureException sends an error to Sentry.
func CaptureException(err error) {
	sentry.CaptureException(err)
}

// CaptureMessage sends a message to Sentry.
func CaptureMessage(msg string) {
	sentry.CaptureMessage(msg)
}

// RecoverWithSentry captures panic value as a Sentry event with stack trace.
func RecoverWithSentry(recovered interface{}) {
	if recovered == nil {
		return
	}
	hub := sentry.CurrentHub().Clone()
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelFatal)
	})
	switch v := recovered.(type) {
	case error:
		hub.Recover(v)
	default:
		hub.Recover(fmt.Errorf("panic: %v", v))
	}
	hub.Flush(2 * time.Second)
}
