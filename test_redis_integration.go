// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

//go:build integration
// +build integration

package main

import (
	"context"
	"fmt"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/state"
	"github.com/sirupsen/logrus"
)

// This is a manual integration test for Redis operations
// Run this with: go run test_redis_integration.go
// Requires: Redis running on localhost:6379

func main() {
	logrus.SetLevel(logrus.DebugLevel)
	logrus.Infof("Starting Redis integration test...")

	ctx := context.Background()

	// Initialize Redis client
	client, err := state.InitRedisClient(ctx)
	if err != nil {
		logrus.Fatalf("Failed to initialize Redis: %v", err)
	}
	defer client.Close()

	testUserID := fmt.Sprintf("test-user-%d", time.Now().Unix())
	logrus.Infof("Testing with user ID: %s", testUserID)

	// Test 1: Get state for new player
	logrus.Infof("\n=== Test 1: Get state for new player ===")
	state1, err := state.GetChurnState(ctx, client, testUserID)
	if err != nil {
		logrus.Fatalf("GetChurnState failed: %v", err)
	}
	logrus.Infof("✓ Got new player state: Sessions=%+v", state1.Sessions)

	// Test 2: Update state
	logrus.Infof("\n=== Test 2: Update player state ===")
	state1.Sessions.ThisWeek = 5
	state1.Sessions.LastWeek = 3
	state1.Challenge.Active = true
	state1.Challenge.WinsNeeded = 3
	state1.Intervention.TotalTriggered = 1

	err = state.UpdateChurnState(ctx, client, testUserID, state1)
	if err != nil {
		logrus.Fatalf("UpdateChurnState failed: %v", err)
	}
	logrus.Infof("✓ Updated player state")

	// Test 3: Retrieve updated state
	logrus.Infof("\n=== Test 3: Retrieve updated state ===")
	state2, err := state.GetChurnState(ctx, client, testUserID)
	if err != nil {
		logrus.Fatalf("GetChurnState failed: %v", err)
	}
	logrus.Infof("✓ Retrieved state: Sessions.ThisWeek=%d, Challenge.Active=%v, Intervention.TotalTriggered=%d",
		state2.Sessions.ThisWeek, state2.Challenge.Active, state2.Intervention.TotalTriggered)

	// Verify values match
	if state2.Sessions.ThisWeek != 5 {
		logrus.Fatalf("❌ ThisWeek mismatch: got %d, expected 5", state2.Sessions.ThisWeek)
	}
	if !state2.Challenge.Active {
		logrus.Fatalf("❌ Challenge should be active")
	}

	// Test 4: Test weekly reset logic
	logrus.Infof("\n=== Test 4: Test weekly reset logic ===")
	state2.Sessions.LastReset = time.Now().Add(-8 * 24 * time.Hour) // 8 days ago
	resetOccurred := state.CheckWeeklyReset(state2, time.Now())
	if !resetOccurred {
		logrus.Fatalf("❌ Weekly reset should have occurred")
	}
	logrus.Infof("✓ Weekly reset occurred: ThisWeek=%d, LastWeek=%d",
		state2.Sessions.ThisWeek, state2.Sessions.LastWeek)

	// Test 5: Test churn detection
	logrus.Infof("\n=== Test 5: Test churn detection ===")
	state2.Sessions.ThisWeek = 0
	state2.Sessions.LastWeek = 5
	state2.Sessions.LastReset = time.Now().Add(-8 * 24 * time.Hour)
	isChurning := state.IsChurning(state2, time.Now())
	if !isChurning {
		logrus.Fatalf("❌ Player should be detected as churning")
	}
	logrus.Infof("✓ Player correctly identified as churning")

	// Test 6: Test intervention trigger logic
	logrus.Infof("\n=== Test 6: Test intervention trigger logic ===")
	state2.Challenge.Active = false
	state2.Intervention.CooldownUntil = time.Time{} // No cooldown
	shouldTrigger := state.ShouldTriggerIntervention(state2, time.Now())
	if !shouldTrigger {
		logrus.Fatalf("❌ Intervention should be triggered")
	}
	logrus.Infof("✓ Intervention correctly triggered")

	// Test 7: Test cooldown logic
	logrus.Infof("\n=== Test 7: Test cooldown logic ===")
	cooldownDuration := 48 * time.Hour
	state.SetInterventionCooldown(state2, time.Now(), cooldownDuration)
	canTrigger := state.CanTriggerIntervention(state2, time.Now())
	if canTrigger {
		logrus.Fatalf("❌ Intervention should be blocked by cooldown")
	}
	logrus.Infof("✓ Cooldown correctly prevents intervention")

	// Test 8: Test challenge creation and progress
	logrus.Infof("\n=== Test 8: Test challenge creation and progress ===")
	state.CreateChallenge(state2, 3, 10, time.Now().Add(7*24*time.Hour), "churn_detected")
	if !state2.Challenge.Active {
		logrus.Fatalf("❌ Challenge should be active")
	}
	logrus.Infof("✓ Challenge created successfully")

	// Update challenge progress
	completed := state.UpdateChallengeProgress(state2, 12, time.Now()) // 2 wins
	if completed {
		logrus.Fatalf("❌ Challenge should not be completed yet")
	}
	logrus.Infof("✓ Challenge progress updated: %d/%d wins",
		state2.Challenge.WinsCurrent, state2.Challenge.WinsNeeded)

	completed = state.UpdateChallengeProgress(state2, 13, time.Now()) // 3 wins
	if !completed {
		logrus.Fatalf("❌ Challenge should be completed")
	}
	logrus.Infof("✓ Challenge completed!")

	// Test 9: Clean up - delete state
	logrus.Infof("\n=== Test 9: Clean up ===")
	err = state.DeleteChurnState(ctx, client, testUserID)
	if err != nil {
		logrus.Fatalf("DeleteChurnState failed: %v", err)
	}
	logrus.Infof("✓ Deleted test player state")

	// Verify deletion
	state3, err := state.GetChurnState(ctx, client, testUserID)
	if err != nil {
		logrus.Fatalf("GetChurnState after delete failed: %v", err)
	}
	if state3.Sessions.ThisWeek != 0 || state3.Sessions.LastWeek != 0 {
		logrus.Fatalf("❌ State should be reset after deletion")
	}
	logrus.Infof("✓ Verified state was deleted (got new state)")

	logrus.Infof("\n==================================================")
	logrus.Infof("✅ All Redis integration tests passed!")
	logrus.Infof("==================================================")
}
