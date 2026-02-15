package service

import (
	"context"
)

// Service interfaces for external dependencies that rules/actions can use.
// These interfaces enable lazy loading of expensive data after threshold checks.
//
// You may not need to have interface and go with direct struct usage,
// but having interfaces allows easier mocking for unit tests.

type EntitlementGranter interface {
	// GrantEntitlement grants an entitlement/item to a player
	GrantEntitlement(ctx context.Context, userID, itemID string, quantity int) error
}

type UserStatisticUpdater interface {
	// UpdatePlayerStat updates a player's statistic
	UpdateStatComebackChallenge(ctx context.Context, userID string) error
}

// StateStore defines the interface for accessing player churn state.
// This allows for easier testing and different storage implementations.
type StateStore interface {
	GetChurnState(ctx context.Context, userID string) (*ChurnState, error)
	UpdateChurnState(ctx context.Context, userID string, state *ChurnState) error
}
