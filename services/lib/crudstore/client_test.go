package crudstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/services/lib/eventstore"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"log"
	"os"
	"testing"
)

var postgresURI string

func TestMain(m *testing.M) {
	dbURI := os.Getenv("DB_URI")
	if dbURI == "" {
		log.Fatalf("DB_URI is required parameter")
	}

	postgresURI = dbURI
	code := m.Run()
	os.Exit(code)
}

type User struct {
	Originator *common.Originator
	Email      string
	FirstName  string
	LastName   string
	Active     bool
}

func TestCrudAdd(t *testing.T) {
	sqlStore, err := eventstore.NewSqlStore("postgres", postgresURI)
	assert.NoError(t, err)
	assert.NotNil(t, sqlStore)

	// cleaning up all of the data we added during those tests
	t.Cleanup(func() {
		assert.NoErrorf(t, sqlStore.Cleanup(), "cleanup failed")
	})

	crudStore, err := NewCrudStoreProvider(context.Background(), sqlStore)
	assert.NoError(t, err)
	assert.NotNil(t, crudStore)

	client := NewStructCrudStoreClient(crudStore)

	testCases := []struct {
		Name      string
		inputUser interface{}
		// if not supplied will be checked against the one in the inputUser
		expectedUser *User
		equalCB      func(expectedUser *User, actualUser *User) error
		err          error
	}{
		{
			Name: "non pointer",
			inputUser: User{
				Originator: &common.Originator{
					Id:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				},
				Email: "makkalotsomething@gmail.com",
			},
			err: InvalidArgumentError,
		},
		{
			Name: "success",
			inputUser: &User{
				Originator: &common.Originator{
					Id:      uuid.Must(uuid.NewV4()).String(),
					Version: "1",
				},
				Email:  "makkalotwork@gmail.com",
				Active: true,
			},
		},
		{
			Name: "success empty originator",
			inputUser: &User{
				Email:  "makkalotoriginator@gmail.com",
				Active: true,
			},
			equalCB: func(expectedUser *User, actualUser *User) error {
				if actualUser.Originator == nil {
					return fmt.Errorf("equalcb : originator is empty")
				}

				if actualUser.Originator.Id == "" || actualUser.Originator.Version == "" {
					return fmt.Errorf("equalCB: empty fields in originator : %+v", actualUser.Originator)
				}

				if actualUser.Active != expectedUser.Active || actualUser.Email != expectedUser.Email {
					return fmt.Errorf("equalCB failed : actual : %+v", actualUser)
				}

				return nil
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.Name, func(tt *testing.T) {
			originator, err := client.Create(tc.inputUser)

			if tc.err != nil {
				if !errors.Is(err, tc.err) {
					tt.Errorf("expected error : %v, received : %v", tc.err, err)
				}
			} else {
				assert.NoError(tt, err)
				assert.NotNil(tt, originator)

				var fetchedUser User
				err = client.Get(originator, &fetchedUser, false)
				assert.NoError(tt, err)
				//check if all the fields are in place

				if tc.expectedUser != nil && tc.equalCB == nil  {
					assert.Equal(tt, tc.expectedUser, fetchedUser)
				} else if tc.equalCB != nil {
					var err error
					if tc.expectedUser != nil {
						err = tc.equalCB(tc.expectedUser, &fetchedUser)
					} else {
						err = tc.equalCB(tc.inputUser.(*User), &fetchedUser)
					}
					assert.NoError(tt, err)
				} else {
					assert.Equal(tt, tc.inputUser, &fetchedUser)
				}
			}
		})
	}
}
