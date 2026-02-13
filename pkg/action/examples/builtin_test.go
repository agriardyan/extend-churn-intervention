package examples

import (
	"context"
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
)

// mockStateStore is a mock implementation for testing
type mockStateStore struct {
	updateCalled bool
	updateError  error
}

func (m *mockStateStore) GetChurnState(ctx context.Context, userID string) (*state.ChurnState, error) {
	return &state.ChurnState{}, nil
}

func (m *mockStateStore) UpdateChurnState(ctx context.Context, userID string, state *state.ChurnState) error {
	m.updateCalled = true
	return m.updateError
}

// mockEntitlementGranter is a mock implementation for testing
type mockEntitlementGranter struct {
	grantCalled bool
	grantError  error
	lastItemID  string
	lastQty     int32
}

func (m *mockEntitlementGranter) GrantEntitlement(ctx context.Context, userID, itemID string, quantity int) error {
	m.grantCalled = true
	m.lastItemID = itemID
	m.lastQty = int32(quantity)
	return m.grantError
}

func TestComebackChallengeAction_Execute(t *testing.T) {
	mockStore := &mockStateStore{}
	config := action.ActionConfig{
		ID:      "test_challenge",
		Type:    ComebackChallengeActionID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"wins_needed":    3,
			"duration_days":  7,
			"cooldown_hours": 48,
		},
	}

	act := NewComebackChallengeAction(config, mockStore)

	playerState := &state.ChurnState{
		Challenge: state.ChallengeState{Active: false},
	}
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  playerState,
	}

	trigger := rule.NewTrigger("rage_quit", "test-user", "rage quit detected", 10)
	trigger.Metadata["current_wins"] = 5

	err := act.Execute(context.Background(), trigger, playerCtx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !playerState.Challenge.Active {
		t.Error("Expected challenge to be active")
	}

	if playerState.Challenge.WinsNeeded != 3 {
		t.Errorf("Expected wins needed 3, got %d", playerState.Challenge.WinsNeeded)
	}

	if playerState.Challenge.WinsAtStart != 5 {
		t.Errorf("Expected wins at start 5, got %d", playerState.Challenge.WinsAtStart)
	}

	if playerState.Challenge.TriggerReason != "rage_quit" {
		t.Errorf("Expected trigger reason 'rage_quit', got '%s'", playerState.Challenge.TriggerReason)
	}

	if !mockStore.updateCalled {
		t.Error("Expected state store update to be called")
	}
}

func TestComebackChallengeAction_Execute_ChallengeAlreadyActive(t *testing.T) {
	mockStore := &mockStateStore{}
	config := action.ActionConfig{
		ID:      "test_challenge",
		Type:    ComebackChallengeActionID,
		Enabled: true,
	}

	act := NewComebackChallengeAction(config, mockStore)

	playerState := &state.ChurnState{
		Challenge: state.ChallengeState{Active: true}, // Already active
	}
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  playerState,
	}

	trigger := rule.NewTrigger("rage_quit", "test-user", "rage quit detected", 10)

	err := act.Execute(context.Background(), trigger, playerCtx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if mockStore.updateCalled {
		t.Error("Did not expect state store update when challenge already active")
	}
}

func TestComebackChallengeAction_Execute_NoPlayerContext(t *testing.T) {
	mockStore := &mockStateStore{}
	config := action.ActionConfig{
		ID:      "test_challenge",
		Type:    ComebackChallengeActionID,
		Enabled: true,
	}

	act := NewComebackChallengeAction(config, mockStore)

	trigger := rule.NewTrigger("rage_quit", "test-user", "rage quit detected", 10)

	err := act.Execute(context.Background(), trigger, nil)
	if err != action.ErrMissingPlayerContext {
		t.Errorf("Expected ErrMissingPlayerContext, got %v", err)
	}
}

func TestComebackChallengeAction_Rollback(t *testing.T) {
	mockStore := &mockStateStore{}
	config := action.ActionConfig{
		ID:      "test_challenge",
		Type:    ComebackChallengeActionID,
		Enabled: true,
	}

	act := NewComebackChallengeAction(config, mockStore)

	playerState := &state.ChurnState{
		Challenge: state.ChallengeState{
			Active:        true,
			TriggerReason: "rage_quit",
		},
		Intervention: state.InterventionState{
			TotalTriggered: 1,
			CooldownUntil:  time.Now().Add(48 * time.Hour),
		},
	}
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  playerState,
	}

	trigger := rule.NewTrigger("rage_quit", "test-user", "rage quit detected", 10)

	err := act.Rollback(context.Background(), trigger, playerCtx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if playerState.Challenge.Active {
		t.Error("Expected challenge to be inactive after rollback")
	}

	if playerState.Intervention.TotalTriggered != 0 {
		t.Errorf("Expected intervention count 0, got %d", playerState.Intervention.TotalTriggered)
	}

	if !mockStore.updateCalled {
		t.Error("Expected state store update to be called")
	}
}

func TestGrantItemAction_Execute(t *testing.T) {
	mockGranter := &mockEntitlementGranter{}
	config := action.ActionConfig{
		ID:      "test_grant",
		Type:    GrantItemActionID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"item_id":  "speed_booster",
			"quantity": 2,
		},
	}

	act := NewGrantItemAction(config, mockGranter, "test-namespace")

	trigger := rule.NewTrigger("challenge_complete", "test-user", "challenge completed", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	err := act.Execute(context.Background(), trigger, playerCtx)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if !mockGranter.grantCalled {
		t.Error("Expected item granter to be called")
	}

	if mockGranter.lastItemID != "speed_booster" {
		t.Errorf("Expected item ID 'speed_booster', got '%s'", mockGranter.lastItemID)
	}

	if mockGranter.lastQty != 2 {
		t.Errorf("Expected quantity 2, got %d", mockGranter.lastQty)
	}
}

func TestGrantItemAction_Execute_TestMode(t *testing.T) {
	config := action.ActionConfig{
		ID:      "test_grant",
		Type:    GrantItemActionID,
		Enabled: true,
		Parameters: map[string]interface{}{
			"item_id":  "speed_booster",
			"quantity": 1,
		},
	}

	// No granter provided (test mode)
	act := NewGrantItemAction(config, nil, "test-namespace")

	trigger := rule.NewTrigger("challenge_complete", "test-user", "challenge completed", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	err := act.Execute(context.Background(), trigger, playerCtx)
	if err != nil {
		t.Fatalf("Unexpected error in test mode: %v", err)
	}
}

func TestGrantItemAction_Execute_NoItemID(t *testing.T) {
	mockGranter := &mockEntitlementGranter{}
	config := action.ActionConfig{
		ID:      "test_grant",
		Type:    GrantItemActionID,
		Enabled: true,
		Parameters: map[string]interface{}{
			// No item_id configured
			"quantity": 1,
		},
	}

	act := NewGrantItemAction(config, mockGranter, "test-namespace")

	trigger := rule.NewTrigger("challenge_complete", "test-user", "challenge completed", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	err := act.Execute(context.Background(), trigger, playerCtx)
	if err == nil {
		t.Error("Expected error when item_id not configured")
	}
}

func TestGrantItemAction_Rollback(t *testing.T) {
	mockGranter := &mockEntitlementGranter{}
	config := action.ActionConfig{
		ID:      "test_grant",
		Type:    GrantItemActionID,
		Enabled: true,
	}

	act := NewGrantItemAction(config, mockGranter, "test-namespace")

	trigger := rule.NewTrigger("challenge_complete", "test-user", "challenge completed", 10)
	playerCtx := &signal.PlayerContext{
		UserID: "test-user",
		State:  &state.ChurnState{},
	}

	err := act.Rollback(context.Background(), trigger, playerCtx)
	if err != action.ErrRollbackNotSupported {
		t.Errorf("Expected ErrRollbackNotSupported, got %v", err)
	}
}
