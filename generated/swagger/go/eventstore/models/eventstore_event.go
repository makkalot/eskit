package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// EventstoreEvent Event is what you operate on with event store
// It's the smalled bit that the event store is aware of
// swagger:model eventstoreEvent
type EventstoreEvent struct {

	// this is the event type that this is related to
	// event type should be in the format of `Entity.Created` so that store can infer the
	// partition this event belongs to
	EventType string `json:"event_type,omitempty"`

	// utc unix timestamp of the event occurence
	OccuredOn int64 `json:"occured_on,omitempty"`

	// The object this event belongs to
	Originator *CommonOriginator `json:"originator,omitempty"`

	// the data of the event is inside the payload
	Payload string `json:"payload,omitempty"`
}

// Validate validates this eventstore event
func (m *EventstoreEvent) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateOriginator(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *EventstoreEvent) validateOriginator(formats strfmt.Registry) error {

	if swag.IsZero(m.Originator) { // not required
		return nil
	}

	if m.Originator != nil {

		if err := m.Originator.Validate(formats); err != nil {
			if ve, ok := err.(*errors.Validation); ok {
				return ve.ValidateName("originator")
			}
			return err
		}
	}

	return nil
}
