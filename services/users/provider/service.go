package provider

import (
	"context"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	eskitcommon "github.com/makkalot/eskit/services/lib/common"

	"github.com/makkalot/eskit/services/clients"

	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/crudstore"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceProvider struct {
	crud *clients.CrudStoreClient
}

func NewUserServiceProvider(crudStoreEndpoint string) (*UserServiceProvider, error) {
	ctx := context.Background()
	var crudConn *grpc.ClientConn

	if err := eskitcommon.RetryNormal(func() error {
		var err error
		crudConn, err = grpc.Dial(crudStoreEndpoint, grpc.WithInsecure())
		if err != nil {
			return err
		}

		return nil
	}); err != nil {
		return nil, err
	}

	crudGRPC := crudstore.NewCrudStoreServiceClient(crudConn)
	_, err := crudGRPC.RegisterType(ctx, &crudstore.RegisterTypeRequest{
		Spec: &crudstore.CrudEntitySpec{
			EntityType: clients.EntityTypeFromMsg(&users.User{}),
		},
		SkipDuplicate: true,
	})

	if err != nil {
		return nil, fmt.Errorf("type registration for use entity type failed : %v", err)
	}

	crudClient, err := clients.NewCrudStoreWithActiveConn(ctx, crudGRPC)
	if err != nil {
		return nil, err
	}

	return &UserServiceProvider{
		crud: crudClient,
	}, nil
}

func (u *UserServiceProvider) Healtz(ctx context.Context, request *users.HealthRequest) (*users.HealthResponse, error) {
	return &users.HealthResponse{}, nil
}

func (u *UserServiceProvider) Create(ctx context.Context, request *users.CreateRequest) (*users.CreateResponse, error) {
	originator := &common.Originator{
		Id:      uuid.Must(uuid.NewV4()).String(),
		Version: "1",
	}

	user := &users.User{
		Originator: originator,
		Email:      request.Email,
		FirstName:  request.FirstName,
		LastName:   request.LastName,
	}

	createdOriginator, err := u.crud.Create(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "creation failed")
	}

	user.Originator = createdOriginator

	return &users.CreateResponse{
		User: user,
	}, nil
}

func (u *UserServiceProvider) Get(ctx context.Context, req *users.GetRequest) (*users.GetResponse, error) {
	if req.Originator == nil {
		return nil, status.Error(codes.InvalidArgument, "missing originator")
	}

	retrievedUser := &users.User{}
	if err := u.crud.Get(req.Originator, retrievedUser, req.FetchDeleted); err != nil {
		return nil, err
	}

	return &users.GetResponse{
		User: retrievedUser,
	}, nil
}

func (u *UserServiceProvider) Update(ctx context.Context, req *users.UpdateRequest) (*users.UpdateResponse, error) {
	if req.Originator == nil {
		return nil, status.Error(codes.InvalidArgument, "missing originator")
	}

	retrievedUser := &users.User{}
	if err := u.crud.Get(req.Originator, retrievedUser, false); err != nil {
		return nil, err
	}

	if req.LastName != "" {
		retrievedUser.LastName = req.LastName
	}

	if req.FirstName != "" {
		retrievedUser.FirstName = req.FirstName
	}

	if req.Email != "" {
		retrievedUser.Email = req.Email
	}

	if req.Active != retrievedUser.Active {
		retrievedUser.Active = req.Active
	}

	updatedOriginator, err := u.crud.Update(retrievedUser)
	if err != nil {
		return nil, err
	}

	retrievedUser.Originator = updatedOriginator

	return &users.UpdateResponse{
		User: retrievedUser,
	}, nil

}

func (u *UserServiceProvider) Delete(ctx context.Context, req *users.DeleteRequest) (*users.DeleteResponse, error) {
	if req.Originator == nil {
		return nil, status.Error(codes.InvalidArgument, "missing originator")
	}

	deletedOriginator, err := u.crud.Delete(req.Originator, &users.User{})
	if err != nil {
		return nil, err
	}

	return &users.DeleteResponse{
		Originator: deletedOriginator,
	}, nil
}
