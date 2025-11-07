package provider

import (
	"encoding/json"
	"errors"
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/types"
	"net/http"
	"strconv"
)

// REST API Request/Response types

type CreateCamConfigRequest struct {
	CameraID   string  `json:"cameraId"`
	Name       string  `json:"name"`
	Gamma      float64 `json:"gamma"`
	Exposure   int     `json:"exposure"`
	Saturation int     `json:"saturation"`
	Sharpness  int     `json:"sharpness"`
	Gain       int     `json:"gain"`
}

type UpdateCamConfigRequest struct {
	CameraID   string  `json:"cameraId,omitempty"`
	Name       string  `json:"name,omitempty"`
	Gamma      float64 `json:"gamma,omitempty"`
	Exposure   int     `json:"exposure,omitempty"`
	Saturation int     `json:"saturation,omitempty"`
	Sharpness  int     `json:"sharpness,omitempty"`
	Gain       int     `json:"gain,omitempty"`
}

type CamConfigResponse struct {
	Originator *types.Originator `json:"originator"`
	CameraID   string            `json:"cameraId"`
	Name       string            `json:"name"`
	Gamma      float64           `json:"gamma"`
	Exposure   int               `json:"exposure"`
	Saturation int               `json:"saturation"`
	Sharpness  int               `json:"sharpness"`
	Gain       int               `json:"gain"`
}

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type HealthResponse struct {
	Status string `json:"status"`
}

// Helper functions

func writeJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func writeError(w http.ResponseWriter, status int, err string, message string) {
	writeJSON(w, status, ErrorResponse{
		Error:   err,
		Message: message,
	})
}

func camConfigToResponse(c *CamConfig) *CamConfigResponse {
	if c == nil {
		return nil
	}
	return &CamConfigResponse{
		Originator: c.Originator,
		CameraID:   c.CameraID,
		Name:       c.Name,
		Gamma:      c.Gamma,
		Exposure:   c.Exposure,
		Saturation: c.Saturation,
		Sharpness:  c.Sharpness,
		Gain:       c.Gain,
	}
}

// REST Handlers

func (s *CamConfigServiceProvider) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (s *CamConfigServiceProvider) CreateCamConfigHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateCamConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON request body")
		return
	}

	// Create native camconfig
	nativeConfig := &CamConfig{
		CameraID:   req.CameraID,
		Name:       req.Name,
		Gamma:      req.Gamma,
		Exposure:   req.Exposure,
		Saturation: req.Saturation,
		Sharpness:  req.Sharpness,
		Gain:       req.Gain,
	}

	// Use library with native types
	_, err := s.crudStore.Create(nativeConfig)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	// Return created config
	writeJSON(w, http.StatusCreated, camConfigToResponse(nativeConfig))
}

func (s *CamConfigServiceProvider) GetCamConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID and Version from query parameters
	id := r.URL.Query().Get("id")
	versionStr := r.URL.Query().Get("version")
	fetchDeleted := r.URL.Query().Get("fetchDeleted") == "true"

	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}

	var version uint64
	if versionStr != "" {
		v, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "Invalid 'version' parameter")
			return
		}
		version = v
	}

	nativeOriginator := &types.Originator{
		ID:      id,
		Version: version,
	}

	// Use library with native types
	retrievedConfig := &CamConfig{}
	if err := s.crudStore.Get(nativeOriginator, retrievedConfig, fetchDeleted); err != nil {
		if errors.Is(err, crudstore.RecordNotFound) || errors.Is(err, crudstore.RecordDeleted) {
			writeError(w, http.StatusNotFound, "not_found", "Camera config not found or deleted")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, camConfigToResponse(retrievedConfig))
}

func (s *CamConfigServiceProvider) UpdateCamConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID and Version from query parameters
	id := r.URL.Query().Get("id")
	versionStr := r.URL.Query().Get("version")

	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}

	var version uint64
	if versionStr != "" {
		v, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "Invalid 'version' parameter")
			return
		}
		version = v
	}

	var req UpdateCamConfigRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON request body")
		return
	}

	nativeOriginator := &types.Originator{
		ID:      id,
		Version: version,
	}

	// Get existing config with native types
	retrievedConfig := &CamConfig{}
	if err := s.crudStore.Get(nativeOriginator, retrievedConfig, false); err != nil {
		if errors.Is(err, crudstore.RecordNotFound) || errors.Is(err, crudstore.RecordDeleted) {
			writeError(w, http.StatusNotFound, "not_found", "Camera config not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Update fields from request (only if provided)
	if req.CameraID != "" {
		retrievedConfig.CameraID = req.CameraID
	}
	if req.Name != "" {
		retrievedConfig.Name = req.Name
	}
	if req.Gamma != 0 {
		retrievedConfig.Gamma = req.Gamma
	}
	if req.Exposure != 0 {
		retrievedConfig.Exposure = req.Exposure
	}
	if req.Saturation != 0 {
		retrievedConfig.Saturation = req.Saturation
	}
	if req.Sharpness != 0 {
		retrievedConfig.Sharpness = req.Sharpness
	}
	if req.Gain != 0 {
		retrievedConfig.Gain = req.Gain
	}

	// Update using library with native types
	updatedOriginator, err := s.crudStore.Update(retrievedConfig)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	retrievedConfig.Originator = updatedOriginator
	writeJSON(w, http.StatusOK, camConfigToResponse(retrievedConfig))
}

func (s *CamConfigServiceProvider) DeleteCamConfigHandler(w http.ResponseWriter, r *http.Request) {
	// Get ID and Version from query parameters
	id := r.URL.Query().Get("id")
	versionStr := r.URL.Query().Get("version")

	if id == "" {
		writeError(w, http.StatusBadRequest, "invalid_request", "Missing 'id' parameter")
		return
	}

	var version uint64
	if versionStr != "" {
		v, err := strconv.ParseUint(versionStr, 10, 64)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid_request", "Invalid 'version' parameter")
			return
		}
		version = v
	}

	nativeOriginator := &types.Originator{
		ID:      id,
		Version: version,
	}

	// Delete using library with native types
	deletedOriginator, err := s.crudStore.Delete(nativeOriginator, &CamConfig{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]*types.Originator{
		"originator": deletedOriginator,
	})
}
