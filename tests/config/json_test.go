package config

import (
	"os"
	"testing"
	"time"

	"github.com/genov8/breakr/config"
)

func TestLoadConfigJSON(t *testing.T) {
	jsonData := `{
		"failure_threshold": 2,
		"reset_timeout": "3s",
		"execution_timeout": "1s",
		"window_size": "5s",
		"failure_codes": [500, 502, 503]
	}`

	tmpFile, err := os.CreateTemp("", "config-*.json")
	if err != nil {
		t.Fatalf("Error creating temp file: %v", err)
	}
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	if _, err := tmpFile.Write([]byte(jsonData)); err != nil {
		t.Fatalf("Error writing to temp file: %v", err)
	}
	_ = tmpFile.Close()

	conf, err := config.LoadConfigJSON(tmpFile.Name())
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
		if len(conf.FailureCodes) != 3 || conf.FailureCodes[0] != 500 {
			t.Errorf("Expected default FailureCodes [500, 502, 503], got %v", conf.FailureCodes)
		}
	}
}
