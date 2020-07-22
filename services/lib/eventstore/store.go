package provider

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	common2 "github.com/makkalot/eskit/services/lib/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
	"strings"
	"time"
)

var (
	lastStreamID = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "eskit_events_stream_last_id",
			Help: "LastID in the stream",
		}, []string{
			"application_id", "partition_id",
		})

	streamCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eskit_events_stream_count",
			Help: "Stream Count",
		}, []string{
			"application_id", "partition_id",
		})
)

type ErrDuplicate struct {
	msg string
}

func (e *ErrDuplicate) Error() string {
	return fmt.Sprintf("duplicate error : %s", e.msg)
}

type Store interface {
	Append(event *store.Event) error
	Get(originator *common.Originator, fromVersion bool) ([]*store.Event, error)
	Logs(fromID uint64, size uint32, pipelineID string) ([]*store.AppLogEntry, error)
}

type StoredEvent struct {
	OriginatorID      string `gorm:"primary_key; not null"`
	OriginatorVersion uint   `gorm:"primary_key; not null"`
	EventType         string `gorm:"type:varchar(255); not null; index"`
	Payload           string `gorm:"type:text"`
	CreatedAt         time.Timer
}

type StoredLogEntry struct {
	ID            uint64 `gorm:"primary_key; AUTO_INCREMENT; not null"`
	ApplicationID string `gorm:"type:varchar(255); not null; index:index_app_partition; default:'consumer'"`
	PartitionID   string `gorm:"type:varchar(255); not null; index:index_app_partition"`
	EventPayload  string `gorm:"type:text"`
	CreatedAt     time.Timer
}

type SqlStore struct {
	db    *gorm.DB
	dbURI string
}

func NewSqlStore(dialect string, dbURI string) (*SqlStore, error) {
	var db *gorm.DB

	err := common2.RetryNormal(func() error {
		var err error
		db, err = gorm.Open(dialect, dbURI)
		if err != nil {
			return fmt.Errorf("connecting to db : %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if result := db.AutoMigrate(&StoredEvent{}); result.Error != nil {
		return nil, fmt.Errorf("migrate stored_events : %v", result.Error)
	}

	if result := db.AutoMigrate(&StoredLogEntry{}); result.Error != nil {
		return nil, fmt.Errorf("migrate log_entries : %v", result.Error)
	}

	return &SqlStore{
		db:    db,
		dbURI: dbURI,
	}, nil
}

func (estore *SqlStore) Cleanup() error {
	if result := estore.db.DropTableIfExists(&StoredEvent{}); result.Error != nil {
		return fmt.Errorf("drop table failed : %v", result.Error)
	}

	if result := estore.db.DropTableIfExists(&StoredLogEntry{}); result.Error != nil {
		return fmt.Errorf("drop table failed : %v", result.Error)
	}

	if result := estore.db.AutoMigrate(&StoredEvent{}); result.Error != nil {
		return fmt.Errorf("migrate stored_events : %v", result.Error)
	}

	if result := estore.db.AutoMigrate(&StoredLogEntry{}); result.Error != nil {
		return fmt.Errorf("migrate log_entries : %v", result.Error)
	}

	return nil
}

func (estore *SqlStore) Append(event *store.Event) error {

	intVersion, err := strconv.ParseUint(event.Originator.Version, 10, 64)
	if err != nil {
		return err
	}
	storedEvent := &StoredEvent{
		OriginatorID:      event.Originator.Id,
		OriginatorVersion: uint(intVersion),
		EventType:         event.EventType,
		Payload:           event.Payload,
	}

	//log.Println("stored event : ", spew.Sdump(storedEvent))

	entityType := common2.ExtractEntityType(event)
	jsonEvent, err := json.Marshal(event)
	if err != nil {
		return err
	}

	storedLogEntry := &StoredLogEntry{
		PartitionID:  entityType,
		EventPayload: string(jsonEvent),
	}

	tx := estore.db.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	if tx.Error != nil {
		return err
	}

	if err := tx.Create(storedEvent).Error; err != nil {
		tx.Rollback()
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return &ErrDuplicate{msg: "stored_event"}
		}
		return fmt.Errorf("inserting stored event : %v", err)
	}

	result := tx.Create(storedLogEntry)
	if err := result.Error; err != nil {
		tx.Rollback()
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return &ErrDuplicate{msg: "stored_log_entry"}
		}
		return fmt.Errorf("inserting stored log entry: %v", err)
	} else {
		g := lastStreamID.With(prometheus.Labels{"application_id": storedLogEntry.ApplicationID, "partition_id": storedLogEntry.PartitionID})
		g.Set(float64(storedLogEntry.ID))

		c := streamCounter.With(prometheus.Labels{"application_id": storedLogEntry.ApplicationID, "partition_id": storedLogEntry.PartitionID})
		c.Inc()
	}

	return tx.Commit().Error
}

func (estore *SqlStore) Get(originator *common.Originator, fromVersion bool) ([]*store.Event, error) {
	storedEvents := []*StoredEvent{}
	q := estore.db.Where("originator_id = ?", originator.Id)
	if originator.Version != "" {
		if !fromVersion {
			q = q.Where("originator_version <= ?", originator.Version)
		} else {
			q = q.Where("originator_version >= ?", originator.Version)
		}
	}

	results := q.Order("originator_version").Find(&storedEvents)
	if err := results.Error; err != nil {
		return nil, fmt.Errorf("fetch : %v", err)
	}

	var events []*store.Event
	for _, es := range storedEvents {
		events = append(events, &store.Event{
			Originator: &common.Originator{
				Id:      es.OriginatorID,
				Version: strconv.Itoa(int(es.OriginatorVersion)),
			},
			EventType: es.EventType,
			Payload:   es.Payload,
		})
	}

	return events, nil
}

func (estore *SqlStore) Logs(fromID uint64, size uint32, pipelineID string) ([]*store.AppLogEntry, error) {
	storedLogs := []*StoredLogEntry{}
	q := estore.db.Where("id >= ?", uint64(fromID))
	if size == 0 {
		size = 20
	}

	if pipelineID != "" {
		q = estore.db.Where("partition_id = ?", pipelineID)
	}

	results := q.Order("id").Limit(size).Find(&storedLogs)
	if err := results.Error; err != nil {
		return nil, fmt.Errorf("fetch : %v", err)
	}

	var logs []*store.AppLogEntry
	for _, sl := range storedLogs {
		event := &store.Event{}
		if err := json.Unmarshal([]byte(sl.EventPayload), event); err != nil {
			return nil, fmt.Errorf("unmarshall : %v", err)
		}
		logs = append(logs, &store.AppLogEntry{
			Id:    strconv.FormatUint(sl.ID, 10),
			Event: event,
		})
	}

	return logs, nil
}

type InMemoryStore struct {
	eventStore map[string][]*store.Event
	logs       []*store.AppLogEntry
}

func NewInMemoryStore() Store {
	return &InMemoryStore{
		eventStore: map[string][]*store.Event{},
		logs:       []*store.AppLogEntry{},
	}
}

func (s *InMemoryStore) Append(event *store.Event) error {
	if s.eventStore[event.Originator.Id] == nil {
		s.eventStore[event.Originator.Id] = []*store.Event{event}
		return s.appendLog(event)
	}

	events := s.eventStore[event.Originator.Id]
	latestEvent := events[len(events)-1]
	latestVersion := latestEvent.Originator.Version
	latestVersionInt, err := strconv.ParseInt(latestVersion, 10, 64)
	if err != nil {
		return err
	}

	newVersion := event.Originator.Version
	newVersionInt, err := strconv.ParseInt(newVersion, 10, 64)
	if err != nil {
		return err
	}

	if newVersionInt <= latestVersionInt {
		//log.Println("current store is like : ", spew.Sdump(s.eventStore))
		return &ErrDuplicate{msg: fmt.Sprintf("you apply version : %d but there's a newer version : %d for %s", newVersionInt, latestVersionInt, event.Originator.Id)}
	}

	s.eventStore[event.Originator.Id] = append(s.eventStore[event.Originator.Id], event)

	return s.appendLog(event)
}

func (s *InMemoryStore) appendLog(event *store.Event) error {
	var latestID string
	if len(s.logs) == 0 {
		latestID = "1"
	} else {
		latestLog := s.logs[len(s.logs)-1]
		latestIDInt, err := strconv.ParseInt(latestLog.Id, 10, 64)
		if err != nil {
			return err
		}
		latestIDInt++
		latestID = strconv.Itoa(int(latestIDInt))
	}

	s.logs = append(s.logs, &store.AppLogEntry{
		Id:    latestID,
		Event: event,
	})

	return nil
}

func (s *InMemoryStore) Get(originator *common.Originator, fromVersion bool) ([]*store.Event, error) {
	//log.Println("event store : ", spew.Sdump(s.eventStore))

	events := s.eventStore[originator.Id]
	if events == nil || len(events) == 0 {
		return []*store.Event{}, nil
	}

	eventVersion := originator.Version
	if eventVersion == "" {
		return events, nil
	}

	eventVersionInt, err := strconv.ParseInt(eventVersion, 10, 64)
	if err != nil {
		return nil, err
	}

	var results []*store.Event
	for _, e := range events {
		currentOriginator := e.Originator
		currentVersion := currentOriginator.Version
		currentVersionInt, err := strconv.ParseInt(currentVersion, 10, 64)
		if err != nil {
			return nil, err
		}

		if !fromVersion {
			if currentVersionInt <= eventVersionInt {
				results = append(results, e)
			}
		} else {
			if currentVersionInt >= eventVersionInt {
				results = append(results, e)
			}
		}
	}

	return results, nil
}

func (s *InMemoryStore) Logs(fromID uint64, size uint32, pipelineID string) ([]*store.AppLogEntry, error) {
	if s.logs == nil || len(s.logs) == 0 {
		return []*store.AppLogEntry{}, nil
	}

	if len(s.logs)-int(fromID) < 0 {
		return []*store.AppLogEntry{}, nil
	}

	//log.Println("fetching : ")
	if fromID > 0 {
		fromID--
	}

	var results []*store.AppLogEntry
	//log.Println("logs : ", spew.Sdump(s.logs))
	if len(s.logs)-int(fromID) > int(size) {
		//log.Printf("fetching : %d:%d\n", int(fromID), int(fromID)+int(size))
		results = s.logs[fromID : int(fromID)+int(size)]
	} else {
		results = s.logs[fromID:]
	}

	if pipelineID == "" {
		return results, nil
	}

	var finalResults []*store.AppLogEntry
	for _, r := range results {
		if common2.ExtractEntityType(r.Event) == pipelineID {
			finalResults = append(finalResults, r)
		}
	}

	return finalResults, nil
}
