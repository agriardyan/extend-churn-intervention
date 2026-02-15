// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package service

import (
	"time"
)

// ChurnState represents anti-churn detection and intervention history for a player.
// This is the ONLY state that the anti-churn system owns and manages.
type ChurnState struct {
	Sessions            SessionState         `json:"sessions"`
	SignalHistory       []ChurnSignal        `json:"signalHistory"`
	InterventionHistory []InterventionRecord `json:"interventionHistory"`
	Cooldown            CooldownState        `json:"cooldown"`
}

// SessionState caches session data from IAM events (read-only cache).
// We DO NOT maintain these counts - we only cache what we see from events.
// The IAM/Session service is the source of truth.
type SessionState struct {
	ThisWeek  int       `json:"thisWeek"`
	LastWeek  int       `json:"lastWeek"`
	LastReset time.Time `json:"lastReset"`
}

// ChurnSignal represents a detected churn risk signal.
// This is our record of what behavioral patterns we detected.
type ChurnSignal struct {
	Type       string                 `json:"type"`       // e.g., "losing_streak", "rage_quit", "session_decline"
	DetectedAt time.Time              `json:"detectedAt"` // When we detected this signal
	Severity   string                 `json:"severity"`   // "low", "medium", "high"
	Metadata   map[string]interface{} `json:"metadata"`   // Signal-specific data (e.g., streak length, decline percentage)
}

// InterventionRecord tracks interventions we executed.
// This is our record of what we DID to re-engage the player.
// We DO NOT track the intervention's internal progress (e.g., challenge wins) - that's owned by other systems.
type InterventionRecord struct {
	ID          string                 `json:"id"`          // Unique intervention ID
	Type        string                 `json:"type"`        // e.g., "comeback_challenge", "grant_item"
	TriggeredBy string                 `json:"triggeredBy"` // Rule ID that triggered this intervention
	TriggeredAt time.Time              `json:"triggeredAt"` // When we created this intervention
	ExpiresAt   *time.Time             `json:"expiresAt"`   // Time-limited interventions (null if not applicable)
	Outcome     string                 `json:"outcome"`     // "active", "completed", "expired", "failed"
	OutcomeAt   *time.Time             `json:"outcomeAt"`   // When the outcome was determined
	Metadata    map[string]interface{} `json:"metadata"`    // Intervention-specific data
}

// CooldownState tracks intervention throttling to prevent spam.
// This ensures we don't overwhelm players with too many interventions.
type CooldownState struct {
	LastInterventionAt time.Time       `json:"lastInterventionAt"` // When we last intervened
	CooldownUntil      time.Time       `json:"cooldownUntil"`      // When cooldown expires
	InterventionCounts map[string]int  `json:"interventionCounts"` // Count per intervention type (for rate limiting)
	LastSignalAt       map[string]time.Time `json:"lastSignalAt"` // Last detection time per signal type
}

// IsOnCooldown returns whether the player is currently in a cooldown period.
func (c *CooldownState) IsOnCooldown() bool {
	return time.Now().Before(c.CooldownUntil)
}

// GetActiveInterventions returns all interventions that are currently active.
func (cs *ChurnState) GetActiveInterventions() []InterventionRecord {
	var active []InterventionRecord
	for _, intervention := range cs.InterventionHistory {
		if intervention.Outcome == "active" {
			// Check if expired
			if intervention.ExpiresAt != nil && time.Now().After(*intervention.ExpiresAt) {
				continue // Expired
			}
			active = append(active, intervention)
		}
	}
	return active
}

// GetInterventionByID finds an intervention by its ID.
func (cs *ChurnState) GetInterventionByID(id string) *InterventionRecord {
	for i := range cs.InterventionHistory {
		if cs.InterventionHistory[i].ID == id {
			return &cs.InterventionHistory[i]
		}
	}
	return nil
}

// AddSignal records a new churn signal detection.
func (cs *ChurnState) AddSignal(signalType, severity string, metadata map[string]interface{}) {
	signal := ChurnSignal{
		Type:       signalType,
		DetectedAt: time.Now(),
		Severity:   severity,
		Metadata:   metadata,
	}
	cs.SignalHistory = append(cs.SignalHistory, signal)

	// Update last signal detection time
	if cs.Cooldown.LastSignalAt == nil {
		cs.Cooldown.LastSignalAt = make(map[string]time.Time)
	}
	cs.Cooldown.LastSignalAt[signalType] = time.Now()
}

// AddIntervention records a new intervention execution.
func (cs *ChurnState) AddIntervention(id, interventionType, triggeredBy string, expiresAt *time.Time, metadata map[string]interface{}) {
	intervention := InterventionRecord{
		ID:          id,
		Type:        interventionType,
		TriggeredBy: triggeredBy,
		TriggeredAt: time.Now(),
		ExpiresAt:   expiresAt,
		Outcome:     "active",
		Metadata:    metadata,
	}
	cs.InterventionHistory = append(cs.InterventionHistory, intervention)

	// Update cooldown state
	cs.Cooldown.LastInterventionAt = time.Now()
	if cs.Cooldown.InterventionCounts == nil {
		cs.Cooldown.InterventionCounts = make(map[string]int)
	}
	cs.Cooldown.InterventionCounts[interventionType]++
}

// UpdateInterventionOutcome updates the outcome of an intervention.
func (cs *ChurnState) UpdateInterventionOutcome(id, outcome string) bool {
	intervention := cs.GetInterventionByID(id)
	if intervention == nil {
		return false
	}
	intervention.Outcome = outcome
	now := time.Now()
	intervention.OutcomeAt = &now
	return true
}
