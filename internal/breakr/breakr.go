package breakr

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/genov8/breakr/config"
	"github.com/genov8/breakr/metrics"
)

type Breaker struct {
	mu              sync.Mutex
	state           State
	config          config.Config
	failures        []time.Time
	lastFailureTime time.Time
	metrics         *metrics.Metrics
}

func New(cfg config.Config) *Breaker {
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}

	b := &Breaker{
		state:  Closed,
		config: cfg,
	}

	if cfg.Metrics != nil {
		b.metrics = cfg.Metrics
		b.metrics.SetState(b.state.String())
	}

	return b
}

func (b *Breaker) State() State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *Breaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	return b.runWithContext(context.Background(), func(ctx context.Context) (interface{}, error) {
		return fn()
	})
}

func (b *Breaker) ExecuteCtx(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	return b.runWithContext(ctx, fn)
}

func (b *Breaker) reset() {
	b.failures = []time.Time{}
	b.state = Closed
}

func (b *Breaker) startResetTimer() {
	go func() {
		time.Sleep(b.config.ResetTimeout)
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.state == Open {
			b.state = HalfOpen
			b.cleanUpFailures()
		}
	}()
}

func (b *Breaker) isFailure(err error) bool {
	if err == nil {
		return false
	}

	if len(b.config.FailureCodes) == 0 {
		return true
	}

	var httpErr interface{ Code() int }
	if errors.As(err, &httpErr) {
		for _, code := range b.config.FailureCodes {
			if httpErr.Code() == code {
				return true
			}
		}
		return false
	}

	return true
}

func (b *Breaker) cleanUpFailures() {
	if b.config.WindowSize == 0 {
		return
	}

	cutoff := time.Now().Add(-b.config.WindowSize)
	newFailures := make([]time.Time, 0, len(b.failures))

	for _, ts := range b.failures {
		if ts.After(cutoff) {
			newFailures = append(newFailures, ts)
		}
	}

	b.failures = newFailures
}

func (b *Breaker) shouldTrip() bool {
	if b.config.WindowSize > 0 {
		cutoff := time.Now().Add(-b.config.WindowSize)
		count := 0
		for _, ts := range b.failures {
			if ts.After(cutoff) {
				count++
			}
		}
		return count >= b.config.FailureThreshold
	}
	return len(b.failures) >= b.config.FailureThreshold
}
