package breakr

import (
	"context"
	"time"
)

func (b *Breaker) runWithContext(ctx context.Context, fn func(ctx context.Context) (interface{}, error)) (interface{}, error) {
	start := time.Now()

	b.mu.Lock()
	stateAtStart := b.state

	if b.state == Open {
		if time.Since(b.lastFailureTime) > b.config.ResetTimeout {
			b.setState(HalfOpen)
			b.cleanUpFailures()
			stateAtStart = HalfOpen
		} else {
			b.mu.Unlock()

			if b.metrics != nil {
				b.metrics.ObserveBlocked(stateAtStart.String())
			}
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
		d := time.Since(start)

		b.mu.Lock()
		b.cleanUpFailures()
		now := time.Now()
		b.failures = append(b.failures, now)
		b.lastFailureTime = now

		if b.state == HalfOpen || b.shouldTrip() {
			b.setState(Open)
			b.startResetTimer()
		}
		b.mu.Unlock()

		if b.metrics != nil {
			b.metrics.ObserveTimeout(stateAtStart.String(), d)
		}
		return nil, ctx.Err()

	case result := <-resultChan:
		d := time.Since(start)

		b.mu.Lock()
		b.reset()
		b.mu.Unlock()

		if b.metrics != nil {
			b.metrics.ObserveSuccess(stateAtStart.String(), d)
		}

		return result, nil

	case err := <-errChan:
		d := time.Since(start)

		b.mu.Lock()
		if !b.isFailure(err) {
			b.mu.Unlock()

			if b.metrics != nil {
				b.metrics.ObserveIgnored(stateAtStart.String(), d)
			}
			return nil, err
		}

		b.cleanUpFailures()
		now := time.Now()
		b.failures = append(b.failures, now)
		b.lastFailureTime = now

		if b.state == HalfOpen || b.shouldTrip() {
			b.setState(Open)
			b.startResetTimer()
		}
		b.mu.Unlock()

		if b.metrics != nil {
			b.metrics.ObserveError(stateAtStart.String(), d)
		}

		return nil, err
	}
}
