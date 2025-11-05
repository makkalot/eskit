package provider

import (
	"encoding/json"
	"fmt"
	"github.com/makkalot/eskit/lib/types"
	"html/template"
	"net/http"
	"path/filepath"
	"strconv"
	"strings"
)

var templates *template.Template

// InitTemplates loads HTML templates
func InitTemplates(templateDir string) error {
	var err error
	templates, err = template.ParseGlob(filepath.Join(templateDir, "*.html"))
	return err
}

// WebIndexHandler displays the list of camera configurations
func (s *CamConfigServiceProvider) WebIndexHandler(w http.ResponseWriter, r *http.Request) {
	// List all configs
	var configs []*CamConfig
	_, err := s.crudStore.ListWithPagination(&configs, "", 100)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to list configs: %v", err), http.StatusInternalServerError)
		return
	}

	data := map[string]interface{}{
		"Configs": configs,
	}

	if err := templates.ExecuteTemplate(w, "index.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// WebCreateHandler displays the create form or handles form submission
func (s *CamConfigServiceProvider) WebCreateHandler(w http.ResponseWriter, r *http.Request) {
	if r.Method == http.MethodGet {
		// Show create form
		if err := templates.ExecuteTemplate(w, "form.html", map[string]interface{}{
			"Title":  "Create Camera Configuration",
			"Action": "/web/create",
		}); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Handle POST - create new config
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	gamma, _ := strconv.ParseFloat(r.FormValue("gamma"), 64)
	exposure, _ := strconv.Atoi(r.FormValue("exposure"))
	saturation, _ := strconv.Atoi(r.FormValue("saturation"))
	sharpness, _ := strconv.Atoi(r.FormValue("sharpness"))
	gain, _ := strconv.Atoi(r.FormValue("gain"))

	config := &CamConfig{
		CameraID:   r.FormValue("cameraId"),
		Name:       r.FormValue("name"),
		Gamma:      gamma,
		Exposure:   exposure,
		Saturation: saturation,
		Sharpness:  sharpness,
		Gain:       gain,
	}

	_, err := s.crudStore.Create(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to create config: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/web/", http.StatusSeeOther)
}

// WebEditHandler displays the edit form or handles form submission
func (s *CamConfigServiceProvider) WebEditHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	// Get current config
	originator := &types.Originator{ID: id}
	config := &CamConfig{}
	if err := s.crudStore.Get(originator, config, false); err != nil {
		http.Error(w, fmt.Sprintf("Failed to get config: %v", err), http.StatusInternalServerError)
		return
	}

	if r.Method == http.MethodGet {
		// Show edit form
		data := map[string]interface{}{
			"Title":  "Edit Camera Configuration",
			"Action": "/web/edit?id=" + id,
			"Config": config,
		}
		if err := templates.ExecuteTemplate(w, "form.html", data); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
		}
		return
	}

	// Handle POST - update config
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Failed to parse form", http.StatusBadRequest)
		return
	}

	config.CameraID = r.FormValue("cameraId")
	config.Name = r.FormValue("name")
	config.Gamma, _ = strconv.ParseFloat(r.FormValue("gamma"), 64)
	config.Exposure, _ = strconv.Atoi(r.FormValue("exposure"))
	config.Saturation, _ = strconv.Atoi(r.FormValue("saturation"))
	config.Sharpness, _ = strconv.Atoi(r.FormValue("sharpness"))
	config.Gain, _ = strconv.Atoi(r.FormValue("gain"))

	_, err := s.crudStore.Update(config)
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to update config: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/web/", http.StatusSeeOther)
}

// WebDeleteHandler handles deletion
func (s *CamConfigServiceProvider) WebDeleteHandler(w http.ResponseWriter, r *http.Request) {
	id := r.URL.Query().Get("id")
	if id == "" {
		http.Error(w, "Missing id parameter", http.StatusBadRequest)
		return
	}

	originator := &types.Originator{ID: id}
	_, err := s.crudStore.Delete(originator, &CamConfig{})
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to delete config: %v", err), http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, "/web/", http.StatusSeeOther)
}

// AuditLogEntry represents a parsed audit log entry for display
type AuditLogEntry struct {
	ID         uint64
	EventType  string
	CameraID   string
	ConfigID   string
	Version    uint64
	OccurredOn string
	Changes    []FieldChange
}

// FieldChange represents a change to a field
type FieldChange struct {
	Field    string
	OldValue string
	NewValue string
}

// WebAuditLogHandler displays the audit log
func (s *CamConfigServiceProvider) WebAuditLogHandler(w http.ResponseWriter, r *http.Request) {
	// Get filter parameters
	filterID := r.URL.Query().Get("id")

	// Get application logs
	logs, err := s.eventStore.Logs(0, 1000, "")
	if err != nil {
		http.Error(w, fmt.Sprintf("Failed to get logs: %v", err), http.StatusInternalServerError)
		return
	}

	// Parse logs and filter for CamConfig events
	var auditEntries []AuditLogEntry
	previousStates := make(map[string]*CamConfig) // Store previous states by ID

	for _, log := range logs {
		event := log.Event

		// Only show CamConfig events
		if !strings.HasPrefix(event.EventType, "CamConfig.") {
			continue
		}

		// Apply filter if specified
		if filterID != "" && event.Originator.ID != filterID {
			continue
		}

		entry := AuditLogEntry{
			ID:         log.ID,
			EventType:  event.EventType,
			ConfigID:   event.Originator.ID,
			Version:    event.Originator.Version,
			OccurredOn: event.OccurredOn.Format("2006-01-02 15:04:05"),
		}

		// Parse the payload to get the config state
		var currentConfig CamConfig
		if err := json.Unmarshal([]byte(event.Payload), &currentConfig); err == nil {
			entry.CameraID = currentConfig.CameraID

			// Calculate changes if this is an update
			if strings.HasSuffix(event.EventType, ".Updated") {
				if prevConfig, exists := previousStates[event.Originator.ID]; exists {
					entry.Changes = calculateChanges(prevConfig, &currentConfig)
				}
			}

			// Store current state for next iteration
			previousStates[event.Originator.ID] = &currentConfig
		}

		auditEntries = append(auditEntries, entry)
	}

	data := map[string]interface{}{
		"Entries":  auditEntries,
		"FilterID": filterID,
	}

	if err := templates.ExecuteTemplate(w, "audit.html", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// calculateChanges compares two configs and returns the changes
func calculateChanges(old, new *CamConfig) []FieldChange {
	var changes []FieldChange

	if old.CameraID != new.CameraID {
		changes = append(changes, FieldChange{"Camera ID", old.CameraID, new.CameraID})
	}
	if old.Name != new.Name {
		changes = append(changes, FieldChange{"Name", old.Name, new.Name})
	}
	if old.Gamma != new.Gamma {
		changes = append(changes, FieldChange{"Gamma", fmt.Sprintf("%.2f", old.Gamma), fmt.Sprintf("%.2f", new.Gamma)})
	}
	if old.Exposure != new.Exposure {
		changes = append(changes, FieldChange{"Exposure", strconv.Itoa(old.Exposure), strconv.Itoa(new.Exposure)})
	}
	if old.Saturation != new.Saturation {
		changes = append(changes, FieldChange{"Saturation", strconv.Itoa(old.Saturation), strconv.Itoa(new.Saturation)})
	}
	if old.Sharpness != new.Sharpness {
		changes = append(changes, FieldChange{"Sharpness", strconv.Itoa(old.Sharpness), strconv.Itoa(new.Sharpness)})
	}
	if old.Gain != new.Gain {
		changes = append(changes, FieldChange{"Gain", strconv.Itoa(old.Gain), strconv.Itoa(new.Gain)})
	}

	return changes
}
