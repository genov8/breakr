package config

import (
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

func LoadConfigYAML(filePath string) (*Config, error) {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var rawConfig map[string]interface{}
	if err := yaml.Unmarshal(data, &rawConfig); err != nil {
		return nil, err
	}

	config := &Config{}

	if v, ok := rawConfig["failure_threshold"].(int); ok {
		config.FailureThreshold = v
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
			if num, ok := code.(int); ok {
				config.FailureCodes = append(config.FailureCodes, num)
			}
		}
	}

	return config, nil
}
