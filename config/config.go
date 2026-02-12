package config

import (
	"errors"
	"time"

	"github.com/genov8/breakr/metrics"
)

type Config struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	ExecutionTimeout time.Duration
	WindowSize       time.Duration
	FailureCodes     []int
	Metrics          *metrics.Metrics
}

func (c Config) Validate() error {
	if c.FailureThreshold <= 0 {
		return errors.New("FailureThreshold must be > 0")
	}
	if c.ResetTimeout <= 0 {
		return errors.New("ResetTimeout must be > 0")
	}
	if c.ExecutionTimeout <= 0 {
		return errors.New("ExecutionTimeout must be > 0")
	}
	return nil
}
