package types

// Originator identifies an entity and its version for event sourcing.
// It combines a unique identifier with a version number for optimistic locking.
type Originator struct {
	// ID is the unique identifier of the entity
	ID string `json:"id" gorm:"column:id"`

	// Version is the version number of the entity, used for optimistic locking
	// and ensuring events are applied in the correct order
	Version string `json:"version" gorm:"column:version"`
}

// NewOriginator creates a new Originator with the given ID and version
func NewOriginator(id, version string) *Originator {
	return &Originator{
		ID:      id,
		Version: version,
	}
}
