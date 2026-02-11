package handler

import (
	"context"
	"fmt"
	"time"

	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/fulfillment"
	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/social"
	"github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclient/user_statistic"
	"github.com/AccelByte/extends-anti-churn/pkg/common"
	pb_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/state"
	"github.com/go-redis/redis/v8"

	"github.com/sirupsen/logrus"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

// Statistic listens for statistic update events
type Statistic struct {
	pb_social.UnimplementedStatisticStatItemUpdatedServiceServer

	fulfillmentService platform.FulfillmentService
	statisticsService  social.UserStatisticService
	redisClient        *redis.Client
	namespace          string
}

// NewStatistic creates a new Statistic event listener
func NewStatistic(
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository,
	redisClient *redis.Client,
	namespace string,
) *Statistic {
	listener := &Statistic{
		redisClient: redisClient,
		namespace:   namespace,
	}

	// Only initialize AGS services if config is provided (for production)
	// Tests can pass nil to skip AGS service initialization
	if configRepo != nil {
		listener.fulfillmentService = platform.FulfillmentService{
			Client:           factory.NewPlatformClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		}
		listener.statisticsService = social.UserStatisticService{
			Client:           factory.NewSocialClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		}
	}

	return listener
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
	statCode := msg.GetPayload().GetStatCode()
	latestValue := msg.GetPayload().GetLatestValue()

	if userID == "" {
		userID = msg.GetPayload().GetUserId()
		if userID == "" {
			logrus.Warnf("received statistic event with empty user_id")
			return &emptypb.Empty{}, nil
		}
	}

	logrus.Infof("received stat update event: userId=%s statCode=%s value=%v namespace=%s",
		userID, statCode, latestValue, msg.GetNamespace())

	// Route to appropriate handler based on stat code
	var err error
	switch statCode {
	case StatCodeRageQuit:
		err = s.handleRageQuit(ctx, userID, int(latestValue))
	case StatCodeMatchWins:
		err = s.handleMatchWin(ctx, userID, int(latestValue))
	case StatCodeLosingStreak:
		err = s.handleLosingStreak(ctx, userID, int(latestValue))
	default:
		// Ignore other stat codes
		logrus.Debugf("ignoring stat code: %s", statCode)
		return &emptypb.Empty{}, nil
	}

	if err != nil {
		logrus.Errorf("failed to handle stat %s for user %s: %v", statCode, userID, err)
		return &emptypb.Empty{}, status.Errorf(codes.Internal,
			"failed to handle stat update: %v", err)
	}

	logrus.Infof("successfully processed stat update for user %s: %s=%v",
		userID, statCode, latestValue)
	return &emptypb.Empty{}, nil
}

// handleRageQuit processes rage quit events
func (s *Statistic) handleRageQuit(ctx context.Context, userID string, rageQuitCount int) error {
	logrus.Infof("handling rage quit for user %s: count=%d", userID, rageQuitCount)

	// Get current churn state
	churnState, err := state.GetChurnState(ctx, s.redisClient, userID)
	if err != nil {
		return err
	}

	now := time.Now()

	// Check if weekly reset is needed
	state.CheckWeeklyReset(churnState, now)

	// Check if rage quit count exceeds threshold
	if rageQuitCount < RageQuitThreshold {
		logrus.Debugf("rage quit count %d below threshold %d, no intervention needed",
			rageQuitCount, RageQuitThreshold)
		// Still save state in case weekly reset occurred
		return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
	}

	// Check if intervention can be triggered (not session-based, just cooldown + no active challenge)
	if churnState.Challenge.Active {
		logrus.Debugf("intervention not triggered for user %s: challenge already active", userID)
		return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
	}

	if !state.CanTriggerIntervention(churnState, now) {
		logrus.Debugf("intervention not triggered for user %s: still in cooldown", userID)
		return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
	}

	// Trigger intervention
	logrus.Infof("triggering intervention for rage quit: userId=%s count=%d",
		userID, rageQuitCount)

	// Get current win count from AGS Statistics (winsAtStart for challenge)
	var currentWins int
	if s.statisticsService.Client != nil {
		var err error
		currentWins, err = s.getCurrentWinCount(ctx, userID)
		if err != nil {
			logrus.Warnf("failed to fetch current wins for user %s: %v, using 0", userID, err)
			currentWins = 0
		}
	} else {
		// In tests without AGS services, use 0
		currentWins = 0
	}

	// Create challenge
	expiresAt := now.Add(ChallengeDurationDays * 24 * time.Hour)
	state.CreateChallenge(churnState, ChallengeWinsNeeded, currentWins, expiresAt, "rage_quit")

	// Set cooldown
	cooldownDuration := InterventionCooldownHours * time.Hour
	state.SetInterventionCooldown(churnState, now, cooldownDuration)

	logrus.Infof("created rage quit challenge for user %s (no immediate reward, must complete challenge)", userID)

	// Save state
	return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
}

// handleMatchWin processes match win events
func (s *Statistic) handleMatchWin(ctx context.Context, userID string, totalWins int) error {
	logrus.Infof("handling match win for user %s: totalWins=%d", userID, totalWins)

	// Get current churn state
	churnState, err := state.GetChurnState(ctx, s.redisClient, userID)
	if err != nil {
		return err
	}

	now := time.Now()

	// Check if there's an active challenge
	if !churnState.Challenge.Active {
		logrus.Debugf("no active challenge for user %s, skipping win processing", userID)
		return nil
	}

	// Update challenge progress
	completed := state.UpdateChallengeProgress(churnState, totalWins, now)

	if completed {
		logrus.Infof("challenge completed for user %s! Wins: %d/%d",
			userID, churnState.Challenge.WinsCurrent, churnState.Challenge.WinsNeeded)

		// Grant speed booster entitlement as reward
		if s.fulfillmentService.Client != nil {
			speedBoosterID := common.GetEnv(EnvSpeedBoosterItemID, DefaultSpeedBoosterItemID)
			if err := grantEntitlement(s.fulfillmentService, s.namespace, userID, speedBoosterID); err != nil {
				logrus.Errorf("failed to grant speed booster to user %s: %v", userID, err)
				// Continue anyway - challenge is marked complete in Redis
			} else {
				logrus.Infof("granted speed booster reward to user %s", userID)
			}
		} else {
			logrus.Infof("[TEST MODE] would grant speed booster to user %s", userID)
		}

		// Clear challenge data from Redis to free up memory
		state.ResetChallenge(churnState)
	} else if churnState.Challenge.Active {
		logrus.Infof("challenge progress for user %s: %d/%d wins",
			userID, churnState.Challenge.WinsCurrent, churnState.Challenge.WinsNeeded)
	}

	// Save state
	return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
}

// handleLosingStreak processes losing streak events
func (s *Statistic) handleLosingStreak(ctx context.Context, userID string, currentStreak int) error {
	logrus.Infof("handling losing streak for user %s: streak=%d", userID, currentStreak)

	// Get current churn state
	churnState, err := state.GetChurnState(ctx, s.redisClient, userID)
	if err != nil {
		return err
	}

	now := time.Now()

	// Check if weekly reset is needed
	state.CheckWeeklyReset(churnState, now)

	// Check if losing streak exceeds threshold
	if currentStreak < LosingStreakThreshold {
		logrus.Debugf("losing streak %d below threshold %d, no intervention needed",
			currentStreak, LosingStreakThreshold)
		// Still save state in case weekly reset occurred
		return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
	}

	// Check if intervention can be triggered (not session-based, just cooldown + no active challenge)
	if churnState.Challenge.Active {
		logrus.Debugf("intervention not triggered for user %s: challenge already active", userID)
		return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
	}

	if !state.CanTriggerIntervention(churnState, now) {
		logrus.Debugf("intervention not triggered for user %s: still in cooldown", userID)
		return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
	}

	// Trigger intervention
	logrus.Infof("triggering intervention for losing streak: userId=%s streak=%d",
		userID, currentStreak)

	// Get current win count from AGS Statistics (winsAtStart for challenge)
	var currentWins int
	if s.statisticsService.Client != nil {
		var err error
		currentWins, err = s.getCurrentWinCount(ctx, userID)
		if err != nil {
			logrus.Warnf("failed to fetch current wins for user %s: %v, using 0", userID, err)
			currentWins = 0
		}
	} else {
		// In tests without AGS services, use 0
		currentWins = 0
	}

	// Create challenge
	expiresAt := now.Add(ChallengeDurationDays * 24 * time.Hour)
	state.CreateChallenge(churnState, ChallengeWinsNeeded, currentWins, expiresAt, "losing_streak")

	// Set cooldown
	cooldownDuration := InterventionCooldownHours * time.Hour
	state.SetInterventionCooldown(churnState, now, cooldownDuration)

	logrus.Infof("created losing streak challenge for user %s (no immediate reward, must complete challenge)", userID)

	// Save state
	return state.UpdateChurnState(ctx, s.redisClient, userID, churnState)
}

// getCurrentWinCount fetches the current total win count for a user from AGS Statistics
func (s *Statistic) getCurrentWinCount(ctx context.Context, userID string) (int, error) {
	// Query AGS Statistics for the current win count
	statCode := StatCodeMatchWins
	input := &user_statistic.GetUserStatItemsParams{
		Namespace: s.namespace,
		UserID:    userID,
		StatCodes: &statCode,
	}

	stats, err := s.statisticsService.GetUserStatItemsShort(input)
	if err != nil {
		return 0, err
	}

	// Find the match wins stat
	if stats != nil && stats.Data != nil {
		for _, stat := range stats.Data {
			if stat.StatCode != nil && *stat.StatCode == StatCodeMatchWins {
				if stat.Value != nil {
					return int(*stat.Value), nil
				}
			}
		}
	}

	// Stat not found or no value, return 0
	return 0, nil
}

// grantEntitlement grants an item entitlement to a user via AGS Platform API
// This is a shared helper function used by intervention and challenge completion logic
func grantEntitlement(
	fulfillmentService platform.FulfillmentService,
	namespace string,
	userID string,
	itemID string,
) error {
	quantity := int32(1)

	input := &fulfillment.FulfillItemParams{
		Namespace: namespace,
		UserID:    userID,
		Body: &platformclientmodels.FulfillmentRequest{
			ItemID:   itemID,
			Quantity: &quantity,
			Source:   platformclientmodels.FulfillmentRequestSourceREWARD,
		},
	}

	fulfillmentResponse, err := fulfillmentService.FulfillItemShort(input)

	if err != nil {
		return fmt.Errorf("failed to fulfill item: %w", err)
	}

	if fulfillmentResponse == nil {
		return fmt.Errorf("could not grant item to user: empty response")
	}

	return nil
}
