package builtin

import (
	"context"

	"github.com/AccelByte/extend-churn-intervention/pkg/action"
	"github.com/AccelByte/extend-churn-intervention/pkg/rule"
	"github.com/AccelByte/extend-churn-intervention/pkg/signal"
	"github.com/sirupsen/logrus"
)

const (
	// SendEmailActionID is the identifier for send email notification action
	SendEmailActionID = "send_email_notification_after_granting_item"
)

// SendEmailAction sends an email notification to a player.
// This is a no-op placeholder for email integration.
type SendEmailAction struct {
	config action.ActionConfig
}

func NewSendEmailAction(config action.ActionConfig) *SendEmailAction {
	return &SendEmailAction{
		config: config,
	}
}

func (a *SendEmailAction) ID() string {
	return a.config.ID
}

func (a *SendEmailAction) Name() string {
	return "Send Email"
}

func (a *SendEmailAction) Config() action.ActionConfig {
	return a.config
}

func (a *SendEmailAction) Execute(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	logrus.Infof("[NO-OP] SendEmailAction for user %s triggered by rule %s", trigger.UserID, trigger.RuleID)
	logrus.Infof("[NO-OP] Would send email notification about granted item to user %s", trigger.UserID)

	// This is only placeholder for email sending logic.
	// In a real implementation, you would integrate with IAM and an email service provider here:
	// 1. Retrieve player's email from IAM
	// 2. Send email notification via email service provider (SendGrid, AWS SES, etc.)
	return nil
}

func (a *SendEmailAction) Rollback(ctx context.Context, trigger *rule.Trigger, playerCtx *signal.PlayerContext) error {
	return action.ErrRollbackNotSupported
}
