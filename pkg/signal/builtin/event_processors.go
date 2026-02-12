package builtin

import (
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// RegisterBuiltinEventProcessors registers all built-in event processors.
func RegisterBuiltinEventProcessors(
	eventRegistry *signal.EventProcessorRegistry,
	mapperRegistry *signal.MapperRegistry,
) {
	eventRegistry.Register(&OAuthEventProcessor{})
	eventRegistry.Register(NewStatEventProcessor(mapperRegistry))
}
