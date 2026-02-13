package examples

import (
	"context"
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalExamples "github.com/AccelByte/extends-anti-churn/pkg/signal/examples"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

func TestChallengeCompletionRule_NoActiveChallenge(t *testing.T) {
	config := rule.RuleConfig{
		ID:      "test-challenge-completion",
		Type:    ChallengeCompletionRuleID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"wins_needed":  3,
			"challenge_id": "test-challenge",
		},
	}

	r := NewChallengeCompletionRule(config)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State: &state.ChurnState{
			Challenge: state.ChallengeState{
				Active: false, // No active challenge
			},
		},
	}

	sig := signalExamples.NewWinSignal("test-user", time.Now(), 10, playerCtx)

	matched, trigger, err := r.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matched {
		t.Error("expected no match when challenge is not active")
	}

	if trigger != nil {
		t.Error("expected nil trigger when challenge is not active")
	}
}

func TestChallengeCompletionRule_ChallengeExpired(t *testing.T) {
	config := rule.RuleConfig{
		ID:      "test-challenge-completion",
		Type:    ChallengeCompletionRuleID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"wins_needed":  3,
			"challenge_id": "test-challenge",
		},
	}

	r := NewChallengeCompletionRule(config)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State: &state.ChurnState{
			Challenge: state.ChallengeState{
				Active:      true,
				WinsNeeded:  3,
				WinsAtStart: 5,
				ExpiresAt:   time.Now().Add(-24 * time.Hour), // Expired yesterday
			},
		},
	}

	sig := signalExamples.NewWinSignal("test-user", time.Now(), 10, playerCtx)

	matched, trigger, err := r.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matched {
		t.Error("expected no match when challenge is expired")
	}

	if trigger != nil {
		t.Error("expected nil trigger when challenge is expired")
	}
}

func TestChallengeCompletionRule_ChallengeNotYetCompleted(t *testing.T) {
	config := rule.RuleConfig{
		ID:      "test-challenge-completion",
		Type:    ChallengeCompletionRuleID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"wins_needed":  3,
			"challenge_id": "test-challenge",
		},
	}

	r := NewChallengeCompletionRule(config)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State: &state.ChurnState{
			Challenge: state.ChallengeState{
				Active:      true,
				WinsNeeded:  3,
				WinsAtStart: 5, // Started with 5 wins
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
		},
	}

	// Player now has 7 wins total (5 at start + 2 new wins)
	// Need 3 wins, only achieved 2 so far
	sig := signalExamples.NewWinSignal("test-user", time.Now(), 7, playerCtx)

	matched, trigger, err := r.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matched {
		t.Error("expected no match when challenge is not yet completed")
	}

	if trigger != nil {
		t.Error("expected nil trigger when challenge is not yet completed")
	}
}

func TestChallengeCompletionRule_ChallengeCompleted(t *testing.T) {
	config := rule.RuleConfig{
		ID:      "test-challenge-completion",
		Type:    ChallengeCompletionRuleID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"wins_needed":  3,
			"challenge_id": "comeback-challenge",
		},
	}

	r := NewChallengeCompletionRule(config)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State: &state.ChurnState{
			Challenge: state.ChallengeState{
				Active:        true,
				WinsNeeded:    3,
				WinsAtStart:   5, // Started with 5 wins
				ExpiresAt:     time.Now().Add(7 * 24 * time.Hour),
				TriggerReason: "rage_quit",
			},
		},
	}

	// Player now has 8 wins total (5 at start + 3 new wins)
	// Exactly met the requirement
	sig := signalExamples.NewWinSignal("test-user", time.Now(), 8, playerCtx)

	matched, trigger, err := r.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !matched {
		t.Fatal("expected match when challenge is completed")
	}

	if trigger == nil {
		t.Fatal("expected trigger when challenge is completed")
	}

	if trigger.RuleID != r.ID() {
		t.Errorf("expected trigger rule ID %s, got %s", r.ID(), trigger.RuleID)
	}

	if trigger.UserID != "test-user" {
		t.Errorf("expected trigger user ID 'test-user', got %s", trigger.UserID)
	}

	if trigger.Metadata["wins_achieved"] != 3 {
		t.Errorf("expected wins_achieved 3, got %v", trigger.Metadata["wins_achieved"])
	}

	if trigger.Metadata["wins_needed"] != 3 {
		t.Errorf("expected wins_needed 3, got %v", trigger.Metadata["wins_needed"])
	}

	if trigger.Metadata["challenge_id"] != "comeback-challenge" {
		t.Errorf("expected challenge_id 'comeback-challenge', got %v", trigger.Metadata["challenge_id"])
	}
}

func TestChallengeCompletionRule_ChallengeExceeded(t *testing.T) {
	config := rule.RuleConfig{
		ID:      "test-challenge-completion",
		Type:    ChallengeCompletionRuleID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"wins_needed":  3,
			"challenge_id": "comeback-challenge",
		},
	}

	r := NewChallengeCompletionRule(config)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State: &state.ChurnState{
			Challenge: state.ChallengeState{
				Active:      true,
				WinsNeeded:  3,
				WinsAtStart: 5, // Started with 5 wins
				ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
			},
		},
	}

	// Player now has 10 wins total (5 at start + 5 new wins)
	// Exceeded the requirement (only needed 3)
	sig := signalExamples.NewWinSignal("test-user", time.Now(), 10, playerCtx)

	matched, trigger, err := r.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if !matched {
		t.Fatal("expected match when challenge is exceeded")
	}

	if trigger == nil {
		t.Fatal("expected trigger when challenge is exceeded")
	}

	if trigger.Metadata["wins_achieved"] != 5 {
		t.Errorf("expected wins_achieved 5, got %v", trigger.Metadata["wins_achieved"])
	}
}

func TestChallengeCompletionRule_NoPlayerContext(t *testing.T) {
	config := rule.RuleConfig{
		ID:      "test-challenge-completion",
		Type:    ChallengeCompletionRuleID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"wins_needed":  3,
			"challenge_id": "test-challenge",
		},
	}

	r := NewChallengeCompletionRule(config)

	// Signal without player context
	sig := signalExamples.NewWinSignal("test-user", time.Now(), 10, nil)

	matched, trigger, err := r.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if matched {
		t.Error("expected no match when player context is missing")
	}

	if trigger != nil {
		t.Error("expected nil trigger when player context is missing")
	}
}
