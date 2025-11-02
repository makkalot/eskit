package types

// Originator identifies an entity and its version for event sourcing.
// It combines a unique identifier with a version number for optimistic locking.
type Originator struct {
	// ID is the unique identifier of the entity
	ID string `json:"id" gorm:"column:id"`

	// Version is the version number of the entity, used for optimistic locking
	// and ensuring events are applied in the correct order.
	// Version starts at 1 for newly created entities and increments with each update.
	Version uint64 `json:"version" gorm:"column:version"`
}

// NewOriginator creates a new Originator with the given ID and version
func NewOriginator(id string, version uint64) *Originator {
	return &Originator{
		ID:      id,
		Version: version,
	}
}
