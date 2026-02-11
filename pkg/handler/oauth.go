package handler

import (
	"context"
	"time"

	"github.com/AccelByte/extends-anti-churn/pkg/common"
	pb_iam "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
	"github.com/go-redis/redis/v8"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// OAuth listens for OAuth token generated events
type OAuth struct {
	pb_iam.UnimplementedOauthTokenOauthTokenGeneratedServiceServer

	redisClient *redis.Client
	namespace   string
}

// NewOAuth creates a new OAuth event listener
func NewOAuth(redisClient *redis.Client, namespace string) *OAuth {
	return &OAuth{
		redisClient: redisClient,
		namespace:   namespace,
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

	// Get current churn state
	churnState, err := state.GetChurnState(ctx, s.redisClient, userID)
	if err != nil {
		logrus.Errorf("failed to get churn state for user %s: %v", userID, err)
		return &emptypb.Empty{}, status.Errorf(codes.Internal,
			"failed to get churn state: %v", err)
	}

	now := time.Now()

	// Check if weekly reset is needed
	if state.CheckWeeklyReset(churnState, now) {
		logrus.Infof("weekly reset performed for user %s", userID)
	}

	// Increment this week's session count
	churnState.Sessions.ThisWeek++
	logrus.Debugf("incremented session count for user %s: thisWeek=%d, lastWeek=%d",
		userID, churnState.Sessions.ThisWeek, churnState.Sessions.LastWeek)

	// Check if player is churning and intervention should be triggered
	if state.ShouldTriggerIntervention(churnState, now) {
		logrus.Infof("triggering intervention for churning player: userId=%s", userID)

		// For session-based churn, we don't have a current wins count yet
		// We'll need to fetch it from AGS Statistics in Phase 4
		// For now, we'll create the challenge with placeholder values
		expiresAt := now.Add(ChallengeDurationDays * 24 * time.Hour)
		state.CreateChallenge(churnState, ChallengeWinsNeeded, 0, expiresAt, "session_decline")

		// Set cooldown period
		cooldownDuration := InterventionCooldownHours * time.Hour
		state.SetInterventionCooldown(churnState, now, cooldownDuration)

		logrus.Infof("created challenge for user %s: winsNeeded=%d, expiresAt=%v",
			userID, ChallengeWinsNeeded, expiresAt)
	}

	// Save updated state
	if err := state.UpdateChurnState(ctx, s.redisClient, userID, churnState); err != nil {
		logrus.Errorf("failed to update churn state for user %s: %v", userID, err)
		return &emptypb.Empty{}, status.Errorf(codes.Internal,
			"failed to update churn state: %v", err)
	}

	logrus.Infof("successfully processed OAuth event for user %s", userID)
	return &emptypb.Empty{}, nil
}
