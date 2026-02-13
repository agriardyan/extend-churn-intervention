package signal

import (
	"context"
	"fmt"
	"sync"
)

// EventProcessor processes raw events into signals.
// Implementations handle specific event types (OAuth, Stat, etc).
type EventProcessor interface {
	// EventType returns the type of event this processor handles.
	// Examples: "oauth_token_generated", "stat_item_updated"
	EventType() string

	// Process converts a raw event into a signal with enriched context.
	Process(ctx context.Context, event interface{}, contextLoader PlayerContextLoader) (Signal, error)
}

// PlayerContextLoader provides player context for event processing.
// This allows event processors to enrich signals without direct state store access.
type PlayerContextLoader interface {
	Load(ctx context.Context, userID string) (*PlayerContext, error)
}

// EventProcessorRegistry manages registered event processors.
type EventProcessorRegistry struct {
	mu         sync.RWMutex
	processors map[string]EventProcessor
}

// NewEventProcessorRegistry creates a new event processor registry.
func NewEventProcessorRegistry() *EventProcessorRegistry {
	return &EventProcessorRegistry{
		processors: make(map[string]EventProcessor),
	}
}

// Register adds an event processor to the registry.
func (r *EventProcessorRegistry) Register(processor EventProcessor) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.processors[processor.EventType()] = processor
}

// Get retrieves an event processor by event type.
func (r *EventProcessorRegistry) Get(eventType string) EventProcessor {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.processors[eventType]
}

// GetAll returns all registered event processors.
func (r *EventProcessorRegistry) GetAll() map[string]EventProcessor {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make(map[string]EventProcessor, len(r.processors))
	for k, v := range r.processors {
		result[k] = v
	}
	return result
}

// Count returns the number of registered event processors.
func (r *EventProcessorRegistry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.processors)
}

// Unregister removes an event processor from the registry.
func (r *EventProcessorRegistry) Unregister(eventType string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.processors[eventType]; !exists {
		return fmt.Errorf("event processor for type '%s' not found", eventType)
	}

	delete(r.processors, eventType)
	return nil
}
