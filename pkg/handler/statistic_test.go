package handler

import (
	"context"
	"testing"
	"time"

	pb_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

func TestStatistic_OnMessage_RageQuit(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	// Create mock repos (nil is fine for tests that don't call AGS APIs)
	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-rage"

	// Pre-create player state without an active challenge
	// This simulates a player who has been playing for a while
	churnState := &state.ChurnState{
		Sessions: state.SessionState{
			ThisWeek:  2,
			LastWeek:  5,
			LastReset: time.Now().Add(-2 * 24 * time.Hour),
		},
		Challenge: state.ChallengeState{
			Active: false,
		},
		Intervention: state.InterventionState{},
	}
	state.UpdateChurnState(ctx, client, userID, churnState)

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeRageQuit,
			LatestValue: float64(RageQuitThreshold), // Exactly at threshold
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify intervention was triggered
	updatedState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if !updatedState.Challenge.Active {
		t.Error("Challenge should be active for rage quit threshold")
	}
	if updatedState.Challenge.TriggerReason != "rage_quit" {
		t.Errorf("TriggerReason = %s, expected 'rage_quit'",
			updatedState.Challenge.TriggerReason)
	}
}

func TestStatistic_OnMessage_RageQuitBelowThreshold(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-low-rage"

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeRageQuit,
			LatestValue: float64(RageQuitThreshold - 1), // Below threshold
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify intervention was NOT triggered
	churnState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if churnState.Challenge.Active {
		t.Error("Challenge should not be active below rage quit threshold")
	}
}

func TestStatistic_OnMessage_MatchWin_NoChallenge(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-win-no-challenge"

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeMatchWins,
			LatestValue: 10,
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Should not create a challenge, just ignore
	churnState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if churnState.Challenge.Active {
		t.Error("Challenge should not be active without prior trigger")
	}
}

func TestStatistic_OnMessage_MatchWin_ChallengeProgress(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-win-progress"

	// Create active challenge
	churnState := &state.ChurnState{
		Challenge: state.ChallengeState{
			Active:      true,
			WinsNeeded:  3,
			WinsCurrent: 0,
			WinsAtStart: 10,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		},
	}
	state.UpdateChurnState(ctx, client, userID, churnState)

	// Send win update
	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeMatchWins,
			LatestValue: 12, // 2 wins since challenge start
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify progress updated
	updatedState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if !updatedState.Challenge.Active {
		t.Error("Challenge should still be active")
	}
	if updatedState.Challenge.WinsCurrent != 2 {
		t.Errorf("WinsCurrent = %d, expected 2", updatedState.Challenge.WinsCurrent)
	}
}

func TestStatistic_OnMessage_MatchWin_ChallengeCompleted(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-win-complete"

	// Create active challenge
	churnState := &state.ChurnState{
		Challenge: state.ChallengeState{
			Active:      true,
			WinsNeeded:  3,
			WinsCurrent: 0,
			WinsAtStart: 10,
			ExpiresAt:   time.Now().Add(24 * time.Hour),
		},
	}
	state.UpdateChurnState(ctx, client, userID, churnState)

	// Send win update that completes challenge
	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeMatchWins,
			LatestValue: 13, // 3 wins since challenge start
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify challenge completed and reset
	updatedState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if updatedState.Challenge.Active {
		t.Error("Challenge should be completed and inactive")
	}
	// After completion, challenge data is reset to free Redis memory
	if updatedState.Challenge.WinsCurrent != 0 {
		t.Errorf("WinsCurrent = %d, expected 0 (reset after completion)", updatedState.Challenge.WinsCurrent)
	}
	if updatedState.Challenge.WinsNeeded != 0 {
		t.Errorf("WinsNeeded = %d, expected 0 (reset after completion)", updatedState.Challenge.WinsNeeded)
	}
	if updatedState.Challenge.TriggerReason != "" {
		t.Errorf("TriggerReason = %s, expected empty (reset after completion)", updatedState.Challenge.TriggerReason)
	}
}

func TestStatistic_OnMessage_LosingStreak(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-losing"

	// Pre-create player state without an active challenge
	churnState := &state.ChurnState{
		Sessions: state.SessionState{
			ThisWeek:  2,
			LastWeek:  5,
			LastReset: time.Now().Add(-2 * 24 * time.Hour),
		},
		Challenge: state.ChallengeState{
			Active: false,
		},
		Intervention: state.InterventionState{},
	}
	state.UpdateChurnState(ctx, client, userID, churnState)

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeLosingStreak,
			LatestValue: float64(LosingStreakThreshold), // Exactly at threshold
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify intervention was triggered
	updatedState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if !updatedState.Challenge.Active {
		t.Error("Challenge should be active for losing streak threshold")
	}
	if updatedState.Challenge.TriggerReason != "losing_streak" {
		t.Errorf("TriggerReason = %s, expected 'losing_streak'",
			updatedState.Challenge.TriggerReason)
	}
}

func TestStatistic_OnMessage_LosingStreakBelowThreshold(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-low-losing"

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeLosingStreak,
			LatestValue: float64(LosingStreakThreshold - 1), // Below threshold
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify intervention was NOT triggered
	churnState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if churnState.Challenge.Active {
		t.Error("Challenge should not be active below losing streak threshold")
	}
}

func TestStatistic_OnMessage_InterventionCooldown(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-cooldown"

	// Create state with active cooldown
	churnState := &state.ChurnState{
		Intervention: state.InterventionState{
			CooldownUntil:  time.Now().Add(24 * time.Hour),
			TotalTriggered: 1,
		},
	}
	state.UpdateChurnState(ctx, client, userID, churnState)

	// Try to trigger intervention via rage quit
	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    StatCodeRageQuit,
			LatestValue: float64(RageQuitThreshold),
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify intervention was blocked by cooldown
	updatedState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if updatedState.Challenge.Active {
		t.Error("Challenge should not be active during cooldown")
	}
	if updatedState.Intervention.TotalTriggered != 1 {
		t.Errorf("TotalTriggered should remain 1, got %d",
			updatedState.Intervention.TotalTriggered)
	}
}

func TestStatistic_OnMessage_IgnoreUnknownStatCode(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewStatistic(nil, nil, client, "test-namespace")
	ctx := context.Background()

	msg := &pb_social.StatItemUpdated{
		UserId:    "test-user-unknown",
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    "unknown-stat-code",
			LatestValue: 100,
		},
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() should not error on unknown stat code: %v", err)
	}
}
