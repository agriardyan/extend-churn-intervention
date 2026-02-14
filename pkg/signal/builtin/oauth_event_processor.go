package builtin

import (
	"context"
	"fmt"
	"time"

	oauth "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	"github.com/sirupsen/logrus"
)

// OAuthEventProcessor processes OAuth token generation events into login signals.
type OAuthEventProcessor struct{}

func (p *OAuthEventProcessor) EventType() string {
	return "oauth_token_generated"
}

func (p *OAuthEventProcessor) Process(ctx context.Context, event interface{}, playerContextLoader signal.PlayerContextLoader) (signal.Signal, error) {
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

	// Load player context
	playerCtx, err := playerContextLoader.Load(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to load player context for user %s: %w", userID, err)
	}

	// Create login signal
	loginSignal := NewLoginSignal(userID, time.Now(), playerCtx)

	logrus.Debugf("processed OAuth event for user %s into LoginSignal", userID)
	return loginSignal, nil
}
