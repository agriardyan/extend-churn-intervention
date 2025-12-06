// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package state

import (
	"testing"
	"time"
)

func TestCheckWeeklyReset(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name           string
		state          *ChurnState
		now            time.Time
		expectReset    bool
		expectedValues map[string]int
	}{
		{
			name: "no reset needed - less than 7 days",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  5,
					LastWeek:  3,
					LastReset: now.Add(-3 * 24 * time.Hour),
				},
			},
			now:         now,
			expectReset: false,
		},
		{
			name: "reset needed - exactly 7 days",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  5,
					LastWeek:  3,
					LastReset: now.Add(-7 * 24 * time.Hour),
				},
			},
			now:         now,
			expectReset: true,
			expectedValues: map[string]int{
				"thisWeek": 0,
				"lastWeek": 5,
			},
		},
		{
			name: "reset needed - more than 7 days",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  10,
					LastWeek:  7,
					LastReset: now.Add(-14 * 24 * time.Hour),
				},
			},
			now:         now,
			expectReset: true,
			expectedValues: map[string]int{
				"thisWeek": 0,
				"lastWeek": 10,
			},
		},
		{
			name: "reset cancels active challenge",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  5,
					LastWeek:  3,
					LastReset: now.Add(-8 * 24 * time.Hour),
				},
				Challenge: ChallengeState{
					Active:    true,
					ExpiresAt: now.Add(24 * time.Hour), // Challenge still valid
				},
			},
			now:         now,
			expectReset: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CheckWeeklyReset(tt.state, tt.now)

			if result != tt.expectReset {
				t.Errorf("CheckWeeklyReset() = %v, expected %v", result, tt.expectReset)
			}

			if tt.expectReset {
				if tt.expectedValues != nil {
					if tt.state.Sessions.ThisWeek != tt.expectedValues["thisWeek"] {
						t.Errorf("ThisWeek = %d, expected %d",
							tt.state.Sessions.ThisWeek, tt.expectedValues["thisWeek"])
					}
					if tt.state.Sessions.LastWeek != tt.expectedValues["lastWeek"] {
						t.Errorf("LastWeek = %d, expected %d",
							tt.state.Sessions.LastWeek, tt.expectedValues["lastWeek"])
					}
				}

				// Check that LastReset was updated
				if !tt.state.Sessions.LastReset.Equal(tt.now) {
					t.Errorf("LastReset was not updated to now")
				}

				// If challenge was active and hasn't expired, it should be cancelled
				if tt.state.Challenge.Active && tt.now.Before(tt.state.Challenge.ExpiresAt) {
					t.Errorf("Active challenge should be cancelled on weekly reset")
				}
			}
		})
	}
}

func TestCanTriggerIntervention(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		state    *ChurnState
		now      time.Time
		expected bool
	}{
		{
			name: "no cooldown set - allow intervention",
			state: &ChurnState{
				Intervention: InterventionState{},
			},
			now:      now,
			expected: true,
		},
		{
			name: "cooldown expired - allow intervention",
			state: &ChurnState{
				Intervention: InterventionState{
					CooldownUntil: now.Add(-1 * time.Hour),
				},
			},
			now:      now,
			expected: true,
		},
		{
			name: "still in cooldown - block intervention",
			state: &ChurnState{
				Intervention: InterventionState{
					CooldownUntil: now.Add(2 * time.Hour),
				},
			},
			now:      now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CanTriggerIntervention(tt.state, tt.now)
			if result != tt.expected {
				t.Errorf("CanTriggerIntervention() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestSetInterventionCooldown(t *testing.T) {
	now := time.Now()
	cooldownDuration := 48 * time.Hour

	state := &ChurnState{
		Intervention: InterventionState{
			TotalTriggered: 2,
		},
	}

	SetInterventionCooldown(state, now, cooldownDuration)

	// Check LastTimestamp was set
	if !state.Intervention.LastTimestamp.Equal(now) {
		t.Errorf("LastTimestamp = %v, expected %v", state.Intervention.LastTimestamp, now)
	}

	// Check CooldownUntil was set correctly
	expectedCooldownUntil := now.Add(cooldownDuration)
	if !state.Intervention.CooldownUntil.Equal(expectedCooldownUntil) {
		t.Errorf("CooldownUntil = %v, expected %v",
			state.Intervention.CooldownUntil, expectedCooldownUntil)
	}

	// Check TotalTriggered was incremented
	if state.Intervention.TotalTriggered != 3 {
		t.Errorf("TotalTriggered = %d, expected 3", state.Intervention.TotalTriggered)
	}
}

func TestIsChurning(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		state    *ChurnState
		now      time.Time
		expected bool
	}{
		{
			name: "churning - was active last week, not this week",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  0,
					LastWeek:  5,
					LastReset: now.Add(-8 * 24 * time.Hour),
				},
			},
			now:      now,
			expected: true,
		},
		{
			name: "not churning - active both weeks",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  3,
					LastWeek:  5,
					LastReset: now.Add(-8 * 24 * time.Hour),
				},
			},
			now:      now,
			expected: false,
		},
		{
			name: "not churning - not enough time passed",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  0,
					LastWeek:  5,
					LastReset: now.Add(-3 * 24 * time.Hour),
				},
			},
			now:      now,
			expected: false,
		},
		{
			name: "not churning - never active",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  0,
					LastWeek:  0,
					LastReset: now.Add(-10 * 24 * time.Hour),
				},
			},
			now:      now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsChurning(tt.state, tt.now)
			if result != tt.expected {
				t.Errorf("IsChurning() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestShouldTriggerIntervention(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name     string
		state    *ChurnState
		now      time.Time
		expected bool
	}{
		{
			name: "should trigger - churning, no active challenge, no cooldown",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  0,
					LastWeek:  5,
					LastReset: now.Add(-8 * 24 * time.Hour),
				},
				Challenge: ChallengeState{
					Active: false,
				},
				Intervention: InterventionState{},
			},
			now:      now,
			expected: true,
		},
		{
			name: "should not trigger - not churning",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  3,
					LastWeek:  5,
					LastReset: now.Add(-8 * 24 * time.Hour),
				},
				Challenge: ChallengeState{
					Active: false,
				},
				Intervention: InterventionState{},
			},
			now:      now,
			expected: false,
		},
		{
			name: "should not trigger - active challenge exists",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  0,
					LastWeek:  5,
					LastReset: now.Add(-8 * 24 * time.Hour),
				},
				Challenge: ChallengeState{
					Active: true,
				},
				Intervention: InterventionState{},
			},
			now:      now,
			expected: false,
		},
		{
			name: "should not trigger - in cooldown",
			state: &ChurnState{
				Sessions: SessionState{
					ThisWeek:  0,
					LastWeek:  5,
					LastReset: now.Add(-8 * 24 * time.Hour),
				},
				Challenge: ChallengeState{
					Active: false,
				},
				Intervention: InterventionState{
					CooldownUntil: now.Add(24 * time.Hour),
				},
			},
			now:      now,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ShouldTriggerIntervention(tt.state, tt.now)
			if result != tt.expected {
				t.Errorf("ShouldTriggerIntervention() = %v, expected %v", result, tt.expected)
			}
		})
	}
}

func TestCreateChallenge(t *testing.T) {
	state := &ChurnState{}
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)

	CreateChallenge(state, 3, 10, expiresAt, "churn_detected")

	if !state.Challenge.Active {
		t.Error("Challenge should be active")
	}
	if state.Challenge.WinsNeeded != 3 {
		t.Errorf("WinsNeeded = %d, expected 3", state.Challenge.WinsNeeded)
	}
	if state.Challenge.WinsCurrent != 0 {
		t.Errorf("WinsCurrent = %d, expected 0", state.Challenge.WinsCurrent)
	}
	if state.Challenge.WinsAtStart != 10 {
		t.Errorf("WinsAtStart = %d, expected 10", state.Challenge.WinsAtStart)
	}
	if !state.Challenge.ExpiresAt.Equal(expiresAt) {
		t.Errorf("ExpiresAt = %v, expected %v", state.Challenge.ExpiresAt, expiresAt)
	}
	if state.Challenge.TriggerReason != "churn_detected" {
		t.Errorf("TriggerReason = %s, expected 'churn_detected'", state.Challenge.TriggerReason)
	}
}

func TestUpdateChallengeProgress(t *testing.T) {
	now := time.Now()

	tests := []struct {
		name            string
		state           *ChurnState
		newWinCount     int
		now             time.Time
		expectCompleted bool
		expectActive    bool
	}{
		{
			name: "challenge in progress - not completed",
			state: &ChurnState{
				Challenge: ChallengeState{
					Active:      true,
					WinsNeeded:  3,
					WinsCurrent: 0,
					WinsAtStart: 10,
					ExpiresAt:   now.Add(24 * time.Hour),
				},
			},
			newWinCount:     12, // 2 wins since start
			now:             now,
			expectCompleted: false,
			expectActive:    true,
		},
		{
			name: "challenge completed",
			state: &ChurnState{
				Challenge: ChallengeState{
					Active:      true,
					WinsNeeded:  3,
					WinsCurrent: 0,
					WinsAtStart: 10,
					ExpiresAt:   now.Add(24 * time.Hour),
				},
			},
			newWinCount:     13, // 3 wins since start
			now:             now,
			expectCompleted: true,
			expectActive:    false,
		},
		{
			name: "challenge expired",
			state: &ChurnState{
				Challenge: ChallengeState{
					Active:      true,
					WinsNeeded:  3,
					WinsCurrent: 0,
					WinsAtStart: 10,
					ExpiresAt:   now.Add(-1 * time.Hour), // Expired
				},
			},
			newWinCount:     12,
			now:             now,
			expectCompleted: false,
			expectActive:    false,
		},
		{
			name: "inactive challenge - no update",
			state: &ChurnState{
				Challenge: ChallengeState{
					Active:      false,
					WinsNeeded:  3,
					WinsCurrent: 0,
					WinsAtStart: 10,
					ExpiresAt:   now.Add(24 * time.Hour),
				},
			},
			newWinCount:     13,
			now:             now,
			expectCompleted: false,
			expectActive:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := UpdateChallengeProgress(tt.state, tt.newWinCount, tt.now)

			if result != tt.expectCompleted {
				t.Errorf("UpdateChallengeProgress() = %v, expected %v", result, tt.expectCompleted)
			}

			if tt.state.Challenge.Active != tt.expectActive {
				t.Errorf("Challenge.Active = %v, expected %v",
					tt.state.Challenge.Active, tt.expectActive)
			}
		})
	}
}
