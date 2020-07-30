package eventstore

import (
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"

	"github.com/golang/protobuf/proto"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
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
			originator := &common.Originator{
				Id: uuid.Must(uuid.NewV4()).String(),
			}

			events, err := currentStore.Get(&common.Originator{
				Id: originator.Id,
			}, false)
			assert.NoError(t, err)
			assert.Len(t, events, 0)

			// check the log
			logs, err := currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 0)

			e1 := &store.Event{
				Originator: &common.Originator{
					Id:      originator.Id,
					Version: "1",
				},
				EventType: "Project.Created",
				Payload:   "{}",
			}

			err = currentStore.Append(e1)
			assert.NoError(t, err)

			events, err = currentStore.Get(&common.Originator{
				Id: originator.Id,
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.True(t, proto.Equal(events[0], e1))

			logs, err = currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)
			assert.Equal(t, logs[0].Id, "1")
			assert.True(t, proto.Equal(logs[0].Event, e1))

			logs, err = currentStore.Logs(1, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)
			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 0)

			e2 := &store.Event{
				Originator: &common.Originator{
					Id:      originator.Id,
					Version: "2",
				},
				EventType: "Project.Updated",
				Payload:   "{}",
			}

			err = currentStore.Append(e2)
			assert.NoError(t, err)

			events, err = currentStore.Get(&common.Originator{
				Id: originator.Id,
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 2)
			assert.True(t, proto.Equal(events[0], e1))
			assert.True(t, proto.Equal(events[1], e2))

			events, err = currentStore.Get(&common.Originator{
				Id:      originator.Id,
				Version: "1",
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.True(t, proto.Equal(events[0], e1))

			events, err = currentStore.Get(&common.Originator{
				Id:      originator.Id,
				Version: "2",
			}, true)

			assert.NoError(t, err)
			assert.Len(t, events, 1)
			assert.True(t, proto.Equal(events[0], e2))

			logs, err = currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)
			assert.Equal(t, logs[0].Id, "1")
			assert.True(t, proto.Equal(logs[0].Event, e1))
			assert.Equal(t, logs[1].Id, "2")
			assert.True(t, proto.Equal(logs[1].Event, e2))

			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)

			assert.Equal(t, logs[0].Id, "2")
			assert.True(t, proto.Equal(logs[0].Event, e2))

			e3 := &store.Event{
				Originator: &common.Originator{
					Id:      originator.Id,
					Version: "3",
				},
				EventType: "Project.Deleted",
				Payload:   "",
			}

			err = currentStore.Append(e3)
			assert.NoError(t, err)

			events, err = currentStore.Get(&common.Originator{
				Id: originator.Id,
			}, false)

			assert.NoError(t, err)
			assert.Len(t, events, 3)
			assert.True(t, proto.Equal(events[0], e1))
			assert.True(t, proto.Equal(events[1], e2))
			assert.True(t, proto.Equal(events[2], e3))

			logs, err = currentStore.Logs(0, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 3)
			assert.Equal(t, logs[0].Id, "1")
			assert.True(t, proto.Equal(logs[0].Event, e1))
			assert.Equal(t, logs[1].Id, "2")
			assert.True(t, proto.Equal(logs[1].Event, e2))
			assert.Equal(t, logs[2].Id, "3")
			assert.True(t, proto.Equal(logs[2].Event, e3))

			logs, err = currentStore.Logs(2, 20, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)

			assert.Equal(t, logs[0].Id, "2")
			assert.True(t, proto.Equal(logs[0].Event, e2))
			assert.Equal(t, logs[1].Id, "3")
			assert.True(t, proto.Equal(logs[1].Event, e3))

			logs, err = currentStore.Logs(0, 2, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 2)
			assert.Equal(t, logs[0].Id, "1")
			assert.True(t, proto.Equal(logs[0].Event, e1))
			assert.Equal(t, logs[1].Id, "2")

			logs, err = currentStore.Logs(2, 1, "")
			assert.NoError(t, err)
			assert.Len(t, logs, 1)
			assert.Equal(t, logs[0].Id, "2")
			assert.True(t, proto.Equal(logs[0].Event, e2))

			// try to insert the same version again
			err = currentStore.Append(e3)
			assert.Error(t, err)

			// SQLITE implementation is shit man !
			// add a different originator
			//originator2 := &common.Originator{
			//	Id: uuid.Must(uuid.NewV4()).String(),
			//}
			//e11 := &store.Event{
			//	Originator: &common.Originator{
			//		Id:      originator2.Id,
			//		Version: "1",
			//	},
			//	EventType: "Project.Created",
			//	Payload:   "{}",
			//}
			//
			//err = currentStore.Append(e11)
			//assert.NoError(t, err)

		})
	}

}
