package builtin

import (
	"github.com/AccelByte/extends-anti-churn/pkg/service"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

type EventProcessorDependencies struct {
	LoginTrackingStore service.LoginSessionTracker
}

// RegisterEventProcessors registers all built-in event processors.
// Add your dependencies parameters to the `deps` as needed.
// Add your configuration parameters to the `config` as needed.
func RegisterEventProcessors(
	registry *signal.EventProcessorRegistry,
	stateStore service.StateStore,
	namespace string,
	deps *EventProcessorDependencies,
) {
	registry.Register(NewOAuthEventProcessor(stateStore, deps.LoginTrackingStore, namespace))
	registry.Register(NewRageQuitEventProcessor(stateStore, namespace))
	registry.Register(NewLosingStreakEventProcessor(stateStore, namespace))
}
