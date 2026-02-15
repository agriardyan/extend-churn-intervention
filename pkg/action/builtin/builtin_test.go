package builtin

import (
	"context"
	"testing"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
)

// mockStateStore is a mock implementation for testing
type mockStateStore struct {
	updateCalled bool
	updateError  error
}

func (m *mockStateStore) GetChurnState(ctx context.Context, userID string) (*service.ChurnState, error) {
	return &service.ChurnState{}, nil
}

func (m *mockStateStore) UpdateChurnState(ctx context.Context, userID string, state *service.ChurnState) error {
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

// mockUserStatUpdater is a mock implementation for testing
type mockUserStatUpdater struct {
	updateCalled bool
	updateError  error
}

func (m *mockUserStatUpdater) UpdateStatComebackChallenge(ctx context.Context, userID string) error {
	m.updateCalled = true
	return m.updateError
}

func TestComebackChallengeAction_Execute(t *testing.T) {
	mockStore := &mockStateStore{}
	mockStatUpdater := &mockUserStatUpdater{}
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

	act := NewDispatchComebackChallengeAction(config, mockStore, mockStatUpdater)

	playerState := &service.ChurnState{
		SignalHistory:       []service.ChurnSignal{},
		InterventionHistory: []service.InterventionRecord{},
		Cooldown: service.CooldownState{
			InterventionCounts: make(map[string]int),
			LastSignalAt:       make(map[string]time.Time),
		},
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

	// Check intervention was recorded
	if len(playerState.InterventionHistory) != 1 {
		t.Errorf("Expected 1 intervention, got %d", len(playerState.InterventionHistory))
	}

	intervention := playerState.InterventionHistory[0]
	if intervention.Type != ComebackChallengeActionID {
		t.Errorf("Expected intervention type %s, got %s", ComebackChallengeActionID, intervention.Type)
	}

	if intervention.TriggeredBy != "rage_quit" {
		t.Errorf("Expected triggered by 'rage_quit', got '%s'", intervention.TriggeredBy)
	}

	if intervention.Outcome != "active" {
		t.Errorf("Expected outcome 'active', got '%s'", intervention.Outcome)
	}

	// Check cooldown was set
	if !playerState.Cooldown.IsOnCooldown() {
		t.Error("Expected cooldown to be set")
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

	act := NewDispatchComebackChallengeAction(config, mockStore, &mockUserStatUpdater{})

	// Create state with an active comeback challenge intervention
	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)
	playerState := &service.ChurnState{
		SignalHistory: []service.ChurnSignal{},
		InterventionHistory: []service.InterventionRecord{
			{
				ID:          "existing-intervention",
				Type:        ComebackChallengeActionID,
				TriggeredBy: "previous-rule",
				TriggeredAt: now,
				ExpiresAt:   &expiresAt,
				Outcome:     "active",
			},
		},
		Cooldown: service.CooldownState{
			InterventionCounts: make(map[string]int),
		},
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
		t.Error("Did not expect state store update when comeback challenge already active")
	}

	// Verify no new intervention was added
	if len(playerState.InterventionHistory) != 1 {
		t.Errorf("Expected only 1 intervention (existing), got %d", len(playerState.InterventionHistory))
	}
}

func TestComebackChallengeAction_Execute_NoPlayerContext(t *testing.T) {
	mockStore := &mockStateStore{}
	config := action.ActionConfig{
		ID:      "test_challenge",
		Type:    ComebackChallengeActionID,
		Enabled: true,
	}

	act := NewDispatchComebackChallengeAction(config, mockStore, &mockUserStatUpdater{})

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

	act := NewDispatchComebackChallengeAction(config, mockStore, &mockUserStatUpdater{})

	now := time.Now()
	expiresAt := now.Add(7 * 24 * time.Hour)
	playerState := &service.ChurnState{
		SignalHistory: []service.ChurnSignal{},
		InterventionHistory: []service.InterventionRecord{
			{
				ID:          "active-intervention",
				Type:        ComebackChallengeActionID,
				TriggeredBy: "rage_quit",
				TriggeredAt: now,
				ExpiresAt:   &expiresAt,
				Outcome:     "active",
			},
		},
		Cooldown: service.CooldownState{
			CooldownUntil:      now.Add(48 * time.Hour),
			InterventionCounts: map[string]int{ComebackChallengeActionID: 1},
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

	// Check intervention was marked as failed
	if len(playerState.InterventionHistory) != 1 {
		t.Fatalf("Expected 1 intervention, got %d", len(playerState.InterventionHistory))
	}

	intervention := playerState.InterventionHistory[0]
	if intervention.Outcome != "failed" {
		t.Errorf("Expected intervention outcome 'failed', got '%s'", intervention.Outcome)
	}

	// Check cooldown was reset
	if playerState.Cooldown.IsOnCooldown() {
		t.Error("Expected cooldown to be reset after rollback")
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
		State:  &service.ChurnState{},
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
		State:  &service.ChurnState{},
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
		State:  &service.ChurnState{},
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
		State:  &service.ChurnState{},
	}

	err := act.Rollback(context.Background(), trigger, playerCtx)
	if err != action.ErrRollbackNotSupported {
		t.Errorf("Expected ErrRollbackNotSupported, got %v", err)
	}
}
