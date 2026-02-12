package signal

import (
	"sync"
	"time"
)

// SignalMapper maps stat events to domain signals.
// This allows extending the signal processor with custom stat-to-signal mappings.
type SignalMapper interface {
	// StatCode returns the stat code this mapper handles (e.g., "rse-rage-quit").
	StatCode() string

	// MapToSignal converts a stat value into a domain signal.
	MapToSignal(userID string, timestamp time.Time, value float64, context *PlayerContext) Signal
}

// MapperRegistry manages registered signal mappers.
// It provides thread-safe registration and lookup of mappers.
type MapperRegistry struct {
	mappers map[string]SignalMapper
	mu      sync.RWMutex
}

// NewMapperRegistry creates a new empty mapper registry.
func NewMapperRegistry() *MapperRegistry {
	return &MapperRegistry{
		mappers: make(map[string]SignalMapper),
	}
}

// Register adds a mapper to the registry.
// If a mapper for the same stat code already exists, it will be replaced.
func (r *MapperRegistry) Register(mapper SignalMapper) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.mappers[mapper.StatCode()] = mapper
}

// Get returns a mapper for the given stat code.
// Returns nil if no mapper is registered for that stat code.
func (r *MapperRegistry) Get(statCode string) SignalMapper {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.mappers[statCode]
}

// GetAll returns all registered mappers.
func (r *MapperRegistry) GetAll() []SignalMapper {
	r.mu.RLock()
	defer r.mu.RUnlock()

	mappers := make([]SignalMapper, 0, len(r.mappers))
	for _, mapper := range r.mappers {
		mappers = append(mappers, mapper)
	}
	return mappers
}

// Count returns the number of registered mappers.
func (r *MapperRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.mappers)
}
