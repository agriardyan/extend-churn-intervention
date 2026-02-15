// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package bootstrap

import (
	"fmt"

	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	ruleBuiltin "github.com/AccelByte/extends-anti-churn/pkg/rule/builtin"
	"github.com/sirupsen/logrus"
)

// InitRuleEngine creates and initializes a rule engine with rules from pipeline config.
//
// ============================================================
// DEVELOPER: Register custom rule types here.
// ============================================================
// Rules evaluate signals and decide whether to trigger actions.
// Each rule type defines detection logic (e.g., losing streak,
// session decline, rage quit).
//
// Steps to add a new rule:
// 1. Create your rule in pkg/rule/builtin/ (see examples)
// 2. Implement the Rule interface
// 3. Register the rule type in pkg/rule/builtin/init.go
// 4. Add rule configuration to config/pipeline.yaml
//
// The builtin rules detect:
// - Losing streaks → comeback challenges
// - Rage quits → re-engagement
// - Session decline → intervention
// - Challenge completion → rewards
// ============================================================
func InitRuleEngine(pipelineConfig *pipeline.Config) (*rule.Engine, *rule.Registry, error) {
	// ============================================================
	// DEVELOPER: Builtin rule type registration
	// ============================================================
	// This registers all rule factories defined in pkg/rule/builtin/init.go
	// To add new rule types, modify pkg/rule/builtin/init.go
	// ============================================================
	ruleBuiltin.RegisterRules()

	// ============================================================
	// DEVELOPER: Register custom rule types below
	// ============================================================
	// If you have custom rule types outside pkg/rule/builtin/,
	// register them here:
	//
	// rule.RegisterRuleType("my_custom_rule", func(cfg rule.RuleConfig) (rule.Rule, error) {
	//     return mycustom.NewMyRule(cfg), nil
	// })
	// ============================================================

	// Convert pipeline configs to rule configs
	ruleConfigs := convertRuleConfigs(pipelineConfig.Rules)

	// Create registry and register rules
	registry := rule.NewRegistry()
	if err := rule.RegisterRules(registry, ruleConfigs); err != nil {
		return nil, nil, fmt.Errorf("failed to register rules: %w", err)
	}

	logrus.Infof("registered %d rules", len(ruleConfigs))

	engine := rule.NewEngine(registry)
	logrus.Infof("initialized rule engine")

	return engine, registry, nil
}

func convertRuleConfigs(configs []pipeline.RuleConfig) []rule.RuleConfig {
	result := make([]rule.RuleConfig, len(configs))
	for i, rc := range configs {
		result[i] = rule.RuleConfig{
			ID:         rc.ID,
			Type:       rc.Type,
			Enabled:    rc.Enabled,
			Parameters: rc.Parameters,
		}
	}
	return result
}
