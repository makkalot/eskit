package provider

import (
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/eventstore"
)

// AdminProvider holds dependencies for the admin service
type AdminProvider struct {
	eventStore eventstore.Store
	crudStore  crudstore.CrudStore
}

// NewAdminProvider creates a new admin service provider
func NewAdminProvider(eventStore eventstore.Store, crudStore crudstore.CrudStore) *AdminProvider {
	return &AdminProvider{
		eventStore: eventStore,
		crudStore:  crudStore,
	}
}

// EventStore returns the event store instance
func (p *AdminProvider) EventStore() eventstore.Store {
	return p.eventStore
}

// CrudStore returns the crud store instance
func (p *AdminProvider) CrudStore() crudstore.CrudStore {
	return p.crudStore
}
