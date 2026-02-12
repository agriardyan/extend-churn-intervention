package pipeline

import "testing"

func TestNewPipeline(t *testing.T) {
	pipeline := NewPipeline("test_pipeline")

	if pipeline == nil {
		t.Fatal("Expected non-nil pipeline")
	}

	if pipeline.Name != "test_pipeline" {
		t.Errorf("Expected name 'test_pipeline', got '%s'", pipeline.Name)
	}

	if pipeline.Actions == nil {
		t.Error("Expected non-nil Actions map")
	}
}

func TestPipeline_AddRule(t *testing.T) {
	pipeline := NewPipeline("test")

	pipeline.AddRule("rule1").AddRule("rule2")

	if len(pipeline.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(pipeline.Rules))
	}

	if pipeline.Rules[0] != "rule1" || pipeline.Rules[1] != "rule2" {
		t.Errorf("Expected rules [rule1 rule2], got %v", pipeline.Rules)
	}
}

func TestPipeline_AddActions(t *testing.T) {
	pipeline := NewPipeline("test")

	pipeline.AddActions("rule1", "action1", "action2")
	pipeline.AddActions("rule2", "action3")

	actions1 := pipeline.GetActions("rule1")
	if len(actions1) != 2 {
		t.Errorf("Expected 2 actions for rule1, got %d", len(actions1))
	}

	actions2 := pipeline.GetActions("rule2")
	if len(actions2) != 1 {
		t.Errorf("Expected 1 action for rule2, got %d", len(actions2))
	}

	// Test adding more actions to existing rule
	pipeline.AddActions("rule1", "action4")
	actions1 = pipeline.GetActions("rule1")
	if len(actions1) != 3 {
		t.Errorf("Expected 3 actions for rule1 after adding, got %d", len(actions1))
	}
}

func TestPipeline_GetActions(t *testing.T) {
	pipeline := NewPipeline("test")

	pipeline.AddActions("rule1", "action1", "action2")

	// Test existing rule
	actions := pipeline.GetActions("rule1")
	if len(actions) != 2 {
		t.Errorf("Expected 2 actions, got %d", len(actions))
	}

	// Test non-existent rule
	noActions := pipeline.GetActions("non_existent")
	if noActions != nil {
		t.Errorf("Expected nil for non-existent rule, got %v", noActions)
	}
}

func TestPipeline_SetConcurrent(t *testing.T) {
	pipeline := NewPipeline("test")

	if pipeline.Concurrent {
		t.Error("Expected Concurrent to be false by default")
	}

	pipeline.SetConcurrent(true)

	if !pipeline.Concurrent {
		t.Error("Expected Concurrent to be true after SetConcurrent(true)")
	}
}

func TestPipeline_Chaining(t *testing.T) {
	// Test that all methods return the pipeline for chaining
	pipeline := NewPipeline("test").
		AddRule("rule1").
		AddRule("rule2").
		AddActions("rule1", "action1", "action2").
		SetConcurrent(true)

	if len(pipeline.Rules) != 2 {
		t.Errorf("Expected 2 rules, got %d", len(pipeline.Rules))
	}

	if len(pipeline.GetActions("rule1")) != 2 {
		t.Errorf("Expected 2 actions for rule1, got %d", len(pipeline.GetActions("rule1")))
	}

	if !pipeline.Concurrent {
		t.Error("Expected Concurrent to be true")
	}
}
