package crudstore

import (
	"context"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/services/lib/eventstore"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

type User struct{
	Originator           *common.Originator
	Email                string
	FirstName            string
	LastName             string
	Active               bool
}

func TestCrudAdd(t *testing.T){
	sqlStore, err := eventstore.NewSqlStore("sqlite3", "estore.db")
	assert.NoError(t, err)
	assert.NotNil(t, sqlStore)

	// cleaning up all of the data we added during those tests
	t.Cleanup(func() {
		if _, err := os.Stat("estore.db"); err == nil {
			assert.NoError(t, os.Remove("estore.db"))
		}
	})

	crudStore, err := NewCrudStoreProvider(context.Background(), sqlStore)
	assert.NoError(t, err)
	assert.NotNil(t, crudStore)

	client := NewStructCrudStoreClient(crudStore)

	user := &User{
		Originator: &common.Originator{
			Id:                   uuid.Must(uuid.NewV4()).String(),
			Version:              "1",
		},
		Email:      "makkalotwork@gmail.com",
		FirstName:  "",
		LastName:   "",
		Active:     false,
	}

	originator, err := client.Create(user)
	assert.NoError(t, err)
	assert.NotNil(t, originator)

	// Todo: then try to fetch it here we miss the Get in the client.

}