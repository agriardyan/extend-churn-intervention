// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package bootstrap

import (
	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/sirupsen/logrus"
)

// InitPipeline creates and initializes the pipeline manager with rule-to-action mappings.
//
// ============================================================
// DEVELOPER: Configure rule-to-action mappings
// ============================================================
// The pipeline orchestrates the flow:
// Events → Signals → Rules → Actions
//
// Rule-to-action mappings are configured in config/pipeline.yaml:
//
// rules:
//   - id: my-rule
//     type: my_rule_type
//     actions: [action1, action2]  # ← Actions to execute
//
// When a rule triggers:
// 1. The pipeline looks up the action IDs from the mapping
// 2. Executes each action in sequence
// 3. If any action fails, remaining actions are rolled back
//
// To modify mappings, edit config/pipeline.yaml, not this file.
// ============================================================
func InitPipeline(
	processor *signal.Processor,
	ruleEngine *rule.Engine,
	actionExecutor *action.Executor,
	pipelineConfig *pipeline.Config,
) *pipeline.Manager {
	// ============================================================
	// Build rule-to-actions mapping from config
	// ============================================================
	// This extracts the 'actions' field from each rule in
	// config/pipeline.yaml and creates a map for quick lookup.
	// ============================================================
	ruleActions := make(map[string][]string)
	for _, rc := range pipelineConfig.Rules {
		if len(rc.Actions) > 0 {
			ruleActions[rc.ID] = rc.Actions
		}
	}

	logrus.Infof("configured %d rule-to-action mappings", len(ruleActions))

	manager := pipeline.NewManager(processor, ruleEngine, actionExecutor, ruleActions, nil)
	logrus.Infof("initialized pipeline manager")

	return manager
}
