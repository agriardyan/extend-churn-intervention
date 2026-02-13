package service

// Data structures returned by service interfaces

// Dependencies holds all external service dependencies that rules/actions can use.
// Components receive this struct and can access only the services they need.
type Dependencies struct {
	GrantEntitlementService EntitlementGranter
	StatUpdateService       StatUpdater
}

// NewDependencies creates a new dependencies container.
// Services can be nil if not needed - components should handle nil gracefully.
func NewDependencies() *Dependencies {
	return &Dependencies{}
}

// WithEntitlementGranter sets the entitlement granter service
func (d *Dependencies) WithGrantEntitlementService(service EntitlementGranter) *Dependencies {
	d.GrantEntitlementService = service
	return d
}

// WithStatUpdater sets the stat updater service
func (d *Dependencies) WithStatUpdaterService(service StatUpdater) *Dependencies {
	d.StatUpdateService = service
	return d
}
