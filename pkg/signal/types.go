package signal

import "time"

// Signal type constants
const (
	TypeLogin      = "login"
	TypeLogout     = "logout"
	TypeWin        = "win"
	TypeLoss       = "loss"
	TypeRageQuit   = "rage_quit"
	TypeStatUpdate = "stat_update"
)

// LoginSignal represents a player login event.
type LoginSignal struct {
	BaseSignal
}

// NewLoginSignal creates a new login signal.
func NewLoginSignal(userID string, timestamp time.Time, context *PlayerContext) *LoginSignal {
	metadata := map[string]interface{}{
		"event": "oauth_token_generated",
	}
	return &LoginSignal{
		BaseSignal: NewBaseSignal(TypeLogin, userID, timestamp, metadata, context),
	}
}

// LogoutSignal represents a player logout event.
type LogoutSignal struct {
	BaseSignal
}

// NewLogoutSignal creates a new logout signal.
func NewLogoutSignal(userID string, timestamp time.Time, context *PlayerContext) *LogoutSignal {
	metadata := map[string]interface{}{
		"event": "logout",
	}
	return &LogoutSignal{
		BaseSignal: NewBaseSignal(TypeLogout, userID, timestamp, metadata, context),
	}
}

// WinSignal represents a player winning a match.
type WinSignal struct {
	BaseSignal
	TotalWins int
}

// NewWinSignal creates a new win signal.
func NewWinSignal(userID string, timestamp time.Time, totalWins int, context *PlayerContext) *WinSignal {
	metadata := map[string]interface{}{
		"total_wins": totalWins,
		"stat_code":  "rse-match-wins",
	}
	return &WinSignal{
		BaseSignal: NewBaseSignal(TypeWin, userID, timestamp, metadata, context),
		TotalWins:  totalWins,
	}
}

// LossSignal represents a player losing a match.
type LossSignal struct {
	BaseSignal
	CurrentStreak int
}

// NewLossSignal creates a new loss signal.
func NewLossSignal(userID string, timestamp time.Time, currentStreak int, context *PlayerContext) *LossSignal {
	metadata := map[string]interface{}{
		"current_streak": currentStreak,
		"stat_code":      "rse-current-losing-streak",
	}
	return &LossSignal{
		BaseSignal:    NewBaseSignal(TypeLoss, userID, timestamp, metadata, context),
		CurrentStreak: currentStreak,
	}
}

// RageQuitSignal represents a player rage quitting.
type RageQuitSignal struct {
	BaseSignal
	QuitCount    int
	MatchContext map[string]interface{}
}

// NewRageQuitSignal creates a new rage quit signal.
func NewRageQuitSignal(userID string, timestamp time.Time, quitCount int, context *PlayerContext) *RageQuitSignal {
	metadata := map[string]interface{}{
		"quit_count": quitCount,
		"stat_code":  "rse-rage-quit",
	}
	return &RageQuitSignal{
		BaseSignal:   NewBaseSignal(TypeRageQuit, userID, timestamp, metadata, context),
		QuitCount:    quitCount,
		MatchContext: make(map[string]interface{}),
	}
}

// StatUpdateSignal represents a generic statistic update event.
// This is used as a fallback for stat codes that don't have specific signal types.
type StatUpdateSignal struct {
	BaseSignal
	StatCode string
	Value    float64
}

// NewStatUpdateSignal creates a new stat update signal.
func NewStatUpdateSignal(userID string, timestamp time.Time, statCode string, value float64, context *PlayerContext) *StatUpdateSignal {
	metadata := map[string]interface{}{
		"stat_code": statCode,
		"value":     value,
	}
	return &StatUpdateSignal{
		BaseSignal: NewBaseSignal(TypeStatUpdate, userID, timestamp, metadata, context),
		StatCode:   statCode,
		Value:      value,
	}
}
