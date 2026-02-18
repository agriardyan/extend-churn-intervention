// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package bootstrap

import (
	"fmt"

	"github.com/AccelByte/extend-churn-intervention/pkg/action"
	actionBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/action/builtin"
	"github.com/AccelByte/extend-churn-intervention/pkg/pipeline"
	"github.com/sirupsen/logrus"
)

// InitActionExecutor creates and initializes an action executor with actions from pipeline config.
//
// ============================================================
// DEVELOPER: Register custom action types here.
// ============================================================
// Actions execute interventions when rules trigger.
// Each action type defines what to do (e.g., create challenge,
// grant item, send notification).
//
// Steps to add a new action:
// 1. Create your action in pkg/action/builtin/ (see examples)
// 2. Implement the Action interface
// 3. Register the action type in pkg/action/builtin/init.go
// 4. Add action configuration to config/pipeline.yaml
// 5. Map it to rules in config/pipeline.yaml
//
// The builtin actions:
// - dispatch_comeback_challenge → creates time-limited challenges
// - grant_item → grants entitlements/items to players
//
// IMPORTANT: Actions may need external service dependencies
// (e.g., itemGranter for granting items). Pass dependencies
// through the Dependencies struct.
// ============================================================
func InitActionExecutor(
	pipelineConfig *pipeline.Config,
	deps *actionBuiltin.Dependencies,
) (*action.Executor, *action.Registry, error) {
	// ============================================================
	// DEVELOPER: Action registration
	// ============================================================
	// This registers all action factories defined in pkg/action/builtin/init.go
	// To add new action types, modify pkg/action/builtin/init.go
	// ============================================================
	actionBuiltin.RegisterActions(deps)

	// Convert pipeline configs to action configs
	actionConfigs := convertActionConfigs(pipelineConfig.Actions)

	// Create registry and register actions
	registry := action.NewRegistry()
	if err := action.RegisterActions(registry, actionConfigs); err != nil {
		return nil, nil, fmt.Errorf("failed to register actions: %w", err)
	}

	logrus.Infof("registered %d actions", len(actionConfigs))

	executor := action.NewExecutor(registry)
	logrus.Infof("initialized action executor")

	return executor, registry, nil
}

func convertActionConfigs(configs []pipeline.ActionConfig) []action.ActionConfig {
	result := make([]action.ActionConfig, len(configs))
	for i, ac := range configs {
		result[i] = action.ActionConfig{
			ID:         ac.ID,
			Type:       ac.Type,
			Enabled:    ac.Enabled,
			Parameters: ac.Parameters,
		}
	}
	return result
}
