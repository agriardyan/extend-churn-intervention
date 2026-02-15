package pipeline_test

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"
)

func TestValidate_UnknownActionReference(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pipeline.yaml")

	// Config with rule referencing non-existent action
	configContent := `
rules:
  - id: test-rule
    type: rage_quit
    enabled: true
    actions: [dispatch-comeback-challenge, non-existent-action]
    parameters:
      threshold: 3

actions:
  - id: dispatch-comeback-challenge
    type: dispatch_comeback_challenge
    enabled: true
    parameters:
      wins_needed: 3
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	_, err = pipeline.LoadConfig(configPath)
	if err == nil {
		t.Fatal("expected error for unknown action reference, got nil")
	}

	if !strings.Contains(err.Error(), "unknown action") {
		t.Errorf("expected error about unknown action, got: %v", err)
	}
}

func TestValidate_ValidActionReferences(t *testing.T) {
	tmpDir := t.TempDir()
	configPath := filepath.Join(tmpDir, "pipeline.yaml")

	// Config with valid action references
	configContent := `
rules:
  - id: test-rule
    type: rage_quit
    enabled: true
    actions: [dispatch-comeback-challenge, grant-item]
    parameters:
      threshold: 3

actions:
  - id: dispatch-comeback-challenge
    type: dispatch_comeback_challenge
    enabled: true
    parameters:
      wins_needed: 3
  - id: grant-item
    type: grant_item
    enabled: true
    parameters:
      item_id: TEST_ITEM
`

	err := os.WriteFile(configPath, []byte(configContent), 0644)
	if err != nil {
		t.Fatalf("failed to write config: %v", err)
	}

	config, err := pipeline.LoadConfig(configPath)
	if err != nil {
		t.Fatalf("expected no error, got: %v", err)
	}

	if len(config.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(config.Rules))
	}

	if len(config.Rules[0].Actions) != 2 {
		t.Errorf("expected 2 actions in rule, got %d", len(config.Rules[0].Actions))
	}

	if config.Rules[0].Actions[0] != "dispatch-comeback-challenge" {
		t.Errorf("expected first action to be 'dispatch-comeback-challenge', got %s", config.Rules[0].Actions[0])
	}

	if config.Rules[0].Actions[1] != "grant-item" {
		t.Errorf("expected second action to be 'grant-item', got %s", config.Rules[0].Actions[1])
	}
}
