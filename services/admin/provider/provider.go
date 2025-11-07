package provider

import (
	"github.com/makkalot/eskit/lib/eventstore"
)

// AdminProvider holds dependencies for the admin service
type AdminProvider struct {
	eventStore eventstore.Store
}

// NewAdminProvider creates a new admin service provider
func NewAdminProvider(eventStore eventstore.Store) *AdminProvider {
	return &AdminProvider{
		eventStore: eventStore,
	}
}

// EventStore returns the event store instance
func (p *AdminProvider) EventStore() eventstore.Store {
	return p.eventStore
}
