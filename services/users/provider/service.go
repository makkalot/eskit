package provider

import (
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/types"
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
