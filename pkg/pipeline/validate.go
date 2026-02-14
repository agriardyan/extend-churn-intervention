package pipeline

import (
	"fmt"
	"strings"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
)

// ValidateWiring validates that the pipeline is correctly wired.
// It checks that:
// - All enabled rules in config have registered instances
// - All enabled actions in config have registered instances
// - All action references in rules exist in the config
//
// This catches common mistakes like:
// - Forgetting to register a rule type factory
// - Typos in rule/action IDs or types
// - Missing action definitions
func ValidateWiring(ruleRegistry *rule.Registry, actionRegistry *action.Registry, config *Config) error {
	var errors []string

	// Check that every enabled rule in config has a registered instance
	for _, rc := range config.Rules {
		if !rc.Enabled {
			continue
		}

		r := ruleRegistry.Get(rc.ID)
		if r == nil {
			errors = append(errors, fmt.Sprintf("rule '%s' (type=%s) is enabled in config but not registered", rc.ID, rc.Type))
		}
	}

	// Check that every enabled action in config has a registered instance
	for _, ac := range config.Actions {
		if !ac.Enabled {
			continue
		}

		a := actionRegistry.Get(ac.ID)
		if a == nil {
			errors = append(errors, fmt.Sprintf("action '%s' (type=%s) is enabled in config but not registered", ac.ID, ac.Type))
		}
	}

	// Note: The validation for "action references in rules exist in config"
	// is already handled by Config.Validate() during config loading

	if len(errors) > 0 {
		return fmt.Errorf("pipeline wiring validation failed:\n  - %s", strings.Join(errors, "\n  - "))
	}

	return nil
}
