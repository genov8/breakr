package breakr

import (
	"github.com/genov8/breakr/internal"
	"sync"
)

type Breaker struct {
	mu     sync.Mutex
	state  internal.State
	config Config
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
		b.mu.Unlock()
		return nil, ErrCircuitOpen
	}
	b.mu.Unlock()

	result, err := fn()

	b.mu.Lock()
	defer b.mu.Unlock()

	if err != nil {
		b.state = internal.Open
		return nil, err
	}

	b.state = internal.Closed
	return result, nil
}
