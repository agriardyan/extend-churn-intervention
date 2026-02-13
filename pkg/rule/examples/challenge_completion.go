package examples

import (
	"context"
	"fmt"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalExamples "github.com/AccelByte/extends-anti-churn/pkg/signal/examples"
	"github.com/sirupsen/logrus"
)

const (
	// ChallengeCompletionRuleID is the identifier for challenge completion rule
	ChallengeCompletionRuleID = "challenge_completion"

	// DefaultWinsNeeded is the default number of wins needed to complete a challenge
	DefaultWinsNeeded = 3
)

// ChallengeCompletionRule detects when a player completes an active challenge.
// It monitors match wins and checks if the player has achieved the required wins
// for their active challenge.
type ChallengeCompletionRule struct {
	config      rule.RuleConfig
	winsNeeded  int
	challengeID string
}

// NewChallengeCompletionRule creates a new challenge completion detection rule.
func NewChallengeCompletionRule(config rule.RuleConfig) *ChallengeCompletionRule {
	winsNeeded := config.GetInt("wins_needed", DefaultWinsNeeded)
	challengeID := config.GetString("challenge_id", "")

	logrus.Infof("creating challenge completion rule: winsNeeded=%d, challengeID=%s",
		winsNeeded, challengeID)

	return &ChallengeCompletionRule{
		config:      config,
		winsNeeded:  winsNeeded,
		challengeID: challengeID,
	}
}

// ID returns the rule identifier.
func (r *ChallengeCompletionRule) ID() string {
	return r.config.ID
}

// Name returns the rule name.
func (r *ChallengeCompletionRule) Name() string {
	return "Challenge Completion Detection"
}

// SignalTypes returns the signal types this rule handles.
func (r *ChallengeCompletionRule) SignalTypes() []string {
	return []string{signalExamples.TypeMatchWin}
}

// Config returns the rule configuration.
func (r *ChallengeCompletionRule) Config() rule.RuleConfig {
	return r.config
}

// Evaluate checks if the player has an active challenge and has achieved the required wins.
func (r *ChallengeCompletionRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
	// Type assert to WinSignal
	winSig, ok := sig.(*signalExamples.WinSignal)
	if !ok {
		return false, nil, fmt.Errorf("expected WinSignal, got %T", sig)
	}

	// Get player context from signal
	playerCtx := winSig.Context()
	if playerCtx == nil || playerCtx.State == nil {
		logrus.Debugf("no player context available for user %s", winSig.UserID())
		return false, nil, nil
	}

	challenge := &playerCtx.State.Challenge

	// Check if challenge is active
	if !challenge.Active {
		logrus.Debugf("user %s has no active challenge", winSig.UserID())
		return false, nil, nil
	}

	// Check if challenge has expired
	if time.Now().After(challenge.ExpiresAt) {
		logrus.Debugf("challenge for user %s has expired", winSig.UserID())
		return false, nil, nil
	}

	// Calculate wins achieved since challenge started
	winsAchieved := winSig.TotalWins - challenge.WinsAtStart

	logrus.Debugf("evaluating challenge completion for user %s: winsAchieved=%d, winsNeeded=%d, totalWins=%d, winsAtStart=%d",
		winSig.UserID(), winsAchieved, challenge.WinsNeeded, winSig.TotalWins, challenge.WinsAtStart)

	// Check if player has achieved the required wins
	if winsAchieved >= challenge.WinsNeeded {
		trigger := rule.NewTrigger(r.ID(), sig.UserID(), "Challenge completed", r.config.Priority)
		trigger.Metadata["wins_achieved"] = winsAchieved
		trigger.Metadata["wins_needed"] = challenge.WinsNeeded
		trigger.Metadata["total_wins"] = winSig.TotalWins
		trigger.Metadata["wins_at_start"] = challenge.WinsAtStart
		trigger.Metadata["challenge_id"] = r.challengeID
		trigger.Metadata["trigger_reason"] = challenge.TriggerReason

		logrus.Infof("challenge completion rule triggered for user %s: winsAchieved=%d/%d",
			sig.UserID(), winsAchieved, challenge.WinsNeeded)

		return true, trigger, nil
	}

	logrus.Debugf("challenge not yet completed for user %s: %d/%d wins",
		winSig.UserID(), winsAchieved, challenge.WinsNeeded)

	return false, nil, nil
}
