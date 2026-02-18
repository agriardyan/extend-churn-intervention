package builtin

import (
	"context"
	"fmt"
	"time"

	"github.com/AccelByte/extend-churn-intervention/pkg/rule"
	"github.com/AccelByte/extend-churn-intervention/pkg/service"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
	signalBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/signal/builtin"
	"github.com/sirupsen/logrus"
)

const (
	// SessionDeclineRuleID is the identifier for session decline detection rule
	SessionDeclineRuleID = "session_decline"
)

// SessionDeclineRule detects when a player's session frequency declines week-over-week.
// This rule uses LoginSessionTracker to access session tracking data.
type SessionDeclineRule struct {
	config         rule.RuleConfig
	sessionTracker service.LoginSessionTracker
}

// NewSessionDeclineRule creates a new session decline detection rule.
func NewSessionDeclineRule(config rule.RuleConfig, sessionTracker service.LoginSessionTracker) *SessionDeclineRule {
	logrus.Infof("creating session decline rule")
	return &SessionDeclineRule{
		config:         config,
		sessionTracker: sessionTracker,
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
	// Get player context (for cooldown and intervention checks)
	playerCtx := sig.Context()
	if playerCtx == nil || playerCtx.State == nil {
		logrus.Debugf("no player context for user %s, skipping session decline check", sig.UserID())
		return false, nil, nil
	}

	now := sig.Timestamp()
	playerState := playerCtx.State

	// Load session tracking data from session tracker service
	sessionData, err := r.sessionTracker.GetSessionData(ctx, sig.UserID())
	if err != nil {
		logrus.Errorf("failed to load session data for user %s: %v", sig.UserID(), err)
		return false, nil, err
	}

	// Check if player is churning using the map-based data
	churning := isChurning(sessionData, now)
	if !churning {
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

	// Get current and previous week for metadata
	currentWeek := getYearWeek(now)
	previousWeek := getYearWeek(now.Add(-7 * 24 * time.Hour))

	trigger := rule.NewTrigger(r.ID(), sig.UserID(), "Session frequency declined", r.config.Priority)
	trigger.Metadata["current_week"] = currentWeek
	trigger.Metadata["previous_week"] = previousWeek
	trigger.Metadata["current_week_sessions"] = sessionData.LoginCount[currentWeek]
	trigger.Metadata["previous_week_sessions"] = sessionData.LoginCount[previousWeek]

	logrus.Infof("session decline rule triggered for user %s: currentWeek=%s (%d), previousWeek=%s (%d)",
		sig.UserID(), currentWeek, sessionData.LoginCount[currentWeek], previousWeek, sessionData.LoginCount[previousWeek])

	return true, trigger, nil
}

// getYearWeek returns the year-week string in format "YYYYWW" (e.g., "202610" for week 10 of 2026)
func getYearWeek(t time.Time) string {
	year, week := t.ISOWeek()
	return fmt.Sprintf("%04d%02d", year, week)
}

// isChurning determines if a player is exhibiting churn behavior using map-based data.
// A player is churning if:
// - They had activity in a recent past week (within last 2 weeks)
// - They have no activity in the current week (or very low activity)
//
// This approach handles multi-week absences naturally by checking the map for any recent activity.
func isChurning(data *service.SessionTrackingData, now time.Time) bool {
	if len(data.LoginCount) == 0 {
		return false // No data, can't determine churn
	}

	currentWeek := getYearWeek(now)
	previousWeek := getYearWeek(now.Add(-7 * 24 * time.Hour))

	currentCount := data.LoginCount[currentWeek]
	previousCount := data.LoginCount[previousWeek]

	// Check if they had activity last week but not this week
	if previousCount > 0 && currentCount == 0 {
		logrus.Infof("player is churning: previousWeek=%s (%d), currentWeek=%s (%d)",
			previousWeek, previousCount, currentWeek, currentCount)
		return true
	}

	// Alternative: Check if they had activity in any recent week but none currently
	// This handles multi-week absences where they might have been gone for 2+ weeks
	if currentCount == 0 {
		// Check if they had any activity in the past 2 weeks
		for week, count := range data.LoginCount {
			if week != currentWeek && count > 0 {
				logrus.Infof("player is churning: had activity in week %s (%d), none in currentWeek=%s",
					week, count, currentWeek)
				return true
			}
		}
	}

	return false
}
