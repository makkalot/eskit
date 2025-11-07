package provider

import (
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/eventstore"
	"github.com/makkalot/eskit/lib/types"
)

// CamConfig represents a camera configuration with various settings
type CamConfig struct {
	Originator *types.Originator
	CameraID   string  // Unique identifier for the camera
	Name       string  // Friendly name for the camera
	Gamma      float64 // Gamma correction value
	Exposure   int     // Exposure time in microseconds
	Saturation int     // Saturation level (0-100)
	Sharpness  int     // Sharpness level (0-100)
	Gain       int     // Gain value (0-100)
}

type CamConfigServiceProvider struct {
	crudStore  crudstore.Client
	eventStore eventstore.Store
}

func NewCamConfigServiceProvider(crudstore crudstore.Client, eventStore eventstore.Store) (*CamConfigServiceProvider, error) {
	return &CamConfigServiceProvider{
		crudStore:  crudstore,
		eventStore: eventStore,
	}, nil
}
