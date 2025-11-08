package provider

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/jinzhu/gorm"
	"github.com/makkalot/eskit/lib/types"
)

var templates *template.Template

// InitTemplates loads all HTML templates
func InitTemplates(templatePath string) error {
	funcMap := template.FuncMap{
		"hasSuffix": strings.HasSuffix,
	}

	var err error
	templates, err = template.New("").Funcs(funcMap).ParseGlob(templatePath + "/*.html")
	if err != nil {
		return fmt.Errorf("parsing templates: %w", err)
	}
	return nil
}

// StoredEvent represents a row in the stored_events table
type StoredEvent struct {
	OriginatorID      string `gorm:"column:originator_id"`
	OriginatorVersion uint64 `gorm:"column:originator_version"`
	EventType         string `gorm:"column:event_type"`
	Payload           string `gorm:"column:payload"`
}

func (StoredEvent) TableName() string {
	return "stored_events"
}

// StoredLogEntry represents a row in the stored_log_entries table
type StoredLogEntry struct {
	ID           uint64 `gorm:"column:id;primaryKey"`
	PartitionID  string `gorm:"column:partition_id"`
	EventPayload string `gorm:"column:event_payload"`
}

func (StoredLogEntry) TableName() string {
	return "stored_log_entries"
}

// EventRow represents an event for display
type EventRow struct {
	OriginatorID      string
	OriginatorVersion uint64
	EventType         string
	Payload           string
	OccurredOn        time.Time
}

// EventsPageData holds data for the events page
type EventsPageData struct {
	Title        string
	Events       []EventRow
	OriginatorID string
	EventType    string
	DateFrom     string
	DateTo       string
	Page         int
	NextPage     int
	HasMore      bool
}

// AppLogRow represents an app log entry for display
type AppLogRow struct {
	ID                uint64
	PartitionID       string
	OriginatorID      string
	OriginatorVersion uint64
	EventType         string
	Payload           string
	OccurredOn        time.Time
}

// AppLogPageData holds data for the applog page
type AppLogPageData struct {
	Title       string
	Entries     []AppLogRow
	PartitionID string
	EventType   string
	DateFrom    string
	DateTo      string
	FromID      uint64
	NextID      uint64
	HasMore     bool
}

// EntityType represents an entity type with count
type EntityType struct {
	Name  string
	Count int
}

// CrudEntity represents a CRUD entity for display
type CrudEntity struct {
	OriginatorID string
	Version      uint64
	Data         string
	UpdatedAt    time.Time
	IsDeleted    bool
}

// CrudPageData holds data for the CRUD page
type CrudPageData struct {
	Title           string
	EntityTypes     []EntityType
	SelectedType    string
	Entities        []CrudEntity
	ShowingEntities bool
}

// EntityEvent represents an event in entity history
type EntityEvent struct {
	Version    uint64
	EventType  string
	Payload    string
	OccurredOn time.Time
}

// CrudEntityDetailData holds data for entity detail page
type CrudEntityDetailData struct {
	Title        string
	EntityType   string
	OriginatorID string
	Entity       CrudEntity
	Events       []EntityEvent
}

// HomeHandler redirects to the events page
func (p *AdminProvider) HomeHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/events", http.StatusFound)
	}
}

// EventsHandler handles the raw events query page
func (p *AdminProvider) EventsHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		originatorID := r.URL.Query().Get("originator_id")
		eventType := r.URL.Query().Get("event_type")
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")
		page, _ := strconv.Atoi(r.URL.Query().Get("page"))
		if page < 0 {
			page = 0
		}

		// Build query (no date filtering in SQL since created_at doesn't exist)
		query := db.Model(&StoredEvent{})

		if originatorID != "" {
			query = query.Where("originator_id = ?", originatorID)
		}
		if eventType != "" {
			query = query.Where("event_type LIKE ?", "%"+eventType+"%")
		}

		// Get all matching events (we'll filter by date in memory)
		var events []StoredEvent
		if err := query.Order("originator_version DESC").
			Find(&events).Error; err != nil {
			log.Printf("Error querying events: %v", err)
			http.Error(w, "Error querying events", http.StatusInternalServerError)
			return
		}

		// Parse timestamps and apply date filters
		var eventRows []EventRow
		var dateFromTime, dateToTime time.Time
		if dateFrom != "" {
			dateFromTime, _ = time.Parse("2006-01-02T15:04", dateFrom)
		}
		if dateTo != "" {
			dateToTime, _ = time.Parse("2006-01-02T15:04", dateTo)
		}

		for _, e := range events {
			// Parse timestamp from event payload
			var payloadData map[string]interface{}
			timestamp := time.Time{}
			if err := json.Unmarshal([]byte(e.Payload), &payloadData); err == nil {
				if occurredOnStr, ok := payloadData["occurredOn"].(string); ok {
					timestamp, _ = time.Parse(time.RFC3339, occurredOnStr)
				}
			}

			// Apply date filters
			if !dateFromTime.IsZero() && timestamp.Before(dateFromTime) {
				continue
			}
			if !dateToTime.IsZero() && timestamp.After(dateToTime) {
				continue
			}

			eventRows = append(eventRows, EventRow{
				OriginatorID:      e.OriginatorID,
				OriginatorVersion: e.OriginatorVersion,
				EventType:         e.EventType,
				Payload:           e.Payload,
				OccurredOn:        timestamp,
			})
		}

		// Apply pagination
		pageSize := 50
		offset := page * pageSize
		hasMore := len(eventRows) > offset+pageSize

		endIdx := offset + pageSize
		if endIdx > len(eventRows) {
			endIdx = len(eventRows)
		}
		if offset > len(eventRows) {
			offset = len(eventRows)
		}
		paginatedEvents := eventRows[offset:endIdx]

		data := EventsPageData{
			Title:        "Raw Events",
			Events:       paginatedEvents,
			OriginatorID: originatorID,
			EventType:    eventType,
			DateFrom:     dateFrom,
			DateTo:       dateTo,
			Page:         page,
			NextPage:     page + 1,
			HasMore:      hasMore,
		}

		// Check if this is an HTMX request for the table only
		if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "events-table" {
			if err := templates.ExecuteTemplate(w, "events-table", data); err != nil {
				log.Printf("Error executing events-table template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		} else {
			if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
				log.Printf("Error executing layout template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
}

// AppLogHandler handles the application log query page
func (p *AdminProvider) AppLogHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		partitionID := r.URL.Query().Get("partition_id")
		eventType := r.URL.Query().Get("event_type")
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")
		fromID, _ := strconv.ParseUint(r.URL.Query().Get("from_id"), 10, 64)

		// Build query
		query := db.Model(&StoredLogEntry{})

		if partitionID != "" {
			query = query.Where("partition_id = ?", partitionID)
		}
		if fromID > 0 {
			query = query.Where("id >= ?", fromID)
		}

		// For event type and date filtering, we need to parse the JSON payload
		pageSize := 50

		var entries []StoredLogEntry
		if err := query.Order("id ASC").
			Limit(pageSize + 1).
			Find(&entries).Error; err != nil {
			log.Printf("Error querying app log: %v", err)
			http.Error(w, "Error querying app log", http.StatusInternalServerError)
			return
		}

		// Check if there are more results
		hasMore := len(entries) > pageSize
		if hasMore {
			entries = entries[:pageSize]
		}

		// Parse event payloads and apply filters
		var logRows []AppLogRow
		for _, entry := range entries {
			var event types.Event
			if err := json.Unmarshal([]byte(entry.EventPayload), &event); err != nil {
				log.Printf("Error parsing event payload: %v", err)
				continue
			}

			// Apply event type filter
			if eventType != "" && !strings.Contains(event.EventType, eventType) {
				continue
			}

			// Apply date filters
			if dateFrom != "" {
				if t, err := time.Parse("2006-01-02T15:04", dateFrom); err == nil {
					if event.OccurredOn.Before(t) {
						continue
					}
				}
			}
			if dateTo != "" {
				if t, err := time.Parse("2006-01-02T15:04", dateTo); err == nil {
					if event.OccurredOn.After(t) {
						continue
					}
				}
			}

			logRows = append(logRows, AppLogRow{
				ID:                entry.ID,
				PartitionID:       entry.PartitionID,
				OriginatorID:      event.Originator.ID,
				OriginatorVersion: event.Originator.Version,
				EventType:         event.EventType,
				Payload:           event.Payload,
				OccurredOn:        event.OccurredOn,
			})
		}

		nextID := uint64(0)
		if len(entries) > 0 {
			nextID = entries[len(entries)-1].ID + 1
		}

		data := AppLogPageData{
			Title:       "Application Log",
			Entries:     logRows,
			PartitionID: partitionID,
			EventType:   eventType,
			DateFrom:    dateFrom,
			DateTo:      dateTo,
			FromID:      fromID,
			NextID:      nextID,
			HasMore:     hasMore,
		}

		// Check if this is an HTMX request for the table only
		if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "applog-table" {
			if err := templates.ExecuteTemplate(w, "applog-table", data); err != nil {
				log.Printf("Error executing applog-table template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		} else {
			if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
				log.Printf("Error executing layout template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
}

// CrudHandler handles the CRUD entities page
func (p *AdminProvider) CrudHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityType := r.URL.Query().Get("type")

		if entityType == "" {
			// Show entity types list
			entityTypes, err := p.discoverEntityTypes(db)
			if err != nil {
				log.Printf("Error discovering entity types: %v", err)
				http.Error(w, "Error discovering entity types", http.StatusInternalServerError)
				return
			}

			data := CrudPageData{
				Title:           "CRUD Entities",
				EntityTypes:     entityTypes,
				ShowingEntities: false,
			}

			if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		} else {
			// Show entities of a specific type
			entities, err := p.getEntitiesOfType(db, entityType)
			if err != nil {
				log.Printf("Error getting entities: %v", err)
				http.Error(w, "Error getting entities", http.StatusInternalServerError)
				return
			}

			data := CrudPageData{
				Title:           "CRUD Entities",
				SelectedType:    entityType,
				Entities:        entities,
				ShowingEntities: true,
			}

			if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
				log.Printf("Error executing template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
}

// CrudEntityDetailHandler handles the entity detail page with event history
func (p *AdminProvider) CrudEntityDetailHandler(db *gorm.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityType := r.URL.Query().Get("type")
		originatorID := r.URL.Query().Get("id")

		if entityType == "" || originatorID == "" {
			http.Error(w, "Missing type or id parameter", http.StatusBadRequest)
			return
		}

		// Get all events for this entity
		var storedEvents []StoredEvent
		if err := db.Where("originator_id = ?", originatorID).
			Order("originator_version ASC").
			Find(&storedEvents).Error; err != nil {
			log.Printf("Error getting entity events: %v", err)
			http.Error(w, "Error getting entity events", http.StatusInternalServerError)
			return
		}

		if len(storedEvents) == 0 {
			http.Error(w, "Entity not found", http.StatusNotFound)
			return
		}

		// Replay events to get current state
		currentState := make(map[string]interface{})
		isDeleted := false
		var lastEvent StoredEvent
		var lastTimestamp time.Time

		for _, event := range storedEvents {
			lastEvent = event

			// Parse timestamp from payload
			var payloadData map[string]interface{}
			if err := json.Unmarshal([]byte(event.Payload), &payloadData); err == nil {
				if occurredOnStr, ok := payloadData["occurredOn"].(string); ok {
					lastTimestamp, _ = time.Parse(time.RFC3339, occurredOnStr)
				}
			}

			if strings.HasSuffix(event.EventType, ".Created") {
				// Full object
				currentState = payloadData
			} else if strings.HasSuffix(event.EventType, ".Updated") {
				// Merge patch
				for k, v := range payloadData {
					if k != "occurredOn" {
						currentState[k] = v
					}
				}
			} else if strings.HasSuffix(event.EventType, ".Deleted") {
				isDeleted = true
			}
		}

		// Convert current state to JSON
		stateJSON, _ := json.MarshalIndent(currentState, "", "  ")

		// Convert events to template data
		eventHistory := make([]EntityEvent, len(storedEvents))
		for i, e := range storedEvents {
			// Parse timestamp for this event
			var payloadData map[string]interface{}
			eventTimestamp := time.Time{}
			if err := json.Unmarshal([]byte(e.Payload), &payloadData); err == nil {
				if occurredOnStr, ok := payloadData["occurredOn"].(string); ok {
					eventTimestamp, _ = time.Parse(time.RFC3339, occurredOnStr)
				}
			}

			eventHistory[i] = EntityEvent{
				Version:    e.OriginatorVersion,
				EventType:  e.EventType,
				Payload:    e.Payload,
				OccurredOn: eventTimestamp,
			}
		}

		data := CrudEntityDetailData{
			Title:        "Entity Details",
			EntityType:   entityType,
			OriginatorID: originatorID,
			Entity: CrudEntity{
				OriginatorID: originatorID,
				Version:      lastEvent.OriginatorVersion,
				Data:         string(stateJSON),
				UpdatedAt:    lastTimestamp,
				IsDeleted:    isDeleted,
			},
			Events: eventHistory,
		}

		if err := templates.ExecuteTemplate(w, "layout", data); err != nil {
			log.Printf("Error executing template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

// discoverEntityTypes queries the database to find all unique entity types
func (p *AdminProvider) discoverEntityTypes(db *gorm.DB) ([]EntityType, error) {
	// Query distinct partition IDs from stored_log_entries
	var partitions []string
	if err := db.Table("stored_log_entries").
		Select("DISTINCT partition_id").
		Pluck("partition_id", &partitions).Error; err != nil {
		return nil, err
	}

	// Count entities for each partition
	var entityTypes []EntityType
	for _, partition := range partitions {
		// Count unique originator IDs for this partition
		type CountResult struct {
			Count int
		}
		var result CountResult
		if err := db.Table("stored_events").
			Select("COUNT(DISTINCT originator_id) as count").
			Where("event_type LIKE ?", partition+".%").
			Scan(&result).Error; err != nil {
			log.Printf("Error counting entities for partition %s: %v", partition, err)
			continue
		}

		entityTypes = append(entityTypes, EntityType{
			Name:  partition,
			Count: result.Count,
		})
	}

	return entityTypes, nil
}

// getEntitiesOfType retrieves all entities of a specific type with their current state
func (p *AdminProvider) getEntitiesOfType(db *gorm.DB, entityType string) ([]CrudEntity, error) {
	// Get all unique originator IDs for this entity type
	var originatorIDs []string
	if err := db.Table("stored_events").
		Where("event_type LIKE ?", entityType+".%").
		Group("originator_id").
		Pluck("originator_id", &originatorIDs).Error; err != nil {
		return nil, err
	}

	var entities []CrudEntity

	for _, id := range originatorIDs {
		// Get all events for this entity
		var storedEvents []StoredEvent
		if err := db.Where("originator_id = ?", id).
			Order("originator_version ASC").
			Find(&storedEvents).Error; err != nil {
			log.Printf("Error getting events for entity %s: %v", id, err)
			continue
		}

		if len(storedEvents) == 0 {
			continue
		}

		// Replay events to get current state
		currentState := make(map[string]interface{})
		isDeleted := false
		var lastEvent StoredEvent
		var lastTimestamp time.Time

		for _, event := range storedEvents {
			lastEvent = event

			// Parse timestamp from payload
			var payloadData map[string]interface{}
			if err := json.Unmarshal([]byte(event.Payload), &payloadData); err == nil {
				if occurredOnStr, ok := payloadData["occurredOn"].(string); ok {
					lastTimestamp, _ = time.Parse(time.RFC3339, occurredOnStr)
				}
			}

			if strings.HasSuffix(event.EventType, ".Created") {
				currentState = payloadData
			} else if strings.HasSuffix(event.EventType, ".Updated") {
				for k, v := range payloadData {
					if k != "occurredOn" {
						currentState[k] = v
					}
				}
			} else if strings.HasSuffix(event.EventType, ".Deleted") {
				isDeleted = true
			}
		}

		stateJSON, _ := json.Marshal(currentState)

		entities = append(entities, CrudEntity{
			OriginatorID: id,
			Version:      lastEvent.OriginatorVersion,
			Data:         string(stateJSON),
			UpdatedAt:    lastTimestamp,
			IsDeleted:    isDeleted,
		})
	}

	return entities, nil
}

// SetupRoutes configures all HTTP routes for the admin service
func (p *AdminProvider) SetupRoutes(db *gorm.DB, mux *http.ServeMux) {
	// Static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	mux.HandleFunc("/", p.HomeHandler(db))
	mux.HandleFunc("/events", p.EventsHandler(db))
	mux.HandleFunc("/applog", p.AppLogHandler(db))
	mux.HandleFunc("/crud", p.CrudHandler(db))
	mux.HandleFunc("/crud/entity", p.CrudEntityDetailHandler(db))
}
