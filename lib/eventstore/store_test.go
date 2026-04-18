package eventstore

import (
	"os"
	"testing"
	"time"

	"github.com/makkalot/eskit/lib/types"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestSqlStore(tm *testing.T) {

	sqlStore, err := NewSqlStore("sqlite3", "estore.db")
	assert.NoError(tm, err)
	assert.NotNil(tm, sqlStore)

	memoryStore := NewInMemoryStore()

	tmpFile, err := os.CreateTemp("", "events-*.jsonpl")
	assert.NoError(tm, err)
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	fileStore, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(tm, err)
	assert.NotNil(tm, fileStore)

	testCases := []struct {
		name  string
		store Store
	}{
		{
			"sql store",
			sqlStore,
		},
		{
			"inmemory store",
			memoryStore,
		},
		{
			"file store",
			fileStore,
		},
	}

	tm.Cleanup(func() {
		if _, err := os.Stat("estore.db"); err == nil {
			assert.NoError(tm, os.Remove("estore.db"))
		}
		assert.NoError(tm, fileStore.Cleanup())
	})

	for _, tc := range testCases {
		currentStore := tc.store
		tm.Run(tc.name, func(t *testing.T) {
			originator := &types.Originator{
				ID: uuid.Must(uuid.NewV4()).String(),
			}

			events, err := currentStore.Get(&types.Originator{
				ID: originator.ID,
			}, false)
			assert.NoError(t, err)
			assert.Len(t, events, 0)

			// check the log
			logs, err := currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 0)

			e1 := &types.Event{
				Originator: &types.Originator{
					ID:      originator.ID,
					Version: 1,
				},
				EventType:  "Project.Created",
				Payload:    "{}",
				OccurredOn: time.Now().UTC(),
			}

			err = currentStore.Append(e1)
			assert.NoError(t, err)

			events, err = currentStore.Get(&types.Originator{
				ID: originator.ID,
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.Equal(t, e1.Originator.ID, events[0].Originator.ID)
			assert.Equal(t, e1.Originator.Version, events[0].Originator.Version)
			assert.Equal(t, e1.EventType, events[0].EventType)
			assert.Equal(t, e1.Payload, events[0].Payload)

			logs, err = currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)
			assert.Equal(t, logs[0].ID, uint64(1))
			assert.Equal(t, e1.Originator.ID, logs[0].Event.Originator.ID)

			logs, err = currentStore.Logs(1, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)
			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 0)

			e2 := &types.Event{
				Originator: &types.Originator{
					ID:      originator.ID,
					Version: 2,
				},
				EventType:  "Project.Updated",
				Payload:    "{}",
				OccurredOn: time.Now().UTC(),
			}

			err = currentStore.Append(e2)
			assert.NoError(t, err)

			events, err = currentStore.Get(&types.Originator{
				ID: originator.ID,
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 2)
			assert.Equal(t, e1.EventType, events[0].EventType)
			assert.Equal(t, e2.EventType, events[1].EventType)

			events, err = currentStore.Get(&types.Originator{
				ID:      originator.ID,
				Version: 1,
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.Equal(t, e1.EventType, events[0].EventType)

			events, err = currentStore.Get(&types.Originator{
				ID:      originator.ID,
				Version: 2,
			}, true)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.Equal(t, e2.EventType, events[0].EventType)

			logs, err = currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)
			assert.Equal(t, logs[0].ID, uint64(1))
			assert.Equal(t, e1.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, uint64(2))
			assert.Equal(t, e2.EventType, logs[1].Event.EventType)

			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)

			assert.Equal(t, logs[0].ID, uint64(2))
			assert.Equal(t, e2.EventType, logs[0].Event.EventType)

			e3 := &types.Event{
				Originator: &types.Originator{
					ID:      originator.ID,
					Version: 3,
				},
				EventType:  "Project.Deleted",
				Payload:    "",
				OccurredOn: time.Now().UTC(),
			}

			err = currentStore.Append(e3)
			assert.NoError(t, err)

			events, err = currentStore.Get(&types.Originator{
				ID: originator.ID,
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 3)
			assert.Equal(t, e1.EventType, events[0].EventType)
			assert.Equal(t, e2.EventType, events[1].EventType)
			assert.Equal(t, e3.EventType, events[2].EventType)

			logs, err = currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 3)
			assert.Equal(t, logs[0].ID, uint64(1))
			assert.Equal(t, e1.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, uint64(2))
			assert.Equal(t, e2.EventType, logs[1].Event.EventType)
			assert.Equal(t, logs[2].ID, uint64(3))
			assert.Equal(t, e3.EventType, logs[2].Event.EventType)

			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)

			assert.Equal(t, logs[0].ID, uint64(2))
			assert.Equal(t, e2.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, uint64(3))
			assert.Equal(t, e3.EventType, logs[1].Event.EventType)

			logs, err = currentStore.Logs(0, 2, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)
			assert.Equal(t, logs[0].ID, uint64(1))
			assert.Equal(t, e1.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, uint64(2))

			logs, err = currentStore.Logs(2, 1, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)
			assert.Equal(t, logs[0].ID, uint64(2))
			assert.Equal(t, e2.EventType, logs[0].Event.EventType)

			// try to insert the same version again
			err = currentStore.Append(e3)
			assert.Error(t, err)
		})
	}
}

func TestFileStoreReopen(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "events-reopen-*.jsonl")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	t.Cleanup(func() {
		os.Remove(tmpFilePath)
	})

	originator := &types.Originator{
		ID: uuid.Must(uuid.NewV4()).String(),
	}

	e1 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 1},
		EventType:  "Project.Created",
		Payload:    `{"name":"test"}`,
		OccurredOn: time.Now().UTC(),
	}
	e2 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 2},
		EventType:  "Project.Updated",
		Payload:    `{"name":"updated"}`,
		OccurredOn: time.Now().UTC(),
	}

	// Write two events with the first store instance.
	store1, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(t, err)
	assert.NoError(t, store1.Append(e1))
	assert.NoError(t, store1.Append(e2))
	store1.file.Close()
	store1.read_file.Close()

	// Reopen the same file and verify the state is restored.
	store2, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(t, err)
	defer func() {
		store2.file.Close()
		store2.read_file.Close()
	}()

	// All previously written events should be readable.
	events, err := store2.Get(&types.Originator{ID: originator.ID}, false)
	assert.NoError(t, err)
	assert.Len(t, events, 2)
	assert.Equal(t, e1.EventType, events[0].EventType)
	assert.Equal(t, e1.Originator.Version, events[0].Originator.Version)
	assert.Equal(t, e2.EventType, events[1].EventType)
	assert.Equal(t, e2.Originator.Version, events[1].Originator.Version)

	// Log entries should be restored too.
	logs, err := store2.Logs(0, 20, "")
	assert.NoError(t, err)
	assert.Len(t, logs, 2)
	assert.Equal(t, uint64(1), logs[0].ID)
	assert.Equal(t, e1.EventType, logs[0].Event.EventType)
	assert.Equal(t, uint64(2), logs[1].ID)
	assert.Equal(t, e2.EventType, logs[1].Event.EventType)

	// Appending a new event should continue from the correct version.
	e3 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 3},
		EventType:  "Project.Deleted",
		Payload:    "",
		OccurredOn: time.Now().UTC(),
	}
	assert.NoError(t, store2.Append(e3))

	events, err = store2.Get(&types.Originator{ID: originator.ID}, false)
	assert.NoError(t, err)
	assert.Len(t, events, 3)
	assert.Equal(t, e3.EventType, events[2].EventType)

	// Duplicate append should be rejected.
	assert.Error(t, store2.Append(e3))
}
