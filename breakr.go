package breakr

import (
	"errors"
	"fmt"
	"github.com/genov8/breakr/config"
	"github.com/genov8/breakr/internal"
	"sync"
	"time"
)

type Breaker struct {
	mu              sync.Mutex
	state           internal.State
	config          config.Config
	failures        []time.Time
	lastFailureTime time.Time
}

func New(cfg config.Config) *Breaker {
	if err := cfg.Validate(); err != nil {
		panic(fmt.Sprintf("invalid config: %v", err))
	}

	return &Breaker{
		state:  internal.Closed,
		config: cfg,
	}
}

func (b *Breaker) State() internal.State {
	b.mu.Lock()
	defer b.mu.Unlock()
	return b.state
}

func (b *Breaker) Execute(fn func() (interface{}, error)) (interface{}, error) {
	b.mu.Lock()

	if b.state == internal.Open {
		if time.Since(b.lastFailureTime) > b.config.ResetTimeout {
			b.state = internal.HalfOpen
			b.cleanUpFailures()
		} else {
			b.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	}

	b.mu.Unlock()

	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := fn()
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()

	select {
	case result := <-resultChan:
		b.mu.Lock()
		defer b.mu.Unlock()
		b.reset()
		return result, nil

	case err := <-errChan:
		b.mu.Lock()
		defer b.mu.Unlock()

		if !b.isFailure(err) {
			return nil, err
		}

		b.cleanUpFailures()
		now := time.Now()
		b.failures = append(b.failures, now)
		b.lastFailureTime = time.Now()

		if b.state == internal.HalfOpen {
			b.state = internal.Open
			b.startResetTimer()
			return nil, err
		}

		if b.shouldTrip() {
			b.state = internal.Open
			b.startResetTimer()
		}

		return nil, err

	case <-time.After(b.config.ExecutionTimeout):
		b.mu.Lock()
		defer b.mu.Unlock()

		b.cleanUpFailures()
		now := time.Now()
		b.failures = append(b.failures, now)
		b.lastFailureTime = time.Now()

		if b.state == internal.HalfOpen {
			b.state = internal.Open
			b.startResetTimer()
			return nil, errors.New("execution timed out")
		}

		if b.shouldTrip() {
			b.state = internal.Open
			b.startResetTimer()
		}

		return nil, errors.New("execution timed out")
	}
}

func (b *Breaker) reset() {
	b.failures = []time.Time{}
	b.state = internal.Closed
}

func (b *Breaker) startResetTimer() {
	go func() {
		time.Sleep(b.config.ResetTimeout)
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.state == internal.Open {
			b.state = internal.HalfOpen
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
