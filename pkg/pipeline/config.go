package pipeline

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

// Config represents the complete pipeline configuration.
type Config struct {
	Rules   []RuleConfig   `yaml:"rules"`
	Actions []ActionConfig `yaml:"actions"`
}

// RuleConfig represents a rule configuration entry.
type RuleConfig struct {
	ID         string                 `yaml:"id"`
	Type       string                 `yaml:"type"`
	Enabled    bool                   `yaml:"enabled"`
	Actions    []string               `yaml:"actions,omitempty"` // Action IDs to execute when rule triggers
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
}

// ActionConfig represents an action configuration entry.
type ActionConfig struct {
	ID         string                 `yaml:"id"`
	Type       string                 `yaml:"type"`
	Enabled    bool                   `yaml:"enabled"`
	Parameters map[string]interface{} `yaml:"parameters,omitempty"`
}

// LoadConfig loads pipeline configuration from a YAML file.
// Supports environment variable expansion in the form ${VAR_NAME} or ${VAR_NAME:default}.
func LoadConfig(path string) (*Config, error) {
	// Read file
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", path, err)
	}

	// Expand environment variables
	expanded := expandEnvVars(string(data))

	// Parse YAML
	var config Config
	if err := yaml.Unmarshal([]byte(expanded), &config); err != nil {
		return nil, fmt.Errorf("failed to parse YAML config: %w", err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration for common errors.
func (c *Config) Validate() error {
	// Check for duplicate rule IDs
	ruleIDs := make(map[string]bool)
	for _, rule := range c.Rules {
		if rule.ID == "" {
			return fmt.Errorf("rule with empty ID found")
		}
		if ruleIDs[rule.ID] {
			return fmt.Errorf("duplicate rule ID: %s", rule.ID)
		}
		ruleIDs[rule.ID] = true

		if rule.Type == "" {
			return fmt.Errorf("rule %s has empty type", rule.ID)
		}
	}

	// Check for duplicate action IDs
	actionIDs := make(map[string]bool)
	for _, action := range c.Actions {
		if action.ID == "" {
			return fmt.Errorf("action with empty ID found")
		}
		if actionIDs[action.ID] {
			return fmt.Errorf("duplicate action ID: %s", action.ID)
		}
		actionIDs[action.ID] = true

		if action.Type == "" {
			return fmt.Errorf("action %s has empty type", action.ID)
		}
	}

	// Validate that all action references in rules exist
	for _, rule := range c.Rules {
		for _, actionID := range rule.Actions {
			if !actionIDs[actionID] {
				return fmt.Errorf("rule %s references unknown action: %s", rule.ID, actionID)
			}
		}
	}

	return nil
}

// expandEnvVars expands environment variables in the format ${VAR} or ${VAR:default}.
func expandEnvVars(s string) string {
	return os.Expand(s, func(key string) string {
		// Support ${VAR:default} syntax
		parts := strings.SplitN(key, ":", 2)
		varName := parts[0]
		defaultValue := ""
		if len(parts) == 2 {
			defaultValue = parts[1]
		}

		value := os.Getenv(varName)
		if value == "" {
			return defaultValue
		}
		return value
	})
}
