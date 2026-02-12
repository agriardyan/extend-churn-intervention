package action

import (
	"context"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Action performs operations in response to triggers.
// Actions are registered in a Registry and executed by the Executor.
type Action interface {
	// ID returns unique action identifier.
	ID() string

	// Name returns human-readable action name.
	Name() string

	// Execute performs the action.
	// The action can access trigger data and player context to perform its operation.
	// Returns error if the action fails.
	Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error

	// Rollback undoes the action (optional, can return ErrRollbackNotSupported).
	// This is called if a subsequent action in a pipeline fails and rollback is enabled.
	Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error

	// Config returns the action's configuration.
	Config() ActionConfig
}

// ActionResult represents the outcome of an action execution.
type ActionResult struct {
	ActionID string
	Success  bool
	Error    error
	Metadata map[string]interface{}
}

// NewActionResult creates a successful action result.
func NewActionResult(actionID string) *ActionResult {
	return &ActionResult{
		ActionID: actionID,
		Success:  true,
		Metadata: make(map[string]interface{}),
	}
}

// NewActionError creates a failed action result with an error.
func NewActionError(actionID string, err error) *ActionResult {
	return &ActionResult{
		ActionID: actionID,
		Success:  false,
		Error:    err,
		Metadata: make(map[string]interface{}),
	}
}

// WithMetadata adds metadata to the result and returns it for chaining.
func (r *ActionResult) WithMetadata(key string, value interface{}) *ActionResult {
	r.Metadata[key] = value
	return r
}
