// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package state

import (
	"context"
	"encoding/json"
	"testing"
	"time"

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

func TestGetChurnState_NewPlayer(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "test-user-123"

	state, err := GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("GetChurnState() error = %v", err)
	}

	// Should return new state for non-existent player
	if state == nil {
		t.Fatal("GetChurnState() returned nil state")
	}

	// Check default values
	if state.Sessions.ThisWeek != 0 {
		t.Errorf("Sessions.ThisWeek = %d, expected 0", state.Sessions.ThisWeek)
	}
	if state.Sessions.LastWeek != 0 {
		t.Errorf("Sessions.LastWeek = %d, expected 0", state.Sessions.LastWeek)
	}
	if state.Challenge.Active {
		t.Error("Challenge.Active should be false for new player")
	}
	if state.Intervention.TotalTriggered != 0 {
		t.Errorf("Intervention.TotalTriggered = %d, expected 0", state.Intervention.TotalTriggered)
	}
}

func TestGetChurnState_ExistingPlayer(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "test-user-456"

	// Create a state and save it
	expectedState := &ChurnState{
		Sessions: SessionState{
			ThisWeek:  5,
			LastWeek:  3,
			LastReset: time.Now(),
		},
		Challenge: ChallengeState{
			Active:     true,
			WinsNeeded: 3,
		},
		Intervention: InterventionState{
			TotalTriggered: 2,
		},
	}

	// Manually insert into Redis
	data, _ := json.Marshal(expectedState)
	key := makeKey(userID)
	client.Set(ctx, key, data, DefaultTTL)

	// Retrieve state
	state, err := GetChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("GetChurnState() error = %v", err)
	}

	// Verify values
	if state.Sessions.ThisWeek != expectedState.Sessions.ThisWeek {
		t.Errorf("Sessions.ThisWeek = %d, expected %d",
			state.Sessions.ThisWeek, expectedState.Sessions.ThisWeek)
	}
	if state.Challenge.Active != expectedState.Challenge.Active {
		t.Errorf("Challenge.Active = %v, expected %v",
			state.Challenge.Active, expectedState.Challenge.Active)
	}
	if state.Intervention.TotalTriggered != expectedState.Intervention.TotalTriggered {
		t.Errorf("Intervention.TotalTriggered = %d, expected %d",
			state.Intervention.TotalTriggered, expectedState.Intervention.TotalTriggered)
	}
}

func TestUpdateChurnState(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "test-user-789"

	state := &ChurnState{
		Sessions: SessionState{
			ThisWeek:  10,
			LastWeek:  7,
			LastReset: time.Now(),
		},
		Challenge: ChallengeState{
			Active:     true,
			WinsNeeded: 5,
		},
		Intervention: InterventionState{
			TotalTriggered: 3,
		},
	}

	// Update state
	err := UpdateChurnState(ctx, client, userID, state)
	if err != nil {
		t.Fatalf("UpdateChurnState() error = %v", err)
	}

	// Verify it was saved
	key := makeKey(userID)
	data, err := client.Get(ctx, key).Result()
	if err != nil {
		t.Fatalf("failed to get key from Redis: %v", err)
	}

	var retrievedState ChurnState
	if err := json.Unmarshal([]byte(data), &retrievedState); err != nil {
		t.Fatalf("failed to unmarshal state: %v", err)
	}

	if retrievedState.Sessions.ThisWeek != state.Sessions.ThisWeek {
		t.Errorf("Sessions.ThisWeek = %d, expected %d",
			retrievedState.Sessions.ThisWeek, state.Sessions.ThisWeek)
	}
	if retrievedState.Challenge.WinsNeeded != state.Challenge.WinsNeeded {
		t.Errorf("Challenge.WinsNeeded = %d, expected %d",
			retrievedState.Challenge.WinsNeeded, state.Challenge.WinsNeeded)
	}
}

func TestDeleteChurnState(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "test-user-delete"

	// Create and save a state
	state := &ChurnState{
		Sessions: SessionState{
			ThisWeek: 5,
		},
	}
	UpdateChurnState(ctx, client, userID, state)

	// Verify it exists
	key := makeKey(userID)
	exists, _ := client.Exists(ctx, key).Result()
	if exists != 1 {
		t.Fatal("State should exist before deletion")
	}

	// Delete state
	err := DeleteChurnState(ctx, client, userID)
	if err != nil {
		t.Fatalf("DeleteChurnState() error = %v", err)
	}

	// Verify it's gone
	exists, _ = client.Exists(ctx, key).Result()
	if exists != 0 {
		t.Error("State should not exist after deletion")
	}
}

func TestMakeKey(t *testing.T) {
	userID := "test-user"
	expected := KeyPrefix + userID

	result := makeKey(userID)
	if result != expected {
		t.Errorf("makeKey() = %s, expected %s", result, expected)
	}
}

func TestUpdateChurnState_TTL(t *testing.T) {
	client, mr := setupTestRedis(t)
	defer mr.Close()

	ctx := context.Background()
	userID := "test-user-ttl"

	state := &ChurnState{
		Sessions: SessionState{
			ThisWeek: 1,
		},
	}

	// Update state
	err := UpdateChurnState(ctx, client, userID, state)
	if err != nil {
		t.Fatalf("UpdateChurnState() error = %v", err)
	}

	// Check TTL was set
	key := makeKey(userID)
	ttl, err := client.TTL(ctx, key).Result()
	if err != nil {
		t.Fatalf("failed to get TTL: %v", err)
	}

	// TTL should be approximately 30 days
	expectedTTL := DefaultTTL
	// Allow 1 second tolerance for test execution time
	if ttl < expectedTTL-time.Second || ttl > expectedTTL {
		t.Errorf("TTL = %v, expected approximately %v", ttl, expectedTTL)
	}
}
