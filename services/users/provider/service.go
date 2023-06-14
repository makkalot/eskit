package provider

import (
	"context"
	"errors"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"github.com/makkalot/eskit/lib/crudstore"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type UserServiceProvider struct {
	crudStore crudstore.Client
}

type CreateRequest struct {
	Email     string
	FirstName string
	LastName  string
}

type User struct {
	Email     string
	FirstName string
	LastName  string
}

func NewUserServiceProvider(crudstore crudstore.Client) (*UserServiceProvider, error) {
	return &UserServiceProvider{crudStore: crudstore}, nil
}

func (u *UserServiceProvider) Create(ctx context.Context, request *CreateRequest) (*users.CreateResponse, error) {
	user := &User{
		Email:     request.Email,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	}

	_, err := u.crudStore.Create(user)
	if err != nil {
		return nil, status.Error(codes.Internal, "creation failed")
	}

	return &users.CreateResponse{
		User: user,
	}, nil
}

func (u *UserServiceProvider) Get(ctx context.Context, req *users.GetRequest) (*users.GetResponse, error) {
	if req.Originator == nil {
		return nil, status.Error(codes.InvalidArgument, "missing originator")
	}

	retrievedUser := &users.User{}
	if err := u.crudStore.Get(req.Originator, retrievedUser, req.FetchDeleted); err != nil {
		if errors.Is(err, crudstore.RecordNotFound) || errors.Is(err, crudstore.RecordDeleted) {
			return nil, status.Error(codes.NotFound, "deleted or not found")
		}
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
	if err := u.crudStore.Get(req.Originator, retrievedUser, false); err != nil {
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

	updatedOriginator, err := u.crudStore.Update(retrievedUser)
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

	deletedOriginator, err := u.crudStore.Delete(req.Originator, &users.User{})
	if err != nil {
		return nil, err
	}

	return &users.DeleteResponse{
		Originator: deletedOriginator,
	}, nil
}
