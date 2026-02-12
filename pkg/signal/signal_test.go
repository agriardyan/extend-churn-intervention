package signal

import (
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

func TestBaseSignal(t *testing.T) {
	timestamp := time.Now()
	metadata := map[string]interface{}{
		"test_key": "test_value",
	}
	playerCtx := &PlayerContext{
		UserID:    "user123",
		State:     &state.ChurnState{},
		Namespace: "test-namespace",
	}

	signal := NewBaseSignal("test_type", "user123", timestamp, metadata, playerCtx)

	if signal.Type() != "test_type" {
		t.Errorf("Expected type 'test_type', got '%s'", signal.Type())
	}

	if signal.UserID() != "user123" {
		t.Errorf("Expected userID 'user123', got '%s'", signal.UserID())
	}

	if !signal.Timestamp().Equal(timestamp) {
		t.Errorf("Expected timestamp %v, got %v", timestamp, signal.Timestamp())
	}

	if signal.Metadata()["test_key"] != "test_value" {
		t.Errorf("Expected metadata test_key='test_value', got '%v'", signal.Metadata()["test_key"])
	}

	if signal.Context() != playerCtx {
		t.Errorf("Expected context to match")
	}
}

func TestBaseSignal_NilMetadata(t *testing.T) {
	signal := NewBaseSignal("test", "user1", time.Now(), nil, nil)

	if signal.Metadata() == nil {
		t.Error("Expected non-nil metadata map")
	}
}

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
