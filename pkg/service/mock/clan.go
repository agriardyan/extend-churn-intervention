package mock

import (
	"context"
	"fmt"

	"github.com/AccelByte/extends-anti-churn/pkg/rule"
)

// ClanService is a mock implementation of rule.ClanService for testing
type ClanService struct {
	// GetClanActivityFunc allows tests to customize the behavior
	GetClanActivityFunc func(ctx context.Context, clanID string) (*rule.ClanActivity, error)
	IsPlayerInClanFunc  func(ctx context.Context, userID string) (bool, string, error)
	GetClanMembersFunc  func(ctx context.Context, clanID string) ([]string, error)

	// Simple fields for common test scenarios
	Activity *rule.ClanActivity
	Error    error
}

// GetClanActivity returns mocked clan activity data
func (m *ClanService) GetClanActivity(ctx context.Context, clanID string) (*rule.ClanActivity, error) {
	if m.GetClanActivityFunc != nil {
		return m.GetClanActivityFunc(ctx, clanID)
	}
	if m.Error != nil {
		return nil, m.Error
	}
	if m.Activity != nil {
		return m.Activity, nil
	}
	return nil, fmt.Errorf("no mock data configured for clan %s", clanID)
}

// IsPlayerInClan returns mocked player clan membership
func (m *ClanService) IsPlayerInClan(ctx context.Context, userID string) (bool, string, error) {
	if m.IsPlayerInClanFunc != nil {
		return m.IsPlayerInClanFunc(ctx, userID)
	}
	return false, "", nil
}

// GetClanMembers returns mocked clan members list
func (m *ClanService) GetClanMembers(ctx context.Context, clanID string) ([]string, error) {
	if m.GetClanMembersFunc != nil {
		return m.GetClanMembersFunc(ctx, clanID)
	}
	return []string{}, nil
}

// NewClanService creates a new mock clan service with default behavior
func NewClanService() *ClanService {
	return &ClanService{}
}

// WithActivity sets the activity data to return
func (m *ClanService) WithActivity(activity *rule.ClanActivity) *ClanService {
	m.Activity = activity
	return m
}

// WithError sets an error to return
func (m *ClanService) WithError(err error) *ClanService {
	m.Error = err
	return m
}
