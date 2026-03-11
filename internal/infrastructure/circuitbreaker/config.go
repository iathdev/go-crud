package circuitbreaker

import "time"

type BreakerConfig struct {
	Name         string
	MaxRequests  uint32        // max requests allowed in half-open state (default: 5)
	Interval     time.Duration // cyclic period to clear counts in closed state (default: 60s)
	Timeout      time.Duration // duration of open state before transitioning to half-open (default: 30s)
	FailureRatio float64       // failure ratio threshold to trip the breaker (default: 0.6)
	MinRequests  uint32        // minimum requests before evaluating failure ratio (default: 10)
}

func (c BreakerConfig) withDefaults() BreakerConfig {
	if c.MaxRequests == 0 {
		c.MaxRequests = 5
	}
	if c.Interval == 0 {
		c.Interval = 60 * time.Second
	}
	if c.Timeout == 0 {
		c.Timeout = 30 * time.Second
	}
	if c.FailureRatio == 0 {
		c.FailureRatio = 0.6
	}
	if c.MinRequests == 0 {
		c.MinRequests = 10
	}
	return c
}
