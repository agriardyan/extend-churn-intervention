package examples

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

const (
	TypeOauthTokenGenerated = "oauth_token_generated"
)

// OauthTokenGeneratedSignal represents a player login event.
type OauthTokenGeneratedSignal struct {
	signalType string
	userID     string
	timestamp  time.Time
	metadata   map[string]interface{}
	context    *signal.PlayerContext
}

// NewOauthTokenGeneratedSignal creates a new login signal.
func NewOauthTokenGeneratedSignal(userID string, timestamp time.Time, context *signal.PlayerContext) *OauthTokenGeneratedSignal {
	metadata := map[string]interface{}{
		"event": "oauth_token_generated",
	}
	return &OauthTokenGeneratedSignal{
		signalType: TypeOauthTokenGenerated,
		userID:     userID,
		timestamp:  timestamp,
		metadata:   metadata,
		context:    context,
	}
}

// Type implements Signal interface.
func (s *OauthTokenGeneratedSignal) Type() string {
	return s.signalType
}

// UserID implements Signal interface.
func (s *OauthTokenGeneratedSignal) UserID() string {
	return s.userID
}

// Timestamp implements Signal interface.
func (s *OauthTokenGeneratedSignal) Timestamp() time.Time {
	return s.timestamp
}

// Metadata implements Signal interface.
func (s *OauthTokenGeneratedSignal) Metadata() map[string]interface{} {
	return s.metadata
}

// Context implements Signal interface.
func (s *OauthTokenGeneratedSignal) Context() *signal.PlayerContext {
	return s.context
}
