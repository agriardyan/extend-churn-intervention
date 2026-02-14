package builtin

import (
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// RegisterEventProcessors registers all built-in event processors.
func RegisterEventProcessors(registry *signal.EventProcessorRegistry) {
	registry.Register(&OAuthEventProcessor{})
	registry.Register(&RageQuitEventProcessor{})
	registry.Register(&LosingStreakEventProcessor{})
	registry.Register(&MatchWinEventProcessor{})
}
