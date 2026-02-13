package examples

import (
	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
)

// Dependencies holds dependencies needed by built-in actions.
type Dependencies struct {
	StateStore         service.StateStore
	EntitlementGranter service.EntitlementGranter
	Namespace          string
}

// init registers all built-in action types with the factory
func init() {
	// Note: Built-in actions require dependencies, so they cannot be registered
	// directly in init(). Instead, use RegisterBuiltinActions() after creating
	// the dependencies.
}

// RegisterActions registers built-in action factories with dependencies.
func RegisterActions(deps *Dependencies) {
	// Register comeback challenge action
	action.RegisterActionType(ComebackChallengeActionID, func(config action.ActionConfig) (action.Action, error) {
		return NewComebackChallengeAction(config, deps.StateStore), nil
	})

	// Register grant item action
	action.RegisterActionType(GrantItemActionID, func(config action.ActionConfig) (action.Action, error) {
		return NewGrantItemAction(config, deps.EntitlementGranter, deps.Namespace), nil
	})
}
