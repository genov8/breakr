package breakr

import (
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
		} else {
			b.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	}

	b.mu.Unlock()

	result, err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.failures++
		b.lastFailureTime = time.Now()

		if b.failures >= b.config.FailureThreshold {
			b.state = internal.Open
		}
		return nil, err
	}

	b.reset()
	return result, nil
}

func (b *Breaker) reset() {
	b.failures = 0
	b.state = internal.Closed
}
