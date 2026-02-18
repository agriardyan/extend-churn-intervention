package handler

import (
	"context"

	"github.com/AccelByte/extend-churn-intervention/pkg/common"
	pb_iam "github.com/AccelByte/extend-churn-intervention/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	"github.com/AccelByte/extend-churn-intervention/pkg/pipeline"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// OAuth listens for OAuth token generated events
type OAuth struct {
	pb_iam.UnimplementedOauthTokenOauthTokenGeneratedServiceServer

	pipelineManager *pipeline.Manager
	namespace       string
}

// NewOAuth creates a new OAuth event listener
func NewOAuth(pipelineManager *pipeline.Manager, namespace string) *OAuth {
	return &OAuth{
		pipelineManager: pipelineManager,
		namespace:       namespace,
	}
}

// OnMessage handles oauthTokenGenerated events
// This is called when a player logs in and gets a new OAuth token
func (s *OAuth) OnMessage(
	ctx context.Context,
	msg *pb_iam.OauthTokenGenerated,
) (*emptypb.Empty, error) {
	scope := common.GetScopeFromContext(ctx, "OAuth.OnMessage")
	defer scope.Finish()

	// Extract user ID from the event
	userID := msg.GetUserId()
	if userID == "" {
		logrus.Warnf("received OAuth event with empty user_id")
		return &emptypb.Empty{}, nil
	}

	logrus.Infof("received OAuth token generated event: userId=%s namespace=%s",
		userID, msg.GetNamespace())

	// Process event through pipeline
	if err := s.pipelineManager.ProcessOAuthEvent(ctx, msg); err != nil {
		logrus.Errorf("pipeline processing failed for user %s: %v", userID, err)
		return &emptypb.Empty{}, status.Errorf(codes.Internal,
			"pipeline processing failed: %v", err)
	}

	logrus.Infof("successfully processed OAuth event for user %s", userID)
	return &emptypb.Empty{}, nil
}
