package builtin

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/AccelByte/extend-churn-intervention/pkg/rule"
	"github.com/AccelByte/extend-churn-intervention/pkg/service"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
	signalBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/signal/builtin"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
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
				State:  &service.ChurnState{},
			}
			sig := signalBuiltin.NewRageQuitSignal("test-user", time.Now(), int(tt.statValue), playerCtx)

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
		State:  &service.ChurnState{},
	}
	sig := signalBuiltin.NewLoginSignal("test-user", time.Now(), playerCtx)

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
				State:  &service.ChurnState{},
			}
			sig := signalBuiltin.NewLosingStreakSignal("test-user", time.Now(), int(tt.statValue), playerCtx)

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
	expiresAt := now.Add(7 * 24 * time.Hour)

	// Helper to get yearWeek format
	getYearWeek := func(t time.Time) string {
		year, week := t.ISOWeek()
		return fmt.Sprintf("%04d%02d", year, week)
	}

	currentWeek := getYearWeek(now)
	lastWeek := getYearWeek(now.Add(-7 * 24 * time.Hour))
	twoWeeksAgo := getYearWeek(now.Add(-14 * 24 * time.Hour))
	threeWeeksAgo := getYearWeek(now.Add(-21 * 24 * time.Hour))
	fourWeeksAgo := getYearWeek(now.Add(-28 * 24 * time.Hour))

	tests := []struct {
		name                string
		loginCountData      map[string]int // yearWeek -> count
		cooldownState       service.CooldownState
		interventionHistory []service.InterventionRecord
		expectTrigger       bool
	}{
		{
			name: "session decline detected",
			loginCountData: map[string]int{
				lastWeek: 5, // Active last week
				// currentWeek: 0 (implicit - no entry means 0)
			},
			cooldownState: service.CooldownState{
				CooldownUntil: time.Time{}, // No cooldown
			},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       true,
		},
		{
			name: "no decline - active this week",
			loginCountData: map[string]int{
				lastWeek:    5, // Active last week
				currentWeek: 3, // Still active this week
			},
			cooldownState:       service.CooldownState{},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       false,
		},
		{
			name:           "no decline - inactive last week",
			loginCountData: map[string]int{
				// No entries - never active
			},
			cooldownState:       service.CooldownState{},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       false,
		},
		{
			name: "decline but in cooldown",
			loginCountData: map[string]int{
				lastWeek: 5, // Active last week
				// currentWeek: 0 (no activity)
			},
			cooldownState: service.CooldownState{
				CooldownUntil: now.Add(24 * time.Hour), // Still in cooldown
			},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       false,
		},
		{
			name: "decline but comeback challenge active",
			loginCountData: map[string]int{
				lastWeek: 5, // Active last week
				// currentWeek: 0 (no activity)
			},
			cooldownState: service.CooldownState{},
			interventionHistory: []service.InterventionRecord{
				{
					ID:          "existing-challenge",
					Type:        "dispatch_comeback_challenge",
					TriggeredBy: "previous-rule",
					TriggeredAt: now,
					ExpiresAt:   &expiresAt,
					Outcome:     "active",
				},
			},
			expectTrigger: false,
		},
		{
			name: "multi-week absence - had activity 2 weeks ago",
			loginCountData: map[string]int{
				twoWeeksAgo: 7, // Active 2 weeks ago
				// lastWeek: 0, currentWeek: 0 (absent for 2 weeks)
			},
			cooldownState:       service.CooldownState{},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       true, // Should detect churn from any recent week
		},
		{
			name: "3-week absence - vacation/busy period",
			loginCountData: map[string]int{
				threeWeeksAgo: 15, // Very active 3 weeks ago
				// Absent for 3 weeks (vacation, busy at work, etc.)
			},
			cooldownState:       service.CooldownState{},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       true, // 4-week retention catches this!
		},
		{
			name: "4-week absence - at retention boundary",
			loginCountData: map[string]int{
				fourWeeksAgo: 20, // Very active 4 weeks ago
				// Absent for 4 weeks
			},
			cooldownState:       service.CooldownState{},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       true, // Still within 4-week retention window
		},
		{
			name: "gradual decline over 3 weeks",
			loginCountData: map[string]int{
				threeWeeksAgo: 10, // High activity
				twoWeeksAgo:   5,  // Declining
				lastWeek:      2,  // Very low
				// currentWeek: 0 (churned)
			},
			cooldownState:       service.CooldownState{},
			interventionHistory: []service.InterventionRecord{},
			expectTrigger:       true, // Detect gradual churn pattern
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

			mr, _ := miniredis.Run()
			defer mr.Close()
			redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
			defer redisClient.Close()

			// Create login session tracker
			sessionTracker := service.NewRedisLoginSessionTrackingStore(redisClient, service.RedisLoginSessionTrackingStoreConfig{})

			// Set up session data in Redis using map structure
			if len(tt.loginCountData) > 0 {
				sessionData := &service.SessionTrackingData{
					LoginCount: tt.loginCountData,
				}
				sessionTracker.SaveSessionData(context.Background(), "test-user", sessionData)
			}

			rule := NewSessionDeclineRule(config, sessionTracker)

			playerState := &service.ChurnState{
				Cooldown:            tt.cooldownState,
				InterventionHistory: tt.interventionHistory,
				SignalHistory:       []service.ChurnSignal{},
			}

			playerCtx := &signal.PlayerContext{
				UserID: "test-user",
				State:  playerState,
			}
			sig := signalBuiltin.NewLoginSignal("test-user", now, playerCtx)

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

	mr, _ := miniredis.Run()
	defer mr.Close()
	redisClient := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	defer redisClient.Close()
	sessionTracker := service.NewRedisLoginSessionTrackingStore(redisClient, service.RedisLoginSessionTrackingStoreConfig{})
	rule := NewSessionDeclineRule(config, sessionTracker)

	// Create signal without player context
	sig := signalBuiltin.NewLoginSignal("test-user", time.Now(), nil)

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
