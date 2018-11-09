package clients

import (
	"testing"
	"context"
	"github.com/makkalot/eskit/services/clients/mocks"
	"github.com/makkalot/eskit/generated/grpc/go/crudstore"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"github.com/golang/protobuf/jsonpb"
)

func TestCrudAdd(t *testing.T) {

	ctx := context.Background()
	entityType := EntityTypeFromMsg(&users.User{})
	originator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "1",
	}

	user := &users.User{
		Originator: originator,
		Email:      "testeskit@gmail.com",
		FirstName:  "test",
		LastName:   "eskit",
	}

	marshaller := &jsonpb.Marshaler{}
	payloadJSON, err := marshaller.MarshalToString(user)
	assert.NoError(t, err)

	mockCrud := &mocks.CrudStoreServiceClient{}
	mockCrud.On("Create", ctx, &crudstore.CreateRequest{
		EntityType: entityType,
		Originator: originator,
		Payload:    payloadJSON,
	}).Return(&crudstore.CreateResponse{
		Originator: originator,
	}, nil)

	crud, err := NewCrudStoreWithActiveConn(ctx, mockCrud)
	assert.NoError(t, err)
	assert.NotNil(t, crud)

	createOriginator, err := crud.Create(user)
	assert.NoError(t, err)
	assert.Equal(t, createOriginator, originator)
}

func TestCrudGet(t *testing.T) {
	ctx := context.Background()
	originator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "1",
	}

	user := &users.User{
		Originator: originator,
		Email:      "testeskit@gmail.com",
		FirstName:  "test",
		LastName:   "eskit",
	}

	marshaller := &jsonpb.Marshaler{}
	payloadJSON, err := marshaller.MarshalToString(user)
	assert.NoError(t, err)

	getRequest := &crudstore.GetRequest{
		EntityType: EntityTypeFromMsg(user),
		Originator: originator,
		Deleted:    false,
	}

	getResp := &crudstore.GetResponse{
		Originator: originator,
		Payload:    payloadJSON,
	}

	mockCrud := &mocks.CrudStoreServiceClient{}
	mockCrud.On("Get", ctx, getRequest).Return(getResp, nil)

	crud, err := NewCrudStoreWithActiveConn(ctx, mockCrud)
	assert.NoError(t, err)
	assert.NotNil(t, crud)

	retrieveUser := &users.User{}
	err = crud.Get(originator, retrieveUser, false)
	assert.NoError(t, err)
	assert.Equal(t, user, retrieveUser)
}

func TestUpdate(t *testing.T) {

	ctx := context.Background()
	entityType := EntityTypeFromMsg(&users.User{})
	originator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "2",
	}

	user := &users.User{
		Originator: originator,
		Email:      "testeskit@gmail.com",
		FirstName:  "test",
		LastName:   "eskit",
	}

	marshaller := &jsonpb.Marshaler{}
	payloadJSON, err := marshaller.MarshalToString(user)
	assert.NoError(t, err)

	mockCrud := &mocks.CrudStoreServiceClient{}
	mockCrud.On("Update", ctx, &crudstore.UpdateRequest{
		EntityType: entityType,
		Originator: originator,
		Payload:    payloadJSON,
	}).Return(&crudstore.UpdateResponse{
		Originator: originator,
	}, nil)

	crud, err := NewCrudStoreWithActiveConn(ctx, mockCrud)
	assert.NoError(t, err)
	assert.NotNil(t, crud)

	updateOriginator, err := crud.Update(user)
	assert.NoError(t, err)
	assert.Equal(t, updateOriginator, originator)

}

func TestDelete(t *testing.T) {
	ctx := context.Background()
	entityType := EntityTypeFromMsg(&users.User{})
	originator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "2",
	}

	mockCrud := &mocks.CrudStoreServiceClient{}
	mockCrud.On("Delete", ctx, &crudstore.DeleteRequest{
		EntityType: entityType,
		Originator: originator,
	}).Return(&crudstore.DeleteResponse{
		Originator: originator,
	}, nil)

	crud, err := NewCrudStoreWithActiveConn(ctx, mockCrud)
	assert.NoError(t, err)
	assert.NotNil(t, crud)

	deleteOriginator, err := crud.Delete(originator, &users.User{})
	assert.NoError(t, err)
	assert.Equal(t, deleteOriginator, originator)

}

func TestList(t *testing.T) {

	ctx := context.Background()
	entityType := EntityTypeFromMsg(&users.User{})

	firstOirignator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "2",
	}

	secondOriginator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "2",
	}

	thirdOriginator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "2",
	}

	forthOriginator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "2",
	}

	userList := []*users.User{
		{
			Email:     "firstuser@gmail.com",
			FirstName: "First",
		},
		{
			Email:     "seconduser@gmail.com",
			FirstName: "Second",
		},
		{
			Email:     "third@gmail.com",
			FirstName: "Third",
		},
		{
			Email:     "forth@gmail.com",
			FirstName: "Forth",
		},
	}

	marshaller := &jsonpb.Marshaler{}
	firstUserJSON, err := marshaller.MarshalToString(userList[0])
	assert.NoError(t, err)

	secondUserJSON, err := marshaller.MarshalToString(userList[1])
	assert.NoError(t, err)

	thirdUserJSON, err := marshaller.MarshalToString(userList[2])
	assert.NoError(t, err)

	forthUserJSON, err := marshaller.MarshalToString(userList[3])
	assert.NoError(t, err)

	listItems := []*crudstore.ListResponseItem{
		{EntityType: entityType, Originator: firstOirignator, Payload: firstUserJSON},
		{EntityType: entityType, Originator: secondOriginator, Payload: secondUserJSON},
		{EntityType: entityType, Originator: thirdOriginator, Payload: thirdUserJSON},
		{EntityType: entityType, Originator: forthOriginator, Payload: forthUserJSON},
	}

	mockCrud := &mocks.CrudStoreServiceClient{}

	mockCrud.On("List", ctx, &crudstore.ListRequest{
		EntityType: entityType,
	}).Return(&crudstore.ListResponse{
		Results: listItems,
	}, nil)

	crud, err := NewCrudStoreWithActiveConn(ctx, mockCrud)
	assert.NoError(t, err)
	assert.NotNil(t, crud)

	var result []*users.User
	err = crud.List(&result)
	assert.NoError(t, err)
	assert.Len(t, result, 4)

	expectedUserList := []*users.User{
		{
			Originator: firstOirignator,
			Email:      userList[0].Email,
			FirstName:  userList[0].FirstName,
		},
		{
			Originator: secondOriginator,
			Email:      userList[1].Email,
			FirstName:  userList[1].FirstName,
		},
		{
			Originator: thirdOriginator,
			Email:      userList[2].Email,
			FirstName:  userList[2].FirstName,
		},
		{
			Originator: forthOriginator,
			Email:      userList[3].Email,
			FirstName:  userList[3].FirstName,
		},
	}
	assert.Equal(t, result, expectedUserList)
}
