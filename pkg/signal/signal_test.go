package signal

import (
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

func TestStatUpdateSignal(t *testing.T) {
	timestamp := time.Now()
	playerCtx := &PlayerContext{
		UserID: "user123",
		State:  &state.ChurnState{},
	}

	signal := NewStatUpdateSignal("user123", timestamp, "custom-stat", 99.5, playerCtx)

	if signal.Type() != TypeStatUpdate {
		t.Errorf("Expected type '%s', got '%s'", TypeStatUpdate, signal.Type())
	}

	if signal.StatCode != "custom-stat" {
		t.Errorf("Expected StatCode 'custom-stat', got '%s'", signal.StatCode)
	}

	if signal.Value != 99.5 {
		t.Errorf("Expected Value 99.5, got %f", signal.Value)
	}

	if signal.Metadata()["stat_code"] != "custom-stat" {
		t.Errorf("Expected stat_code in metadata")
	}

	if signal.Metadata()["value"] != 99.5 {
		t.Errorf("Expected value in metadata")
	}
}
