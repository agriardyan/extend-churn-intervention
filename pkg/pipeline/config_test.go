package pipeline

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	// Create a temporary config file
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pipeline.yaml")

	configContent := `
rules:
  - id: rage-quit
    type: rage_quit
    enabled: true
    parameters:
      threshold: 3
    cooldown:
      duration: 24h
      per_user: true

  - id: losing-streak
    type: losing_streak
    enabled: true
    parameters:
      threshold: 5

actions:
  - id: comeback-challenge
    type: comeback_challenge
    enabled: true
    parameters:
      wins_needed: 3
      duration_days: 7

  - id: grant-item
    type: grant_item
    enabled: true
    parameters:
      item_id: "COMEBACK_REWARD"
      quantity: 1
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	// Load config
	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Validate rules
	if len(config.Rules) != 2 {
		t.Errorf("expected 2 rules, got %d", len(config.Rules))
	}

	if config.Rules[0].ID != "rage-quit" {
		t.Errorf("expected first rule ID 'rage-quit', got '%s'", config.Rules[0].ID)
	}

	if config.Rules[0].Type != "rage_quit" {
		t.Errorf("expected first rule type 'rage_quit', got '%s'", config.Rules[0].Type)
	}

	if !config.Rules[0].Enabled {
		t.Error("expected first rule to be enabled")
	}

	if threshold, ok := config.Rules[0].Parameters["threshold"].(int); !ok || threshold != 3 {
		t.Errorf("expected threshold parameter to be 3, got %v", config.Rules[0].Parameters["threshold"])
	}

	// Validate actions
	if len(config.Actions) != 2 {
		t.Errorf("expected 2 actions, got %d", len(config.Actions))
	}

	if config.Actions[0].ID != "comeback-challenge" {
		t.Errorf("expected first action ID 'comeback-challenge', got '%s'", config.Actions[0].ID)
	}
}

func TestLoadConfig_EnvVarExpansion(t *testing.T) {
	// Set environment variable
	os.Setenv("TEST_THRESHOLD", "10")
	defer os.Unsetenv("TEST_THRESHOLD")

	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pipeline.yaml")

	configContent := `
rules:
  - id: test-rule
    type: rage_quit
    enabled: true
    parameters:
      threshold: ${TEST_THRESHOLD}

actions: []
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Note: YAML unmarshaling will interpret "10" as int
	if threshold, ok := config.Rules[0].Parameters["threshold"].(int); !ok || threshold != 10 {
		t.Errorf("expected threshold to be 10, got %v (type: %T)", config.Rules[0].Parameters["threshold"], config.Rules[0].Parameters["threshold"])
	}
}

func TestLoadConfig_EnvVarDefault(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pipeline.yaml")

	configContent := `
rules:
  - id: test-rule
    type: rage_quit
    enabled: true
    parameters:
      threshold: ${NONEXISTENT_VAR:5}

actions: []
`

	if err := os.WriteFile(configPath, []byte(configContent), 0644); err != nil {
		t.Fatalf("failed to write test config: %v", err)
	}

	config, err := LoadConfig(configPath)
	if err != nil {
		t.Fatalf("LoadConfig failed: %v", err)
	}

	// Default value should be used
	if threshold, ok := config.Rules[0].Parameters["threshold"].(int); !ok || threshold != 5 {
		t.Errorf("expected threshold to be 5, got %v", config.Rules[0].Parameters["threshold"])
	}
}

func TestValidate_DuplicateRuleID(t *testing.T) {
	config := &Config{
		Rules: []RuleConfig{
			{ID: "duplicate", Type: "rage_quit", Enabled: true},
			{ID: "duplicate", Type: "losing_streak", Enabled: true},
		},
		Actions: []ActionConfig{},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected validation error for duplicate rule ID")
	}
}

func TestValidate_DuplicateActionID(t *testing.T) {
	config := &Config{
		Rules: []RuleConfig{},
		Actions: []ActionConfig{
			{ID: "duplicate", Type: "comeback_challenge", Enabled: true},
			{ID: "duplicate", Type: "grant_item", Enabled: true},
		},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected validation error for duplicate action ID")
	}
}

func TestValidate_EmptyRuleID(t *testing.T) {
	config := &Config{
		Rules: []RuleConfig{
			{ID: "", Type: "rage_quit", Enabled: true},
		},
		Actions: []ActionConfig{},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected validation error for empty rule ID")
	}
}

func TestValidate_EmptyRuleType(t *testing.T) {
	config := &Config{
		Rules: []RuleConfig{
			{ID: "test", Type: "", Enabled: true},
		},
		Actions: []ActionConfig{},
	}

	err := config.Validate()
	if err == nil {
		t.Error("expected validation error for empty rule type")
	}
}
