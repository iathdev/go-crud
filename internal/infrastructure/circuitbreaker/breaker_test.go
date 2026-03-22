package circuitbreaker

import (
	"errors"
	"testing"
	"time"

	apperr "learning-go/internal/shared/error"

	"github.com/sony/gobreaker/v2"
)

func TestBreaker_Execute_Success(t *testing.T) {
	b := NewBreaker(BreakerConfig{Name: "test"}, nil)

	result, err := b.Execute(func() (any, error) {
		return "hello", nil
	})

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if result != "hello" {
		t.Fatalf("expected 'hello', got %v", result)
	}
}

func TestBreaker_Execute_PassesThroughErrors(t *testing.T) {
	b := NewBreaker(BreakerConfig{Name: "test"}, nil)
	expectedErr := errors.New("some error")

	_, err := b.Execute(func() (any, error) {
		return nil, expectedErr
	})

	if !errors.Is(err, expectedErr) {
		t.Fatalf("expected %v, got %v", expectedErr, err)
	}
}

func TestBreaker_TripsAfterFailures(t *testing.T) {
	b := NewBreaker(BreakerConfig{
		Name:         "test",
		MaxRequests:  1,
		Interval:     10 * time.Second,
		Timeout:      5 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  5,
	}, nil)

	infraErr := errors.New("connection refused")

	// Generate enough failures to trip the breaker
	for i := 0; i < 10; i++ {
		b.Execute(func() (any, error) {
			return nil, infraErr
		})
	}

	if b.State() != gobreaker.StateOpen {
		t.Fatalf("expected StateOpen, got %v", b.State())
	}

	// Next call should fail fast with ErrServiceUnavailable
	_, err := b.Execute(func() (any, error) {
		return "should not execute", nil
	})

	appErr, ok := apperr.IsAppError(err)
	if !ok || appErr.Code() != apperr.CodeServiceUnavailable {
		t.Fatalf("expected CodeServiceUnavailable, got %v", err)
	}
}

func TestBreaker_IsSuccessful_BusinessErrorsDoNotTrip(t *testing.T) {
	notFoundErr := apperr.NotFound("common.not_found")

	b := NewBreaker(BreakerConfig{
		Name:         "test",
		MaxRequests:  1,
		Interval:     10 * time.Second,
		Timeout:      5 * time.Second,
		FailureRatio: 0.6,
		MinRequests:  5,
	}, func(err error) bool {
		if err == nil {
			return true
		}
		appErr, ok := apperr.IsAppError(err)
		return ok && appErr.Code() == apperr.CodeNotFound
	})

	// Generate many "not found" errors - these should NOT trip the breaker
	for i := 0; i < 20; i++ {
		b.Execute(func() (any, error) {
			return nil, notFoundErr
		})
	}

	if b.State() != gobreaker.StateClosed {
		t.Fatalf("expected StateClosed (business errors should not trip), got %v", b.State())
	}
}

func TestBreaker_RecoveryFromOpen(t *testing.T) {
	b := NewBreaker(BreakerConfig{
		Name:         "test",
		MaxRequests:  1,
		Interval:     10 * time.Second,
		Timeout:      1 * time.Second, // Short timeout for test
		FailureRatio: 0.6,
		MinRequests:  5,
	}, nil)

	infraErr := errors.New("connection refused")

	// Trip the breaker
	for i := 0; i < 10; i++ {
		b.Execute(func() (any, error) {
			return nil, infraErr
		})
	}

	if b.State() != gobreaker.StateOpen {
		t.Fatalf("expected StateOpen, got %v", b.State())
	}

	// Wait for timeout to transition to half-open
	time.Sleep(1500 * time.Millisecond)

	// Successful call should close the breaker
	result, err := b.Execute(func() (any, error) {
		return "recovered", nil
	})

	if err != nil {
		t.Fatalf("expected no error after recovery, got %v", err)
	}
	if result != "recovered" {
		t.Fatalf("expected 'recovered', got %v", result)
	}

	if b.State() != gobreaker.StateClosed {
		t.Fatalf("expected StateClosed after recovery, got %v", b.State())
	}
}

func TestRegistry_RegisterAndGet(t *testing.T) {
	r := NewRegistry()
	b := NewBreaker(BreakerConfig{Name: "postgres"}, nil)

	r.Register(b)
	got := r.Get("postgres")

	if got != b {
		t.Fatal("expected same breaker instance")
	}
}

func TestRegistry_Get_Panics_WhenNotFound(t *testing.T) {
	r := NewRegistry()

	defer func() {
		if recover() == nil {
			t.Fatal("expected panic for unregistered breaker")
		}
	}()

	r.Get("nonexistent")
}
