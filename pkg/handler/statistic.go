package handler

import (
	"context"

	"github.com/AccelByte/extends-anti-churn/pkg/common"
	pb_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/pipeline"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Statistic listens for statistic update events
type Statistic struct {
	pb_social.UnimplementedStatisticStatItemUpdatedServiceServer

	pipelineManager *pipeline.Manager
	namespace       string
}

// NewStatistic creates a new Statistic event listener
func NewStatistic(pipelineManager *pipeline.Manager, namespace string) *Statistic {
	return &Statistic{
		pipelineManager: pipelineManager,
		namespace:       namespace,
	}
}

// OnMessage handles statItemUpdated events
// This is called when a player's statistic is updated (rage quits, wins, losing streak)
func (s *Statistic) OnMessage(
	ctx context.Context,
	msg *pb_social.StatItemUpdated,
) (*emptypb.Empty, error) {
	scope := common.GetScopeFromContext(ctx, "Statistic.OnMessage")
	defer scope.Finish()

	// Extract fields from the event
	userID := msg.GetUserId()
	if userID == "" {
		userID = msg.GetPayload().GetUserId()
		if userID == "" {
			logrus.Warnf("received statistic event with empty user_id")
			return &emptypb.Empty{}, nil
		}
	}

	statCode := msg.GetPayload().GetStatCode()
	latestValue := msg.GetPayload().GetLatestValue()

	logrus.Infof("received stat update event: userId=%s statCode=%s value=%v namespace=%s",
		userID, statCode, latestValue, msg.GetNamespace())

	// Process event through pipeline
	if err := s.pipelineManager.ProcessStatEvent(ctx, msg); err != nil {
		logrus.Errorf("pipeline processing failed for user %s: %v", userID, err)
		return &emptypb.Empty{}, status.Errorf(codes.Internal,
			"pipeline processing failed: %v", err)
	}

	logrus.Infof("successfully processed stat update for user %s: %s=%v",
		userID, statCode, latestValue)
	return &emptypb.Empty{}, nil
}
