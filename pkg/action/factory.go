package action

import (
	"fmt"

	"github.com/sirupsen/logrus"
)

// ActionFactory is a function that creates an action from a configuration.
type ActionFactory func(config ActionConfig) (Action, error)

// factories stores registered action factories by type
var factories = make(map[string]ActionFactory)

// RegisterActionType registers a factory function for an action type.
// This allows external packages to register their action types without creating import cycles.
func RegisterActionType(actionType string, factory ActionFactory) {
	factories[actionType] = factory
	logrus.Debugf("registered action type: %s", actionType)
}

// CreateAction creates an action instance based on the configuration.
// Returns an error if the action type is unknown.
func CreateAction(config ActionConfig) (Action, error) {
	if !config.Enabled {
		logrus.Infof("skipping disabled action: %s", config.ID)
		return nil, nil
	}

	logrus.Infof("creating action: id=%s, type=%s", config.ID, config.Type)

	factory, exists := factories[config.Type]
	if !exists {
		return nil, fmt.Errorf("unknown action type: %s", config.Type)
	}

	return factory(config)
}

// CreateActions creates multiple action instances from a list of configurations.
// Returns all successfully created actions and any errors encountered.
func CreateActions(configs []ActionConfig) ([]Action, []error) {
	var actions []Action
	var errors []error

	for _, config := range configs {
		action, err := CreateAction(config)
		if err != nil {
			errors = append(errors, fmt.Errorf("failed to create action %s: %w", config.ID, err))
			continue
		}

		if action != nil {
			actions = append(actions, action)
		}
	}

	return actions, errors
}

// RegisterActions registers multiple actions with the provided registry.
// This is a convenience function for setting up actions.
func RegisterActions(registry *Registry, configs []ActionConfig) error {
	actions, errors := CreateActions(configs)

	if len(errors) > 0 {
		logrus.Warnf("encountered %d errors while creating actions", len(errors))
		for _, err := range errors {
			logrus.Warnf("action creation error: %v", err)
		}
	}

	for _, action := range actions {
		if err := registry.Register(action); err != nil {
			return fmt.Errorf("failed to register action %s: %w", action.ID(), err)
		}
	}

	logrus.Infof("registered %d actions", len(actions))
	return nil
}
