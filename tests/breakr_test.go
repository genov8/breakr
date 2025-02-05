package tests

import (
	"errors"
	"github.com/genov8/breakr"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	cb := breakr.New(breakr.Config{
		FailureThreshold: 2,
		ResetTimeout:     time.Second,
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
}
