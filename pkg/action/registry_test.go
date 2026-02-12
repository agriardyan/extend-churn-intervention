package action

import (
	"context"
	"testing"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// mockAction is a simple action implementation for testing
type mockAction struct {
	id     string
	name   string
	config ActionConfig
}

func (m *mockAction) ID() string   { return m.id }
func (m *mockAction) Name() string { return m.name }
func (m *mockAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	return nil
}
func (m *mockAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	return ErrRollbackNotSupported
}
func (m *mockAction) Config() ActionConfig { return m.config }

func TestNewRegistry(t *testing.T) {
	registry := NewRegistry()

	if registry == nil {
		t.Fatal("Expected non-nil registry")
	}

	if registry.Count() != 0 {
		t.Errorf("Expected empty registry, got count %d", registry.Count())
	}
}

func TestRegistry_Register(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{
		id:     "test_action",
		name:   "Test Action",
		config: ActionConfig{ID: "test_action", Enabled: true},
	}

	err := registry.Register(action)
	if err != nil {
		t.Fatalf("Failed to register action: %v", err)
	}

	if registry.Count() != 1 {
		t.Errorf("Expected count 1, got %d", registry.Count())
	}

	// Try to register same action again
	err = registry.Register(action)
	if err == nil {
		t.Error("Expected error when registering duplicate action")
	}
}

func TestRegistry_Get(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{
		id:     "test_action",
		name:   "Test Action",
		config: ActionConfig{ID: "test_action", Enabled: true},
	}

	registry.Register(action)

	retrieved := registry.Get("test_action")
	if retrieved == nil {
		t.Fatal("Expected to retrieve action")
	}

	if retrieved.ID() != "test_action" {
		t.Errorf("Expected action ID 'test_action', got '%s'", retrieved.ID())
	}

	// Try to get non-existent action
	notFound := registry.Get("non_existent")
	if notFound != nil {
		t.Error("Expected nil for non-existent action")
	}
}

func TestRegistry_GetEnabled(t *testing.T) {
	registry := NewRegistry()

	enabledAction := &mockAction{
		id:     "enabled_action",
		name:   "Enabled Action",
		config: ActionConfig{ID: "enabled_action", Enabled: true},
	}
	disabledAction := &mockAction{
		id:     "disabled_action",
		name:   "Disabled Action",
		config: ActionConfig{ID: "disabled_action", Enabled: false},
	}

	registry.Register(enabledAction)
	registry.Register(disabledAction)

	// Should get enabled action
	retrieved := registry.GetEnabled("enabled_action")
	if retrieved == nil {
		t.Error("Expected to retrieve enabled action")
	}

	// Should not get disabled action
	notRetrieved := registry.GetEnabled("disabled_action")
	if notRetrieved != nil {
		t.Error("Expected nil for disabled action")
	}
}

func TestRegistry_Unregister(t *testing.T) {
	registry := NewRegistry()
	action := &mockAction{
		id:     "test_action",
		name:   "Test Action",
		config: ActionConfig{ID: "test_action", Enabled: true},
	}

	registry.Register(action)

	err := registry.Unregister("test_action")
	if err != nil {
		t.Fatalf("Failed to unregister action: %v", err)
	}

	if registry.Count() != 0 {
		t.Errorf("Expected count 0 after unregister, got %d", registry.Count())
	}

	// Try to unregister non-existent action
	err = registry.Unregister("non_existent")
	if err == nil {
		t.Error("Expected error when unregistering non-existent action")
	}
}

func TestRegistry_GetAll(t *testing.T) {
	registry := NewRegistry()

	action1 := &mockAction{id: "action1", config: ActionConfig{ID: "action1", Enabled: true}}
	action2 := &mockAction{id: "action2", config: ActionConfig{ID: "action2", Enabled: true}}
	action3 := &mockAction{id: "action3", config: ActionConfig{ID: "action3", Enabled: false}}

	registry.Register(action1)
	registry.Register(action2)
	registry.Register(action3)

	allActions := registry.GetAll()
	if len(allActions) != 3 {
		t.Errorf("Expected 3 actions, got %d", len(allActions))
	}
}

func TestRegistry_GetAllEnabled(t *testing.T) {
	registry := NewRegistry()

	action1 := &mockAction{id: "action1", config: ActionConfig{ID: "action1", Enabled: true}}
	action2 := &mockAction{id: "action2", config: ActionConfig{ID: "action2", Enabled: true}}
	action3 := &mockAction{id: "action3", config: ActionConfig{ID: "action3", Enabled: false}}

	registry.Register(action1)
	registry.Register(action2)
	registry.Register(action3)

	enabledActions := registry.GetAllEnabled()
	if len(enabledActions) != 2 {
		t.Errorf("Expected 2 enabled actions, got %d", len(enabledActions))
	}
}

func TestActionConfig_GetParameterHelpers(t *testing.T) {
	config := ActionConfig{
		Parameters: map[string]interface{}{
			"int_value":    42,
			"float_value":  3.14,
			"string_value": "test",
			"bool_value":   true,
			"slice_value":  []string{"a", "b", "c"},
		},
	}

	// Test GetParameterInt
	if val := config.GetParameterInt("int_value", 0); val != 42 {
		t.Errorf("Expected int 42, got %d", val)
	}
	if val := config.GetParameterInt("missing", 99); val != 99 {
		t.Errorf("Expected default 99, got %d", val)
	}

	// Test GetParameterFloat
	if val := config.GetParameterFloat("float_value", 0.0); val != 3.14 {
		t.Errorf("Expected float 3.14, got %f", val)
	}

	// Test GetParameterString
	if val := config.GetParameterString("string_value", ""); val != "test" {
		t.Errorf("Expected string 'test', got '%s'", val)
	}

	// Test GetParameterBool
	if val := config.GetParameterBool("bool_value", false); val != true {
		t.Errorf("Expected bool true, got %v", val)
	}

	// Test GetParameterStringSlice
	slice := config.GetParameterStringSlice("slice_value", nil)
	if len(slice) != 3 {
		t.Errorf("Expected slice length 3, got %d", len(slice))
	}
	if slice[0] != "a" || slice[1] != "b" || slice[2] != "c" {
		t.Errorf("Expected slice [a b c], got %v", slice)
	}
}
