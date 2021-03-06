package models

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"strconv"

	strfmt "github.com/go-openapi/strfmt"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/swag"
)

// CrudstoreListResponse crudstore list response
// swagger:model crudstoreListResponse
type CrudstoreListResponse struct {

	// next page id
	NextPageID string `json:"next_page_id,omitempty"`

	// results
	Results []*CrudstoreListResponseItem `json:"results"`
}

// Validate validates this crudstore list response
func (m *CrudstoreListResponse) Validate(formats strfmt.Registry) error {
	var res []error

	if err := m.validateResults(formats); err != nil {
		// prop
		res = append(res, err)
	}

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}

func (m *CrudstoreListResponse) validateResults(formats strfmt.Registry) error {

	if swag.IsZero(m.Results) { // not required
		return nil
	}

	for i := 0; i < len(m.Results); i++ {

		if swag.IsZero(m.Results[i]) { // not required
			continue
		}

		if m.Results[i] != nil {

			if err := m.Results[i].Validate(formats); err != nil {
				if ve, ok := err.(*errors.Validation); ok {
					return ve.ValidateName("results" + "." + strconv.Itoa(i))
				}
				return err
			}
		}

	}

	return nil
}
