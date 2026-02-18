// Copyright (c) 2025 AccelByte Inc. All Rights Reserved.
// This is licensed software from AccelByte Inc, for limitations
// and restrictions contact your company contract manager.

package bootstrap

import (
	"github.com/AccelByte/extend-churn-intervention/pkg/service"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
	signalBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/signal/builtin"
	"github.com/sirupsen/logrus"
)

// InitSignalProcessor creates and initializes a signal processor with builtin event processors.
//
// ============================================================
// DEVELOPER: Register custom event processors here.
// ============================================================
// Event processors normalize raw events into signals.
// Each processor handles a specific event type (e.g., OAuth events,
// stat updates) and enriches them with player context.
//
// Steps to add a new event processor:
// 1. Create your processor in pkg/signal/builtin/ (see examples)
// 2. Implement the EventProcessor interface
// 3. Register it in pkg/signal/builtin/event_processors.go
// 4. The registration function is called below automatically
//
// The builtin processors handle:
// - OAuth login events → login signals
// - Stat updates (match wins, losses, streaks) → game signals
// - Custom stat codes → custom signals
// ============================================================
func InitSignalProcessor(
	stateStore service.StateStore,
	loginTrackingStore service.LoginSessionTracker,
	namespace string,
) *signal.Processor {
	processor := signal.NewProcessor(stateStore, namespace)

	// ============================================================
	// DEVELOPER: Builtin event processor registration
	// ============================================================
	// This registers all event processors defined in pkg/signal/builtin/
	// To add new processors, modify pkg/signal/builtin/event_processors.go
	// ============================================================
	signalBuiltin.RegisterEventProcessors(
		processor.GetEventProcessorRegistry(),
		processor.GetStateStore(),
		processor.GetNamespace(),

		// DEVELOPER: Pass dependencies needed by event processors here
		// ============================================================
		// If your event processors need external dependencies (e.g., login session tracking),
		// add them to the EventProcessorDependencies struct in
		// pkg/signal/builtin/event_processors.go and pass them here.
		// ============================================================
		&signalBuiltin.EventProcessorDependencies{
			LoginTrackingStore: loginTrackingStore,
		},
	)

	logrus.Infof("initialized signal processor with %d event processors",
		processor.GetEventProcessorRegistry().Count())

	// ============================================================
	// DEVELOPER: Register custom event processors below
	// ============================================================
	// If you have custom processors outside pkg/signal/builtin/,
	// register them here:
	//
	// customProcessor := mycustom.NewMyEventProcessor(stateStore, namespace)
	// processor.GetEventProcessorRegistry().Register(customProcessor)
	// ============================================================

	return processor
}
