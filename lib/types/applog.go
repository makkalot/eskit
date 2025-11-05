package types

// AppLogEntry represents an entry in the application log.
// The application log is a sequential stream of all events in the system,
// similar to Kafka's event log. Each entry has a unique sequential ID.
type AppLogEntry struct {
	// ID is the sequential ID of this entry in the application log.
	// IDs are auto-incremented and monotonically increasing (1, 2, 3, ...).
	ID uint64 `json:"id" gorm:"column:id;primaryKey"`

	// Event is the event stored in this log entry
	Event *Event `json:"event" gorm:"embedded;embeddedPrefix:event_"`
}

// NewAppLogEntry creates a new AppLogEntry with the given ID and event
func NewAppLogEntry(id uint64, event *Event) *AppLogEntry {
	return &AppLogEntry{
		ID:    id,
		Event: event,
	}
}
