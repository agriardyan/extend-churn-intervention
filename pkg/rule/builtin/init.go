package builtin

import (
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
)

// RegisterBuiltinRules registers all built-in rule types with the factory.
func RegisterBuiltinRules() {
	rule.RegisterRuleType(RageQuitRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
		return NewRageQuitRule(config), nil
	})

	rule.RegisterRuleType(LosingStreakRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
		return NewLosingStreakRule(config), nil
	})

	rule.RegisterRuleType(SessionDeclineRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
		return NewSessionDeclineRule(config), nil
	})

	rule.RegisterRuleType(ChallengeCompletionRuleID, func(config rule.RuleConfig) (rule.Rule, error) {
		return NewChallengeCompletionRule(config), nil
	})
}
