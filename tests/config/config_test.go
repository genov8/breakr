package config

import (
	"github.com/genov8/breakr/config"
	"testing"
	"time"
)

func TestConfigValidation(t *testing.T) {
	tests := []struct {
		name    string
		cfg     config.Config
		wantErr bool
	}{
		{
			name: "valid config",
			cfg: config.Config{
				FailureThreshold: 3,
				ResetTimeout:     2 * time.Second,
				ExecutionTimeout: 1 * time.Second,
			},
			wantErr: false,
		},
		{
			name: "invalid FailureThreshold",
			cfg: config.Config{
				FailureThreshold: 0,
				ResetTimeout:     2 * time.Second,
				ExecutionTimeout: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid ResetTimeout",
			cfg: config.Config{
				FailureThreshold: 3,
				ResetTimeout:     0,
				ExecutionTimeout: 1 * time.Second,
			},
			wantErr: true,
		},
		{
			name: "invalid ExecutionTimeout",
			cfg: config.Config{
				FailureThreshold: 3,
				ResetTimeout:     2 * time.Second,
				ExecutionTimeout: 0,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.cfg.Validate()
			if (err != nil) != tt.wantErr {
				t.Errorf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
