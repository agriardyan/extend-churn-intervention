package service

import (
	"context"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/fulfillment"
	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/social"
	"github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclient/user_statistic"
	"github.com/AccelByte/accelbyte-go-sdk/social-sdk/pkg/socialclientmodels"
)

type EntitlementService struct {
	fulfillmentClient *platform.FulfillmentService
	cfg               EntitlementServiceConfig
}

type EntitlementServiceConfig struct {
	Namespace string
}

func NewEntitlementService(
	fulfillmentClient *platform.FulfillmentService,
	cfg EntitlementServiceConfig,
) *EntitlementService {
	return &EntitlementService{
		fulfillmentClient: fulfillmentClient,
		cfg:               cfg,
	}
}

func (s *EntitlementService) GrantEntitlement(
	ctx context.Context,
	userID string,
	itemID string,
	quantity int,
) error {
	qnty := int32(quantity)

	namespace := s.cfg.Namespace
	fulfillmentService := s.fulfillmentClient

	input := &fulfillment.FulfillItemParams{
		Namespace: namespace,
		UserID:    userID,
		Body: &platformclientmodels.FulfillmentRequest{
			ItemID:   itemID,
			Quantity: &qnty,
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

type StatisticService struct {
	statisticsService *social.UserStatisticService
	cfg               StatisticServiceConfig
}

type StatisticServiceConfig struct {
	Namespace string
}

func NewStatisticService(
	statisticsService *social.UserStatisticService,
	cfg StatisticServiceConfig,
) *StatisticService {
	return &StatisticService{
		statisticsService: statisticsService,
		cfg:               cfg,
	}
}

func (s *StatisticService) UpdateStatComebackChallenge(ctx context.Context, userID string) error {
	namespace := s.cfg.Namespace
	statisticsService := s.statisticsService

	statCode := "rse-comeback-challenge"
	input := &user_statistic.IncUserStatItemValueParams{
		Namespace: namespace,
		UserID:    userID,
		StatCode:  statCode,
		Body: &socialclientmodels.StatItemInc{
			Inc: 1,
		},
	}

	_, err := statisticsService.IncUserStatItemValueShort(input)
	if err != nil {
		return fmt.Errorf("failed to increment user %s statistic %s: %w", userID, statCode, err)
	}

	return nil
}
