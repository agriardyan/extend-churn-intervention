package builtin

import (
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
)

// Dependencies holds dependencies needed by built-in rules.
type Dependencies struct {
	LoginSessionTracker service.LoginSessionTracker
}

// RegisterRules registers all built-in rule types with the factory.
func RegisterRules(deps *Dependencies) {
	rule.RegisterRuleType(RageQuitRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
		return NewRageQuitRule(config), nil
	})

	rule.RegisterRuleType(LosingStreakRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
		return NewLosingStreakRule(config), nil
	})

	rule.RegisterRuleType(SessionDeclineRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
		return NewSessionDeclineRule(config, deps.LoginSessionTracker), nil
	})
}
