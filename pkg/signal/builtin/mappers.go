package builtin

import (
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// RegisterBuiltinMappers registers all built-in signal mappers with the registry.
func RegisterBuiltinMappers(registry *signal.MapperRegistry) {
	registry.Register(&RageQuitMapper{})
	registry.Register(&MatchWinMapper{})
	registry.Register(&LosingStreakMapper{})
}
