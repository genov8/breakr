package config

import (
	"github.com/genov8/breakr/config"
	"os"
	"testing"
	"time"
)

func TestLoadConfigYAML(t *testing.T) {
	yamlData := `
failure_threshold: 2
reset_timeout: "3s"
execution_timeout: "1s"
window_size: "5s"
failure_codes:
  - 400
  - 500
`

	tmpFile, err := os.CreateTemp("", "config-*.yaml")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(yamlData)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	tmpFile.Close()

	conf, err := config.LoadConfigYAML(tmpFile.Name())
	if err != nil {
		t.Fatalf("Error loading config: %v", err)
	}

	if conf.FailureThreshold != 2 {
		t.Errorf("Expected FailureThreshold 2, got %d", conf.FailureThreshold)
	}
	if conf.ResetTimeout != 3*time.Second {
		t.Errorf("Expected ResetTimeout 3s, got %s", conf.ResetTimeout)
	}
	if conf.ExecutionTimeout != 1*time.Second {
		t.Errorf("Expected ExecutionTimeout 1s, got %s", conf.ExecutionTimeout)
	}
	if conf.WindowSize != 5*time.Second {
		t.Errorf("Expected WindowSize 5s, got %s", conf.WindowSize)
	}

	if conf.FailureCodes != nil {
		if len(conf.FailureCodes) != 2 || conf.FailureCodes[0] != 400 || conf.FailureCodes[1] != 500 {
			t.Errorf("Expected FailureCodes [400, 500], got %v", conf.FailureCodes)
		}
	}
}
