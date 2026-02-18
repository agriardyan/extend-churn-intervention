package builtin

import (
	"context"
	"fmt"

	"github.com/AccelByte/extend-churn-intervention/pkg/rule"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
	signalBuiltin "github.com/AccelByte/extend-churn-intervention/pkg/signal/builtin"
	"github.com/sirupsen/logrus"
)

const (
	// RageQuitRuleID is the identifier for rage quit detection rule
	RageQuitRuleID = "rage_quit"

	// DefaultRageQuitThreshold is the default number of rage quits to trigger
	DefaultRageQuitThreshold = 3
)

// RageQuitRule detects when a player exhibits rage quit behavior.
// A rage quit is tracked via the "rse-rage-quit" stat code.
type RageQuitRule struct {
	config    rule.RuleConfig
	threshold int
}

// NewRageQuitRule creates a new rage quit detection rule.
func NewRageQuitRule(config rule.RuleConfig) *RageQuitRule {
	threshold := config.GetInt("threshold", DefaultRageQuitThreshold)

	logrus.Infof("creating rage quit rule with threshold=%d", threshold)

	return &RageQuitRule{
		config:    config,
		threshold: threshold,
	}
}

// ID returns the rule identifier.
func (r *RageQuitRule) ID() string {
	return r.config.ID
}

// Name returns the rule name.
func (r *RageQuitRule) Name() string {
	return "Rage Quit Detection"
}

// SignalTypes returns the signal types this rule handles.
func (r *RageQuitRule) SignalTypes() []string {
	return []string{signalBuiltin.TypeRageQuit}
}

// Config returns the rule configuration.
func (r *RageQuitRule) Config() rule.RuleConfig {
	return r.config
}

// Evaluate checks if the player has reached the rage quit threshold.
func (r *RageQuitRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
	// Type assert to RageQuitSignal
	rageQuitSig, ok := sig.(*signalBuiltin.RageQuitSignal)
	if !ok {
		return false, nil, fmt.Errorf("expected RageQuitSignal, got %T", sig)
	}

	logrus.Debugf("evaluating rage quit for user %s: count=%d, threshold=%d",
		rageQuitSig.UserID(), rageQuitSig.QuitCount, r.threshold)

	// Check if rage quit count meets or exceeds threshold
	if rageQuitSig.QuitCount >= r.threshold {
		trigger := rule.NewTrigger(r.ID(), sig.UserID(), "Rage quit threshold reached", r.config.Priority)
		trigger.Metadata["rage_quit_count"] = rageQuitSig.QuitCount
		trigger.Metadata["threshold"] = r.threshold
		trigger.Metadata["stat_code"] = "rse-rage-quit"

		logrus.Infof("rage quit rule triggered for user %s: count=%d, threshold=%d",
			sig.UserID(), rageQuitSig.QuitCount, r.threshold)

		return true, trigger, nil
	}

	return false, nil, nil
}
