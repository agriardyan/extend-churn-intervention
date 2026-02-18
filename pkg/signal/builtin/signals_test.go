package builtin

import (
	"testing"
	"time"

	"github.com/AccelByte/extend-churn-intervention/pkg/service"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
)

func TestLoginSignal(t *testing.T) {
	timestamp := time.Now()
	playerCtx := &signal.PlayerContext{
		UserID: "user123",
		State:  &service.ChurnState{},
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

func TestRageQuitSignal(t *testing.T) {
	timestamp := time.Now()
	playerCtx := &signal.PlayerContext{
		UserID: "user123",
		State:  &service.ChurnState{},
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
		State:  &service.ChurnState{},
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
