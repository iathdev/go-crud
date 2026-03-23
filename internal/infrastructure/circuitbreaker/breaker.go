package circuitbreaker

import (
	"errors"

	apperr "learning-go/internal/shared/error"
	"learning-go/internal/shared/logger"

	"github.com/sony/gobreaker/v2"
	"go.uber.org/zap"
)

type Breaker struct {
	cb   *gobreaker.CircuitBreaker[any]
	name string
}

func NewBreaker(cfg BreakerConfig, isSuccessful func(err error) bool) *Breaker {
	cfg = cfg.withDefaults()

	settings := gobreaker.Settings{
		Name:        cfg.Name,
		MaxRequests: cfg.MaxRequests,
		Interval:    cfg.Interval,
		Timeout:     cfg.Timeout,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			if counts.Requests < cfg.MinRequests {
				return false
			}
			failureRatio := float64(counts.TotalFailures) / float64(counts.Requests)
			return failureRatio >= cfg.FailureRatio
		},
		IsSuccessful: isSuccessful,
		OnStateChange: func(name string, from gobreaker.State, to gobreaker.State) {
			logger.Warn("[SERVER] circuit breaker state change",
				zap.String("breaker", name),
				zap.String("from", from.String()),
				zap.String("to", to.String()),
			)
		},
	}

	return &Breaker{
		cb:   gobreaker.NewCircuitBreaker[any](settings),
		name: cfg.Name,
	}
}

func (breaker *Breaker) Name() string {
	return breaker.name
}

func (breaker *Breaker) Execute(fn func() (any, error)) (any, error) {
	result, err := breaker.cb.Execute(fn)
	if err != nil {
		if errors.Is(err, gobreaker.ErrOpenState) || errors.Is(err, gobreaker.ErrTooManyRequests) {
			return nil, apperr.ServiceUnavailable("common.service_unavailable", err)
		}
		return nil, err
	}
	return result, nil
}

func (breaker *Breaker) State() gobreaker.State {
	return breaker.cb.State()
}
