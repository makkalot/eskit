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

	"github.com/makkalot/eskit/lib/common"
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
func (p *AdminProvider) HomeHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, "/events", http.StatusFound)
	}
}

// EventsHandler handles the raw events query page using eventstore.Logs()
func (p *AdminProvider) EventsHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		originatorID := r.URL.Query().Get("originator_id")
		eventType := r.URL.Query().Get("event_type")
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")
		fromID, _ := strconv.ParseUint(r.URL.Query().Get("from_id"), 10, 64)

		// Parse date filters
		var dateFromTime, dateToTime time.Time
		if dateFrom != "" {
			dateFromTime, _ = time.Parse("2006-01-02T15:04", dateFrom)
		}
		if dateTo != "" {
			dateToTime, _ = time.Parse("2006-01-02T15:04", dateTo)
		}

		// Use eventstore.Logs() to get events
		// Note: Logs() returns all events in the system, we filter in memory
		pageSize := 50
		logs, err := p.eventStore.Logs(fromID, uint32(pageSize*10), "")
		if err != nil {
			log.Printf("Error querying logs: %v", err)
			http.Error(w, "Error querying logs", http.StatusInternalServerError)
			return
		}

		// Filter and convert to EventRows
		var eventRows []EventRow
		for _, logEntry := range logs {
			event := logEntry.Event

			// Filter by originator ID
			if originatorID != "" && event.Originator.ID != originatorID {
				continue
			}

			// Filter by event type
			if eventType != "" && !strings.Contains(event.EventType, eventType) {
				continue
			}

			// Filter by date
			if !dateFromTime.IsZero() && event.OccurredOn.Before(dateFromTime) {
				continue
			}
			if !dateToTime.IsZero() && event.OccurredOn.After(dateToTime) {
				continue
			}

			eventRows = append(eventRows, EventRow{
				OriginatorID:      event.Originator.ID,
				OriginatorVersion: event.Originator.Version,
				EventType:         event.EventType,
				Payload:           event.Payload,
				OccurredOn:        event.OccurredOn,
			})
		}

		// Apply pagination
		hasMore := len(eventRows) >= pageSize
		if len(eventRows) > pageSize {
			eventRows = eventRows[:pageSize]
		}

		nextID := fromID
		if len(logs) > 0 {
			nextID = logs[len(logs)-1].ID + 1
		}

		data := EventsPageData{
			Title:        "Raw Events",
			Events:       eventRows,
			OriginatorID: originatorID,
			EventType:    eventType,
			DateFrom:     dateFrom,
			DateTo:       dateTo,
			Page:         int(fromID / uint64(pageSize)),
			NextPage:     int(nextID / uint64(pageSize)),
			HasMore:      hasMore,
		}

		// Check if this is an HTMX request for the table only
		if r.Header.Get("HX-Request") == "true" && r.Header.Get("HX-Target") == "events-table" {
			if err := templates.ExecuteTemplate(w, "events-table", data); err != nil {
				log.Printf("Error executing events-table template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		} else {
			if err := templates.ExecuteTemplate(w, "events.html", data); err != nil {
				log.Printf("Error executing events template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
}

// AppLogHandler handles the application log query page using eventstore.Logs()
func (p *AdminProvider) AppLogHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Parse query parameters
		partitionID := r.URL.Query().Get("partition_id")
		eventType := r.URL.Query().Get("event_type")
		dateFrom := r.URL.Query().Get("date_from")
		dateTo := r.URL.Query().Get("date_to")
		fromID, _ := strconv.ParseUint(r.URL.Query().Get("from_id"), 10, 64)

		// Parse date filters
		var dateFromTime, dateToTime time.Time
		if dateFrom != "" {
			dateFromTime, _ = time.Parse("2006-01-02T15:04", dateFrom)
		}
		if dateTo != "" {
			dateToTime, _ = time.Parse("2006-01-02T15:04", dateTo)
		}

		// Use eventstore.Logs() with partition filter
		pageSize := 50
		logs, err := p.eventStore.Logs(fromID, uint32(pageSize+1), partitionID)
		if err != nil {
			log.Printf("Error querying app log: %v", err)
			http.Error(w, "Error querying app log", http.StatusInternalServerError)
			return
		}

		log.Printf("AppLog query returned %d entries (partition_id=%s, fromID=%d)", len(logs), partitionID, fromID)

		// Check if there are more results
		hasMore := len(logs) > pageSize
		if hasMore {
			logs = logs[:pageSize]
		}

		// Filter and convert to AppLogRows
		var logRows []AppLogRow
		for _, logEntry := range logs {
			event := logEntry.Event

			// Filter by event type
			if eventType != "" && !strings.Contains(event.EventType, eventType) {
				continue
			}

			// Filter by date
			if !dateFromTime.IsZero() && event.OccurredOn.Before(dateFromTime) {
				continue
			}
			if !dateToTime.IsZero() && event.OccurredOn.After(dateToTime) {
				continue
			}

			logRows = append(logRows, AppLogRow{
				ID:                logEntry.ID,
				PartitionID:       common.ExtractEntityType(event),
				OriginatorID:      event.Originator.ID,
				OriginatorVersion: event.Originator.Version,
				EventType:         event.EventType,
				Payload:           event.Payload,
				OccurredOn:        event.OccurredOn,
			})
		}

		nextID := fromID
		if len(logs) > 0 {
			nextID = logs[len(logs)-1].ID + 1
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
			if err := templates.ExecuteTemplate(w, "applog.html", data); err != nil {
				log.Printf("Error executing applog template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
}

// CrudHandler handles the CRUD entities page using GetPartitions() and List()
func (p *AdminProvider) CrudHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityType := r.URL.Query().Get("type")

		if entityType == "" {
			// Show entity types list using GetPartitions()
			partitions, err := p.eventStore.GetPartitions()
			if err != nil {
				log.Printf("Error getting partitions: %v", err)
				http.Error(w, "Error getting partitions", http.StatusInternalServerError)
				return
			}

			// Count entities for each partition using List()
			var entityTypes []EntityType
			for _, partition := range partitions {
				// Get a sample list to count entities
				originators, _, err := p.crudStore.List(partition, "0", 1000)
				if err != nil {
					log.Printf("Error listing entities for partition %s: %v", partition, err)
					continue
				}

				entityTypes = append(entityTypes, EntityType{
					Name:  partition,
					Count: len(originators),
				})
			}

			data := CrudPageData{
				Title:           "CRUD Entities",
				EntityTypes:     entityTypes,
				ShowingEntities: false,
			}

			if err := templates.ExecuteTemplate(w, "crud.html", data); err != nil {
				log.Printf("Error executing crud template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		} else {
			// Show entities of a specific type using List()
			originators, _, err := p.crudStore.List(entityType, "0", 100)
			if err != nil {
				log.Printf("Error listing entities: %v", err)
				http.Error(w, "Error listing entities", http.StatusInternalServerError)
				return
			}

			// Get full details for each entity using Get()
			var entities []CrudEntity
			for _, originator := range originators {
				payload, latestOriginator, err := p.crudStore.Get(originator, false)
				if err != nil {
					// Skip deleted or not found entities
					log.Printf("Skipping entity %s: %v", originator.ID, err)
					continue
				}

				// Parse the payload to get timestamp
				var payloadData map[string]interface{}
				updatedAt := time.Time{}
				if err := json.Unmarshal([]byte(payload), &payloadData); err == nil {
					if occurredOnStr, ok := payloadData["occurredOn"].(string); ok {
						updatedAt, _ = time.Parse(time.RFC3339, occurredOnStr)
					}
				}

				// Pretty print the JSON
				var prettyJSON []byte
				if err := json.Unmarshal([]byte(payload), &payloadData); err == nil {
					prettyJSON, _ = json.MarshalIndent(payloadData, "", "  ")
				} else {
					prettyJSON = []byte(payload)
				}

				entities = append(entities, CrudEntity{
					OriginatorID: originator.ID,
					Version:      latestOriginator.Version,
					Data:         string(prettyJSON),
					UpdatedAt:    updatedAt,
					IsDeleted:    false,
				})
			}

			data := CrudPageData{
				Title:           "CRUD Entities",
				SelectedType:    entityType,
				Entities:        entities,
				ShowingEntities: true,
			}

			if err := templates.ExecuteTemplate(w, "crud.html", data); err != nil {
				log.Printf("Error executing crud template: %v", err)
				http.Error(w, "Error rendering template", http.StatusInternalServerError)
			}
		}
	}
}

// CrudEntityDetailHandler handles the entity detail page using eventstore.Get()
func (p *AdminProvider) CrudEntityDetailHandler() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		entityType := r.URL.Query().Get("type")
		originatorID := r.URL.Query().Get("id")

		if entityType == "" || originatorID == "" {
			http.Error(w, "Missing type or id parameter", http.StatusBadRequest)
			return
		}

		// Get all events for this entity using eventstore.Get()
		originator := &types.Originator{ID: originatorID}
		events, err := p.eventStore.Get(originator, false)
		if err != nil {
			log.Printf("Error getting entity events: %v", err)
			http.Error(w, "Error getting entity events", http.StatusInternalServerError)
			return
		}

		if len(events) == 0 {
			http.Error(w, "Entity not found", http.StatusNotFound)
			return
		}

		// Get current state using crudstore.Get()
		payload, latestOriginator, err := p.crudStore.Get(originator, true) // true to include deleted
		isDeleted := false
		if err != nil {
			// Check if it's deleted
			if strings.Contains(err.Error(), "deleted") {
				isDeleted = true
				// Try to reconstruct state from events
				payload = "{}"
			} else {
				log.Printf("Error getting entity state: %v", err)
				http.Error(w, "Error getting entity state", http.StatusInternalServerError)
				return
			}
		}

		// Pretty print the current state
		var stateData map[string]interface{}
		var stateJSON []byte
		if err := json.Unmarshal([]byte(payload), &stateData); err == nil {
			stateJSON, _ = json.MarshalIndent(stateData, "", "  ")
		} else {
			stateJSON = []byte(payload)
		}

		// Get last updated time from events
		lastEvent := events[len(events)-1]
		updatedAt := lastEvent.OccurredOn

		// Convert events to template data
		eventHistory := make([]EntityEvent, len(events))
		for i, e := range events {
			eventHistory[i] = EntityEvent{
				Version:    e.Originator.Version,
				EventType:  e.EventType,
				Payload:    e.Payload,
				OccurredOn: e.OccurredOn,
			}
		}

		data := CrudEntityDetailData{
			Title:        "Entity Details",
			EntityType:   entityType,
			OriginatorID: originatorID,
			Entity: CrudEntity{
				OriginatorID: originatorID,
				Version:      latestOriginator.Version,
				Data:         string(stateJSON),
				UpdatedAt:    updatedAt,
				IsDeleted:    isDeleted,
			},
			Events: eventHistory,
		}

		if err := templates.ExecuteTemplate(w, "crud_entity.html", data); err != nil {
			log.Printf("Error executing crud_entity template: %v", err)
			http.Error(w, "Error rendering template", http.StatusInternalServerError)
		}
	}
}

// SetupRoutes configures all HTTP routes for the admin service
func (p *AdminProvider) SetupRoutes(mux *http.ServeMux) {
	// Static files
	fs := http.FileServer(http.Dir("./static"))
	mux.Handle("/static/", http.StripPrefix("/static/", fs))

	// Routes
	mux.HandleFunc("/", p.HomeHandler())
	mux.HandleFunc("/events", p.EventsHandler())
	mux.HandleFunc("/applog", p.AppLogHandler())
	mux.HandleFunc("/crud", p.CrudHandler())
	mux.HandleFunc("/crud/entity", p.CrudEntityDetailHandler())
}
