package eventstore

import (
	"encoding/json"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/makkalot/eskit/lib/common"
	"github.com/makkalot/eskit/lib/types"
	"github.com/prometheus/client_golang/prometheus"
	"strings"
	"time"
)

type StoredEvent struct {
	OriginatorID      string `gorm:"not null; unique_index:idx_originator_composite"`
	OriginatorVersion uint   `gorm:"not null; unique_index:idx_originator_composite"`
	EventType         string `gorm:"type:varchar(255); not null; index"`
	Payload           string `gorm:"type:text"`
	CreatedAt         time.Time
}

type StoredLogEntry struct {
	ID            uint64 `gorm:"primary_key; AUTO_INCREMENT; not null"`
	ApplicationID string `gorm:"type:varchar(255); not null; index:index_app_partition; default:'consumer'"`
	PartitionID   string `gorm:"type:varchar(255); not null; index:index_app_partition"`
	EventPayload  string `gorm:"type:text"`
	CreatedAt     time.Time
}

type SqlStore struct {
	db    *gorm.DB
	dbURI string
}

func NewSqlStore(dialect string, dbURI string) (*SqlStore, error) {
	var db *gorm.DB

	err := common.RetryNormal(func() error {
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

func (estore *SqlStore) Append(event *types.Event) error {
	storedEvent := &StoredEvent{
		OriginatorID:      event.Originator.ID,
		OriginatorVersion: uint(event.Originator.Version),
		EventType:         event.EventType,
		Payload:           event.Payload,
	}

	//log.Println("stored event : ", spew.Sdump(storedEvent))

	entityType := common.ExtractEntityType(event)
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

func (estore *SqlStore) Get(originator *types.Originator, fromVersion bool) ([]*types.Event, error) {
	storedEvents := []*StoredEvent{}
	q := estore.db.Where("originator_id = ?", originator.ID)
	if originator.Version != 0 {
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

	var events []*types.Event
	for _, es := range storedEvents {
		events = append(events, &types.Event{
			Originator: &types.Originator{
				ID:      es.OriginatorID,
				Version: uint64(es.OriginatorVersion),
			},
			EventType: es.EventType,
			Payload:   es.Payload,
		})
	}

	return events, nil
}

func (estore *SqlStore) Logs(fromID uint64, size uint32, pipelineID string) ([]*types.AppLogEntry, error) {
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

	var logs []*types.AppLogEntry
	for _, sl := range storedLogs {
		event := &types.Event{}
		if err := json.Unmarshal([]byte(sl.EventPayload), event); err != nil {
			return nil, fmt.Errorf("unmarshall : %v", err)
		}
		logs = append(logs, &types.AppLogEntry{
			ID:    sl.ID,
			Event: event,
		})
	}

	return logs, nil
}

func (estore *SqlStore) GetPartitions() ([]string, error) {
	var results []struct {
		PartitionID string
	}

	err := estore.db.Table("stored_log_entries").
		Select("DISTINCT partition_id").
		Where("partition_id != ?", "").
		Find(&results).Error

	if err != nil {
		return nil, fmt.Errorf("fetching partitions: %v", err)
	}

	var partitions []string
	for _, r := range results {
		partitions = append(partitions, r.PartitionID)
	}

	return partitions, nil
}
