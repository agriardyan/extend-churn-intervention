package handler

import (
	"context"
	"testing"

	pb_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/alicebob/miniredis/v2"
)

func TestStatistic_OnMessage_ProcessesEvent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewStatistic(pipelineManager, "test-namespace")
	ctx := context.Background()

	userID := "test-user-stat"

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    "match_result",
			LatestValue: 1.0,
		},
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Event should be processed without error
	// Actual rule/action behavior is tested in integration tests
}

func TestStatistic_OnMessage_RageQuitEvent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewStatistic(pipelineManager, "test-namespace")
	ctx := context.Background()

	userID := "test-user-rage"

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    "rage_quit_count",
			LatestValue: 3.0, // Rage quit threshold
		},
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Event should be processed without error
	// Rules will be evaluated by the pipeline
}

func TestStatistic_OnMessage_WinEvent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewStatistic(pipelineManager, "test-namespace")
	ctx := context.Background()

	userID := "test-user-win"

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    "match_result",
			LatestValue: 1.0, // Win
		},
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Event should be processed without error
}

func TestStatistic_OnMessage_LossEvent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewStatistic(pipelineManager, "test-namespace")
	ctx := context.Background()

	userID := "test-user-loss"

	msg := &pb_social.StatItemUpdated{
		UserId:    userID,
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    "match_result",
			LatestValue: 0.0, // Loss
		},
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Event should be processed without error
}

func TestStatistic_OnMessage_InvalidEvent(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewStatistic(pipelineManager, "test-namespace")
	ctx := context.Background()

	// Event with nil payload should be handled gracefully
	msg := &pb_social.StatItemUpdated{
		UserId:    "test-user-nil",
		Namespace: "test-namespace",
		Payload:   nil,
	}

	_, err = listener.OnMessage(ctx, msg)
	// Should not panic, may return error
	// Error handling is acceptable for invalid events
	_ = err
}

func TestStatistic_OnMessage_EmptyStatCode(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewStatistic(pipelineManager, "test-namespace")
	ctx := context.Background()

	msg := &pb_social.StatItemUpdated{
		UserId:    "test-user-empty",
		Namespace: "test-namespace",
		Payload: &pb_social.StatItem{
			StatCode:    "",
			LatestValue: 1.0,
		},
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		// Empty stat code might cause error, which is acceptable
		t.Logf("Expected error for empty stat code: %v", err)
	}
}
