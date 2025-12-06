// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package state

import (
	"time"
)

// ChurnState represents the complete state for a player
type ChurnState struct {
	Sessions     SessionState      `json:"sessions"`
	Challenge    ChallengeState    `json:"challenge"`
	Intervention InterventionState `json:"intervention"`
}

// SessionState tracks player session frequency
type SessionState struct {
	ThisWeek  int       `json:"thisWeek"`
	LastWeek  int       `json:"lastWeek"`
	LastReset time.Time `json:"lastReset"`
}

// ChallengeState tracks active comeback challenge
type ChallengeState struct {
	Active        bool      `json:"active"`
	WinsNeeded    int       `json:"winsNeeded"`
	WinsCurrent   int       `json:"winsCurrent"`
	WinsAtStart   int       `json:"winsAtStart"`
	ExpiresAt     time.Time `json:"expiresAt"`
	TriggerReason string    `json:"triggerReason"`
}

// InterventionState tracks intervention history and cooldowns
type InterventionState struct {
	LastTimestamp  time.Time `json:"lastTimestamp"`
	CooldownUntil  time.Time `json:"cooldownUntil"`
	TotalTriggered int       `json:"totalTriggered"`
}
