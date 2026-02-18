---
name: add-plugin
description: Interactively create new churn intervention plugins (stat listeners, event handlers, rules, actions, or complete flows). Automates code generation, registration, testing, and configuration for the plugin-based architecture.
---

# Add Anti-Churn Plugin

You are helping the user create a new plugin for the churn intervention system. Follow these steps carefully to generate all necessary files, registrations, and tests.

## Step 1: Determine Plugin Type

Ask the user what type of plugin they want to create using AskUserQuestion:

**Question**: "What type of plugin would you like to create?"

**Options**:
1. **Stat Listener** - Listen to a new stat code from game events (e.g., `rse-player-level`)
2. **Event Handler** - Handle a new event type (e.g., party events, purchases)
3. **Rule** - Detect a new churn pattern from signals
4. **Action** - Execute a new intervention when rules trigger
5. **Complete Flow** - Full pipeline: Stat listener + Rule + Action

## Step 2: Gather Required Information

Based on the plugin type, collect the necessary information using AskUserQuestion.

### For Stat Listener:
- **Stat code**: The stat code to listen to (e.g., `rse-player-level`)
- **Signal type name**: Name for the signal type (e.g., `player_level`)
- **Data to extract**: What data should be extracted from the stat event
- **Description**: What this stat tracks

### For Event Handler:
- **Event type**: The event type to handle (e.g., `party_disbanded`)
- **Signal type name**: Name for the signal type (e.g., `party_disbanded`)
- **Data to extract**: What data should be extracted from the event
- **Description**: What this event represents

### For Rule:
- **Rule ID**: Unique identifier (e.g., `stuck_player`)
- **Rule name**: Human-readable name (e.g., "Stuck Player Detection")
- **Signal types**: Which signal types this rule evaluates
- **Detection logic**: Description of the pattern to detect
- **Thresholds**: Any thresholds or parameters needed
- **Priority**: Rule priority (default: 10)

### For Action:
- **Action ID**: Unique identifier (e.g., `send_push_notification`)
- **Action name**: Human-readable name (e.g., "Send Push Notification")
- **Description**: What this action does
- **Parameters**: Parameters needed for execution
- **Dependencies**: External dependencies (e.g., notification service)
- **Supports rollback**: Whether this action can be rolled back

### For Complete Flow:
- Gather all of the above information

## Step 3: Generate Files

Create the necessary files based on the plugin type. Use the PLUGIN_DEVELOPMENT.md guide for detailed examples and patterns.

### For Stat Listener:

1. **Event Processor** (`pkg/signal/builtin/{name}_event_processor.go`)
   - Implement `EventType()` returning the stat code
   - Implement `Process()` to extract data and create signal
   - Load player context from state store
   - Handle errors gracefully

2. **Signal Type** (`pkg/signal/builtin/{name}_signal.go`)
   - Define signal type constant
   - Create signal struct with custom fields
   - Implement signal.Signal interface (Type, UserID, Timestamp, Context methods)
   - Add accessor methods for custom fields

3. **Registration** (update `pkg/signal/builtin/event_processors.go`)
   - Register the event processor in `RegisterEventProcessors()`

4. **Tests** (`pkg/signal/builtin/{name}_test.go`)
   - Test signal creation
   - Test field accessors
   - Test edge cases

### For Rule:

1. **Rule Implementation** (`pkg/rule/builtin/{name}.go`)
   - Define rule ID constant
   - Create rule struct with config and parameters
   - Implement constructor extracting parameters from config
   - Implement rule.Rule interface (ID, Name, SignalTypes, Config, Evaluate)
   - Implement evaluation logic with proper type assertions
   - Check cooldown before triggering
   - Create trigger with metadata
   - Add comprehensive logging

2. **Registration** (update `pkg/rule/builtin/init.go`)
   - Register rule type with factory function

3. **Configuration** (update `config/pipeline.yaml`)
   - Add rule configuration with ID, type, enabled flag, actions, and parameters

4. **Tests** (`pkg/rule/builtin/{name}_test.go`)
   - Test trigger conditions
   - Test non-trigger conditions
   - Test cooldown behavior
   - Test edge cases

### For Action:

1. **Action Implementation** (`pkg/action/builtin/{name}.go`)
   - Define action ID constant
   - Create action struct with config and dependencies
   - Implement constructor
   - Implement action.Action interface (ID, Name, Config, Execute)
   - Implement Execute with proper error handling and logging
   - Implement Rollback (or return ErrRollbackNotSupported)

2. **Registration** (update `pkg/action/builtin/init.go`)
   - Add dependencies to Dependencies struct if needed
   - Register action type with factory function

3. **Configuration** (update `config/pipeline.yaml`)
   - Add action configuration with ID, type, enabled flag, and parameters

4. **Tests** (`pkg/action/builtin/{name}_test.go`)
   - Test successful execution
   - Test error handling
   - Test rollback (if supported)
   - Test parameter validation

## Step 4: Follow Naming Conventions

**IMPORTANT**: Use consistent naming conventions:

- **CamelCase** for Go types: `PlayerLevelEventProcessor`, `StuckPlayerRule`
- **snake_case** for type IDs and signal types: `player_level`, `stuck_player`
- **kebab-case** for config IDs: `stuck-player-detection`, `send-push-notification`

## Step 5: Add Proper Logging

Include appropriate logging at these levels:

- **Info**: Rule triggered, action executed, important events
- **Debug**: Evaluation details, signal processing steps
- **Error**: Failures, validation errors
- **Warning**: Non-critical issues, skipped operations

Example:
```go
logrus.Infof("rule %s triggered for user %s", r.ID(), sig.UserID())
logrus.Debugf("evaluating rule %s for signal type %s", r.ID(), sig.Type())
```

## Step 6: Handle Edge Cases

Ensure all edge cases are handled:

- Nil player context → return false, nil, nil
- Missing signal data → return false, nil, nil
- Invalid type assertions → return false, nil, nil
- Cooldown active → log debug message and return false, nil, nil
- Missing required parameters → return error during construction

## Step 7: Create Comprehensive Tests

Write tests covering:

- **Happy path**: Expected trigger/execution
- **No trigger**: Condition not met
- **Edge cases**: Nil contexts, missing data
- **Error handling**: Invalid inputs, external failures
- **Cooldown**: Verify cooldown prevents re-triggering

Use table-driven tests for multiple scenarios:

```go
tests := []struct {
    name          string
    // test setup
    expectTrigger bool
}{
    {name: "should trigger when...", expectTrigger: true},
    {name: "should not trigger when...", expectTrigger: false},
}
```

## Step 8: Verify Implementation

Run verification checks:

1. **Build Check**:
   ```bash
   go build -o /tmp/churn-intervention-test .
   ```

2. **Test Check**:
   ```bash
   go test ./pkg/signal/builtin/... -v
   go test ./pkg/rule/builtin/... -v
   go test ./pkg/action/builtin/... -v
   ```

3. **Configuration Validation**:
   - Verify `pipeline.yaml` is valid YAML
   - Check all referenced action IDs exist
   - Verify rule registrations match config

## Step 9: Provide Summary

After successful creation, provide a summary:

```
✅ Plugin Created Successfully!

Files created:
- pkg/signal/builtin/{name}_event_processor.go
- pkg/signal/builtin/{name}_signal.go
- pkg/rule/builtin/{name}.go
- pkg/action/builtin/{name}.go
- pkg/rule/builtin/{name}_test.go

Files updated:
- pkg/signal/builtin/event_processors.go
- pkg/rule/builtin/init.go
- pkg/action/builtin/init.go
- config/pipeline.yaml

Configuration:
- Stat code: {stat_code}
- Rule ID: {rule_id}
- Action ID: {action_id}
- Signal type: {signal_type}

Next steps:
1. Review the generated code
2. Customize the implementation logic
3. Add more test cases
4. Run: make test
5. Run: make build
```

## Important Notes

- Always read PLUGIN_DEVELOPMENT.md for detailed examples and patterns
- Follow the existing code style in the codebase
- Ensure all generated code compiles before finishing
- Add godoc comments to all exported types and functions
- Validate parameters and provide sensible defaults
- Keep implementations simple and focused - avoid over-engineering

## Error Handling

If generation fails:
1. Show clear error message
2. Explain what went wrong
3. Suggest how to fix it
4. Clean up any partial files

If tests fail:
1. Show the test output
2. Identify the issue
3. Offer to fix it
4. Re-run tests after fixing

## References

- See PLUGIN_DEVELOPMENT.md for detailed examples and code templates
- See CLAUDE.md for architectural boundaries and system scope
- See existing implementations in `pkg/signal/builtin/`, `pkg/rule/builtin/`, `pkg/action/builtin/`
