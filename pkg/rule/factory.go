package rule

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// RuleFactory is a function that creates a rule from a configuration.
type RuleFactory func(config RuleConfig) (Rule, error)

// factories stores registered rule factories by type
var factories = make(map[string]RuleFactory)

// RegisterRuleType registers a factory function for a rule type.
// This allows external packages to register their rule types without creating import cycles.
func RegisterRuleType(ruleType string, factory RuleFactory) {
	factories[ruleType] = factory
	logrus.Debugf("registered rule type: %s", ruleType)
}

// CreateRule creates a rule instance based on the configuration.
// Returns an error if the rule type is unknown.
func CreateRule(config RuleConfig) (Rule, error) {
	if !config.Enabled {
		logrus.Infof("skipping disabled rule: %s", config.ID)
		return nil, nil
	}

	logrus.Infof("creating rule: id=%s, type=%s, priority=%d", config.ID, config.Type, config.Priority)

	factory, exists := factories[config.Type]
	if !exists {
		return nil, fmt.Errorf("unknown rule type: %s", config.Type)
	}

	return factory(config)
}

// CreateRules creates multiple rule instances from a list of configurations.
// Returns all successfully created rules and any errors encountered.
func CreateRules(configs []RuleConfig) ([]Rule, []error) {
	var rules []Rule
	var errors []error

	for _, config := range configs {
		rule, err := CreateRule(config)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create rule %s: %w", config.ID, err))
			continue
		}

		if rule != nil {
			rules = append(rules, rule)
		}
	}

	return rules, errors
}

// RegisterRules registers multiple rules with the provided registry.
// This is a convenience function for setting up rules.
func RegisterRules(registry *Registry, configs []RuleConfig) error {
	rules, errors := CreateRules(configs)

	if len(errors) > 0 {
		logrus.Warnf("encountered %d errors while creating rules", len(errors))
		for _, err := range errors {
			logrus.Warnf("rule creation error: %v", err)
		}
	}

	for _, rule := range rules {
		if err := registry.Register(rule); err != nil {
			return fmt.Errorf("failed to register rule %s: %w", rule.ID(), err)
		}
	}

	logrus.Infof("registered %d rules", len(rules))
	return nil
}
