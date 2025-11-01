package types

import "time"

// Event represents a domain event in the event sourcing system.
// Events are immutable records of state changes that have occurred.
type Event struct {
	// Originator identifies the entity this event belongs to
	Originator *Originator `json:"originator" gorm:"embedded;embeddedPrefix:originator_"`

	// EventType is the type of event in the format "Entity.Action"
	// (e.g., "User.Created", "Order.Shipped")
	// The store uses this to infer the partition this event belongs to
	EventType string `json:"event_type" gorm:"column:event_type"`

	// Payload is the JSON-encoded data of the event
	Payload string `json:"payload" gorm:"column:payload;type:text"`

	// OccurredOn is the UTC timestamp when the event occurred
	OccurredOn time.Time `json:"occurred_on" gorm:"column:occurred_on"`
}

// NewEvent creates a new Event with the given parameters
func NewEvent(originator *Originator, eventType, payload string) *Event {
	return &Event{
		Originator: originator,
		EventType:  eventType,
		Payload:    payload,
		OccurredOn: time.Now().UTC(),
	}
}

// PartitionKey extracts the partition key from the event type.
// For example, "User.Created" returns "User"
func (e *Event) PartitionKey() string {
	for i, c := range e.EventType {
		if c == '.' {
			return e.EventType[:i]
		}
	}
	return e.EventType
}
