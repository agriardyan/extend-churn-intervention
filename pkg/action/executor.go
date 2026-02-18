package action

import (
	"context"
	"fmt"
	"sync"

	"github.com/AccelByte/extend-churn-intervention/pkg/rule"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
	"github.com/sirupsen/logrus"
)

// Executor executes actions in response to rule triggers.
type Executor struct {
	registry *Registry
	mu       sync.RWMutex
}

// NewExecutor creates a new action executor.
func NewExecutor(registry *Registry) *Executor {
	return &Executor{
		registry: registry,
	}
}

// Execute runs an action in response to a trigger.
func (e *Executor) Execute(ctx context.Context, actionID string, trigger *rule.Trigger, playerCtx *signal.PlayerContext) (*ActionResult, error) {
	action := e.registry.Get(actionID)
	if action == nil {
		return nil, fmt.Errorf("action not found: %s", actionID)
	}

	logrus.Infof("executing action %s for trigger %s (user: %s)", actionID, trigger.RuleID, trigger.UserID)

	err := action.Execute(ctx, trigger, playerCtx)
	if err != nil {
		logrus.Errorf("action %s failed: %v", actionID, err)
		return NewActionError(actionID, err), err
	}

	logrus.Infof("action %s completed successfully", actionID)
	return NewActionResult(actionID), nil
}

// ExecuteMultiple executes multiple actions in sequence.
// If rollbackOnError is true, previously executed actions will be rolled back if a later action fails.
func (e *Executor) ExecuteMultiple(ctx context.Context, actionIDs []string, trigger *rule.Trigger, playerCtx *signal.PlayerContext, rollbackOnError bool) ([]*ActionResult, error) {
	var results []*ActionResult
	var executedActions []Action

	for _, actionID := range actionIDs {
		action := e.registry.Get(actionID)
		if action == nil {
			err := fmt.Errorf("action not found: %s", actionID)
			logrus.Errorf("%v", err)

			if rollbackOnError && len(executedActions) > 0 {
				e.rollbackActions(ctx, executedActions, trigger, playerCtx)
			}

			return results, err
		}

		logrus.Infof("executing action %s for trigger %s (user: %s)", actionID, trigger.RuleID, trigger.UserID)

		err := action.Execute(ctx, trigger, playerCtx)
		if err != nil {
			logrus.Errorf("action %s failed: %v", actionID, err)
			results = append(results, NewActionError(actionID, err))

			if rollbackOnError && len(executedActions) > 0 {
				e.rollbackActions(ctx, executedActions, trigger, playerCtx)
			}

			return results, err
		}

		executedActions = append(executedActions, action)
		results = append(results, NewActionResult(actionID))
		logrus.Infof("action %s completed successfully", actionID)
	}

	return results, nil
}

// rollbackActions rolls back actions in reverse order.
func (e *Executor) rollbackActions(ctx context.Context, actions []Action, trigger *rule.Trigger, playerCtx *signal.PlayerContext) {
	logrus.Warnf("rolling back %d actions", len(actions))

	// Rollback in reverse order
	for i := len(actions) - 1; i >= 0; i-- {
		action := actions[i]
		logrus.Infof("rolling back action %s", action.ID())

		err := action.Rollback(ctx, trigger, playerCtx)
		if err != nil {
			if err == ErrRollbackNotSupported {
				logrus.Warnf("action %s does not support rollback", action.ID())
			} else {
				logrus.Errorf("failed to rollback action %s: %v", action.ID(), err)
			}
		} else {
			logrus.Infof("action %s rolled back successfully", action.ID())
		}
	}
}

// GetRegistry returns the action registry used by this executor.
func (e *Executor) GetRegistry() *Registry {
	return e.registry
}
