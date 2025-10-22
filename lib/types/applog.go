package types

// AppLogEntry represents an entry in the application log.
// The application log is a sequential stream of all events in the system,
// similar to Kafka's event log. Each entry has a unique sequential ID.
type AppLogEntry struct {
	// ID is the sequential ID of this entry in the application log
	// IDs are monotonically increasing strings (e.g., "1", "2", "3", ...)
	ID string `json:"id" gorm:"column:id;primaryKey"`

	// Event is the event stored in this log entry
	Event *Event `json:"event" gorm:"embedded;embeddedPrefix:event_"`
}

// NewAppLogEntry creates a new AppLogEntry with the given ID and event
func NewAppLogEntry(id string, event *Event) *AppLogEntry {
	return &AppLogEntry{
		ID:    id,
		Event: event,
	}
}
