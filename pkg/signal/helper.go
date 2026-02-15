package signal

import (
	"github.com/AccelByte/extends-anti-churn/pkg/service"
)

// BuildPlayerContext creates a PlayerContext from churn state.
// This helper is used by event processors that need to enrich signals with player context.
func BuildPlayerContext(userID, namespace string, churnState *service.ChurnState) *PlayerContext {
	playerContext := &PlayerContext{
		UserID:      userID,
		State:       churnState,
		Namespace:   namespace,
		SessionInfo: make(map[string]interface{}),
	}

	// Add session metadata
	playerContext.SessionInfo["sessions_this_week"] = churnState.Sessions.ThisWeek
	playerContext.SessionInfo["sessions_last_week"] = churnState.Sessions.LastWeek
	playerContext.SessionInfo["active_interventions"] = len(churnState.GetActiveInterventions())
	playerContext.SessionInfo["on_cooldown"] = churnState.Cooldown.IsOnCooldown()

	return playerContext
}
