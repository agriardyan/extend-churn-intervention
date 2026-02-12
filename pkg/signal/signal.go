package signal

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/state"
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
	State       *state.ChurnState
	Namespace   string
	SessionInfo map[string]interface{}
}

// BaseSignal provides common functionality for all signal types.
type BaseSignal struct {
	signalType string
	userID     string
	timestamp  time.Time
	metadata   map[string]interface{}
	context    *PlayerContext
}

// NewBaseSignal creates a new base signal with common fields.
func NewBaseSignal(signalType, userID string, timestamp time.Time, metadata map[string]interface{}, context *PlayerContext) BaseSignal {
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	return BaseSignal{
		signalType: signalType,
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
	}
}

// Type implements Signal interface.
func (b BaseSignal) Type() string {
	return b.signalType
}

// UserID implements Signal interface.
func (b BaseSignal) UserID() string {
	return b.userID
}

// Timestamp implements Signal interface.
func (b BaseSignal) Timestamp() time.Time {
	return b.timestamp
}

// Metadata implements Signal interface.
func (b BaseSignal) Metadata() map[string]interface{} {
	return b.metadata
}

// Context implements Signal interface.
func (b BaseSignal) Context() *PlayerContext {
	return b.context
}
