package builtin

import (
	"time"

	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
)

const (
	TypeLogin = "login"
)

// LoginSignal represents a player login event.
type LoginSignal struct {
	signalType string
	userID     string
	timestamp  time.Time
	metadata   map[string]interface{}
	context    *signal.PlayerContext
}

// NewLoginSignal creates a new login signal.
func NewLoginSignal(userID string, timestamp time.Time, context *signal.PlayerContext) *LoginSignal {
	metadata := map[string]interface{}{
		"event": "login",
	}
	return &LoginSignal{
		signalType: TypeLogin,
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
	}
}

// Type implements Signal interface.
func (s *LoginSignal) Type() string {
	return s.signalType
}

// UserID implements Signal interface.
func (s *LoginSignal) UserID() string {
	return s.userID
}

// Timestamp implements Signal interface.
func (s *LoginSignal) Timestamp() time.Time {
	return s.timestamp
}

// Metadata implements Signal interface.
func (s *LoginSignal) Metadata() map[string]interface{} {
	return s.metadata
}

// Context implements Signal interface.
func (s *LoginSignal) Context() *signal.PlayerContext {
	return s.context
}
