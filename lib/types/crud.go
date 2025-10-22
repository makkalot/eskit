package types

// SchemaSpec defines the schema specification for validating entities.
// It supports versioning and JSON schema validation.
type SchemaSpec struct {
	// SchemaVersion is the version of the schema for this entity type.
	// Used to ensure clients are using the correct schema version.
	SchemaVersion uint64 `json:"schema_version" gorm:"column:schema_version"`

	// JSONSchema is an optional JSON Schema definition for validation.
	// If empty, no validation is performed.
	// If supplied, entities are validated against this schema.
	JSONSchema string `json:"json_schema,omitempty" gorm:"column:json_schema;type:text"`
}

// CrudEntitySpec defines the complete specification for a CRUD entity type.
// This is used to register entity types in the CRUD store.
type CrudEntitySpec struct {
	// EntityType is the type of entity (e.g., "User", "Project", "Order")
	// For entities in different bounded contexts, use namespacing like "com.example.User"
	EntityType string `json:"entity_type" gorm:"column:entity_type;primaryKey"`

	// SchemaSpec defines the schema and validation rules for this entity type
	SchemaSpec *SchemaSpec `json:"schema_spec,omitempty" gorm:"embedded;embeddedPrefix:schema_"`
}

// NewCrudEntitySpec creates a new CrudEntitySpec with the given entity type
func NewCrudEntitySpec(entityType string) *CrudEntitySpec {
	return &CrudEntitySpec{
		EntityType: entityType,
		SchemaSpec: &SchemaSpec{},
	}
}

// CrudEntity represents a complete CRUD entity with its metadata.
// This is used internally by the CRUD store to track entity state.
type CrudEntity struct {
	// EntityType is the type of this entity
	EntityType string `json:"entity_type" gorm:"column:entity_type"`

	// Originator identifies this specific entity instance
	Originator *Originator `json:"originator" gorm:"embedded;embeddedPrefix:originator_"`

	// Payload is the JSON-encoded entity data
	Payload string `json:"payload" gorm:"column:payload;type:text"`

	// Deleted indicates whether this entity has been soft-deleted
	Deleted bool `json:"deleted" gorm:"column:deleted"`
}

// NewCrudEntity creates a new CrudEntity
func NewCrudEntity(entityType string, originator *Originator, payload string, deleted bool) *CrudEntity {
	return &CrudEntity{
		EntityType: entityType,
		Originator: originator,
		Payload:    payload,
		Deleted:    deleted,
	}
}
