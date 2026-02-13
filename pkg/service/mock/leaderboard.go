package mock

import (
	"context"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
)

// LeaderboardService is a mock implementation of rule.LeaderboardService for testing
type LeaderboardService struct {
	// Function fields for custom behavior
	GetPlayerRankFunc        func(ctx context.Context, userID, leaderboardID string) (int, error)
	GetTopPlayersFunc        func(ctx context.Context, leaderboardID string, limit int) ([]rule.LeaderboardEntry, error)
	GetPlayerRankHistoryFunc func(ctx context.Context, userID, leaderboardID string, days int) ([]rule.RankSnapshot, error)

	// Simple fields for common scenarios
	PlayerRank  int
	TopPlayers  []rule.LeaderboardEntry
	RankHistory []rule.RankSnapshot
	Error       error
}

// GetPlayerRank returns mocked player rank
func (m *LeaderboardService) GetPlayerRank(ctx context.Context, userID, leaderboardID string) (int, error) {
	if m.GetPlayerRankFunc != nil {
		return m.GetPlayerRankFunc(ctx, userID, leaderboardID)
	}
	if m.Error != nil {
		return 0, m.Error
	}
	return m.PlayerRank, nil
}

// GetTopPlayers returns mocked top players
func (m *LeaderboardService) GetTopPlayers(ctx context.Context, leaderboardID string, limit int) ([]rule.LeaderboardEntry, error) {
	if m.GetTopPlayersFunc != nil {
		return m.GetTopPlayersFunc(ctx, leaderboardID, limit)
	}
	if m.Error != nil {
		return nil, m.Error
	}
	return m.TopPlayers, nil
}

// GetPlayerRankHistory returns mocked rank history
func (m *LeaderboardService) GetPlayerRankHistory(ctx context.Context, userID, leaderboardID string, days int) ([]rule.RankSnapshot, error) {
	if m.GetPlayerRankHistoryFunc != nil {
		return m.GetPlayerRankHistoryFunc(ctx, userID, leaderboardID, days)
	}
	if m.Error != nil {
		return nil, m.Error
	}
	return m.RankHistory, nil
}

// NewLeaderboardService creates a new mock leaderboard service
func NewLeaderboardService() *LeaderboardService {
	return &LeaderboardService{
		PlayerRank: 0,
	}
}

// WithPlayerRank sets the player rank to return
func (m *LeaderboardService) WithPlayerRank(rank int) *LeaderboardService {
	m.PlayerRank = rank
	return m
}

// WithTopPlayers sets the top players to return
func (m *LeaderboardService) WithTopPlayers(players []rule.LeaderboardEntry) *LeaderboardService {
	m.TopPlayers = players
	return m
}

// WithRankHistory sets the rank history to return
func (m *LeaderboardService) WithRankHistory(history []rule.RankSnapshot) *LeaderboardService {
	m.RankHistory = history
	return m
}

// WithError sets an error to return
func (m *LeaderboardService) WithError(err error) *LeaderboardService {
	m.Error = err
	return m
}
