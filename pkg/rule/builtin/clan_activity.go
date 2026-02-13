package builtin

import (
	"context"
	"fmt"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
	"github.com/AccelByte/extends-anti-churn/pkg/signal"
)

const (
	// ClanActivityRuleID is the rule type identifier for clan activity detection
	ClanActivityRuleID = "clan_activity"
)

// ClanActivityRule detects when players in low-activity clans might churn
// This demonstrates the two-stage evaluation pattern with lazy loading:
// 1. Quick check: Verify player has clan info in signal context
// 2. Lazy load: Only fetch clan activity if player has a clan
type ClanActivityRule struct {
	id          string
	minMembers  int // Minimum active members threshold
	clanService rule.ClanService
}

// NewClanActivityRule creates a new clan activity rule with external dependency
func NewClanActivityRule(config rule.RuleConfig, deps *rule.RuleDependencies) *ClanActivityRule {
	minMembers := 3 // default threshold
	if val, ok := config.Parameters["min_active_members"].(int); ok {
		minMembers = val
	} else if val, ok := config.Parameters["min_active_members"].(float64); ok {
		minMembers = int(val)
	}

	var clanService rule.ClanService
	if deps != nil {
		clanService = deps.ClanService
	}

	return &ClanActivityRule{
		id:          config.ID,
		minMembers:  minMembers,
		clanService: clanService,
	}
}

// ID returns the rule's unique identifier
func (r *ClanActivityRule) ID() string {
	return r.id
}

// Name returns the human-readable name
func (r *ClanActivityRule) Name() string {
	return "Clan Activity Detection"
}

// Description returns what this rule detects
func (r *ClanActivityRule) Description() string {
	return fmt.Sprintf("Detects players in clans with fewer than %d active members in the last 7 days", r.minMembers)
}

// SignalTypes returns the signal types this rule listens to
func (r *ClanActivityRule) SignalTypes() []string {
	return []string{"oauth_token_generated", "match_win"} // Evaluate on login and match completion
}

// Config returns the rule configuration
func (r *ClanActivityRule) Config() rule.RuleConfig {
	return rule.RuleConfig{
		ID:   r.id,
		Type: ClanActivityRuleID,
		Parameters: map[string]interface{}{
			"min_active_members": r.minMembers,
		},
	}
}

// Evaluate checks if the player's clan has low activity
// Two-stage evaluation:
// 1. Quick check: Does player have clan info?
// 2. Expensive check: Fetch clan activity from external service
func (r *ClanActivityRule) Evaluate(ctx context.Context, sig signal.Signal) (bool, *rule.Trigger, error) {
	// Stage 1: Quick checks using data already in signal (no external calls)
	playerCtx := sig.Context()
	if playerCtx == nil || playerCtx.SessionInfo == nil {
		// No player context - can't evaluate
		return false, nil, nil
	}

	// Check if player has clan information in session info
	clanID, hasClan := playerCtx.SessionInfo["clan_id"].(string)
	if !hasClan || clanID == "" {
		// Player not in a clan - rule doesn't apply
		return false, nil, nil
	}

	// Stage 2: Lazy load expensive data (only if quick checks passed)
	if r.clanService == nil {
		// Service not available - can't fetch clan data
		// In production, this would be logged/monitored
		return false, nil, fmt.Errorf("clan service not available")
	}

	// Fetch clan activity from external service (Redis/API)
	clanActivity, err := r.clanService.GetClanActivity(ctx, clanID)
	if err != nil {
		// External service error - decide how to handle
		// Option 1: Return error (fail-safe)
		// Option 2: Return false (fail-open)
		return false, nil, fmt.Errorf("failed to get clan activity: %w", err)
	}

	// Check if clan has low activity
	if clanActivity.ActiveMembersLast7Days < r.minMembers {
		// Trigger detected - clan has low activity
		metadata := map[string]interface{}{
			"clan_id":        clanID,
			"active_members": clanActivity.ActiveMembersLast7Days,
			"total_members":  clanActivity.TotalMembers,
			"threshold":      r.minMembers,
			"activity_rate":  float64(clanActivity.ActiveMembersLast7Days) / float64(clanActivity.TotalMembers),
		}
		trigger := &rule.Trigger{
			RuleID:   r.id,
			UserID:   sig.UserID(),
			Metadata: metadata,
		}
		return true, trigger, nil
	}

	// Clan activity is healthy - no trigger
	return false, nil, nil
}

// Validate checks if the rule configuration is valid
func (r *ClanActivityRule) Validate() error {
	if r.minMembers < 1 {
		return fmt.Errorf("min_active_members must be at least 1, got %d", r.minMembers)
	}
	return nil
}
