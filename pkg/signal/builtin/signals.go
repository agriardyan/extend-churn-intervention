package builtin

import (
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Signal type constants for built-in signals
const (
	TypeLogin        = "login"
	TypeLogout       = "logout"
	TypeMatchWin     = "match_win"
	TypeLosingStreak = "losing_streak"
	TypeRageQuit     = "rage_quit"
)

// LoginSignal represents a player login event.
type LoginSignal struct {
	signal.BaseSignal
}

// NewLoginSignal creates a new login signal.
func NewLoginSignal(userID string, timestamp time.Time, context *signal.PlayerContext) *LoginSignal {
	metadata := map[string]interface{}{
		"event": "oauth_token_generated",
	}
	return &LoginSignal{
		BaseSignal: signal.NewBaseSignal(TypeLogin, userID, timestamp, metadata, context),
	}
}

// LogoutSignal represents a player logout event.
type LogoutSignal struct {
	signal.BaseSignal
}

// NewLogoutSignal creates a new logout signal.
func NewLogoutSignal(userID string, timestamp time.Time, context *signal.PlayerContext) *LogoutSignal {
	metadata := map[string]interface{}{
		"event": "logout",
	}
	return &LogoutSignal{
		BaseSignal: signal.NewBaseSignal(TypeLogout, userID, timestamp, metadata, context),
	}
}

// WinSignal represents a player winning a match.
type WinSignal struct {
	signal.BaseSignal
	TotalWins int
}

// NewWinSignal creates a new win signal.
func NewWinSignal(userID string, timestamp time.Time, totalWins int, context *signal.PlayerContext) *WinSignal {
	metadata := map[string]interface{}{
		"total_wins": totalWins,
		"stat_code":  "rse-match-wins",
	}
	return &WinSignal{
		BaseSignal: signal.NewBaseSignal(TypeMatchWin, userID, timestamp, metadata, context),
		TotalWins:  totalWins,
	}
}

// LossSignal represents a player losing a match.
type LossSignal struct {
	signal.BaseSignal
	CurrentStreak int
}

// NewLossSignal creates a new loss signal.
func NewLossSignal(userID string, timestamp time.Time, currentStreak int, context *signal.PlayerContext) *LossSignal {
	metadata := map[string]interface{}{
		"current_streak": currentStreak,
		"stat_code":      "rse-current-losing-streak",
	}
	return &LossSignal{
		BaseSignal:    signal.NewBaseSignal(TypeLosingStreak, userID, timestamp, metadata, context),
		CurrentStreak: currentStreak,
	}
}

// RageQuitSignal represents a player rage quitting.
type RageQuitSignal struct {
	signal.BaseSignal
	QuitCount    int
	MatchContext map[string]interface{}
}

// NewRageQuitSignal creates a new rage quit signal.
func NewRageQuitSignal(userID string, timestamp time.Time, quitCount int, context *signal.PlayerContext) *RageQuitSignal {
	metadata := map[string]interface{}{
		"quit_count": quitCount,
		"stat_code":  "rse-rage-quit",
	}
	return &RageQuitSignal{
		BaseSignal:   signal.NewBaseSignal(TypeRageQuit, userID, timestamp, metadata, context),
		QuitCount:    quitCount,
		MatchContext: make(map[string]interface{}),
	}
}
