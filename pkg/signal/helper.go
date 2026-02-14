package signal

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

// BuildPlayerContext creates a PlayerContext from churn state.
// This helper is used by event processors that need to enrich signals with player context.
func BuildPlayerContext(userID, namespace string, churnState *state.ChurnState) *PlayerContext {
	playerContext := &PlayerContext{
		UserID:      userID,
		State:       churnState,
		Namespace:   namespace,
		SessionInfo: make(map[string]interface{}),
	}

	// Add session metadata
	playerContext.SessionInfo["sessions_this_week"] = churnState.Sessions.ThisWeek
	playerContext.SessionInfo["sessions_last_week"] = churnState.Sessions.LastWeek
	playerContext.SessionInfo["challenge_active"] = churnState.Challenge.Active
	playerContext.SessionInfo["on_cooldown"] = time.Now().Before(churnState.Intervention.CooldownUntil)

	return playerContext
}
