package builtin

import (
	"context"
	"fmt"
	"time"

	oauth "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/sirupsen/logrus"
)

// OAuthEventProcessor processes OAuth token generation events into login signals.
type OAuthEventProcessor struct {
	stateStore         service.StateStore
	loginTrackingStore service.LoginSessionTracker
	namespace          string
}

// NewOAuthEventProcessor creates a new OAuth event processor.
func NewOAuthEventProcessor(
	stateStore service.StateStore,
	loginTrackingStore service.LoginSessionTracker,
	namespace string,
) *OAuthEventProcessor {
	return &OAuthEventProcessor{
		stateStore:         stateStore,
		loginTrackingStore: loginTrackingStore,
		namespace:          namespace,
	}
}

func (p *OAuthEventProcessor) EventType() string {
	return "oauth_token_generated"
}

func (p *OAuthEventProcessor) Process(ctx context.Context, event interface{}) (signal.Signal, error) {
	oauthEvent, ok := event.(*oauth.OauthTokenGenerated)
	if !ok {
		return nil, fmt.Errorf("expected *oauth.OauthTokenGenerated, got %T", event)
	}

	if oauthEvent == nil {
		return nil, fmt.Errorf("oauth event is nil")
	}

	userID := oauthEvent.GetUserId()
	if userID == "" {
		return nil, fmt.Errorf("user ID is empty in oauth event")
	}

	// Load player context (core churn state)
	churnState, err := p.stateStore.GetChurnState(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load churn state for user %s: %w", userID, err)
	}

	// Increment session count in rule-specific storage
	// This is tracking telemetry for the session_decline rule
	err = p.loginTrackingStore.IncrementSessionCount(ctx, userID)
	if err != nil {
		logrus.Errorf("failed to increment session count for user %s: %v", userID, err)
		// Don't fail the signal processing if session tracking fails
	}

	playerCtx := signal.BuildPlayerContext(userID, p.namespace, churnState)

	// Create login signal
	loginSignal := NewLoginSignal(userID, time.Now(), playerCtx)

	logrus.Debugf("processed OAuth event for user %s into LoginSignal", userID)
	return loginSignal, nil
}
