package handler

import (
	"context"
	"testing"
	"time"

	pb_iam "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
	"github.com/alicebob/miniredis/v2"
	"github.com/go-redis/redis/v8"
)

// setupTestRedis creates a miniredis instance for testing
func setupTestRedis(t *testing.T) (*redis.Client, *miniredis.Miniredis) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}

	client := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	return client, mr
}

func TestOAuth_OnMessage_NewPlayer(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewOAuth(client, "test-namespace")
	ctx := context.Background()

	msg := &pb_iam.OauthTokenGenerated{
		UserId:    "test-user-123",
		Namespace: "test-namespace",
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify state was created and session count incremented
	churnState, err := state.GetChurnState(ctx, client, "test-user-123")
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	if churnState.Sessions.ThisWeek != 1 {
		t.Errorf("Sessions.ThisWeek = %d, expected 1", churnState.Sessions.ThisWeek)
	}
}

func TestOAuth_OnMessage_ChurningPlayer(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewOAuth(client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-churning"

	// Create a churning player (was active last week, not this week, 3 days since reset)
	// They haven't logged in for 3 days, so ThisWeek is still 0
	churnState := &state.ChurnState{
		Sessions: state.SessionState{
			ThisWeek:  0,
			LastWeek:  5,
			LastReset: time.Now().Add(-3 * 24 * time.Hour),
		},
		Challenge: state.ChallengeState{
			Active: false,
		},
		Intervention: state.InterventionState{
			TotalTriggered: 0,
		},
	}
	state.UpdateChurnState(ctx, client, userID, churnState)

	msg := &pb_iam.OauthTokenGenerated{
		UserId:    userID,
		Namespace: "test-namespace",
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify session was incremented
	updatedState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	// After login, they're no longer churning (ThisWeek = 1)
	// So no intervention should be triggered yet (they need 7 days to be considered churning)
	// This test actually shows that OAuth login alone doesn't trigger intervention
	// Intervention is triggered by statistic events (rage quit, losing streak)
	if updatedState.Sessions.ThisWeek != 1 {
		t.Errorf("Sessions.ThisWeek = %d, expected 1", updatedState.Sessions.ThisWeek)
	}
}

func TestOAuth_OnMessage_WeeklyReset(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	listener := NewOAuth(client, "test-namespace")
	ctx := context.Background()

	userID := "test-user-reset"

	// Create player with old last reset date
	churnState := &state.ChurnState{
		Sessions: state.SessionState{
			ThisWeek:  5,
			LastWeek:  3,
			LastReset: time.Now().Add(-10 * 24 * time.Hour),
		},
	}
	state.UpdateChurnState(ctx, client, userID, churnState)

	msg := &pb_iam.OauthTokenGenerated{
		UserId:    userID,
		Namespace: "test-namespace",
	}

	_, err := listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Verify weekly reset occurred
	updatedState, err := state.GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("failed to get churn state: %v", err)
	}

	// After reset: lastWeek should be 5 (old thisWeek), thisWeek should be 1 (new session)
	if updatedState.Sessions.LastWeek != 5 {
		t.Errorf("Sessions.LastWeek = %d, expected 5", updatedState.Sessions.LastWeek)
	}
	if updatedState.Sessions.ThisWeek != 1 {
		t.Errorf("Sessions.ThisWeek = %d, expected 1", updatedState.Sessions.ThisWeek)
	}
}
