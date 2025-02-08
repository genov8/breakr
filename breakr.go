package breakr

import (
	"errors"
	"github.com/genov8/breakr/internal"
	"sync"
	"time"
)

type Breaker struct {
	mu              sync.Mutex
	state           internal.State
	config          Config
	failures        int
	lastFailureTime time.Time
}

func New(config Config) *Breaker {
	return &Breaker{
		state:  internal.Closed,
		config: config,
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
			b.failures = 0
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
		b.failures++
		b.lastFailureTime = time.Now()

		if b.state == internal.HalfOpen {
			b.state = internal.Open
			b.startResetTimer()
			return nil, err
		}

		if b.failures >= b.config.FailureThreshold {
			b.state = internal.Open
			b.startResetTimer()
		}

		return nil, err

	case <-time.After(b.config.ExecutionTimeout):
		b.mu.Lock()
		defer b.mu.Unlock()
		b.failures++
		b.lastFailureTime = time.Now()

		if b.state == internal.HalfOpen {
			b.state = internal.Open
			b.startResetTimer()
			return nil, errors.New("execution timed out")
		}

		if b.failures >= b.config.FailureThreshold {
			b.state = internal.Open
			b.startResetTimer()
		}

		return nil, errors.New("execution timed out")
	}
}

func (b *Breaker) reset() {
	b.failures = 0
	b.state = internal.Closed
}

func (b *Breaker) startResetTimer() {
	go func() {
		time.Sleep(b.config.ResetTimeout)
		b.mu.Lock()
		defer b.mu.Unlock()

		if b.state == internal.Open {
			b.state = internal.HalfOpen
			b.failures = 0
		}
	}()
}
