package builtin

import (
	"context"
	"fmt"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
	signalBuiltin "github.com/AccelByte/extends-anti-churn/pkg/signal/builtin"
	"github.com/sirupsen/logrus"
)

const (
	// LosingStreakRuleID is the identifier for losing streak detection rule
	LosingStreakRuleID = "losing_streak"

	// DefaultLosingStreakThreshold is the default number of consecutive losses to trigger
	DefaultLosingStreakThreshold = 5
)

// LosingStreakRule detects when a player is on a losing streak.
// A losing streak is tracked via the "rse-current-losing-streak" stat code.
type LosingStreakRule struct {
	config    rule.RuleConfig
	threshold int
}

// NewLosingStreakRule creates a new losing streak detection rule.
func NewLosingStreakRule(config rule.RuleConfig) *LosingStreakRule {
	threshold := config.GetInt("threshold", DefaultLosingStreakThreshold)

	logrus.Infof("creating losing streak rule with threshold=%d", threshold)

	return &LosingStreakRule{
		config:    config,
		threshold: threshold,
	}
}

// ID returns the rule identifier.
func (r *LosingStreakRule) ID() string {
	return r.config.ID
}

// Name returns the rule name.
func (r *LosingStreakRule) Name() string {
	return "Losing Streak Detection"
}

// SignalTypes returns the signal types this rule handles.
func (r *LosingStreakRule) SignalTypes() []string {
	return []string{signalBuiltin.TypeLosingStreak}
}

// Config returns the rule configuration.
func (r *LosingStreakRule) Config() rule.RuleConfig {
	return r.config
}

// Evaluate checks if the player has reached the losing streak threshold.
func (r *LosingStreakRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
	// Type assert to LossSignal
	lossSig, ok := sig.(*signalBuiltin.LossSignal)
	if !ok {
		return false, nil, fmt.Errorf("expected LossSignal, got %T", sig)
	}

	logrus.Debugf("evaluating losing streak for user %s: streak=%d, threshold=%d",
		lossSig.UserID(), lossSig.CurrentStreak, r.threshold)

	// Check if losing streak meets or exceeds threshold
	if lossSig.CurrentStreak >= r.threshold {
		trigger := rule.NewTrigger(r.ID(), sig.UserID(), "Losing streak threshold reached", r.config.Priority)
		trigger.Metadata["losing_streak"] = lossSig.CurrentStreak
		trigger.Metadata["threshold"] = r.threshold
		trigger.Metadata["stat_code"] = "rse-current-losing-streak"

		logrus.Infof("losing streak rule triggered for user %s: streak=%d, threshold=%d",
			sig.UserID(), lossSig.CurrentStreak, r.threshold)

		return true, trigger, nil
	}

	return false, nil, nil
}
