package rule

import "time"

// RuleConfig is the base configuration for all rules.
// This is typically loaded from YAML configuration files.
type RuleConfig struct {
	ID         string                 `yaml:"id" json:"id"`
	Name       string                 `yaml:"name" json:"name"`
	Type       string                 `yaml:"type" json:"type"` // e.g., "builtin.rage_quit"
	Enabled    bool                   `yaml:"enabled" json:"enabled"`
	Priority   int                    `yaml:"priority" json:"priority"`
	Cooldown   *CooldownConfig        `yaml:"cooldown,omitempty" json:"cooldown,omitempty"`
	Conditions map[string]interface{} `yaml:"conditions" json:"conditions"`
	Parameters map[string]interface{} `yaml:"parameters" json:"parameters"` // Rule-specific parameters
}

// CooldownConfig defines rate limiting for rule triggers.
type CooldownConfig struct {
	Duration time.Duration `yaml:"duration" json:"duration"`
	Scope    string        `yaml:"scope" json:"scope"` // "global" or "per_user"
}

// GetConditionInt retrieves an integer value from conditions with a default.
func (c *RuleConfig) GetConditionInt(key string, defaultValue int) int {
	if val, ok := c.Conditions[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return defaultValue
}

// GetConditionFloat retrieves a float value from conditions with a default.
func (c *RuleConfig) GetConditionFloat(key string, defaultValue float64) float64 {
	if val, ok := c.Conditions[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}

// GetConditionString retrieves a string value from conditions with a default.
func (c *RuleConfig) GetConditionString(key string, defaultValue string) string {
	if val, ok := c.Conditions[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetConditionBool retrieves a boolean value from conditions with a default.
func (c *RuleConfig) GetConditionBool(key string, defaultValue bool) bool {
	if val, ok := c.Conditions[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}

// GetInt retrieves an integer value from parameters with a default.
func (c *RuleConfig) GetInt(key string, defaultValue int) int {
	if val, ok := c.Parameters[key]; ok {
		if intVal, ok := val.(int); ok {
			return intVal
		}
	}
	return defaultValue
}

// GetFloat retrieves a float value from parameters with a default.
func (c *RuleConfig) GetFloat(key string, defaultValue float64) float64 {
	if val, ok := c.Parameters[key]; ok {
		if floatVal, ok := val.(float64); ok {
			return floatVal
		}
	}
	return defaultValue
}

// GetString retrieves a string value from parameters with a default.
func (c *RuleConfig) GetString(key string, defaultValue string) string {
	if val, ok := c.Parameters[key]; ok {
		if strVal, ok := val.(string); ok {
			return strVal
		}
	}
	return defaultValue
}

// GetBool retrieves a boolean value from parameters with a default.
func (c *RuleConfig) GetBool(key string, defaultValue bool) bool {
	if val, ok := c.Parameters[key]; ok {
		if boolVal, ok := val.(bool); ok {
			return boolVal
		}
	}
	return defaultValue
}
