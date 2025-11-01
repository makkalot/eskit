// Package types provides native Go types for event sourcing operations.
//
// This package contains pure Go types with no gRPC or protobuf dependencies,
// making it suitable for embedded library usage without code generation.
//
// # Main Types
//
// The package defines the following core types:
//
//   - Originator: Identifies an entity and its version for optimistic locking
//   - Event: Represents a domain event with payload and metadata
//   - AppLogEntry: Represents an entry in the application event log
//   - CrudEntity: Represents CRUD entities with full metadata
//   - CrudEntitySpec: Defines the schema for CRUD entities
//
// # Usage
//
// These types are designed to work with the eventstore and crudstore packages:
//
//	import (
//	    "github.com/makkalot/eskit/lib/types"
//	    "github.com/makkalot/eskit/lib/eventstore"
//	)
//
//	// Create an event
//	event := &types.Event{
//	    Originator: &types.Originator{
//	        ID:      "user-123",
//	        Version: "1",
//	    },
//	    EventType:  "User.Created",
//	    Payload:    `{"email":"user@example.com"}`,
//	    OccurredOn: time.Now().UTC(),
//	}
//
//	// Store the event
//	store := eventstore.NewInMemoryStore()
//	store.Append(event)
//
// # Serialization
//
// All types include JSON and GORM tags for easy serialization and database persistence:
//
//   - JSON tags enable encoding/decoding with encoding/json
//   - GORM tags enable database mapping with gorm.io/gorm
//
// # Design
//
// These types follow standard Go conventions:
//
//   - Field names use Go conventions (ID not Id)
//   - Time fields use time.Time (not Unix timestamps)
//   - No proto-generated code or dependencies
//   - Compatible with standard Go tooling and IDEs
package types
