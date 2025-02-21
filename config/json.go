package config

import (
	"encoding/json"
	"os"
	"time"
)

func LoadConfigJSON(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rawConfig map[string]interface{}
	if err := json.Unmarshal(data, &rawConfig); err != nil {
		return nil, err
	}

	config := &Config{}

	if v, ok := rawConfig["failure_threshold"].(float64); ok {
		config.FailureThreshold = int(v)
	}
	if v, ok := rawConfig["reset_timeout"].(string); ok {
		config.ResetTimeout, _ = time.ParseDuration(v)
	}
	if v, ok := rawConfig["execution_timeout"].(string); ok {
		config.ExecutionTimeout, _ = time.ParseDuration(v)
	}
	if v, ok := rawConfig["window_size"].(string); ok {
		config.WindowSize, _ = time.ParseDuration(v)
	}
	if v, ok := rawConfig["failure_codes"].([]interface{}); ok {
		for _, code := range v {
			if num, ok := code.(float64); ok {
				config.FailureCodes = append(config.FailureCodes, int(num))
			}
		}
	}

	return config, nil
}
