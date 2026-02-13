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

func TestRageQuitRule_Evaluate(t *testing.T) {
	tests := []struct {
		name           string
		threshold      int
		statValue      float64
		expectTrigger  bool
		expectRageQuit int
	}{
		{
			name:           "rage quit at threshold",
			threshold:      3,
			statValue:      3,
			expectTrigger:  true,
			expectRageQuit: 3,
		},
		{
			name:           "rage quit above threshold",
			threshold:      3,
			statValue:      5,
			expectTrigger:  true,
			expectRageQuit: 5,
		},
		{
			name:          "rage quit below threshold",
			threshold:     3,
			statValue:     2,
			expectTrigger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := rule.RuleConfig{
				ID:       "test_rage_quit",
				Type:     RageQuitRuleID,
				Enabled:  true,
				Priority: 10,
				Parameters: map[string]interface{}{
					"threshold": tt.threshold,
				},
			}

			rule := NewRageQuitRule(config)

			playerCtx := &signal.PlayerContext{
				UserID: "test-user",
				State:  &state.ChurnState{},
			}
			sig := signalExamples.NewRageQuitSignal("test-user", time.Now(), int(tt.statValue), playerCtx)

			matched, trigger, err := rule.Evaluate(context.Background(), sig)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if matched != tt.expectTrigger {
				t.Errorf("Expected matched=%v, got %v", tt.expectTrigger, matched)
			}

			if tt.expectTrigger {
				if trigger == nil {
					t.Fatal("Expected trigger, got nil")
				}
				if trigger.RuleID != config.ID {
					t.Errorf("Expected rule ID '%s', got '%s'", config.ID, trigger.RuleID)
				}
				if trigger.Metadata["rage_quit_count"] != tt.expectRageQuit {
					t.Errorf("Expected rage_quit_count=%v, got %v", tt.expectRageQuit, trigger.Metadata["rage_quit_count"])
				}
			} else {
				if trigger != nil {
					t.Error("Expected no trigger, got one")
				}
			}
		})
	}
}

func TestRageQuitRule_WrongSignalType(t *testing.T) {
	config := rule.RuleConfig{
		ID:       "test_rage_quit",
		Type:     RageQuitRuleID,
		Enabled:  true,
		Priority: 10,
	}

	rule := NewRageQuitRule(config)

	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}
	sig := signalExamples.NewOauthTokenGeneratedSignal("test-user", time.Now(), playerCtx)

	matched, trigger, err := rule.Evaluate(context.Background(), sig)
	if err == nil {
		t.Error("Expected error for wrong signal type")
	}
	if matched {
		t.Error("Expected no match for wrong signal type")
	}
	if trigger != nil {
		t.Error("Expected no trigger for wrong signal type")
	}
}

func TestLosingStreakRule_Evaluate(t *testing.T) {
	tests := []struct {
		name          string
		threshold     int
		statValue     float64
		expectTrigger bool
		expectStreak  int
	}{
		{
			name:          "losing streak at threshold",
			threshold:     5,
			statValue:     5,
			expectTrigger: true,
			expectStreak:  5,
		},
		{
			name:          "losing streak above threshold",
			threshold:     5,
			statValue:     7,
			expectTrigger: true,
			expectStreak:  7,
		},
		{
			name:          "losing streak below threshold",
			threshold:     5,
			statValue:     3,
			expectTrigger: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := rule.RuleConfig{
				ID:       "test_losing_streak",
				Type:     LosingStreakRuleID,
				Enabled:  true,
				Priority: 10,
				Parameters: map[string]interface{}{
					"threshold": tt.threshold,
				},
			}

			rule := NewLosingStreakRule(config)

			playerCtx := &signal.PlayerContext{
				UserID: "test-user",
				State:  &state.ChurnState{},
			}
			sig := signalExamples.NewLosingStreakSignal("test-user", time.Now(), int(tt.statValue), playerCtx)

			matched, trigger, err := rule.Evaluate(context.Background(), sig)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if matched != tt.expectTrigger {
				t.Errorf("Expected matched=%v, got %v", tt.expectTrigger, matched)
			}

			if tt.expectTrigger {
				if trigger == nil {
					t.Fatal("Expected trigger, got nil")
				}
				if trigger.RuleID != config.ID {
					t.Errorf("Expected rule ID '%s', got '%s'", config.ID, trigger.RuleID)
				}
				if trigger.Metadata["losing_streak"] != tt.expectStreak {
					t.Errorf("Expected losing_streak=%v, got %v", tt.expectStreak, trigger.Metadata["losing_streak"])
				}
			} else {
				if trigger != nil {
					t.Error("Expected no trigger, got one")
				}
			}
		})
	}
}

func TestSessionDeclineRule_Evaluate(t *testing.T) {
	now := time.Now()
	eightDaysAgo := now.Add(-8 * 24 * time.Hour)

	tests := []struct {
		name              string
		sessionState      state.SessionState
		interventionState state.InterventionState
		challengeActive   bool
		expectTrigger     bool
	}{
		{
			name: "session decline detected",
			sessionState: state.SessionState{
				LastWeek:  5,
				ThisWeek:  0,
				LastReset: eightDaysAgo,
			},
			interventionState: state.InterventionState{
				CooldownUntil: time.Time{}, // No cooldown
			},
			challengeActive: false,
			expectTrigger:   true,
		},
		{
			name: "no decline - active this week",
			sessionState: state.SessionState{
				LastWeek:  5,
				ThisWeek:  3,
				LastReset: eightDaysAgo,
			},
			interventionState: state.InterventionState{},
			challengeActive:   false,
			expectTrigger:     false,
		},
		{
			name: "no decline - inactive last week",
			sessionState: state.SessionState{
				LastWeek:  0,
				ThisWeek:  0,
				LastReset: eightDaysAgo,
			},
			interventionState: state.InterventionState{},
			challengeActive:   false,
			expectTrigger:     false,
		},
		{
			name: "decline but in cooldown",
			sessionState: state.SessionState{
				LastWeek:  5,
				ThisWeek:  0,
				LastReset: eightDaysAgo,
			},
			interventionState: state.InterventionState{
				CooldownUntil: now.Add(24 * time.Hour), // Still in cooldown
			},
			challengeActive: false,
			expectTrigger:   false,
		},
		{
			name: "decline but challenge active",
			sessionState: state.SessionState{
				LastWeek:  5,
				ThisWeek:  0,
				LastReset: eightDaysAgo,
			},
			interventionState: state.InterventionState{},
			challengeActive:   true,
			expectTrigger:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := rule.RuleConfig{
				ID:       "test_session_decline",
				Type:     SessionDeclineRuleID,
				Enabled:  true,
				Priority: 10,
			}

			rule := NewSessionDeclineRule(config)

			playerState := &state.ChurnState{
				Sessions:     tt.sessionState,
				Intervention: tt.interventionState,
				Challenge: state.ChallengeState{
					Active: tt.challengeActive,
				},
			}

			playerCtx := &signal.PlayerContext{
				UserID: "test-user",
				State:  playerState,
			}
			sig := signalExamples.NewOauthTokenGeneratedSignal("test-user", now, playerCtx)

			matched, trigger, err := rule.Evaluate(context.Background(), sig)
			if err != nil {
				t.Fatalf("Unexpected error: %v", err)
			}

			if matched != tt.expectTrigger {
				t.Errorf("Expected matched=%v, got %v", tt.expectTrigger, matched)
			}

			if tt.expectTrigger {
				if trigger == nil {
					t.Fatal("Expected trigger, got nil")
				}
				if trigger.RuleID != config.ID {
					t.Errorf("Expected rule ID '%s', got '%s'", config.ID, trigger.RuleID)
				}
			} else {
				if trigger != nil {
					t.Error("Expected no trigger, got one")
				}
			}
		})
	}
}

func TestSessionDeclineRule_NoPlayerContext(t *testing.T) {
	config := rule.RuleConfig{
		ID:       "test_session_decline",
		Type:     SessionDeclineRuleID,
		Enabled:  true,
		Priority: 10,
	}

	rule := NewSessionDeclineRule(config)

	// Create signal without player context
	sig := signalExamples.NewOauthTokenGeneratedSignal("test-user", time.Now(), nil)

	matched, trigger, err := rule.Evaluate(context.Background(), sig)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if matched {
		t.Error("Expected no match without player context")
	}
	if trigger != nil {
		t.Error("Expected no trigger without player context")
	}
}
