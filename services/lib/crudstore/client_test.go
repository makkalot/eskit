package crudstore

import (
	"context"
	"errors"
	"fmt"
	"github.com/jinzhu/copier"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	eskitcommon "github.com/makkalot/eskit/services/lib/common"
	"github.com/makkalot/eskit/services/lib/eventstore"
	uuid "github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

var postgresURI string

func TestMain(m *testing.M) {
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
	sqlStore := eventstore.NewInMemoryStore()
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

				if tc.expectedUser != nil && tc.equalCB == nil {
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

func TestCrudUpdate(t *testing.T) {
	sqlStore := eventstore.NewInMemoryStore()
	assert.NotNil(t, sqlStore)

	// cleaning up all of the data we added during those tests
	t.Cleanup(func() {
		assert.NoErrorf(t, sqlStore.Cleanup(), "cleanup failed")
	})

	crudStore, err := NewCrudStoreProvider(context.Background(), sqlStore)
	assert.NoError(t, err)
	assert.NotNil(t, crudStore)

	client := NewStructCrudStoreClient(crudStore)
	user := User{
		Email:  "makkalotupdate@gmail.com",
		Active: true,
	}
	originator, err := client.Create(&user)

	assert.NoError(t, err, "creation failed")
	assert.NotNil(t, originator, "empty originator")

	user.Active = false
	user.FirstName = "DK"

	updatedOriginator, err := client.Update(&user)
	assert.NoError(t, err, "first update failed")
	assert.NotNil(t, updatedOriginator, "updated originator empty")
	assert.Equal(t, updatedOriginator.Id, originator.Id)
	assert.Equal(t, updatedOriginator.Version, eskitcommon.MustIncrStringInt(originator.Version))
	assert.Equal(t, user.Originator.Version, eskitcommon.MustIncrStringInt(originator.Version))
	assert.Equal(t, user.Originator.Id, originator.Id)

	// get the latest version of
	var updatedUser User
	getErr := client.Get(&common.Originator{Id: originator.Id}, &updatedUser, false)
	assert.NoError(t, getErr, "fetching updated record failed")
	assert.Equal(t, updatedOriginator, updatedUser.Originator, "originator mismatch")
	assert.Equal(t, updatedUser, user, "user mismatch")
	//t.Logf("latest user : %s", spew.Sdump(updatedUser))

	// try to update with the same version it should fail
	var oldCopyUser User
	copyErr := copier.Copy(&oldCopyUser, &updatedUser)
	assert.NoError(t, copyErr, "copy operation failed")
	oldCopyUser.Originator.Version = "1"

	user.LastName = "LastName"
	lastUpdatedOriginator, err := client.Update(&oldCopyUser)
	assert.ErrorIs(t, err, eventstore.ErrDuplicate,  "there should be version duplicate")
	assert.Nil(t, lastUpdatedOriginator, "updated originator empty")

}


// tests the listing here
func TestCrudList(t *testing.T){
	sqlStore := eventstore.NewInMemoryStore()
	assert.NotNil(t, sqlStore)

	// cleaning up all of the data we added during those tests
	t.Cleanup(func() {
		assert.NoErrorf(t, sqlStore.Cleanup(), "cleanup failed")
	})

	crudStore, err := NewCrudStoreProvider(context.Background(), sqlStore)
	assert.NoError(t, err)
	assert.NotNil(t, crudStore)

	client := NewStructCrudStoreClient(crudStore)

	var users []*User
	lastOffsetID, listErr := client.ListWithPagination(&users, "", 10)

	assert.NoError(t, listErr, "listing failed")
	assert.Empty(t, lastOffsetID, "originator should be empty")
	assert.Len(t, users, 0, "users should be at length 0")

	user := User{
		Email:  "makkalolist@gmail.com",
		Active: true,
	}
	originator, err := client.Create(&user)

	assert.NoError(t, err, "creation failed")
	assert.NotNil(t, originator, "empty originator")

	lastOffsetID, listErr = client.ListWithPagination(&users, "", 10)

	assert.NoError(t, listErr, "listing failed")
	assert.NotEmpty(t, lastOffsetID, "originator should not be empty")
	assert.Len(t, users, 1, "users should be at length 1")
	assert.Equal(t, user.Originator, users[0].Originator, "user originator mismatch")

	// update the item and try to list it again
	user.FirstName = "Listing"
	updateOriginator, err := client.Update(&user)
	assert.NoError(t, err, "update failed")
	assert.NotNil(t, updateOriginator, "update originator empty")

	lastOffsetID, listErr = client.ListWithPagination(&users, "", 10)
	assert.NoError(t, listErr, "listing failed")
	assert.NotEmpty(t, lastOffsetID, "originator should not be empty")
	assert.Len(t, users, 1, "users should be at length 1")
	assert.Equal(t, user.Originator, users[0].Originator, "user originator mismatch")
	assert.Equal(t, updateOriginator, users[0].Originator, "user updated originator mismatch")
	assert.Equal(t, user.FirstName, users[0].FirstName, "haven't fetched the last one")

	// now let's add a new item and see
	userTwo := User{
		Email:  "makkalotsecond@gmail.com",
		Active: true,
	}
	originatorTwo, err := client.Create(&userTwo)

	assert.NoError(t, err, "creation for user 2 failed")
	assert.NotNil(t, originatorTwo, "empty originator for user 2")


	// try the pagination as well
	t.Run("pagination test", func(tt *testing.T) {
		lastOffsetID, listErr = client.ListWithPagination(&users, "", 1)
		t.Logf("the last offset id is like : %+v", lastOffsetID)

		assert.NoError(t, listErr, "listing failed")
		assert.NotEmpty(t, lastOffsetID, "originator should not be empty")
		assert.Len(t, users, 1, "users should be at length 2")

		// check the first user details
		assert.Equal(t, user.Originator, users[0].Originator, "user originator mismatch")
		assert.Equal(t, updateOriginator, users[0].Originator, "user updated originator mismatch")
		assert.Equal(t, user.FirstName, users[0].FirstName, "haven't fetched the last one")

		// it should be again the same record
		lastOffsetID, listErr = client.ListWithPagination(&users, lastOffsetID, 1)
		t.Logf("the last offset id is like : %+v", lastOffsetID)

		assert.NoError(t, listErr, "listing failed")
		assert.NotEmpty(t, lastOffsetID, "originator should not be empty")
		assert.Len(t, users, 1, "users should be at length 2")

		// check the first user details
		assert.Equal(t, user.Originator, users[0].Originator, "user originator mismatch")
		assert.Equal(t, updateOriginator, users[0].Originator, "user updated originator mismatch")
		assert.Equal(t, user.FirstName, users[0].FirstName, "haven't fetched the last one")

		// now we should we fetch the last record
		lastOffsetID, listErr = client.ListWithPagination(&users, lastOffsetID, 1)
		t.Logf("the last offset id is like : %+v", lastOffsetID)

		assert.NoError(t, listErr, "listing failed")
		assert.NotEmpty(t, lastOffsetID, "originator should not be empty")
		assert.Len(t, users, 1, "users should be at length 2")

		// check the first user details
		assert.Equal(t, userTwo, *users[0], "haven't fetched the last one")

		// lastly we should get the last item
		lastOffsetID, listErr = client.ListWithPagination(&users, lastOffsetID, 1)
		t.Logf("the last offset id is like : %+v", lastOffsetID)
		assert.NoError(t, listErr, "listing failed")
		assert.Empty(t, lastOffsetID, "originator should not be empty")
		assert.Len(t, users, 0, "users should be at length 2")

	})
}

// tests the deletion
func TestCrudDelete(t *testing.T){
	sqlStore := eventstore.NewInMemoryStore()
	assert.NotNil(t, sqlStore)

	// cleaning up all of the data we added during those tests
	t.Cleanup(func() {
		assert.NoErrorf(t, sqlStore.Cleanup(), "cleanup failed")
	})

	crudStore, err := NewCrudStoreProvider(context.Background(), sqlStore)
	assert.NoError(t, err)
	assert.NotNil(t, crudStore)

	client := NewStructCrudStoreClient(crudStore)

	t.Run("delete non existing user", func(tt *testing.T) {
		deleteOriginator, deleteErr := client.Delete(&common.Originator{Id: "non-existing"}, &User{})
		assert.ErrorIs(tt, deleteErr, RecordNotFound, "unexpected error")
		assert.Nil(tt, deleteOriginator, "non empty originator")
	})

	t.Run("delete user success", func(tt *testing.T) {
		// now try to delete something that's actually there
		userToDelete := User{
			Email:  "makkalotdelete@gmail.com",
			Active: true,
		}
		originatorDel, err := client.Create(&userToDelete)

		assert.NoError(tt, err, "creation for user delete failed")
		assert.NotNil(tt, originatorDel, "empty originator for user delete")

		deleteOriginator, deleteErr := client.Delete(originatorDel, &User{})
		assert.NoError(tt, deleteErr, RecordNotFound, "unexpected error")
		assert.NotNil(tt, deleteOriginator, "empty originator returned")

		// now if we fetch it the item should not come back
		var userRetrieve User
		getErr := client.Get(&common.Originator{Id: originatorDel.Id}, &userRetrieve, false)
		assert.ErrorIs(tt, getErr, RecordDeleted)

		// if we want to get the last version before it was deleted
		getErr = client.Get(&common.Originator{Id: originatorDel.Id}, &userRetrieve, true)
		assert.NoError(tt, getErr, "fetching deleted record failed")
		assert.Equal(tt, userToDelete, userRetrieve)

	})

	t.Run("delete user list", func(tt *testing.T) {
		userListDel := User{
			Email:  "makkalotlist@gmail.com",
			Active: true,
		}
		originatorDel, err := client.Create(&userListDel)

		assert.NoError(tt, err, "creation for user delete failed")
		assert.NotNil(tt, originatorDel, "empty originator for user delete")

		var users []*User
		lastOffsetID, listErr := client.ListWithPagination(&users, "", 0)
		t.Logf("the last offset id is like : %+v", lastOffsetID)
		assert.NoError(tt, listErr, "listing failed")
		assert.Len(tt, users, 1, "should have 1 user")
		assert.Equal(tt, userListDel, *users[0], "user mismatch")

		deleteOriginator, deleteErr := client.Delete(originatorDel, &User{})
		assert.NoError(tt, deleteErr, RecordNotFound, "unexpected error")
		assert.NotNil(tt, deleteOriginator, "empty originator returned")

		// now we should we fetch the last record
		lastOffsetID, listErr = client.ListWithPagination(&users, "", 0)
		t.Logf("the last offset id is like : %+v", lastOffsetID)
		assert.NoError(tt, listErr, "listing failed")
		assert.Len(tt, users, 0, "should have 1 user")

	})
}
