package rule

import (
	"context"
	"sort"

	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
	"github.com/sirupsen/logrus"
)

// Engine evaluates signals against registered rules and returns triggers.
type Engine struct {
	registry *Registry
}

// NewEngine creates a new rule evaluation engine.
func NewEngine(registry *Registry) *Engine {
	return &Engine{
		registry: registry,
	}
}

// Evaluate evaluates a signal against all matching rules.
// Returns a list of triggers for rules that matched.
func (e *Engine) Evaluate(ctx context.Context, sig signal.Signal) ([]*Trigger, error) {
	if sig == nil {
		return nil, nil
	}

	// Get rules that handle this signal type
	rules := e.registry.GetBySignalType(sig.Type())
	if len(rules) == 0 {
		logrus.Debugf("no rules found for signal type '%s'", sig.Type())
		return nil, nil
	}

	logrus.Debugf("evaluating signal type '%s' against %d rules", sig.Type(), len(rules))

	var triggers []*Trigger

	// Evaluate each rule
	for _, rule := range rules {
		matched, trigger, err := rule.Evaluate(ctx, sig)
		if err != nil {
			logrus.Errorf("rule %s evaluation failed: %v", rule.ID(), err)
			// Continue evaluating other rules even if one fails
			continue
		}

		if matched && trigger != nil {
			logrus.Infof("rule %s triggered for user %s: %s", rule.ID(), sig.UserID(), trigger.Reason)
			triggers = append(triggers, trigger)
		}
	}

	// Sort triggers by priority (higher priority first)
	if len(triggers) > 1 {
		sort.Slice(triggers, func(i, j int) bool {
			return triggers[i].Priority > triggers[j].Priority
		})
	}

	return triggers, nil
}

// EvaluateMultiple evaluates multiple signals in sequence.
// This is useful for batch processing.
func (e *Engine) EvaluateMultiple(ctx context.Context, signals []signal.Signal) ([]*Trigger, error) {
	var allTriggers []*Trigger

	for _, sig := range signals {
		triggers, err := e.Evaluate(ctx, sig)
		if err != nil {
			return allTriggers, err
		}
		allTriggers = append(allTriggers, triggers...)
	}

	return allTriggers, nil
}

// GetRegistry returns the rule registry used by this engine.
func (e *Engine) GetRegistry() *Registry {
	return e.registry
}
