package mock

import (
	"context"
	"fmt"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
)

// PlayerHistoryService is a mock implementation of rule.PlayerHistoryService for testing
type PlayerHistoryService struct {
	// GetMatchHistoryFunc is called when GetMatchHistory is invoked
	GetMatchHistoryFunc func(ctx context.Context, userID string, limit int) ([]rule.MatchRecord, error)

	// GetPlaytimeHistoryFunc is called when GetPlaytimeHistory is invoked
	GetPlaytimeHistoryFunc func(ctx context.Context, userID string, days int) ([]rule.PlaytimeSnapshot, error)

	// GetPurchaseHistoryFunc is called when GetPurchaseHistory is invoked
	GetPurchaseHistoryFunc func(ctx context.Context, userID string, limit int) ([]rule.PurchaseRecord, error)

	// Default data
	DefaultMatches   []rule.MatchRecord
	DefaultPlaytime  []rule.PlaytimeSnapshot
	DefaultPurchases []rule.PurchaseRecord
	DefaultError     error

	// Call tracking
	GetMatchHistoryCalls    []GetMatchHistoryCall
	GetPlaytimeHistoryCalls []GetPlaytimeHistoryCall
	GetPurchaseHistoryCalls []GetPurchaseHistoryCall
}

// GetMatchHistoryCall tracks parameters for GetMatchHistory calls
type GetMatchHistoryCall struct {
	UserID string
	Limit  int
}

// GetPlaytimeHistoryCall tracks parameters for GetPlaytimeHistory calls
type GetPlaytimeHistoryCall struct {
	UserID string
	Days   int
}

// GetPurchaseHistoryCall tracks parameters for GetPurchaseHistory calls
type GetPurchaseHistoryCall struct {
	UserID string
	Limit  int
}

// NewPlayerHistoryService creates a new mock PlayerHistoryService with defaults
func NewPlayerHistoryService() *PlayerHistoryService {
	return &PlayerHistoryService{
		DefaultMatches: []rule.MatchRecord{
			{MatchID: "match1", Timestamp: 1234567890, Result: "win", Duration: 1200},
			{MatchID: "match2", Timestamp: 1234567891, Result: "loss", Duration: 900},
			{MatchID: "match3", Timestamp: 1234567892, Result: "win", Duration: 1100},
		},
		DefaultPlaytime: []rule.PlaytimeSnapshot{
			{Date: "2024-01-01", Minutes: 120},
			{Date: "2024-01-02", Minutes: 90},
			{Date: "2024-01-03", Minutes: 150},
		},
		DefaultPurchases: []rule.PurchaseRecord{
			{TransactionID: "tx1", Timestamp: 1234567890, ItemID: "sword_legendary", Amount: 9.99, Currency: "USD"},
			{TransactionID: "tx2", Timestamp: 1234567891, ItemID: "skin_rare", Amount: 4.99, Currency: "USD"},
		},
	}
}

// GetMatchHistory returns recent match results
func (m *PlayerHistoryService) GetMatchHistory(ctx context.Context, userID string, limit int) ([]rule.MatchRecord, error) {
	// Track call
	m.GetMatchHistoryCalls = append(m.GetMatchHistoryCalls, GetMatchHistoryCall{
		UserID: userID,
		Limit:  limit,
	})

	// Use custom function if provided
	if m.GetMatchHistoryFunc != nil {
		return m.GetMatchHistoryFunc(ctx, userID, limit)
	}

	// Use default behavior
	if m.DefaultError != nil {
		return nil, m.DefaultError
	}

	// Return up to limit entries
	if limit < len(m.DefaultMatches) {
		return m.DefaultMatches[:limit], nil
	}
	return m.DefaultMatches, nil
}

// GetPlaytimeHistory returns playtime data over a period
func (m *PlayerHistoryService) GetPlaytimeHistory(ctx context.Context, userID string, days int) ([]rule.PlaytimeSnapshot, error) {
	// Track call
	m.GetPlaytimeHistoryCalls = append(m.GetPlaytimeHistoryCalls, GetPlaytimeHistoryCall{
		UserID: userID,
		Days:   days,
	})

	// Use custom function if provided
	if m.GetPlaytimeHistoryFunc != nil {
		return m.GetPlaytimeHistoryFunc(ctx, userID, days)
	}

	// Use default behavior
	if m.DefaultError != nil {
		return nil, m.DefaultError
	}

	// Return up to days entries
	if days < len(m.DefaultPlaytime) {
		return m.DefaultPlaytime[:days], nil
	}
	return m.DefaultPlaytime, nil
}

// GetPurchaseHistory returns recent purchases
func (m *PlayerHistoryService) GetPurchaseHistory(ctx context.Context, userID string, limit int) ([]rule.PurchaseRecord, error) {
	// Track call
	m.GetPurchaseHistoryCalls = append(m.GetPurchaseHistoryCalls, GetPurchaseHistoryCall{
		UserID: userID,
		Limit:  limit,
	})

	// Use custom function if provided
	if m.GetPurchaseHistoryFunc != nil {
		return m.GetPurchaseHistoryFunc(ctx, userID, limit)
	}

	// Use default behavior
	if m.DefaultError != nil {
		return nil, m.DefaultError
	}

	// Return up to limit entries
	if limit < len(m.DefaultPurchases) {
		return m.DefaultPurchases[:limit], nil
	}
	return m.DefaultPurchases, nil
}

// Reset clears all call tracking
func (m *PlayerHistoryService) Reset() {
	m.GetMatchHistoryCalls = nil
	m.GetPlaytimeHistoryCalls = nil
	m.GetPurchaseHistoryCalls = nil
}

// WithError sets the default error to return
func (m *PlayerHistoryService) WithError(err error) *PlayerHistoryService {
	m.DefaultError = err
	return m
}

// WithLosingStreak creates match history showing consecutive losses
func (m *PlayerHistoryService) WithLosingStreak(count int) *PlayerHistoryService {
	matches := make([]rule.MatchRecord, count)
	for i := 0; i < count; i++ {
		matches[i] = rule.MatchRecord{
			MatchID:   fmt.Sprintf("match%d", i+1),
			Timestamp: int64(1234567890 + i),
			Result:    "loss",
			Duration:  900,
		}
	}
	m.DefaultMatches = matches
	return m
}

// WithWinningStreak creates match history showing consecutive wins
func (m *PlayerHistoryService) WithWinningStreak(count int) *PlayerHistoryService {
	matches := make([]rule.MatchRecord, count)
	for i := 0; i < count; i++ {
		matches[i] = rule.MatchRecord{
			MatchID:   fmt.Sprintf("match%d", i+1),
			Timestamp: int64(1234567890 + i),
			Result:    "win",
			Duration:  1200,
		}
	}
	m.DefaultMatches = matches
	return m
}

// WithDecreasingPlaytime creates playtime showing decline
func (m *PlayerHistoryService) WithDecreasingPlaytime(days int, startMinutes, endMinutes int) *PlayerHistoryService {
	snapshots := make([]rule.PlaytimeSnapshot, days)
	step := (startMinutes - endMinutes) / (days - 1)
	for i := 0; i < days; i++ {
		snapshots[i] = rule.PlaytimeSnapshot{
			Date:    fmt.Sprintf("2024-01-%02d", i+1),
			Minutes: startMinutes - (i * step),
		}
	}
	m.DefaultPlaytime = snapshots
	return m
}

// WithRecentPurchases creates purchase history
func (m *PlayerHistoryService) WithRecentPurchases(count int, totalSpent float64) *PlayerHistoryService {
	purchases := make([]rule.PurchaseRecord, count)
	amountEach := totalSpent / float64(count)
	for i := 0; i < count; i++ {
		purchases[i] = rule.PurchaseRecord{
			TransactionID: fmt.Sprintf("tx%d", i+1),
			Timestamp:     int64(1234567890 + i),
			ItemID:        fmt.Sprintf("item%d", i+1),
			Amount:        amountEach,
			Currency:      "USD",
		}
	}
	m.DefaultPurchases = purchases
	return m
}

// AssertGetMatchHistoryCalled verifies GetMatchHistory was called
func (m *PlayerHistoryService) AssertGetMatchHistoryCalled(userID string) error {
	for _, call := range m.GetMatchHistoryCalls {
		if call.UserID == userID {
			return nil
		}
	}
	return fmt.Errorf("expected GetMatchHistory called with userID=%s, but got calls: %v",
		userID, m.GetMatchHistoryCalls)
}
