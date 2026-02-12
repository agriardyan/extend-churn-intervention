package pipeline

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/AccelByte/extends-anti-churn/pkg/action"
	asyncapi_iam "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/iam/oauth/v1"
	asyncapi_social "github.com/AccelByte/extends-anti-churn/pkg/pb/accelbyte-asyncapi/social/statistic/v1"
	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

// Manager orchestrates the complete anti-churn pipeline:
// Event → Signal → Rules → Actions
type Manager struct {
	processor   *signal.Processor
	engine      *rule.Engine
	executor    *action.Executor
	ruleActions map[string][]string // Maps rule ID to action IDs
	logger      *slog.Logger
}

// NewManager creates a new pipeline manager with all required components.
// ruleActions maps rule IDs to the action IDs they should trigger.
func NewManager(processor *signal.Processor, engine *rule.Engine, executor *action.Executor, ruleActions map[string][]string, logger *slog.Logger) *Manager {
	if logger == nil {
		logger = slog.Default()
	}

	if ruleActions == nil {
		ruleActions = make(map[string][]string)
	}

	return &Manager{
		processor:   processor,
		engine:      engine,
		executor:    executor,
		ruleActions: ruleActions,
		logger:      logger,
	}
}

// ProcessOAuthEvent processes an OAuth event through the complete pipeline.
// Returns any error encountered during pipeline execution.
func (m *Manager) ProcessOAuthEvent(ctx context.Context, event *asyncapi_iam.OauthTokenGenerated) error {
	m.logger.Info("processing OAuth event through pipeline",
		slog.String("user_id", event.GetUserId()),
		slog.String("namespace", event.GetNamespace()))

	// Step 1: Convert event to signal
	sig, err := m.processor.ProcessOAuthEvent(ctx, event)
	if err != nil {
		m.logger.Error("failed to process OAuth event to signal",
			slog.String("user_id", event.GetUserId()),
			slog.String("error", err.Error()))
		return fmt.Errorf("signal processing failed: %w", err)
	}

	if sig == nil {
		m.logger.Debug("OAuth event did not generate a signal, skipping pipeline")
		return nil
	}

	m.logger.Info("OAuth event converted to signal",
		slog.String("signal_type", sig.Type()),
		slog.String("user_id", sig.UserID()))

	// Step 2: Evaluate rules
	return m.evaluateAndExecute(ctx, sig)
}

// ProcessStatEvent processes a statistic event through the complete pipeline.
// Returns any error encountered during pipeline execution.
func (m *Manager) ProcessStatEvent(ctx context.Context, event *asyncapi_social.StatItemUpdated) error {
	m.logger.Info("processing stat event through pipeline",
		slog.String("user_id", event.GetUserId()),
		slog.String("namespace", event.GetNamespace()),
		slog.String("stat_code", event.GetPayload().GetStatCode()))

	// Step 1: Convert event to signal
	sig, err := m.processor.ProcessStatEvent(ctx, event)
	if err != nil {
		m.logger.Error("failed to process stat event to signal",
			slog.String("user_id", event.GetUserId()),
			slog.String("stat_code", event.GetPayload().GetStatCode()),
			slog.String("error", err.Error()))
		return fmt.Errorf("signal processing failed: %w", err)
	}

	if sig == nil {
		m.logger.Debug("stat event did not generate a signal, skipping pipeline")
		return nil
	}

	m.logger.Info("stat event converted to signal",
		slog.String("signal_type", sig.Type()),
		slog.String("user_id", sig.UserID()))

	// Step 2: Evaluate rules
	return m.evaluateAndExecute(ctx, sig)
}

// evaluateAndExecute evaluates rules for a signal and executes triggered actions.
func (m *Manager) evaluateAndExecute(ctx context.Context, sig signal.Signal) error {
	// Step 2: Evaluate rules against the signal
	triggers, err := m.engine.Evaluate(ctx, sig)
	if err != nil {
		m.logger.Error("rule evaluation failed",
			slog.String("signal_type", sig.Type()),
			slog.String("user_id", sig.UserID()),
			slog.String("error", err.Error()))
		return fmt.Errorf("rule evaluation failed: %w", err)
	}

	if len(triggers) == 0 {
		m.logger.Debug("no rules triggered for signal",
			slog.String("signal_type", sig.Type()),
			slog.String("user_id", sig.UserID()))
		return nil
	}

	m.logger.Info("rules triggered",
		slog.Int("trigger_count", len(triggers)),
		slog.String("signal_type", sig.Type()),
		slog.String("user_id", sig.UserID()))

	// Step 3: Execute actions for each trigger
	for _, trigger := range triggers {
		// Get action IDs from rule-to-actions mapping
		actionIDs, ok := m.ruleActions[trigger.RuleID]
		if !ok || len(actionIDs) == 0 {
			m.logger.Info("trigger has no actions configured",
				slog.String("rule_id", trigger.RuleID))
			continue
		}

		m.logger.Info("executing actions for trigger",
			slog.String("rule_id", trigger.RuleID),
			slog.Int("action_count", len(actionIDs)),
			slog.String("user_id", sig.UserID()))

		// Execute all actions for this trigger with rollback support
		results, err := m.executor.ExecuteMultiple(ctx, actionIDs, trigger, sig.Context(), true)
		if err != nil {
			m.logger.Error("action execution encountered error",
				slog.String("rule_id", trigger.RuleID),
				slog.String("error", err.Error()))
		}

		// Log results
		successCount := 0
		failureCount := 0
		for _, result := range results {
			if result.Error != nil {
				failureCount++
				m.logger.Error("action execution failed",
					slog.String("action_id", result.ActionID),
					slog.String("rule_id", trigger.RuleID),
					slog.String("error", result.Error.Error()))
			} else {
				successCount++
			}
		}

		m.logger.Info("action execution completed",
			slog.String("rule_id", trigger.RuleID),
			slog.Int("success_count", successCount),
			slog.Int("failure_count", failureCount))

		// If any action failed and we have a partial failure, log a warning
		if failureCount > 0 && successCount > 0 {
			m.logger.Warn("partial action execution failure",
				slog.String("rule_id", trigger.RuleID),
				slog.Int("success", successCount),
				slog.Int("failed", failureCount))
		}
	}

	return nil
}

// Stats returns pipeline statistics (for observability).
type Stats struct {
	ProcessorStats ProcessorStats `json:"processor"`
	EngineStats    EngineStats    `json:"engine"`
	ExecutorStats  ExecutorStats  `json:"executor"`
}

// ProcessorStats contains signal processor statistics.
type ProcessorStats struct {
	TotalEventsProcessed int64 `json:"total_events_processed"`
	SignalsGenerated     int64 `json:"signals_generated"`
}

// EngineStats contains rule engine statistics.
type EngineStats struct {
	TotalEvaluations  int64 `json:"total_evaluations"`
	TriggersGenerated int64 `json:"triggers_generated"`
}

// ExecutorStats contains action executor statistics.
type ExecutorStats struct {
	TotalActionsExecuted int64 `json:"total_actions_executed"`
	SuccessfulActions    int64 `json:"successful_actions"`
	FailedActions        int64 `json:"failed_actions"`
}

// GetStats returns current pipeline statistics.
// Note: Basic implementation - can be extended with actual metrics tracking.
func (m *Manager) GetStats() Stats {
	return Stats{
		ProcessorStats: ProcessorStats{},
		EngineStats:    EngineStats{},
		ExecutorStats:  ExecutorStats{},
	}
}
