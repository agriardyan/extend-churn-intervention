package action

import "time"

// ActionConfig is the base configuration for all actions.
// This is typically loaded from YAML configuration files.
type ActionConfig struct {
	ID         string                 `yaml:"id" json:"id"`
	Name       string                 `yaml:"name" json:"name"`
	Type       string                 `yaml:"type" json:"type"` // e.g., "builtin.create_challenge"
	Enabled    bool                   `yaml:"enabled" json:"enabled"`
	Async      bool                   `yaml:"async" json:"async"`
	Retry      *RetryConfig           `yaml:"retry,omitempty" json:"retry,omitempty"`
	Parameters map[string]interface{} `yaml:"parameters" json:"parameters"`
}

// RetryConfig defines retry behavior for failed actions.
type RetryConfig struct {
	MaxAttempts int           `yaml:"max_attempts" json:"max_attempts"`
	Delay       time.Duration `yaml:"delay" json:"delay"`
	Backoff     string        `yaml:"backoff" json:"backoff"` // "linear", "exponential"
}

// GetParameterInt retrieves an integer parameter with a default.
func (c *ActionConfig) GetParameterInt(key string, defaultValue int) int {
	if val, ok := c.Parameters[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return defaultValue
}

// GetParameterFloat retrieves a float parameter with a default.
func (c *ActionConfig) GetParameterFloat(key string, defaultValue float64) float64 {
	if val, ok := c.Parameters[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}

// GetParameterString retrieves a string parameter with a default.
func (c *ActionConfig) GetParameterString(key string, defaultValue string) string {
	if val, ok := c.Parameters[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetParameterBool retrieves a boolean parameter with a default.
func (c *ActionConfig) GetParameterBool(key string, defaultValue bool) bool {
	if val, ok := c.Parameters[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// GetParameterStringSlice retrieves a string slice parameter with a default.
func (c *ActionConfig) GetParameterStringSlice(key string, defaultValue []string) []string {
	if val, ok := c.Parameters[key]; ok {
		if sliceVal, ok := val.([]string); ok {
			return sliceVal
		}
		// Try to convert from []interface{}
		if interfaceSlice, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(interfaceSlice))
			for _, item := range interfaceSlice {
				if str, ok := item.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return defaultValue
}
