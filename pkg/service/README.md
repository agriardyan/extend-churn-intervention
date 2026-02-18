# Service Package

This package contains service implementations that provide external dependencies to rules.

## Structure

```
pkg/service/
├── README.md           # This file
├── mock/               # Mock implementations for testing
│   ├── clan.go         # Mock ClanService
│   ├── leaderboard.go  # Mock LeaderboardService
│   └── player.go       # Mock PlayerHistoryService
└── impl/               # Real implementations (future)
    ├── clan.go         # Redis/API-based ClanService
    ├── leaderboard.go  # Redis/API-based LeaderboardService
    └── player.go       # Redis/API-based PlayerHistoryService
```

## Purpose

Rules that need external data (Redis, databases, APIs) should receive dependencies through constructor injection. This package provides:

1. **Service Interfaces**: Defined in `pkg/rule/dependencies.go`
2. **Mock Implementations**: For testing rules in isolation
3. **Real Implementations**: For production use (to be implemented as needed)

## Usage in Rules

```go
type MyRule struct {
    clanService rule.ClanService
}

func NewMyRule(deps *rule.RuleDependencies) Rule {
    return &MyRule{
        clanService: deps.ClanService,
    }
}

func (r *MyRule) Evaluate(ctx context.Context, signal Signal) (bool, map[string]interface{}, error) {
    // Quick checks first (use data already in signal)
    if signal.UserID() == "" {
        return false, nil, nil
    }
    
    // Expensive external call only when needed
    if r.clanService != nil {
        activity, err := r.clanService.GetClanActivity(ctx, "clan-123")
        if err != nil {
            return false, nil, err
        }
        // Use activity data...
    }
    
    return true, nil, nil
}
```

## Testing with Mocks

```go
import "pkg/service/mock"

func TestMyRule(t *testing.T) {
    mockClan := &mock.ClanService{
        Activity: &rule.ClanActivity{
            ActiveMembersLast7Days: 10,
        },
    }
    
    deps := rule.NewRuleDependencies().WithClanService(mockClan)
    rule := NewMyRule(deps)
    
    // Test rule...
}
```

## Implementation Status

- ✅ Service interfaces defined (`pkg/rule/dependencies.go`)
- ⏳ Mock implementations (Phase 4)
- ⏳ Real implementations (Future - as needed per rule)

## Design Principles

1. **Interface-based**: All services are interfaces for testability
2. **Lazy loading**: Expensive calls happen after threshold checks
3. **Nil-safe**: Rules should handle nil services gracefully
4. **Context-aware**: All service methods accept `context.Context`
5. **Error handling**: Services return errors, rules decide how to handle
