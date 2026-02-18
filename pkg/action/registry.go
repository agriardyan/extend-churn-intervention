package action

import (
	"fmt"
	"sync"
)

// Registry manages available actions.
// It provides thread-safe registration and lookup of actions.
type Registry struct {
	actions map[string]Action
	mu      sync.RWMutex
}

// NewRegistry creates a new empty action registry.
func NewRegistry() *Registry {
	return &Registry{
		actions: make(map[string]Action),
	}
}

// Register adds an action to the registry.
// Returns an error if an action with the same ID already exists.
func (r *Registry) Register(action Action) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[action.ID()]; exists {
		return fmt.Errorf("action %s already registered", action.ID())
	}

	r.actions[action.ID()] = action
	return nil
}

// Unregister removes an action from the registry.
// Returns an error if the action doesn't exist.
func (r *Registry) Unregister(actionID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.actions[actionID]; !exists {
		return fmt.Errorf("action %s not found", actionID)
	}

	delete(r.actions, actionID)
	return nil
}

// Get returns an action by ID.
// Returns nil if the action doesn't exist.
func (r *Registry) Get(actionID string) Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.actions[actionID]
}

// GetEnabled returns an action by ID only if it's enabled.
// Returns nil if the action doesn't exist or is disabled.
func (r *Registry) GetEnabled(actionID string) Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	action := r.actions[actionID]
	if action != nil && !action.Config().Enabled {
		return nil
	}

	return action
}

// GetAll returns all registered actions.
func (r *Registry) GetAll() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	actions := make([]Action, 0, len(r.actions))
	for _, action := range r.actions {
		actions = append(actions, action)
	}

	return actions
}

// GetAllEnabled returns all enabled actions.
func (r *Registry) GetAllEnabled() []Action {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var actions []Action
	for _, action := range r.actions {
		if action.Config().Enabled {
			actions = append(actions, action)
		}
	}

	return actions
}

// Count returns the number of registered actions.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.actions)
}
