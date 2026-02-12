package pipeline

// Pipeline connects rules to actions.
// A pipeline defines which actions should be executed when specific rules trigger.
type Pipeline struct {
	Name       string              // Pipeline name
	Rules      []string            // Rule IDs to evaluate
	Actions    map[string][]string // Rule ID → Action IDs mapping
	Concurrent bool                // Execute actions in parallel
}

// NewPipeline creates a new pipeline with the given name.
func NewPipeline(name string) *Pipeline {
	return &Pipeline{
		Name:    name,
		Actions: make(map[string][]string),
	}
}

// AddRule adds a rule to be evaluated in this pipeline.
func (p *Pipeline) AddRule(ruleID string) *Pipeline {
	p.Rules = append(p.Rules, ruleID)
	return p
}

// AddActions associates actions with a rule.
func (p *Pipeline) AddActions(ruleID string, actionIDs ...string) *Pipeline {
	if p.Actions == nil {
		p.Actions = make(map[string][]string)
	}
	p.Actions[ruleID] = append(p.Actions[ruleID], actionIDs...)
	return p
}

// GetActions returns the action IDs for a given rule.
func (p *Pipeline) GetActions(ruleID string) []string {
	return p.Actions[ruleID]
}

// SetConcurrent sets whether actions should be executed in parallel.
func (p *Pipeline) SetConcurrent(concurrent bool) *Pipeline {
	p.Concurrent = concurrent
	return p
}
