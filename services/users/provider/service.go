package provider

import (
	"context"
	"errors"
	"github.com/makkalot/eskit/adapters/proto"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// User represents the internal native user type used with the library
type User struct {
	Originator *types.Originator
	Email      string
	FirstName  string
	LastName   string
	Active     bool
	Workspaces []string
}

type UserServiceProvider struct {
	crudStore crudstore.Client
}

func NewUserServiceProvider(crudstore crudstore.Client) (*UserServiceProvider, error) {
	return &UserServiceProvider{crudStore: crudstore}, nil
}

// userToProto converts native User to proto User
func userToProto(u *User) *users.User {
	if u == nil {
		return nil
	}
	return &users.User{
		Originator: proto.OriginatorToProto(u.Originator),
		Email:      u.Email,
		FirstName:  u.FirstName,
		LastName:   u.LastName,
		Active:     u.Active,
		Workspaces: u.Workspaces,
	}
}

// userFromProto converts proto User to native User
func userFromProto(pb *users.User) *User {
	if pb == nil {
		return nil
	}
	return &User{
		Originator: proto.OriginatorFromProto(pb.Originator),
		Email:      pb.Email,
		FirstName:  pb.FirstName,
		LastName:   pb.LastName,
		Active:     pb.Active,
		Workspaces: pb.Workspaces,
	}
}

func (u *UserServiceProvider) Healtz(ctx context.Context, request *users.HealthRequest) (*users.HealthResponse, error) {
	return &users.HealthResponse{}, nil
}

func (u *UserServiceProvider) Create(ctx context.Context, request *users.CreateRequest) (*users.CreateResponse, error) {
	// Create native user from request
	nativeUser := &User{
		Email:     request.Email,
		FirstName: request.FirstName,
		LastName:  request.LastName,
	}

	// Use library with native types
	_, err := u.crudStore.Create(nativeUser)
	if err != nil {
		return nil, status.Error(codes.Internal, "creation failed")
	}

	// Convert native user back to proto for response
	return &users.CreateResponse{
		User: userToProto(nativeUser),
	}, nil
}

func (u *UserServiceProvider) Get(ctx context.Context, req *users.GetRequest) (*users.GetResponse, error) {
	if req.Originator == nil {
		return nil, status.Error(codes.InvalidArgument, "missing originator")
	}

	// Convert proto Originator to native
	nativeOriginator := proto.OriginatorFromProto(req.Originator)

	// Use library with native types
	retrievedUser := &User{}
	if err := u.crudStore.Get(nativeOriginator, retrievedUser, req.FetchDeleted); err != nil {
		if errors.Is(err, crudstore.RecordNotFound) || errors.Is(err, crudstore.RecordDeleted) {
			return nil, status.Error(codes.NotFound, "deleted or not found")
		}
		return nil, err
	}

	// Convert native user back to proto for response
	return &users.GetResponse{
		User: userToProto(retrievedUser),
	}, nil
}

func (u *UserServiceProvider) Update(ctx context.Context, req *users.UpdateRequest) (*users.UpdateResponse, error) {
	if req.Originator == nil {
		return nil, status.Error(codes.InvalidArgument, "missing originator")
	}

	// Convert proto Originator to native
	nativeOriginator := proto.OriginatorFromProto(req.Originator)

	// Get existing user with native types
	retrievedUser := &User{}
	if err := u.crudStore.Get(nativeOriginator, retrievedUser, false); err != nil {
		return nil, err
	}

	// Update fields from request
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

	// Update using library with native types
	updatedOriginator, err := u.crudStore.Update(retrievedUser)
	if err != nil {
		return nil, err
	}

	retrievedUser.Originator = updatedOriginator

	// Convert native user back to proto for response
	return &users.UpdateResponse{
		User: userToProto(retrievedUser),
	}, nil
}

func (u *UserServiceProvider) Delete(ctx context.Context, req *users.DeleteRequest) (*users.DeleteResponse, error) {
	if req.Originator == nil {
		return nil, status.Error(codes.InvalidArgument, "missing originator")
	}

	// Convert proto Originator to native
	nativeOriginator := proto.OriginatorFromProto(req.Originator)

	// Delete using library with native types
	deletedOriginator, err := u.crudStore.Delete(nativeOriginator, &User{})
	if err != nil {
		return nil, err
	}

	// Convert native Originator back to proto for response
	return &users.DeleteResponse{
		Originator: proto.OriginatorToProto(deletedOriginator),
	}, nil
}
