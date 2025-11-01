package crudstore

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/makkalot/eskit/lib/eventstore"
	"github.com/makkalot/eskit/lib/types"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
)

func TestNewCrudStoreProvider(t *testing.T) {
	estore := eventstore.NewInMemoryStore()
	ctx := context.Background()

	store, err := NewCrudStoreProvider(ctx, estore)
	assert.NoError(t, err)
	assert.NotNil(t, store)
}

func TestCrudStoreProvider_Create(t *testing.T) {
	testCases := []struct {
		name        string
		entityType  string
		originator  *types.Originator
		payload     string
		expectError bool
		errorMsg    string
	}{
		{
			name:       "successful create with version",
			entityType: "User",
			originator: &types.Originator{
				ID:      uuid.Must(uuid.NewV4()).String(),
				Version: "1",
			},
			payload:     `{"name":"test"}`,
			expectError: false,
		},
		{
			name:       "successful create without version",
			entityType: "User",
			originator: &types.Originator{
				ID: uuid.Must(uuid.NewV4()).String(),
			},
			payload:     `{"name":"test"}`,
			expectError: false,
		},
		{
			name:        "nil originator",
			entityType:  "User",
			originator:  nil,
			payload:     `{"name":"test"}`,
			expectError: true,
			errorMsg:    "empty originator",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estore := eventstore.NewInMemoryStore()
			store, err := NewCrudStoreProvider(context.Background(), estore)
			assert.NoError(t, err)

			err = store.Create(tc.entityType, tc.originator, tc.payload)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorMsg != "" {
					assert.Contains(t, err.Error(), tc.errorMsg)
				}
			} else {
				assert.NoError(t, err)

				events, err := estore.Get(tc.originator, false)
				assert.NoError(t, err)
				assert.Len(t, events, 1)
				assert.Equal(t, tc.entityType+".Created", events[0].EventType)
				assert.Equal(t, tc.payload, events[0].Payload)
			}
		})
	}
}

func TestCrudStoreProvider_Update(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(store CrudStore) (*types.Originator, string)
		updateData  func(originator *types.Originator) (string, string)
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "successful update",
			setup: func(store CrudStore) (*types.Originator, string) {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				payload := `{"name":"original","age":30}`
				err := store.Create("User", originator, payload)
				if err != nil {
					panic(err)
				}
				return originator, payload
			},
			updateData: func(originator *types.Originator) (string, string) {
				return "User", `{"name":"updated","age":31}`
			},
			expectError: false,
		},
		{
			name: "update with missing version",
			setup: func(store CrudStore) (*types.Originator, string) {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				payload := `{"name":"original"}`
				err := store.Create("User", originator, payload)
				if err != nil {
					panic(err)
				}
				return originator, payload
			},
			updateData: func(originator *types.Originator) (string, string) {
				originator.Version = ""
				return "User", `{"name":"updated"}`
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return err != nil && err.Error() == "misisng version"
			},
		},
		{
			name: "update non-existing entity",
			setup: func(store CrudStore) (*types.Originator, string) {
				return &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}, ""
			},
			updateData: func(originator *types.Originator) (string, string) {
				return "User", `{"name":"updated"}`
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsErrNotFound(err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estore := eventstore.NewInMemoryStore()
			store, err := NewCrudStoreProvider(context.Background(), estore)
			assert.NoError(t, err)

			originator, _ := tc.setup(store)
			entityType, payload := tc.updateData(originator)

			newOriginator, err := store.Update(entityType, originator, payload)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorCheck != nil {
					assert.True(t, tc.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newOriginator)
				assert.Equal(t, originator.ID, newOriginator.ID)
				assert.Equal(t, "2", newOriginator.Version)
			}
		})
	}
}

func TestCrudStoreProvider_Get(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(store CrudStore, estore eventstore.Store) *types.Originator
		getDeleted  bool
		expectError bool
		errorCheck  func(error) bool
		validate    func(*testing.T, string, *types.Originator)
	}{
		{
			name: "get existing entity",
			setup: func(store CrudStore, estore eventstore.Store) *types.Originator {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				payload := `{"name":"test","age":30}`
				err := store.Create("User", originator, payload)
				if err != nil {
					panic(err)
				}
				return originator
			},
			getDeleted:  false,
			expectError: false,
			validate: func(t *testing.T, payload string, orig *types.Originator) {
				var data map[string]interface{}
				err := json.Unmarshal([]byte(payload), &data)
				assert.NoError(t, err)
				assert.Equal(t, "test", data["name"])
				assert.Equal(t, float64(30), data["age"])
			},
		},
		{
			name: "get non-existing entity",
			setup: func(store CrudStore, estore eventstore.Store) *types.Originator {
				return &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
			},
			getDeleted:  false,
			expectError: true,
			errorCheck: func(err error) bool {
				return IsErrNotFound(err)
			},
		},
		{
			name: "get deleted entity without deleted flag",
			setup: func(store CrudStore, estore eventstore.Store) *types.Originator {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				err := store.Create("User", originator, `{"name":"test"}`)
				if err != nil {
					panic(err)
				}
				_, err = store.Delete("User", &types.Originator{ID: originator.ID})
				if err != nil {
					panic(err)
				}
				return &types.Originator{ID: originator.ID}
			},
			getDeleted:  false,
			expectError: true,
			errorCheck: func(err error) bool {
				return IsErrDeleted(err)
			},
		},
		{
			name: "get deleted entity with deleted flag",
			setup: func(store CrudStore, estore eventstore.Store) *types.Originator {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				err := store.Create("User", originator, `{"name":"test"}`)
				if err != nil {
					panic(err)
				}
				_, err = store.Delete("User", &types.Originator{ID: originator.ID})
				if err != nil {
					panic(err)
				}
				return &types.Originator{ID: originator.ID}
			},
			getDeleted:  true,
			expectError: false,
			validate: func(t *testing.T, payload string, orig *types.Originator) {
				var data map[string]interface{}
				err := json.Unmarshal([]byte(payload), &data)
				assert.NoError(t, err)
				assert.Equal(t, "test", data["name"])
			},
		},
		{
			name: "get with specific version",
			setup: func(store CrudStore, estore eventstore.Store) *types.Originator {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				err := store.Create("User", originator, `{"name":"v1","age":30}`)
				if err != nil {
					panic(err)
				}
				_, err = store.Update("User", originator, `{"name":"v2","age":31}`)
				if err != nil {
					panic(err)
				}
				originator.Version = "1"
				return originator
			},
			getDeleted:  false,
			expectError: false,
			validate: func(t *testing.T, payload string, orig *types.Originator) {
				var data map[string]interface{}
				err := json.Unmarshal([]byte(payload), &data)
				assert.NoError(t, err)
				assert.Equal(t, "v1", data["name"])
			},
		},
		{
			name: "get with version not yet created",
			setup: func(store CrudStore, estore eventstore.Store) *types.Originator {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				err := store.Create("User", originator, `{"name":"v1"}`)
				if err != nil {
					panic(err)
				}
				originator.Version = "5"
				return originator
			},
			getDeleted:  false,
			expectError: true,
			errorCheck: func(err error) bool {
				return IsErrNotFound(err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estore := eventstore.NewInMemoryStore()
			store, err := NewCrudStoreProvider(context.Background(), estore)
			assert.NoError(t, err)

			originator := tc.setup(store, estore)

			payload, resultOrig, err := store.Get(originator, tc.getDeleted)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorCheck != nil {
					assert.True(t, tc.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotEmpty(t, payload)
				assert.NotNil(t, resultOrig)
				if tc.validate != nil {
					tc.validate(t, payload, resultOrig)
				}
			}
		})
	}
}

func TestCrudStoreProvider_Delete(t *testing.T) {
	testCases := []struct {
		name        string
		setup       func(store CrudStore) *types.Originator
		expectError bool
		errorCheck  func(error) bool
	}{
		{
			name: "delete existing entity",
			setup: func(store CrudStore) *types.Originator {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				err := store.Create("User", originator, `{"name":"test"}`)
				if err != nil {
					panic(err)
				}
				return originator
			},
			expectError: false,
		},
		{
			name: "delete non-existing entity",
			setup: func(store CrudStore) *types.Originator {
				return &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsErrNotFound(err)
			},
		},
		{
			name: "delete already deleted entity",
			setup: func(store CrudStore) *types.Originator {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				err := store.Create("User", originator, `{"name":"test"}`)
				if err != nil {
					panic(err)
				}
				_, err = store.Delete("User", &types.Originator{ID: originator.ID})
				if err != nil {
					panic(err)
				}
				return &types.Originator{ID: originator.ID}
			},
			expectError: true,
			errorCheck: func(err error) bool {
				return IsErrDeleted(err)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estore := eventstore.NewInMemoryStore()
			store, err := NewCrudStoreProvider(context.Background(), estore)
			assert.NoError(t, err)

			originator := tc.setup(store)

			newOriginator, err := store.Delete("User", originator)

			if tc.expectError {
				assert.Error(t, err)
				if tc.errorCheck != nil {
					assert.True(t, tc.errorCheck(err))
				}
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, newOriginator)
				assert.Equal(t, originator.ID, newOriginator.ID)
			}
		})
	}
}

func TestCrudStoreProvider_List(t *testing.T) {
	testCases := []struct {
		name       string
		setup      func(store CrudStore) []string
		entityType string
		fromID     string
		size       int
		validate   func(*testing.T, []*types.Originator, string)
	}{
		{
			name: "list empty store",
			setup: func(store CrudStore) []string {
				return []string{}
			},
			entityType: "User",
			fromID:     "",
			size:       10,
			validate: func(t *testing.T, results []*types.Originator, lastID string) {
				assert.Empty(t, results)
			},
		},
		{
			name: "list with single item",
			setup: func(store CrudStore) []string {
				originator := &types.Originator{
					ID:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				}
				err := store.Create("User", originator, `{"name":"user1"}`)
				if err != nil {
					panic(err)
				}
				return []string{originator.ID}
			},
			entityType: "User",
			fromID:     "",
			size:       10,
			validate: func(t *testing.T, results []*types.Originator, lastID string) {
				assert.Len(t, results, 1)
			},
		},
		{
			name: "list with multiple items",
			setup: func(store CrudStore) []string {
				var ids []string
				for i := 0; i < 5; i++ {
					originator := &types.Originator{
						ID:      uuid.Must(uuid.NewV4()).String(),
						Version: "1",
					}
					err := store.Create("User", originator, `{"name":"user"}`)
					if err != nil {
						panic(err)
					}
					ids = append(ids, originator.ID)
				}
				return ids
			},
			entityType: "User",
			fromID:     "",
			size:       10,
			validate: func(t *testing.T, results []*types.Originator, lastID string) {
				assert.Len(t, results, 5)
			},
		},
		{
			name: "list with pagination",
			setup: func(store CrudStore) []string {
				var ids []string
				for i := 0; i < 5; i++ {
					originator := &types.Originator{
						ID:      uuid.Must(uuid.NewV4()).String(),
						Version: "1",
					}
					err := store.Create("User", originator, `{"name":"user"}`)
					if err != nil {
						panic(err)
					}
					ids = append(ids, originator.ID)
				}
				return ids
			},
			entityType: "User",
			fromID:     "",
			size:       2,
			validate: func(t *testing.T, results []*types.Originator, lastID string) {
				assert.Len(t, results, 2)
				assert.NotEmpty(t, lastID)
			},
		},
		{
			name: "list excludes deleted items",
			setup: func(store CrudStore) []string {
				var ids []string
				for i := 0; i < 3; i++ {
					originator := &types.Originator{
						ID:      uuid.Must(uuid.NewV4()).String(),
						Version: "1",
					}
					err := store.Create("User", originator, `{"name":"user"}`)
					if err != nil {
						panic(err)
					}
					if i == 1 {
						_, err = store.Delete("User", originator)
						if err != nil {
							panic(err)
						}
					} else {
						ids = append(ids, originator.ID)
					}
				}
				return ids
			},
			entityType: "User",
			fromID:     "",
			size:       10,
			validate: func(t *testing.T, results []*types.Originator, lastID string) {
				assert.Len(t, results, 2)
			},
		},
		{
			name: "list with default size",
			setup: func(store CrudStore) []string {
				var ids []string
				for i := 0; i < 15; i++ {
					originator := &types.Originator{
						ID:      uuid.Must(uuid.NewV4()).String(),
						Version: "1",
					}
					err := store.Create("User", originator, `{"name":"user"}`)
					if err != nil {
						panic(err)
					}
					ids = append(ids, originator.ID)
				}
				return ids
			},
			entityType: "User",
			fromID:     "",
			size:       0,
			validate: func(t *testing.T, results []*types.Originator, lastID string) {
				assert.Len(t, results, 10)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			estore := eventstore.NewInMemoryStore()
			store, err := NewCrudStoreProvider(context.Background(), estore)
			assert.NoError(t, err)

			tc.setup(store)

			results, lastID, err := store.List(tc.entityType, tc.fromID, tc.size)

			assert.NoError(t, err)
			if tc.validate != nil {
				tc.validate(t, results, lastID)
			}
		})
	}
}

func TestCrudStoreProvider_isEventDeleted(t *testing.T) {
	estore := eventstore.NewInMemoryStore()
	store, err := NewCrudStoreProvider(context.Background(), estore)
	assert.NoError(t, err)

	provider := store.(*CrudStoreProvider)

	testCases := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "deleted event",
			eventType: "User.Deleted",
			expected:  true,
		},
		{
			name:      "deleted event lowercase",
			eventType: "User.deleted",
			expected:  true,
		},
		{
			name:      "created event",
			eventType: "User.Created",
			expected:  false,
		},
		{
			name:      "updated event",
			eventType: "User.Updated",
			expected:  false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &types.Event{
				EventType: tc.eventType,
			}
			result := provider.isEventDeleted(event)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCrudStoreProvider_isCrudEvent(t *testing.T) {
	estore := eventstore.NewInMemoryStore()
	store, err := NewCrudStoreProvider(context.Background(), estore)
	assert.NoError(t, err)

	provider := store.(*CrudStoreProvider)

	testCases := []struct {
		name      string
		eventType string
		expected  bool
	}{
		{
			name:      "created event",
			eventType: "User.Created",
			expected:  true,
		},
		{
			name:      "updated event",
			eventType: "User.Updated",
			expected:  true,
		},
		{
			name:      "deleted event",
			eventType: "User.Deleted",
			expected:  true,
		},
		{
			name:      "custom event",
			eventType: "User.CustomEvent",
			expected:  false,
		},
		{
			name:      "lowercase created",
			eventType: "User.created",
			expected:  true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			event := &types.Event{
				EventType: tc.eventType,
			}
			result := provider.isCrudEvent(event)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsErrNotFound(t *testing.T) {
	assert.True(t, IsErrNotFound(RecordNotFound))
	assert.False(t, IsErrNotFound(RecordDeleted))
	assert.False(t, IsErrNotFound(nil))
}

func TestIsErrDeleted(t *testing.T) {
	assert.True(t, IsErrDeleted(RecordDeleted))
	assert.False(t, IsErrDeleted(RecordNotFound))
	assert.False(t, IsErrDeleted(nil))
}

func TestIsDuplicate(t *testing.T) {
	assert.True(t, IsDuplicate(eventstore.ErrDuplicate))
	assert.False(t, IsDuplicate(RecordNotFound))
	assert.False(t, IsDuplicate(nil))
}
