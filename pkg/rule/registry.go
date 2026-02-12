package rule

import (
	"fmt"
	"sync"
)

// Registry manages available rules.
// It provides thread-safe registration and lookup of rules.
type Registry struct {
	rules map[string]Rule
	mu    sync.RWMutex
}

// NewRegistry creates a new empty rule registry.
func NewRegistry() *Registry {
	return &Registry{
		rules: make(map[string]Rule),
	}
}

// Register adds a rule to the registry.
// Returns an error if a rule with the same ID already exists.
func (r *Registry) Register(rule Rule) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rules[rule.ID()]; exists {
		return fmt.Errorf("rule %s already registered", rule.ID())
	}

	r.rules[rule.ID()] = rule
	return nil
}

// Unregister removes a rule from the registry.
// Returns an error if the rule doesn't exist.
func (r *Registry) Unregister(ruleID string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.rules[ruleID]; !exists {
		return fmt.Errorf("rule %s not found", ruleID)
	}

	delete(r.rules, ruleID)
	return nil
}

// Get returns a rule by ID.
// Returns nil if the rule doesn't exist.
func (r *Registry) Get(ruleID string) Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return r.rules[ruleID]
}

// GetBySignalType returns all enabled rules that handle a specific signal type.
// Rules are returned in no particular order.
func (r *Registry) GetBySignalType(signalType string) []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var matching []Rule
	for _, rule := range r.rules {
		// Skip disabled rules
		if !rule.Config().Enabled {
			continue
		}

		// Check if rule handles this signal type
		signalTypes := rule.SignalTypes()
		if len(signalTypes) == 0 {
			// Empty slice means rule handles all types
			matching = append(matching, rule)
			continue
		}

		for _, st := range signalTypes {
			if st == signalType {
				matching = append(matching, rule)
				break
			}
		}
	}

	return matching
}

// GetAll returns all registered rules.
func (r *Registry) GetAll() []Rule {
	r.mu.RLock()
	defer r.mu.RUnlock()

	rules := make([]Rule, 0, len(r.rules))
	for _, rule := range r.rules {
		rules = append(rules, rule)
	}

	return rules
}

// Count returns the number of registered rules.
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.rules)
}
