package builtin

import (
	"context"
	"fmt"

	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclient/fulfillment"
	"github.com/AccelByte/accelbyte-go-sdk/platform-sdk/pkg/platformclientmodels"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/factory"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/repository"
	"github.com/AccelByte/accelbyte-go-sdk/services-api/pkg/service/platform"
	"github.com/AccelByte/extends-anti-churn/pkg/action"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/sirupsen/logrus"
)

const (
	// GrantItemActionID is the identifier for item grant action
	GrantItemActionID = "grant_item"
)

// ItemGranter is an interface for granting items to users.
// This allows for dependency injection and testing.
type ItemGranter interface {
	GrantItem(ctx context.Context, namespace, userID, itemID string, quantity int32) error
}

// GrantItemAction grants an item or entitlement to a player.
// This action integrates with AccelByte Platform to fulfill items.
type GrantItemAction struct {
	config    action.ActionConfig
	granter   ItemGranter
	namespace string
	itemID    string
	quantity  int32
}

// NewGrantItemAction creates a new grant item action.
func NewGrantItemAction(config action.ActionConfig, granter ItemGranter, namespace string) *GrantItemAction {
	itemID := config.GetParameterString("item_id", "")
	quantity := int32(config.GetParameterInt("quantity", 1))

	logrus.Infof("creating grant item action: itemID=%s, quantity=%d", itemID, quantity)

	return &GrantItemAction{
		config:    config,
		granter:   granter,
		namespace: namespace,
		itemID:    itemID,
		quantity:  quantity,
	}
}

// ID returns the action identifier.
func (a *GrantItemAction) ID() string {
	return a.config.ID
}

// Name returns the action name.
func (a *GrantItemAction) Name() string {
	return "Grant Item"
}

// Config returns the action configuration.
func (a *GrantItemAction) Config() action.ActionConfig {
	return a.config
}

// Execute grants the configured item to the player.
func (a *GrantItemAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	if a.itemID == "" {
		return fmt.Errorf("item_id parameter not configured")
	}

	if a.granter == nil {
		logrus.Warnf("[TEST MODE] would grant item %s (quantity: %d) to user %s",
			a.itemID, a.quantity, trigger.UserID)
		return nil
	}

	logrus.Infof("granting item %s (quantity: %d) to user %s",
		a.itemID, a.quantity, trigger.UserID)

	err := a.granter.GrantItem(ctx, a.namespace, trigger.UserID, a.itemID, a.quantity)
	if err != nil {
		return fmt.Errorf("failed to grant item: %w", err)
	}

	logrus.Infof("successfully granted item %s to user %s", a.itemID, trigger.UserID)
	return nil
}

// Rollback is not supported for item grants (items cannot be taken back).
func (a *GrantItemAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	return action.ErrRollbackNotSupported
}

// AccelByteItemGranter implements ItemGranter using AccelByte Platform API.
type AccelByteItemGranter struct {
	fulfillmentService platform.FulfillmentService
	namespace          string
}

// NewAccelByteItemGranter creates a new AccelByte item granter.
func NewAccelByteItemGranter(
	configRepo repository.ConfigRepository,
	tokenRepo repository.TokenRepository,
	namespace string,
) *AccelByteItemGranter {
	return &AccelByteItemGranter{
		fulfillmentService: platform.FulfillmentService{
			Client:           factory.NewPlatformClient(configRepo),
			ConfigRepository: configRepo,
			TokenRepository:  tokenRepo,
		},
		namespace: namespace,
	}
}

// GrantItem grants an item to a user via AccelByte Platform API.
func (g *AccelByteItemGranter) GrantItem(ctx context.Context, namespace, userID, itemID string, quantity int32) error {
	input := &fulfillment.FulfillItemParams{
		Namespace: namespace,
		UserID:    userID,
		Body: &platformclientmodels.FulfillmentRequest{
			ItemID:   itemID,
			Quantity: &quantity,
			Source:   platformclientmodels.FulfillmentRequestSourceREWARD,
		},
	}

	fulfillmentResponse, err := g.fulfillmentService.FulfillItemShort(input)
	if err != nil {
		return fmt.Errorf("failed to fulfill item: %w", err)
	}

	if fulfillmentResponse == nil {
		return fmt.Errorf("could not grant item to user: empty response")
	}

	return nil
}
