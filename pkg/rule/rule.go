package rule

import (
	"context"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Rule evaluates signals and emits triggers when conditions are met.
// Rules are registered in a Registry and evaluated by the Engine.
type Rule interface {
	// ID returns unique rule identifier.
	ID() string

	// Name returns human-readable rule name.
	Name() string

	// SignalTypes returns which signal types this rule handles.
	// An empty slice means the rule handles all signal types.
	SignalTypes() []string

	// Evaluate checks if the signal matches rule conditions.
	// Returns true and trigger data if rule matches, false otherwise.
	// Returns error only for unexpected failures, not rule mismatches.
	Evaluate(ctx context.Context, sig signal.Signal) (bool, *Trigger, error)

	// Config returns the rule's configuration.
	Config() RuleConfig
}

// Trigger represents a rule match that should execute actions.
type Trigger struct {
	RuleID    string                 // ID of the rule that triggered
	UserID    string                 // Player who triggered the rule
	Timestamp time.Time              // When the trigger occurred
	Reason    string                 // Human-readable reason for the trigger
	Metadata  map[string]interface{} // Rule-specific data for actions
	Priority  int                    // Priority for action ordering (higher = first)
}

// NewTrigger creates a new trigger with the given parameters.
func NewTrigger(ruleID, userID, reason string, priority int) *Trigger {
	return &Trigger{
		RuleID:    ruleID,
		UserID:    userID,
		Timestamp: time.Now(),
		Reason:    reason,
		Metadata:  make(map[string]interface{}),
		Priority:  priority,
	}
}

// WithMetadata adds metadata to the trigger and returns it for chaining.
func (t *Trigger) WithMetadata(key string, value interface{}) *Trigger {
	t.Metadata[key] = value
	return t
}

// WithAllMetadata sets all metadata at once.
func (t *Trigger) WithAllMetadata(metadata map[string]interface{}) *Trigger {
	t.Metadata = metadata
	return t
}
