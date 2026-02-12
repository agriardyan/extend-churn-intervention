package action

import "errors"

var (
	// ErrRollbackNotSupported indicates that an action doesn't support rollback.
	ErrRollbackNotSupported = errors.New("rollback not supported for this action")

	// ErrActionDisabled indicates that an action is disabled in configuration.
	ErrActionDisabled = errors.New("action is disabled")

	// ErrActionNotFound indicates that a requested action doesn't exist in the registry.
	ErrActionNotFound = errors.New("action not found in registry")

	// ErrInvalidConfig indicates that an action's configuration is invalid.
	ErrInvalidConfig = errors.New("invalid action configuration")

	// ErrMaxRetriesExceeded indicates that an action failed after all retry attempts.
	ErrMaxRetriesExceeded = errors.New("maximum retry attempts exceeded")

	// ErrMissingPlayerContext indicates that required player context is missing.
	ErrMissingPlayerContext = errors.New("missing player context")
)
