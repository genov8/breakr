package config

import "time"

type Config struct {
	FailureThreshold int
	ResetTimeout     time.Duration
	ExecutionTimeout time.Duration
	WindowSize       time.Duration
	FailureCodes     []int
}
