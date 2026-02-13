package examples

import (
	"context"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalExamples "github.com/AccelByte/extends-anti-churn/pkg/signal/examples"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
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
	return []string{signalExamples.TypeOauthTokenGenerated}
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
	isChurning := state.IsChurning(playerState, now)

	// Perform weekly reset if needed
	resetOccurred := state.CheckWeeklyReset(playerState, now)
	if resetOccurred {
		logrus.Debugf("weekly reset occurred for user %s", sig.UserID())
	}

	// Use the churn status from before the reset
	if !isChurning {
		return false, nil, nil
	}

	// Check if intervention can be triggered (cooldown check)
	if !state.CanTriggerIntervention(playerState, now) {
		logrus.Debugf("session decline detected for user %s but intervention in cooldown", sig.UserID())
		return false, nil, nil
	}

	// Check if there's already an active challenge
	if playerState.Challenge.Active {
		logrus.Debugf("session decline detected for user %s but challenge already active", sig.UserID())
		return false, nil, nil
	}

	trigger := rule.NewTrigger(r.ID(), sig.UserID(), "Session frequency declined", r.config.Priority)
	trigger.Metadata["last_week_sessions"] = playerState.Sessions.LastWeek
	trigger.Metadata["this_week_sessions"] = playerState.Sessions.ThisWeek
	trigger.Metadata["time_since_reset"] = now.Sub(playerState.Sessions.LastReset).Hours()

	logrus.Infof("session decline rule triggered for user %s: lastWeek=%d, thisWeek=%d",
		sig.UserID(), playerState.Sessions.LastWeek, playerState.Sessions.ThisWeek)

	return true, trigger, nil
}

// SetInterventionCooldown is a helper to set the cooldown after an intervention is triggered.
// This should be called by the action that handles the intervention.
func SetInterventionCooldown(playerState *state.ChurnState, now time.Time, cooldownDuration time.Duration) {
	state.SetInterventionCooldown(playerState, now, cooldownDuration)
}
