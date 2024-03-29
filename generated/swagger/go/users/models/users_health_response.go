package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
)

// UsersHealthResponse users health response
// swagger:model usersHealthResponse
type UsersHealthResponse struct {

	// message
	Message string `json:"message,omitempty"`
}

// Validate validates this users health response
func (m *UsersHealthResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
