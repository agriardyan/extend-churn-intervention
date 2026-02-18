package action_test

import (
	"context"
	"testing"

	"github.com/AccelByte/extend-churn-intervention/pkg/action"
	actionBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/action/builtin"
	"github.com/AccelByte/extend-churn-intervention/pkg/service"
)

func init() {
	// Register builtin actions for all tests
	actionBuiltin.RegisterActions(&actionBuiltin.Dependencies{})
}

// mockStateStore for testing
type mockStateStore struct{}

func (m *mockStateStore) GetChurnState(ctx context.Context, userID string) (*service.ChurnState, error) {
	return &service.ChurnState{}, nil
}

func (m *mockStateStore) UpdateChurnState(ctx context.Context, userID string, state *service.ChurnState) error {
	return nil
}

// mockItemGranter for testing
type mockItemGranter struct{}

func (m *mockItemGranter) GrantEntitlement(ctx context.Context, userID, itemID string, quantity int) error {
	return nil
}

func TestCreateAction_ComebackChallenge(t *testing.T) {
	// Register built-in actions
	deps := &actionBuiltin.Dependencies{
		StateStore:         &mockStateStore{},
		EntitlementGranter: &mockItemGranter{},
		Namespace:          "test",
	}
	actionBuiltin.RegisterActions(deps)

	config := action.ActionConfig{
		ID:      "test_challenge",
		Type:    actionBuiltin.ComebackChallengeActionID,
		Enabled: true,
	}

	act, err := action.CreateAction(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if act == nil {
		t.Fatal("Expected non-nil action")
	}

	if act.ID() != config.ID {
		t.Errorf("Expected action ID '%s', got '%s'", config.ID, act.ID())
	}

	if act.Name() != "Create Comeback Challenge" {
		t.Errorf("Expected action name 'Create Comeback Challenge', got '%s'", act.Name())
	}
}

func TestCreateAction_GrantItem(t *testing.T) {
	// Register built-in actions
	deps := &actionBuiltin.Dependencies{
		StateStore:         &mockStateStore{},
		EntitlementGranter: &mockItemGranter{},
		Namespace:          "test",
	}
	actionBuiltin.RegisterActions(deps)

	config := action.ActionConfig{
		ID:      "test_grant",
		Type:    actionBuiltin.GrantItemActionID,
		Enabled: true,
	}

	act, err := action.CreateAction(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if act == nil {
		t.Fatal("Expected non-nil action")
	}

	if act.ID() != config.ID {
		t.Errorf("Expected action ID '%s', got '%s'", config.ID, act.ID())
	}

	if act.Name() != "Grant Item" {
		t.Errorf("Expected action name 'Grant Item', got '%s'", act.Name())
	}
}

func TestCreateAction_Disabled(t *testing.T) {
	config := action.ActionConfig{
		ID:      "disabled_action",
		Type:    actionBuiltin.ComebackChallengeActionID,
		Enabled: false,
	}

	act, err := action.CreateAction(config)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	if act != nil {
		t.Error("Expected nil action for disabled config")
	}
}

func TestCreateAction_UnknownType(t *testing.T) {
	config := action.ActionConfig{
		ID:      "unknown_action",
		Type:    "unknown_type",
		Enabled: true,
	}

	act, err := action.CreateAction(config)
	if err == nil {
		t.Error("Expected error for unknown action type")
	}

	if act != nil {
		t.Error("Expected nil action for unknown type")
	}
}

func TestCreateActions_Multiple(t *testing.T) {
	// Register built-in actions
	deps := &actionBuiltin.Dependencies{
		StateStore:         &mockStateStore{},
		EntitlementGranter: &mockItemGranter{},
		Namespace:          "test",
	}
	actionBuiltin.RegisterActions(deps)

	configs := []action.ActionConfig{
		{
			ID:      "challenge1",
			Type:    actionBuiltin.ComebackChallengeActionID,
			Enabled: true,
		},
		{
			ID:      "grant1",
			Type:    actionBuiltin.GrantItemActionID,
			Enabled: true,
		},
	}

	actions, errors := action.CreateActions(configs)

	if len(errors) != 0 {
		t.Errorf("Expected no errors, got %d", len(errors))
	}

	if len(actions) != 2 {
		t.Fatalf("Expected 2 actions, got %d", len(actions))
	}
}

func TestCreateActions_WithErrors(t *testing.T) {
	// Register built-in actions
	deps := &actionBuiltin.Dependencies{
		StateStore:         &mockStateStore{},
		EntitlementGranter: &mockItemGranter{},
		Namespace:          "test",
	}
	actionBuiltin.RegisterActions(deps)

	configs := []action.ActionConfig{
		{
			ID:      "valid_action",
			Type:    actionBuiltin.ComebackChallengeActionID,
			Enabled: true,
		},
		{
			ID:      "invalid_action",
			Type:    "unknown_type",
			Enabled: true,
		},
		{
			ID:      "disabled_action",
			Type:    actionBuiltin.GrantItemActionID,
			Enabled: false,
		},
	}

	actions, errors := action.CreateActions(configs)

	if len(errors) != 1 {
		t.Errorf("Expected 1 error, got %d", len(errors))
	}

	// Should have 1 valid action (disabled actions return nil, not error)
	if len(actions) != 1 {
		t.Errorf("Expected 1 action, got %d", len(actions))
	}
}

func TestRegisterActions(t *testing.T) {
	// Register built-in actions
	deps := &actionBuiltin.Dependencies{
		StateStore:         &mockStateStore{},
		EntitlementGranter: &mockItemGranter{},
		Namespace:          "test",
	}
	actionBuiltin.RegisterActions(deps)

	registry := action.NewRegistry()

	configs := []action.ActionConfig{
		{
			ID:      "challenge1",
			Type:    actionBuiltin.ComebackChallengeActionID,
			Enabled: true,
		},
		{
			ID:      "grant1",
			Type:    actionBuiltin.GrantItemActionID,
			Enabled: true,
		},
	}

	err := action.RegisterActions(registry, configs)
	if err != nil {
		t.Fatalf("Unexpected error: %v", err)
	}

	// Verify actions are registered
	allActions := registry.GetAll()
	if len(allActions) != 2 {
		t.Errorf("Expected 2 registered actions, got %d", len(allActions))
	}
}

func TestRegisterActions_DuplicateID(t *testing.T) {
	// Register built-in actions
	deps := &actionBuiltin.Dependencies{
		StateStore:         &mockStateStore{},
		EntitlementGranter: &mockItemGranter{},
		Namespace:          "test",
	}
	actionBuiltin.RegisterActions(deps)

	registry := action.NewRegistry()

	configs := []action.ActionConfig{
		{
			ID:      "same_id",
			Type:    actionBuiltin.ComebackChallengeActionID,
			Enabled: true,
		},
		{
			ID:      "same_id", // Duplicate ID
			Type:    actionBuiltin.GrantItemActionID,
			Enabled: true,
		},
	}

	err := action.RegisterActions(registry, configs)
	if err == nil {
		t.Error("Expected error for duplicate action ID")
	}
}
