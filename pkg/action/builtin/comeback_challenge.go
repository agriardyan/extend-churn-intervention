package builtin

import (
	"context"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
	"github.com/sirupsen/logrus"
)

const (
	// ComebackChallengeActionID is the identifier for comeback challenge creation
	ComebackChallengeActionID = "comeback_challenge"

	// Default parameters
	DefaultWinsNeeded    = 3
	DefaultDurationDays  = 7
	DefaultCooldownHours = 48
)

// ComebackChallengeAction creates a comeback challenge for at-risk players.
// This action creates a time-limited challenge requiring a certain number of wins.
type ComebackChallengeAction struct {
	config        action.ActionConfig
	winsNeeded    int
	durationDays  int
	cooldownHours int
	stateStore    state.StateStore
}

// NewComebackChallengeAction creates a new comeback challenge action.
func NewComebackChallengeAction(config action.ActionConfig, stateStore state.StateStore) *ComebackChallengeAction {
	winsNeeded := config.GetParameterInt("wins_needed", DefaultWinsNeeded)
	durationDays := config.GetParameterInt("duration_days", DefaultDurationDays)
	cooldownHours := config.GetParameterInt("cooldown_hours", DefaultCooldownHours)

	logrus.Infof("creating comeback challenge action: winsNeeded=%d, durationDays=%d, cooldownHours=%d",
		winsNeeded, durationDays, cooldownHours)

	return &ComebackChallengeAction{
		config:        config,
		winsNeeded:    winsNeeded,
		durationDays:  durationDays,
		cooldownHours: cooldownHours,
		stateStore:    stateStore,
	}
}

// ID returns the action identifier.
func (a *ComebackChallengeAction) ID() string {
	return a.config.ID
}

// Name returns the action name.
func (a *ComebackChallengeAction) Name() string {
	return "Create Comeback Challenge"
}

// Config returns the action configuration.
func (a *ComebackChallengeAction) Config() action.ActionConfig {
	return a.config
}

// Execute creates a comeback challenge for the player.
func (a *ComebackChallengeAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	if playerCtx == nil || playerCtx.State == nil {
		return action.ErrMissingPlayerContext
	}

	now := time.Now()
	playerState := playerCtx.State

	// Check if challenge is already active
	if playerState.Challenge.Active {
		logrus.Warnf("challenge already active for user %s, skipping creation", trigger.UserID)
		return nil
	}

	// Get current wins (default to 0 if not available)
	currentWins := 0
	if winsAtStart, ok := trigger.Metadata["current_wins"].(int); ok {
		currentWins = winsAtStart
	}

	// Create challenge
	expiresAt := now.Add(time.Duration(a.durationDays) * 24 * time.Hour)
	state.CreateChallenge(playerState, a.winsNeeded, currentWins, expiresAt, trigger.RuleID)

	// Set intervention cooldown
	cooldownDuration := time.Duration(a.cooldownHours) * time.Hour
	state.SetInterventionCooldown(playerState, now, cooldownDuration)

	logrus.Infof("created comeback challenge for user %s: winsNeeded=%d, winsAtStart=%d, expiresAt=%v, reason=%s",
		trigger.UserID, a.winsNeeded, currentWins, expiresAt, trigger.RuleID)

	// Save updated state
	if a.stateStore != nil {
		if err := a.stateStore.Update(ctx, trigger.UserID, playerState); err != nil {
			logrus.Errorf("failed to save player state after challenge creation: %v", err)
			return err
		}
	}

	return nil
}

// Rollback removes the challenge (if possible).
func (a *ComebackChallengeAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	if playerCtx == nil || playerCtx.State == nil {
		return action.ErrMissingPlayerContext
	}

	playerState := playerCtx.State

	// Only rollback if the challenge was created by this trigger
	if playerState.Challenge.Active && playerState.Challenge.TriggerReason == trigger.RuleID {
		logrus.Infof("rolling back challenge for user %s (reason: %s)", trigger.UserID, trigger.RuleID)
		state.ResetChallenge(playerState)

		// Also reset intervention state
		playerState.Intervention.CooldownUntil = time.Time{}
		playerState.Intervention.TotalTriggered--

		// Save updated state
		if a.stateStore != nil {
			if err := a.stateStore.Update(ctx, trigger.UserID, playerState); err != nil {
				logrus.Errorf("failed to save player state after rollback: %v", err)
				return err
			}
		}

		return nil
	}

	logrus.Warnf("cannot rollback challenge for user %s: challenge not created by this trigger", trigger.UserID)
	return nil
}
