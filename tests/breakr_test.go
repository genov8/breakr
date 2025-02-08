package tests

import (
	"errors"
	"github.com/genov8/breakr"
	"github.com/genov8/breakr/internal"
	"sync"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := breakr.New(breakr.Config{
		FailureThreshold: 2,
		ResetTimeout:     time.Second,
		ExecutionTimeout: 500 * time.Millisecond,
	})

	failFn := func() (interface{}, error) {
		return nil, errors.New("error")
	}

	successFn := func() (interface{}, error) {
		return "success", nil
	}

	_, _ = cb.Execute(failFn)
	_, _ = cb.Execute(failFn)

	if cb.State().String() != "Open" {
		t.Errorf("expected state to be Open, got %s", cb.State().String())
	}

	time.Sleep(1100 * time.Millisecond)

	_, err := cb.Execute(successFn)
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}

	if cb.State().String() != "Closed" {
		t.Errorf("expected state to be Closed, got %s", cb.State().String())
	}

	slowFn := func() (interface{}, error) {
		time.Sleep(700 * time.Millisecond)
		return "success", nil
	}

	_, err = cb.Execute(slowFn)

	if err == nil || err.Error() != "execution timed out" {
		t.Errorf("expected execution timeout error, got %v", err)
	}
}

func TestCircuitBreakerConcurrency(t *testing.T) {
	cb := breakr.New(breakr.Config{
		FailureThreshold: 3,
		ResetTimeout:     time.Second,
		ExecutionTimeout: 500 * time.Millisecond,
	})

	failFn := func() (interface{}, error) {
		return nil, errors.New("error")
	}

	var wg sync.WaitGroup
	const goroutines = 10

	for i := 0; i < goroutines; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			cb.Execute(failFn)
		}()
	}

	wg.Wait()

	if cb.State() != internal.Open {
		t.Errorf("expected Circuit Breaker to be Open, got %v", cb.State())
	}

	time.Sleep(1100 * time.Millisecond)

	if cb.State() != internal.HalfOpen {
		t.Errorf("expected Circuit Breaker to be Half-Open, got %v", cb.State())
	}

	successFn := func() (interface{}, error) {
		return "success", nil
	}

	_, err := cb.Execute(successFn)
	if err != nil {
		t.Errorf("expected success, got error: %v", err)
	}

	if cb.State() != internal.Closed {
		t.Errorf("expected Circuit Breaker to be Closed, got %v", cb.State())
	}
}
