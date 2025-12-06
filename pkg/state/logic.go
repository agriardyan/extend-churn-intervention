// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package state

import (
	"time"

	"github.com/sirupsen/logrus"
)

// CheckWeeklyReset checks if a weekly reset should occur and performs it if needed
// Returns true if a reset occurred, false otherwise
func CheckWeeklyReset(state *ChurnState, now time.Time) bool {
	// Calculate time since last reset
	timeSinceReset := now.Sub(state.Sessions.LastReset)

	// Check if a week has passed (7 days)
	if timeSinceReset >= 7*24*time.Hour {
		logrus.Debugf("weekly reset triggered: %v since last reset", timeSinceReset)

		// Move thisWeek to lastWeek
		state.Sessions.LastWeek = state.Sessions.ThisWeek

		// Reset thisWeek counter
		state.Sessions.ThisWeek = 0

		// Update last reset time
		state.Sessions.LastReset = now

		// If there's an active challenge that hasn't expired, cancel it on weekly reset
		if state.Challenge.Active && now.Before(state.Challenge.ExpiresAt) {
			logrus.Debugf("canceling active challenge due to weekly reset")
			state.Challenge.Active = false
		}

		return true
	}

	return false
}

// CanTriggerIntervention checks if an intervention can be triggered based on cooldown
// Returns true if intervention is allowed, false if still in cooldown
func CanTriggerIntervention(state *ChurnState, now time.Time) bool {
	// If cooldown hasn't been set yet, allow intervention
	if state.Intervention.CooldownUntil.IsZero() {
		return true
	}

	// Check if current time is past the cooldown
	if now.After(state.Intervention.CooldownUntil) {
		logrus.Debugf("intervention cooldown has expired")
		return true
	}

	timeUntilCooldown := state.Intervention.CooldownUntil.Sub(now)
	logrus.Debugf("intervention still in cooldown for %v", timeUntilCooldown)
	return false
}

// SetInterventionCooldown sets the cooldown period for the next intervention
func SetInterventionCooldown(state *ChurnState, now time.Time, cooldownDuration time.Duration) {
	state.Intervention.LastTimestamp = now
	state.Intervention.CooldownUntil = now.Add(cooldownDuration)
	state.Intervention.TotalTriggered++

	logrus.Debugf("intervention cooldown set until %v (duration: %v)",
		state.Intervention.CooldownUntil, cooldownDuration)
}

// IsChurning determines if a player is exhibiting churn behavior
// A player is churning if:
// - They had activity last week (LastWeek > 0)
// - They have no activity this week (ThisWeek == 0)
// - At least 7 days have passed since last reset
func IsChurning(state *ChurnState, now time.Time) bool {
	timeSinceReset := now.Sub(state.Sessions.LastReset)

	// Must be at least 7 days since reset to determine churn
	if timeSinceReset < 7*24*time.Hour {
		return false
	}

	// Was active last week but not this week
	isChurning := state.Sessions.LastWeek > 0 && state.Sessions.ThisWeek == 0

	if isChurning {
		logrus.Debugf("player is churning: lastWeek=%d, thisWeek=%d, timeSinceReset=%v",
			state.Sessions.LastWeek, state.Sessions.ThisWeek, timeSinceReset)
	}

	return isChurning
}

// ShouldTriggerIntervention determines if an intervention should be triggered
// Intervention is triggered if:
// - Player is churning
// - No active challenge exists
// - Not in cooldown period
func ShouldTriggerIntervention(state *ChurnState, now time.Time) bool {
	// Check if player is churning
	if !IsChurning(state, now) {
		return false
	}

	// Don't trigger if there's already an active challenge
	if state.Challenge.Active {
		logrus.Debugf("intervention not triggered: challenge already active")
		return false
	}

	// Check cooldown
	if !CanTriggerIntervention(state, now) {
		logrus.Debugf("intervention not triggered: still in cooldown")
		return false
	}

	logrus.Debugf("intervention should be triggered")
	return true
}

// CreateChallenge creates a new comeback challenge for the player
func CreateChallenge(state *ChurnState, winsNeeded int, currentWins int, expiresAt time.Time, reason string) {
	state.Challenge.Active = true
	state.Challenge.WinsNeeded = winsNeeded
	state.Challenge.WinsCurrent = 0
	state.Challenge.WinsAtStart = currentWins
	state.Challenge.ExpiresAt = expiresAt
	state.Challenge.TriggerReason = reason

	logrus.Infof("created challenge: winsNeeded=%d, winsAtStart=%d, expiresAt=%v, reason=%s",
		winsNeeded, currentWins, expiresAt, reason)
}

// UpdateChallengeProgress updates the progress of an active challenge
// Returns true if the challenge is completed, false otherwise
func UpdateChallengeProgress(state *ChurnState, newWinCount int, now time.Time) bool {
	if !state.Challenge.Active {
		logrus.Warnf("attempted to update inactive challenge")
		return false
	}

	// Check if challenge has expired
	if now.After(state.Challenge.ExpiresAt) {
		logrus.Infof("challenge has expired, deactivating")
		state.Challenge.Active = false
		return false
	}

	// Calculate wins since challenge started
	winsSinceStart := newWinCount - state.Challenge.WinsAtStart
	if winsSinceStart < 0 {
		winsSinceStart = 0 // Handle edge case where wins decreased (shouldn't happen)
	}

	state.Challenge.WinsCurrent = winsSinceStart

	logrus.Debugf("challenge progress: %d/%d wins", state.Challenge.WinsCurrent, state.Challenge.WinsNeeded)

	// Check if challenge is completed
	if state.Challenge.WinsCurrent >= state.Challenge.WinsNeeded {
		logrus.Infof("challenge completed! %d/%d wins achieved",
			state.Challenge.WinsCurrent, state.Challenge.WinsNeeded)
		state.Challenge.Active = false
		return true
	}

	return false
}
