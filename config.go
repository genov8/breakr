package breakr

import "time"

type Config struct {
	FailureThreshold int
	ResetTimeout     time.Duration
}
