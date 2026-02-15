package builtin

import (
	"context"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
	"github.com/sirupsen/logrus"
)

const (
	// SessionDeclineRuleID is the identifier for session decline detection rule
	SessionDeclineRuleID = "session_decline"
)

// SessionDeclineRule detects when a player's session frequency declines week-over-week.
// This rule uses the weekly reset logic to determine if a player was active last week
// but has not been active this week, indicating potential churn.
type SessionDeclineRule struct {
	config rule.RuleConfig
}

// NewSessionDeclineRule creates a new session decline detection rule.
func NewSessionDeclineRule(config rule.RuleConfig) *SessionDeclineRule {
	logrus.Infof("creating session decline rule")
	return &SessionDeclineRule{
		config: config,
	}
}

// ID returns the rule identifier.
func (r *SessionDeclineRule) ID() string {
	return r.config.ID
}

// Name returns the rule name.
func (r *SessionDeclineRule) Name() string {
	return "Session Decline Detection"
}

// SignalTypes returns the signal types this rule handles.
func (r *SessionDeclineRule) SignalTypes() []string {
	return []string{signalBuiltin.TypeLogin}
}

// Config returns the rule configuration.
func (r *SessionDeclineRule) Config() rule.RuleConfig {
	return r.config
}

// Evaluate checks if a player's session frequency has declined.
func (r *SessionDeclineRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
	// Get player context
	playerCtx := sig.Context()
	if playerCtx == nil || playerCtx.State == nil {
		logrus.Debugf("no player context for user %s, skipping session decline check", sig.UserID())
		return false, nil, nil
	}

	now := sig.Timestamp()
	playerState := playerCtx.State

	// Check if player is churning BEFORE doing weekly reset
	// (weekly reset moves ThisWeek to LastWeek, so we need to check first)
	isChurning := isChurning(playerState, now)

	// Perform weekly reset if needed
	resetOccurred := checkWeeklyReset(playerState, now)
	if resetOccurred {
		logrus.Debugf("weekly reset occurred for user %s", sig.UserID())
	}

	// Use the churn status from before the reset
	if !isChurning {
		return false, nil, nil
	}

	// Check if intervention can be triggered (cooldown check)
	if playerState.Cooldown.IsOnCooldown() {
		logrus.Debugf("session decline detected for user %s but intervention in cooldown", sig.UserID())
		return false, nil, nil
	}

	// Check if there's already an active comeback challenge intervention
	activeInterventions := playerState.GetActiveInterventions()
	for _, intervention := range activeInterventions {
		if intervention.Type == "dispatch_comeback_challenge" {
			logrus.Debugf("session decline detected for user %s but comeback challenge already active", sig.UserID())
			return false, nil, nil
		}
	}

	trigger := rule.NewTrigger(r.ID(), sig.UserID(), "Session frequency declined", r.config.Priority)
	trigger.Metadata["last_week_sessions"] = playerState.Sessions.LastWeek
	trigger.Metadata["this_week_sessions"] = playerState.Sessions.ThisWeek
	trigger.Metadata["time_since_reset"] = now.Sub(playerState.Sessions.LastReset).Hours()

	logrus.Infof("session decline rule triggered for user %s: lastWeek=%d, thisWeek=%d",
		sig.UserID(), playerState.Sessions.LastWeek, playerState.Sessions.ThisWeek)

	return true, trigger, nil
}

// checkWeeklyReset checks if a weekly reset should occur and performs it if needed.
// Returns true if a reset occurred, false otherwise.
// NOTE: We only cache session counts - we don't maintain them as source of truth.
func checkWeeklyReset(state *service.ChurnState, now time.Time) bool {
	// Calculate time since last reset
	timeSinceReset := now.Sub(state.Sessions.LastReset)

	// Check if a week hasn't passed (7 days)
	if timeSinceReset < 7*24*time.Hour {
		return false
	}

	logrus.Infof("weekly reset triggered: %v since last reset", timeSinceReset)

	// Move thisWeek to lastWeek
	state.Sessions.LastWeek = state.Sessions.ThisWeek

	// Reset thisWeek counter
	state.Sessions.ThisWeek = 0

	// Update last reset time
	state.Sessions.LastReset = now

	return true
}

// isChurning determines if a player is exhibiting churn behavior.
// A player is churning if:
// - They had activity last week (LastWeek > 0)
// - They have no activity this week (ThisWeek == 0)
// - At least 7 days have passed since last reset
func isChurning(state *service.ChurnState, now time.Time) bool {
	timeSinceReset := now.Sub(state.Sessions.LastReset)

	// Must be at least 7 days since reset to determine churn
	if timeSinceReset < 7*24*time.Hour {
		return false
	}

	// Was active last week but not this week
	churning := state.Sessions.LastWeek > 0 && state.Sessions.ThisWeek == 0

	if churning {
		logrus.Infof("player is churning: lastWeek=%d, thisWeek=%d, timeSinceReset=%v",
			state.Sessions.LastWeek, state.Sessions.ThisWeek, timeSinceReset)
	}

	return churning
}
