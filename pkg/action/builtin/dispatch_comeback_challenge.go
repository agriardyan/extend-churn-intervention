package builtin

import (
	"context"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/sirupsen/logrus"
)

const (
	// ComebackChallengeActionID is the identifier for comeback challenge creation
	ComebackChallengeActionID = "dispatch_comeback_challenge"

	// Default parameters
	DefaultWinsNeeded    = 3
	DefaultDurationDays  = 7
	DefaultCooldownHours = 48
)

// DispatchComebackChallengeAction creates a comeback challenge for at-risk players.
// This action creates a time-limited challenge requiring a certain number of wins.
type DispatchComebackChallengeAction struct {
	config          action.ActionConfig
	winsNeeded      int
	durationDays    int
	cooldownHours   int
	stateStore      service.StateStore
	userStatUpdater service.UserStatisticUpdater
}

// NewDispatchComebackChallengeAction creates a new comeback challenge action.
func NewDispatchComebackChallengeAction(
	config action.ActionConfig,
	stateStore service.StateStore,
	userStatUpdater service.UserStatisticUpdater,
) *DispatchComebackChallengeAction {
	winsNeeded := config.GetParameterInt("wins_needed", DefaultWinsNeeded)
	durationDays := config.GetParameterInt("duration_days", DefaultDurationDays)
	cooldownHours := config.GetParameterInt("cooldown_hours", DefaultCooldownHours)

	logrus.Infof("creating comeback challenge action: winsNeeded=%d, durationDays=%d, cooldownHours=%d",
		winsNeeded, durationDays, cooldownHours)

	return &DispatchComebackChallengeAction{
		config:          config,
		winsNeeded:      winsNeeded,
		durationDays:    durationDays,
		cooldownHours:   cooldownHours,
		stateStore:      stateStore,
		userStatUpdater: userStatUpdater,
	}
}

// ID returns the action identifier.
func (a *DispatchComebackChallengeAction) ID() string {
	return a.config.ID
}

// Name returns the action name.
func (a *DispatchComebackChallengeAction) Name() string {
	return "Create Comeback Challenge"
}

// Config returns the action configuration.
func (a *DispatchComebackChallengeAction) Config() action.ActionConfig {
	return a.config
}

// Execute creates a comeback challenge intervention for the player.
func (a *DispatchComebackChallengeAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	if playerCtx == nil || playerCtx.State == nil {
		return action.ErrMissingPlayerContext
	}

	now := time.Now()
	playerState := playerCtx.State

	// Check if we're on cooldown
	if playerState.Cooldown.IsOnCooldown() {
		logrus.Warnf("intervention on cooldown for user %s, skipping", trigger.UserID)
		return nil
	}

	// Check if there's already an active comeback challenge intervention
	activeInterventions := playerState.GetActiveInterventions()
	for _, intervention := range activeInterventions {
		if intervention.Type == ComebackChallengeActionID {
			logrus.Warnf("comeback challenge already active for user %s, skipping creation", trigger.UserID)
			return nil
		}
	}

	// Calculate expiration time
	expiresAt := now.Add(time.Duration(a.durationDays) * 24 * time.Hour)

	// Record the intervention
	interventionID := trigger.UserID + "-comeback-" + now.Format("20060102150405")
	metadata := map[string]interface{}{
		"wins_needed":     a.winsNeeded,
		"duration_days":   a.durationDays,
		"trigger_rule_id": trigger.RuleID,
	}

	playerState.AddIntervention(interventionID, ComebackChallengeActionID, trigger.RuleID, &expiresAt, metadata)

	// Set cooldown
	cooldownDuration := time.Duration(a.cooldownHours) * time.Hour
	playerState.Cooldown.CooldownUntil = now.Add(cooldownDuration)

	logrus.Infof("created comeback challenge intervention for user %s: id=%s, winsNeeded=%d, expiresAt=%v, reason=%s",
		trigger.UserID, interventionID, a.winsNeeded, expiresAt, trigger.RuleID)

	// Save updated state
	if a.stateStore != nil {
		err := a.stateStore.UpdateChurnState(ctx, trigger.UserID, playerState)
		if err != nil {
			logrus.Errorf("failed to save player state after intervention: %v", err)
			return err
		}
	}

	// This example action utilize extend-challenge-event-handler, which listens for a specific stat update to trigger challenge for a user.
	// The stat `rse-comeback-challenge` will be updated and listened by the extend-challenge-event-handler to trigger the challenge for the user.
	err := a.userStatUpdater.UpdateStatComebackChallenge(ctx, trigger.UserID)
	if err != nil {
		logrus.Errorf("failed to update user statistics: %v", err)
		return err
	}

	return nil
}

// Rollback marks the intervention as failed (if possible).
func (a *DispatchComebackChallengeAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	if playerCtx == nil || playerCtx.State == nil {
		return action.ErrMissingPlayerContext
	}

	playerState := playerCtx.State

	// Find active comeback challenge interventions triggered by this rule
	activeInterventions := playerState.GetActiveInterventions()
	for _, intervention := range activeInterventions {
		if intervention.Type == ComebackChallengeActionID && intervention.TriggeredBy == trigger.RuleID {
			logrus.Infof("rolling back intervention %s for user %s (reason: %s)", intervention.ID, trigger.UserID, trigger.RuleID)

			// Mark intervention as failed
			playerState.UpdateInterventionOutcome(intervention.ID, "failed")

			// Reset cooldown
			playerState.Cooldown.CooldownUntil = time.Time{}

			// TODO revert the challenge creation if possible
			// (e.g., by sending a cancellation event to the challenge service)

			// Save updated state
			if a.stateStore != nil {
				err := a.stateStore.UpdateChurnState(ctx, trigger.UserID, playerState)
				if err != nil {
					logrus.Errorf("failed to save player state after rollback: %v", err)
					return err
				}
			}

			return nil
		}
	}

	logrus.Warnf("cannot rollback intervention for user %s: no active intervention created by this trigger", trigger.UserID)
	return nil
}
