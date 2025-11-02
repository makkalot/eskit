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

type CreateUserRequest struct {
	Email     string `json:"email"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
}

type UpdateUserRequest struct {
	Email      string   `json:"email,omitempty"`
	FirstName  string   `json:"firstName,omitempty"`
	LastName   string   `json:"lastName,omitempty"`
	Active     bool     `json:"active"`
	Workspaces []string `json:"workspaces,omitempty"`
}

type UserResponse struct {
	Originator *types.Originator `json:"originator"`
	Email      string            `json:"email"`
	FirstName  string            `json:"firstName"`
	LastName   string            `json:"lastName"`
	Active     bool              `json:"active"`
	Workspaces []string          `json:"workspaces"`
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

func userToResponse(u *User) *UserResponse {
	if u == nil {
		return nil
	}
	return &UserResponse{
		Originator: u.Originator,
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Active:     u.Active,
		Workspaces: u.Workspaces,
	}
}

// REST Handlers

func (u *UserServiceProvider) HealthHandler(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, HealthResponse{Status: "ok"})
}

func (u *UserServiceProvider) CreateUserHandler(w http.ResponseWriter, r *http.Request) {
	var req CreateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON request body")
		return
	}

	// Create native user
	nativeUser := &User{
		Email:     req.Email,
		FirstName: req.FirstName,
		LastName:  req.LastName,
	}

	// Use library with native types
	_, err := u.crudStore.Create(nativeUser)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "creation_failed", err.Error())
		return
	}

	// Return created user
	writeJSON(w, http.StatusCreated, userToResponse(nativeUser))
}

func (u *UserServiceProvider) GetUserHandler(w http.ResponseWriter, r *http.Request) {
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
	retrievedUser := &User{}
	if err := u.crudStore.Get(nativeOriginator, retrievedUser, fetchDeleted); err != nil {
		if errors.Is(err, crudstore.RecordNotFound) || errors.Is(err, crudstore.RecordDeleted) {
			writeError(w, http.StatusNotFound, "not_found", "User not found or deleted")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, userToResponse(retrievedUser))
}

func (u *UserServiceProvider) UpdateUserHandler(w http.ResponseWriter, r *http.Request) {
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

	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeError(w, http.StatusBadRequest, "invalid_request", "Invalid JSON request body")
		return
	}

	nativeOriginator := &types.Originator{
		ID:      id,
		Version: version,
	}

	// Get existing user with native types
	retrievedUser := &User{}
	if err := u.crudStore.Get(nativeOriginator, retrievedUser, false); err != nil {
		if errors.Is(err, crudstore.RecordNotFound) || errors.Is(err, crudstore.RecordDeleted) {
			writeError(w, http.StatusNotFound, "not_found", "User not found")
			return
		}
		writeError(w, http.StatusInternalServerError, "internal_error", err.Error())
		return
	}

	// Update fields from request
	if req.Email != "" {
		retrievedUser.Email = req.Email
	}
	if req.FirstName != "" {
		retrievedUser.FirstName = req.FirstName
	}
	if req.LastName != "" {
		retrievedUser.LastName = req.LastName
	}
	retrievedUser.Active = req.Active
	if req.Workspaces != nil {
		retrievedUser.Workspaces = req.Workspaces
	}

	// Update using library with native types
	updatedOriginator, err := u.crudStore.Update(retrievedUser)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "update_failed", err.Error())
		return
	}

	retrievedUser.Originator = updatedOriginator
	writeJSON(w, http.StatusOK, userToResponse(retrievedUser))
}

func (u *UserServiceProvider) DeleteUserHandler(w http.ResponseWriter, r *http.Request) {
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
	deletedOriginator, err := u.crudStore.Delete(nativeOriginator, &User{})
	if err != nil {
		writeError(w, http.StatusInternalServerError, "delete_failed", err.Error())
		return
	}

	writeJSON(w, http.StatusOK, map[string]*types.Originator{
		"originator": deletedOriginator,
	})
}
