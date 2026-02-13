package examples

import (
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
)

// RegisterRules registers all built-in rule types with the factory.
// deps can be nil - rules that need dependencies will handle nil gracefully.
func RegisterRules(deps *service.Dependencies) {
	// Rules without external dependencies - pass config directly
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
