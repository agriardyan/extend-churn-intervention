package handler

import (
	"context"
	"testing"

	pb_iam "github.com/AccelByte/extend-churn-intervention/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	"github.com/alicebob/miniredis/v2"
)

func TestOAuth_OnMessage_NewPlayer(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewOAuth(pipelineManager, "test-namespace")
	ctx := context.Background()

	msg := &pb_iam.OauthTokenGenerated{
		UserId:    "test-user-123",
		Namespace: "test-namespace",
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Event should be processed without error
	// Session tracking and rule evaluation happen in the pipeline
}

func TestOAuth_OnMessage_ChurningPlayer(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewOAuth(pipelineManager, "test-namespace")
	ctx := context.Background()

	userID := "test-user-churning"

	msg := &pb_iam.OauthTokenGenerated{
		UserId:    userID,
		Namespace: "test-namespace",
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Event should be processed without error
	// Pipeline will evaluate session decline rules
}

func TestOAuth_OnMessage_WeeklyReset(t *testing.T) {
	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("failed to start miniredis: %v", err)
	}
	defer mr.Close()

	pipelineManager := setupTestPipeline("test-namespace", mr)
	listener := NewOAuth(pipelineManager, "test-namespace")
	ctx := context.Background()

	userID := "test-user-reset"

	msg := &pb_iam.OauthTokenGenerated{
		UserId:    userID,
		Namespace: "test-namespace",
	}

	_, err = listener.OnMessage(ctx, msg)
	if err != nil {
		t.Fatalf("OnMessage() error = %v", err)
	}

	// Event should be processed without error
	// Weekly reset logic would be handled by rules if configured
}
