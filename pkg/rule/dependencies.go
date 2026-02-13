package rule

import "context"

// Service interfaces for external dependencies that rules can use
// These interfaces enable lazy loading of expensive data after threshold checks

// ClanService provides access to clan/guild related data
type ClanService interface {
	// GetClanActivity returns the clan's recent activity metrics
	GetClanActivity(ctx context.Context, clanID string) (*ClanActivity, error)

	// IsPlayerInClan checks if a player is a member of any clan
	IsPlayerInClan(ctx context.Context, userID string) (bool, string, error)

	// GetClanMembers returns list of members in a clan
	GetClanMembers(ctx context.Context, clanID string) ([]string, error)
}

// LeaderboardService provides access to leaderboard and ranking data
type LeaderboardService interface {
	// GetPlayerRank returns the player's current rank in a leaderboard
	GetPlayerRank(ctx context.Context, userID, leaderboardID string) (int, error)

	// GetTopPlayers returns the top N players in a leaderboard
	GetTopPlayers(ctx context.Context, leaderboardID string, limit int) ([]LeaderboardEntry, error)

	// GetPlayerRankHistory returns historical rank data for a player
	GetPlayerRankHistory(ctx context.Context, userID, leaderboardID string, days int) ([]RankSnapshot, error)
}

// PlayerHistoryService provides access to historical player data
type PlayerHistoryService interface {
	// GetMatchHistory returns recent match results for a player
	GetMatchHistory(ctx context.Context, userID string, limit int) ([]MatchRecord, error)

	// GetPlaytimeHistory returns playtime data over a period
	GetPlaytimeHistory(ctx context.Context, userID string, days int) ([]PlaytimeSnapshot, error)

	// GetPurchaseHistory returns recent purchase/transaction data
	GetPurchaseHistory(ctx context.Context, userID string, limit int) ([]PurchaseRecord, error)
}

// Data structures returned by service interfaces

// ClanActivity represents clan activity metrics
type ClanActivity struct {
	ClanID                 string
	ActiveMembersLast7Days int
	TotalMembers           int
	RecentEvents           []ClanEvent
}

// ClanEvent represents a single clan activity event
type ClanEvent struct {
	EventType string
	Timestamp int64
	UserID    string
}

// LeaderboardEntry represents a player's position in a leaderboard
type LeaderboardEntry struct {
	UserID string
	Rank   int
	Score  float64
}

// RankSnapshot represents a player's rank at a point in time
type RankSnapshot struct {
	Timestamp int64
	Rank      int
	Score     float64
}

// MatchRecord represents a single match result
type MatchRecord struct {
	MatchID   string
	Timestamp int64
	Result    string // "win", "loss", "draw"
	Duration  int    // seconds
}

// PlaytimeSnapshot represents playtime at a point in time
type PlaytimeSnapshot struct {
	Date    string // YYYY-MM-DD
	Minutes int
}

// PurchaseRecord represents a purchase/transaction
type PurchaseRecord struct {
	TransactionID string
	Timestamp     int64
	ItemID        string
	Amount        float64
	Currency      string
}

// RuleDependencies holds all external service dependencies that rules can use
// Rules receive this struct and can access only the services they need
type RuleDependencies struct {
	ClanService          ClanService
	LeaderboardService   LeaderboardService
	PlayerHistoryService PlayerHistoryService
}

// NewRuleDependencies creates a new dependencies container
// Services can be nil if not needed - rules should handle nil gracefully
func NewRuleDependencies() *RuleDependencies {
	return &RuleDependencies{}
}

// WithClanService sets the clan service
func (d *RuleDependencies) WithClanService(service ClanService) *RuleDependencies {
	d.ClanService = service
	return d
}

// WithLeaderboardService sets the leaderboard service
func (d *RuleDependencies) WithLeaderboardService(service LeaderboardService) *RuleDependencies {
	d.LeaderboardService = service
	return d
}

// WithPlayerHistoryService sets the player history service
func (d *RuleDependencies) WithPlayerHistoryService(service PlayerHistoryService) *RuleDependencies {
	d.PlayerHistoryService = service
	return d
}
