package builtin

import (
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

func TestLoginSignal(t *testing.T) {
	timestamp := time.Now()
	playerCtx := &signal.PlayerContext{
		UserID: "user123",
		State:  &state.ChurnState{},
	}

	sig := NewLoginSignal("user123", timestamp, playerCtx)

	if sig.Type() != TypeLogin {
		t.Errorf("Expected type '%s', got '%s'", TypeLogin, sig.Type())
	}

	if sig.UserID() != "user123" {
		t.Errorf("Expected userID 'user123', got '%s'", sig.UserID())
	}

	if sig.Metadata()["event"] != "login" {
		t.Errorf("Expected event metadata")
	}
}

func TestWinSignal(t *testing.T) {
	timestamp := time.Now()
	playerCtx := &signal.PlayerContext{
		UserID: "user123",
		State:  &state.ChurnState{},
	}

	sig := NewWinSignal("user123", timestamp, 42, playerCtx)

	if sig.Type() != TypeMatchWin {
		t.Errorf("Expected type '%s', got '%s'", TypeMatchWin, sig.Type())
	}

	if sig.TotalWins != 42 {
		t.Errorf("Expected TotalWins 42, got %d", sig.TotalWins)
	}

	if sig.Metadata()["total_wins"] != 42 {
		t.Errorf("Expected metadata total_wins=42")
	}

	if sig.Metadata()["stat_code"] != "rse-match-wins" {
		t.Errorf("Expected stat_code in metadata")
	}
}

func TestRageQuitSignal(t *testing.T) {
	timestamp := time.Now()
	playerCtx := &signal.PlayerContext{
		UserID: "user123",
		State:  &state.ChurnState{},
	}

	sig := NewRageQuitSignal("user123", timestamp, 5, playerCtx)

	if sig.Type() != TypeRageQuit {
		t.Errorf("Expected type '%s', got '%s'", TypeRageQuit, sig.Type())
	}

	if sig.QuitCount != 5 {
		t.Errorf("Expected QuitCount 5, got %d", sig.QuitCount)
	}

	if sig.Metadata()["quit_count"] != 5 {
		t.Errorf("Expected metadata quit_count=5")
	}

	if sig.MatchContext == nil {
		t.Error("Expected non-nil MatchContext")
	}
}

func TestLossSignal(t *testing.T) {
	timestamp := time.Now()
	playerCtx := &signal.PlayerContext{
		UserID: "user123",
		State:  &state.ChurnState{},
	}

	sig := NewLosingStreakSignal("user123", timestamp, 7, playerCtx)

	if sig.Type() != TypeLosingStreak {
		t.Errorf("Expected type '%s', got '%s'", TypeLosingStreak, sig.Type())
	}

	if sig.CurrentStreak != 7 {
		t.Errorf("Expected CurrentStreak 7, got %d", sig.CurrentStreak)
	}

	if sig.Metadata()["current_streak"] != 7 {
		t.Errorf("Expected metadata current_streak=7")
	}
}
