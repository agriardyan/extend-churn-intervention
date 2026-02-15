package builtin

import (
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// RegisterEventProcessors registers all built-in event processors.
func RegisterEventProcessors(registry *signal.EventProcessorRegistry, stateStore service.StateStore, namespace string) {
	registry.Register(NewOAuthEventProcessor(stateStore, namespace))
	registry.Register(NewRageQuitEventProcessor(stateStore, namespace))
	registry.Register(NewLosingStreakEventProcessor(stateStore, namespace))
}
