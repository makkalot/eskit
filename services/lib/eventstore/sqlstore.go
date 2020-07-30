package eventstore

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	eskitcommon "github.com/makkalot/eskit/services/lib/common"
	"github.com/prometheus/client_golang/prometheus"
	"strconv"
	"strings"
	"time"
)

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

	err := eskitcommon.RetryNormal(func() error {
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

	entityType := eskitcommon.ExtractEntityType(event)
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
			return fmt.Errorf("stored_event: %w", ErrDuplicate)
		}
		return fmt.Errorf("inserting stored event : %v", err)
	}

	result := tx.Create(storedLogEntry)
	if err := result.Error; err != nil {
		tx.Rollback()
		if strings.Contains(strings.ToLower(err.Error()), "unique") {
			return fmt.Errorf("stored_log_entry: %w", ErrDuplicate)
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
