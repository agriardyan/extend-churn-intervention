package examples

import (
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// RegisterEventMappers registers all built-in signal mappers with the registry.
func RegisterEventMappers(registry *signal.MapperRegistry) {
	registry.Register(&RageQuitMapper{})
	registry.Register(&MatchWinMapper{})
	registry.Register(&LosingStreakMapper{})
}
