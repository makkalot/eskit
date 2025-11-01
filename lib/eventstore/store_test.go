package eventstore

import (
	"github.com/makkalot/eskit/lib/types"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
	"time"
)

func TestSqlStore(tm *testing.T) {

	sqlStore, err := NewSqlStore("sqlite3", "estore.db")
	assert.NoError(tm, err)
	assert.NotNil(tm, sqlStore)

	memoryStore := NewInMemoryStore()

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
	}

	tm.Cleanup(func() {
		if _, err := os.Stat("estore.db"); err == nil {
			assert.NoError(tm, os.Remove("estore.db"))
		}
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
					Version: "1",
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
			assert.Equal(t, logs[0].ID, "1")
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
					Version: "2",
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
				Version: "1",
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.Equal(t, e1.EventType, events[0].EventType)

			events, err = currentStore.Get(&types.Originator{
				ID:      originator.ID,
				Version: "2",
			}, true)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.Equal(t, e2.EventType, events[0].EventType)

			logs, err = currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)
			assert.Equal(t, logs[0].ID, "1")
			assert.Equal(t, e1.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, "2")
			assert.Equal(t, e2.EventType, logs[1].Event.EventType)

			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)

			assert.Equal(t, logs[0].ID, "2")
			assert.Equal(t, e2.EventType, logs[0].Event.EventType)

			e3 := &types.Event{
				Originator: &types.Originator{
					ID:      originator.ID,
					Version: "3",
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
			assert.Equal(t, logs[0].ID, "1")
			assert.Equal(t, e1.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, "2")
			assert.Equal(t, e2.EventType, logs[1].Event.EventType)
			assert.Equal(t, logs[2].ID, "3")
			assert.Equal(t, e3.EventType, logs[2].Event.EventType)

			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)

			assert.Equal(t, logs[0].ID, "2")
			assert.Equal(t, e2.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, "3")
			assert.Equal(t, e3.EventType, logs[1].Event.EventType)

			logs, err = currentStore.Logs(0, 2, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)
			assert.Equal(t, logs[0].ID, "1")
			assert.Equal(t, e1.EventType, logs[0].Event.EventType)
			assert.Equal(t, logs[1].ID, "2")

			logs, err = currentStore.Logs(2, 1, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)
			assert.Equal(t, logs[0].ID, "2")
			assert.Equal(t, e2.EventType, logs[0].Event.EventType)

			// try to insert the same version again
			err = currentStore.Append(e3)
			assert.Error(t, err)
		})
	}
}
