package examples

import (
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// RegisterEventProcessors registers all built-in event processors.
func RegisterEventProcessors(
	eventRegistry *signal.EventProcessorRegistry,
	mapperRegistry *signal.MapperRegistry,
) {
	eventRegistry.Register(&OAuthEventProcessor{})
	eventRegistry.Register(NewStatEventProcessor(mapperRegistry))
}
