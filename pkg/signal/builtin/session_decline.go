package builtin

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

const (
	TypeOauthTokenGenerated = "oauth_token_generated"
)

// OauthTokenGeneratedSignal represents a player login event.
type OauthTokenGeneratedSignal struct {
	signal.BaseSignal
}

// NewOauthTokenGeneratedSignal creates a new login signal.
func NewOauthTokenGeneratedSignal(userID string, timestamp time.Time, context *signal.PlayerContext) *OauthTokenGeneratedSignal {
	metadata := map[string]interface{}{
		"event": "oauth_token_generated",
	}
	return &OauthTokenGeneratedSignal{
		BaseSignal: signal.NewBaseSignal(TypeOauthTokenGenerated, userID, timestamp, metadata, context),
	}
}
