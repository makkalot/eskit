package eventstore

import (
	"encoding/json"
	"os"
	"strings"
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

func TestFileStoreMultipleOriginators(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "events-multi-*.jsonl")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	t.Cleanup(func() {
		os.Remove(tmpFilePath)
	})

	store, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(t, err)
	t.Cleanup(func() {
		store.Cleanup()
	})

	originatorA := &types.Originator{ID: uuid.Must(uuid.NewV4()).String()}
	originatorB := &types.Originator{ID: uuid.Must(uuid.NewV4()).String()}

	// Create entity A with 2 events
	eA1 := &types.Event{
		Originator: &types.Originator{ID: originatorA.ID, Version: 1},
		EventType:  "CamConfig.Created",
		Payload:    `{"name":"camera-A"}`,
		OccurredOn: time.Now().UTC(),
	}
	eA2 := &types.Event{
		Originator: &types.Originator{ID: originatorA.ID, Version: 2},
		EventType:  "CamConfig.Updated",
		Payload:    `{"name":"camera-A-v2"}`,
		OccurredOn: time.Now().UTC(),
	}
	assert.NoError(t, store.Append(eA1))
	assert.NoError(t, store.Append(eA2))

	// Create entity B — version 1 should work even though A is at version 2
	eB1 := &types.Event{
		Originator: &types.Originator{ID: originatorB.ID, Version: 1},
		EventType:  "CamConfig.Created",
		Payload:    `{"name":"camera-B"}`,
		OccurredOn: time.Now().UTC(),
	}
	assert.NoError(t, store.Append(eB1))

	// Update entity B to version 2
	eB2 := &types.Event{
		Originator: &types.Originator{ID: originatorB.ID, Version: 2},
		EventType:  "CamConfig.Updated",
		Payload:    `{"name":"camera-B-v2"}`,
		OccurredOn: time.Now().UTC(),
	}
	assert.NoError(t, store.Append(eB2))

	// Verify A's events are intact
	eventsA, err := store.Get(&types.Originator{ID: originatorA.ID}, false)
	assert.NoError(t, err)
	assert.Len(t, eventsA, 2)
	assert.Equal(t, eA1.EventType, eventsA[0].EventType)
	assert.Equal(t, eA2.EventType, eventsA[1].EventType)

	// Verify B's events are intact
	eventsB, err := store.Get(&types.Originator{ID: originatorB.ID}, false)
	assert.NoError(t, err)
	assert.Len(t, eventsB, 2)
	assert.Equal(t, eB1.EventType, eventsB[0].EventType)
	assert.Equal(t, eB2.EventType, eventsB[1].EventType)

	// Trying to append version 1 to B again should fail
	err = store.Append(eB1)
	assert.Error(t, err)

	// Trying to append version 2 to A again should fail
	err = store.Append(eA2)
	assert.Error(t, err)

	// Total logs should be 4
	logs, err := store.Logs(0, 20, "")
	assert.NoError(t, err)
	assert.Len(t, logs, 4)
}

func TestFileStoreGetOne(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "events-getone-*.jsonl")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	t.Cleanup(func() {
		os.Remove(tmpFilePath)
	})

	store, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(t, err)
	t.Cleanup(func() {
		store.Cleanup()
	})

	originator := &types.Originator{ID: uuid.Must(uuid.NewV4()).String()}

	e1 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 1},
		EventType:  "Project.Created",
		Payload:    `{"name":"v1"}`,
		OccurredOn: time.Now().UTC(),
	}
	e2 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 2},
		EventType:  "Project.Updated",
		Payload:    `{"name":"v2"}`,
		OccurredOn: time.Now().UTC(),
	}
	e3 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 3},
		EventType:  "Project.Deleted",
		Payload:    ``,
		OccurredOn: time.Now().UTC(),
	}

	assert.NoError(t, store.Append(e1))
	assert.NoError(t, store.Append(e2))
	assert.NoError(t, store.Append(e3))

	// Version 0 should return the latest event (e3).
	latest, err := store.GetOne(&types.Originator{ID: originator.ID})
	assert.NoError(t, err)
	assert.NotNil(t, latest)
	assert.Equal(t, e3.EventType, latest.EventType)
	assert.Equal(t, uint64(3), latest.Originator.Version)

	// Specific version should return the exact event.
	specific, err := store.GetOne(&types.Originator{ID: originator.ID, Version: 2})
	assert.NoError(t, err)
	assert.NotNil(t, specific)
	assert.Equal(t, e2.EventType, specific.EventType)
	assert.Equal(t, uint64(2), specific.Originator.Version)

	// Non-existent version should return nil.
	missing, err := store.GetOne(&types.Originator{ID: originator.ID, Version: 5})
	assert.NoError(t, err)
	assert.Nil(t, missing)

	// Nil originator should return error.
	_, err = store.GetOne(nil)
	assert.Error(t, err)
}

func TestFileStorePayloadAsJSON(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "events-payload-*.jsonl")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	t.Cleanup(func() {
		os.Remove(tmpFilePath)
	})

	store, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(t, err)
	t.Cleanup(func() {
		store.Cleanup()
	})

	originator := &types.Originator{ID: uuid.Must(uuid.NewV4()).String()}
	e1 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 1},
		EventType:  "Project.Created",
		Payload:    `{"name":"test","count":42}`,
		OccurredOn: time.Now().UTC(),
	}
	e2 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 2},
		EventType:  "Project.Updated",
		Payload:    "plain string",
		OccurredOn: time.Now().UTC(),
	}
	e3 := &types.Event{
		Originator: &types.Originator{ID: originator.ID, Version: 3},
		EventType:  "Project.Deleted",
		Payload:    "",
		OccurredOn: time.Now().UTC(),
	}

	assert.NoError(t, store.Append(e1))
	assert.NoError(t, store.Append(e2))
	assert.NoError(t, store.Append(e3))

	// Verify the raw file contains readable JSON objects, not escaped strings.
	raw, err := os.ReadFile(tmpFilePath)
	assert.NoError(t, err)
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	assert.Len(t, lines, 3)

	var payloadLine1 map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(lines[0]), &payloadLine1))
	payload1, ok := payloadLine1["payload"].(map[string]interface{})
	assert.True(t, ok, "payload should be stored as JSON object")
	assert.Equal(t, "test", payload1["name"])
	assert.Equal(t, float64(42), payload1["count"])

	var payloadLine2 map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(lines[1]), &payloadLine2))
	// Plain string can't be parsed as JSON, so it stays a string.
	assert.Equal(t, "plain string", payloadLine2["payload"])

	var payloadLine3 map[string]interface{}
	assert.NoError(t, json.Unmarshal([]byte(lines[2]), &payloadLine3))
	// Empty payload stored as null.
	assert.Nil(t, payloadLine3["payload"])

	// Verify the API still returns string payloads after reopening.
	store2, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(t, err)
	defer func() {
		store2.file.Close()
		store2.read_file.Close()
	}()

	events, err := store2.Get(&types.Originator{ID: originator.ID}, false)
	assert.NoError(t, err)
	assert.Len(t, events, 3)
	assert.JSONEq(t, `{"name":"test","count":42}`, events[0].Payload)
	assert.Equal(t, "plain string", events[1].Payload)
	assert.Equal(t, "", events[2].Payload)
}

func TestStoreDuplicateVersionZero(t *testing.T) {
	sqlStore, err := NewSqlStore("sqlite3", "estore-dup.db")
	assert.NoError(t, err)
	assert.NotNil(t, sqlStore)

	memoryStore := NewInMemoryStore()

	tmpFile, err := os.CreateTemp("", "events-dup-*.jsonl")
	assert.NoError(t, err)
	tmpFilePath := tmpFile.Name()
	tmpFile.Close()

	fileStore, err := NewFileMemoryStore(tmpFilePath)
	assert.NoError(t, err)
	assert.NotNil(t, fileStore)

	testCases := []struct {
		name  string
		store Store
	}{
		{"sql store", sqlStore},
		{"inmemory store", memoryStore},
		{"file store", fileStore},
	}

	t.Cleanup(func() {
		if _, err := os.Stat("estore-dup.db"); err == nil {
			assert.NoError(t, os.Remove("estore-dup.db"))
		}
		assert.NoError(t, fileStore.Cleanup())
	})

	for _, tc := range testCases {
		currentStore := tc.store
		t.Run(tc.name, func(t *testing.T) {
			originator := &types.Originator{
				ID: uuid.Must(uuid.NewV4()).String(),
			}

			e0 := &types.Event{
				Originator: &types.Originator{
					ID:      originator.ID,
					Version: 0,
				},
				EventType:  "Project.Created",
				Payload:    "{}",
				OccurredOn: time.Now().UTC(),
			}

			// First append with Version 0 should succeed.
			err := currentStore.Append(e0)
			assert.NoError(t, err)

			// Second append with same Version 0 should fail as duplicate.
			err = currentStore.Append(e0)
			assert.Error(t, err)
		})
	}
}
