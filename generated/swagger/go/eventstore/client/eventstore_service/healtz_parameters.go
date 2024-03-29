package eventstore_service

// This file was generated by the swagger tool.
// Editing this file might prove futile when you re-run the swagger generate command

import (
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/go-openapi/errors"
	"github.com/go-openapi/runtime"
	cr "github.com/go-openapi/runtime/client"

	strfmt "github.com/go-openapi/strfmt"
)

// NewHealtzParams creates a new HealtzParams object
// with the default values initialized.
func NewHealtzParams() *HealtzParams {

	return &HealtzParams{

		timeout: cr.DefaultTimeout,
	}
}

// NewHealtzParamsWithTimeout creates a new HealtzParams object
// with the default values initialized, and the ability to set a timeout on a request
func NewHealtzParamsWithTimeout(timeout time.Duration) *HealtzParams {

	return &HealtzParams{

		timeout: timeout,
	}
}

// NewHealtzParamsWithContext creates a new HealtzParams object
// with the default values initialized, and the ability to set a context for a request
func NewHealtzParamsWithContext(ctx context.Context) *HealtzParams {

	return &HealtzParams{

		Context: ctx,
	}
}

// NewHealtzParamsWithHTTPClient creates a new HealtzParams object
// with the default values initialized, and the ability to set a custom HTTPClient for a request
func NewHealtzParamsWithHTTPClient(client *http.Client) *HealtzParams {

	return &HealtzParams{
		HTTPClient: client,
	}
}

/*HealtzParams contains all the parameters to send to the API endpoint
for the healtz operation typically these are written to a http.Request
*/
type HealtzParams struct {
	timeout    time.Duration
	Context    context.Context
	HTTPClient *http.Client
}

// WithTimeout adds the timeout to the healtz params
func (o *HealtzParams) WithTimeout(timeout time.Duration) *HealtzParams {
	o.SetTimeout(timeout)
	return o
}

// SetTimeout adds the timeout to the healtz params
func (o *HealtzParams) SetTimeout(timeout time.Duration) {
	o.timeout = timeout
}

// WithContext adds the context to the healtz params
func (o *HealtzParams) WithContext(ctx context.Context) *HealtzParams {
	o.SetContext(ctx)
	return o
}

// SetContext adds the context to the healtz params
func (o *HealtzParams) SetContext(ctx context.Context) {
	o.Context = ctx
}

// WithHTTPClient adds the HTTPClient to the healtz params
func (o *HealtzParams) WithHTTPClient(client *http.Client) *HealtzParams {
	o.SetHTTPClient(client)
	return o
}

// SetHTTPClient adds the HTTPClient to the healtz params
func (o *HealtzParams) SetHTTPClient(client *http.Client) {
	o.HTTPClient = client
}

// WriteToRequest writes these params to a swagger request
func (o *HealtzParams) WriteToRequest(r runtime.ClientRequest, reg strfmt.Registry) error {

	if err := r.SetTimeout(o.timeout); err != nil {
		return err
	}
	var res []error

	if len(res) > 0 {
		return errors.CompositeValidationError(res...)
	}
	return nil
}
