package breakr

import (
	"context"
	"time"
)

func (b *Breaker) runWithContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	b.mu.Lock()
	if b.state == Open {
		if time.Since(b.lastFailureTime) > b.config.ResetTimeout {
			b.state = HalfOpen
			b.cleanUpFailures()
		} else {
			b.mu.Unlock()
			return nil, ErrCircuitOpen
		}
	}
	b.mu.Unlock()

	if _, ok := ctx.Deadline(); !ok && b.config.ExecutionTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, b.config.ExecutionTimeout)
		defer cancel()
	}

	resultChan := make(chan interface{}, 1)
	errChan := make(chan error, 1)

	go func() {
		result, err := fn(ctx)
		if err != nil {
			errChan <- err
		} else {
			resultChan <- result
		}
	}()

	select {
	case <-ctx.Done():
		b.mu.Lock()
		defer b.mu.Unlock()
		b.cleanUpFailures()
		now := time.Now()
		b.failures = append(b.failures, now)
		b.lastFailureTime = now
		if b.state == HalfOpen {
			b.state = Open
			b.startResetTimer()
		} else if b.shouldTrip() {
			b.state = Open
			b.startResetTimer()
		}
		return nil, ctx.Err()

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
		b.lastFailureTime = now
		if b.state == HalfOpen {
			b.state = Open
			b.startResetTimer()
		} else if b.shouldTrip() {
			b.state = Open
			b.startResetTimer()
		}
		return nil, err
	}
}
