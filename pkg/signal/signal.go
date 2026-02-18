package signal

import (
	"time"

	"github.com/AccelByte/extend-churn-intervention/pkg/service"
)

// Signal represents a normalized domain event with player context.
// Signals are produced by the SignalProcessor from raw gRPC events and
// are consumed by the Rule Engine for evaluation.
type Signal interface {
	// Type returns the signal type identifier (e.g., "login", "rage_quit", "win").
	Type() string

	// UserID returns the player identifier.
	UserID() string

	// Timestamp returns when the signal occurred.
	Timestamp() time.Time

	// Metadata returns additional signal-specific data.
	// This allows rules to access signal-specific information without type assertions.
	Metadata() map[string]interface{}

	// Context returns enriched player context (state, history).
	Context() *PlayerContext
}

// PlayerContext wraps player state with additional metadata.
// This provides rules with all the context they need to make decisions.
type PlayerContext struct {
	UserID      string
	State       *service.ChurnState
	Namespace   string
	SessionInfo map[string]interface{}
}
