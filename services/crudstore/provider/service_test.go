package provider

import (
	"context"
	"github.com/go-test/deep"
	"github.com/golang/protobuf/proto"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	crudstore "github.com/makkalot/eskit/generated/grpc/go/crudstore"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"github.com/makkalot/eskit/services/clients"
	crudstore2 "github.com/makkalot/eskit/services/lib/crudstore"
	eventstore2 "github.com/makkalot/eskit/services/lib/eventstore"
	"github.com/makkalot/eskit/tests/integration/util"
	"github.com/satori/go.uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/grpc/codes"
	"os"
	"testing"
)

func TestCrudStoreProvider_Health(t *testing.T) {

	var client crudstore.CrudStoreServiceClient
	unitC := newUnitTestClient(t)
	client = unitC
	defer unitC.cleanup()

	_, err := client.Healtz(context.Background(), &crudstore.HealthRequest{})
	assert.NoError(t, err)
}

func TestCrudStoreSvcProvider_Create(t *testing.T) {

	ctx := context.Background()

	var client crudstore.CrudStoreServiceClient
	unitC := newUnitTestClient(t)
	client = unitC
	defer unitC.cleanup()

	u1 := uuid.Must(uuid.NewV4())
	user := &users.User{
		Originator: &common.Originator{
			Id: u1.String(),
		},
		Email:     "kevinabi",
		FirstName: "Kevin",
		LastName:  "Abi",
		Active:    true,
	}

	crudClient, err := clients.NewCrudStoreWithActiveConn(
		ctx, client,
	)
	assert.NoError(t, err)
	assert.NotNil(t, crudClient)

	userOriginator, err := crudClient.Create(user)
	assert.NoError(t, err)
	assert.NotNil(t, userOriginator)

	retrievedUser := &users.User{}
	err = crudClient.Get(userOriginator, retrievedUser, false)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(user, retrievedUser))

	// try to create the same one to check for conflicts
	duplicateOriginator, err := crudClient.Create(user)
	assert.Error(t, err)
	assert.NoError(t, util.AssertGrpcCodeErr(err, codes.AlreadyExists))
	assert.Nil(t, duplicateOriginator)

	// try to receive a non existing one
	u2 := uuid.Must(uuid.NewV4())
	retrievedUser = &users.User{}
	err = crudClient.Get(&common.Originator{Id: u2.String()}, retrievedUser, false)
	assert.Error(t, err)
	assert.NoError(t, util.AssertGrpcCodeErr(err, codes.NotFound))
}

func TestCrudStoreProvider_Update(t *testing.T) {
	ctx := context.Background()

	var client crudstore.CrudStoreServiceClient
	unitC := newUnitTestClient(t)
	client = unitC
	defer unitC.cleanup()

	u1 := uuid.Must(uuid.NewV4())
	userV1 := &users.User{
		Originator: &common.Originator{
			Id: u1.String(),
		},
		Email:     "kevinabi@gmail.com",
		FirstName: "Kevin",
		LastName:  "Abi",
		Active:    true,
	}

	crudClient, err := clients.NewCrudStoreWithActiveConn(
		ctx, client,
	)
	assert.NoError(t, err)
	assert.NotNil(t, crudClient)

	userOriginator, err := crudClient.Create(userV1)
	assert.NoError(t, err)
	assert.NotNil(t, userOriginator)

	retrievedUser := &users.User{}
	err = crudClient.Get(userOriginator, retrievedUser, false)
	assert.NoError(t, err)
	retrievedUser.Email = "changed@gmail.com"

	updatedOriginator, err := crudClient.Update(retrievedUser)
	assert.NoError(t, err)

	updatedUser := &users.User{}
	err = crudClient.Get(updatedOriginator, updatedUser, false)
	assert.Equal(t, updatedUser.Email, "changed@gmail.com")

	// try to get the latest version of the userV1 without the version
	noVersionUser := &users.User{}
	err = crudClient.Get(&common.Originator{Id: updatedUser.Originator.Id}, noVersionUser, false)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(noVersionUser, updatedUser))

	// try to get the first version on creation too
	firstUser := &users.User{}
	err = crudClient.Get(userOriginator, firstUser, false)
	assert.NoError(t, err)
	userV1.Originator = userOriginator
	assert.True(t, proto.Equal(firstUser, userV1))

	// try to get a version that does not exist
	nonExistingUser := &users.User{}
	err = crudClient.Get(&common.Originator{Id: userOriginator.Id, Version: "3"}, nonExistingUser, false)
	assert.Error(t, err)
	assert.NoError(t, util.AssertGrpcCodeErr(err, codes.NotFound))

	// try to insert an object with same version so we can test the concurrency
	firstUser.FirstName = "eskitman"
	_, err = crudClient.Update(retrievedUser)
	assert.Error(t, err)
	assert.NoError(t, util.AssertGrpcCodeErr(err, codes.AlreadyExists))

}

func TestCrudStoreProvider_Delete(t *testing.T) {
	ctx := context.Background()

	var client crudstore.CrudStoreServiceClient
	unitC := newUnitTestClient(t)
	client = unitC
	defer unitC.cleanup()

	u1 := uuid.Must(uuid.NewV4())
	userV1 := &users.User{
		Originator: &common.Originator{
			Id: u1.String(),
		},
		Email:     "kevinabi@gmail.com",
		FirstName: "Kevin",
		LastName:  "Abi",
		Active:    true,
	}

	crudClient, err := clients.NewCrudStoreWithActiveConn(
		ctx, client,
	)
	assert.NoError(t, err)
	assert.NotNil(t, crudClient)

	userOriginator, err := crudClient.Create(userV1)
	assert.NoError(t, err)
	assert.NotNil(t, userOriginator)

	deleteOriginator, err := crudClient.Delete(userOriginator, &users.User{})
	assert.NoError(t, err)
	assert.NotNil(t, deleteOriginator)

	versionlessOriginator := &common.Originator{
		Id: deleteOriginator.Id,
	}

	err = crudClient.Get(deleteOriginator, &users.User{}, false)
	assert.Error(t, err)
	assert.NoError(t, util.AssertGrpcCodeErr(err, codes.NotFound))

	err = crudClient.Get(versionlessOriginator, &users.User{}, false)
	assert.Error(t, err)
	assert.NoError(t, util.AssertGrpcCodeErr(err, codes.NotFound))

	//// try to get em via delete = true
	err = crudClient.Get(deleteOriginator, &users.User{}, true)
	assert.NoError(t, err)

	err = crudClient.Get(versionlessOriginator, &users.User{}, true)
	assert.NoError(t, err)
}

func TestCrudStoreSvcProvider_List(t *testing.T) {
	ctx := context.Background()

	var client crudstore.CrudStoreServiceClient
	unitC := newUnitTestClient(t)
	client = unitC
	defer unitC.cleanup()

	crudClient, err := clients.NewCrudStoreWithActiveConn(
		ctx, client,
	)
	assert.NoError(t, err)
	assert.NotNil(t, crudClient)

	u1 := uuid.Must(uuid.NewV4())
	userV1 := &users.User{
		Originator: &common.Originator{
			Id: u1.String(),
		},
		Email:     "kevinabi@gmail.com",
		FirstName: "Kevin",
		LastName:  "Abi",
		Active:    true,
	}

	userOriginator, err := crudClient.Create(userV1)
	assert.NoError(t, err)
	assert.NotNil(t, userOriginator)

	var results []*users.User

	_, err = crudClient.ListWithPagination(&results, "", 0)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0], userV1)

	u2 := uuid.Must(uuid.NewV4())
	user2 := &users.User{
		Originator: &common.Originator{
			Id: u2.String(),
		},
		Email:     "limbo@gmail.com",
		FirstName: "Limbo",
		LastName:  "Abi",
		Active:    true,
	}

	userOriginator2, err := crudClient.Create(user2)
	assert.NoError(t, err)
	assert.NotNil(t, userOriginator2)

	results = nil
	_, err = crudClient.ListWithPagination(&results, "", 0)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, len(results), 2)
	assert.Equal(t, results[0], userV1)
	assert.Equal(t, results[1], user2)

	results = nil
	nextPage, err := crudClient.ListWithPagination(&results, "", 1)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0], userV1)

	results = nil
	nextPage, err = crudClient.ListWithPagination(&results, nextPage, 1)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0], user2)

	results = nil
	_, err = crudClient.ListWithPagination(&results, nextPage, 1)
	assert.NoError(t, err)
	assert.Nil(t, results)

	// now update the user 1 and fetch the list again
	retrievedUser := &users.User{}
	err = crudClient.Get(userOriginator, retrievedUser, false)
	assert.NoError(t, err)
	retrievedUser.Email = "changed@gmail.com"

	updatedOriginator, err := crudClient.Update(retrievedUser)
	assert.NoError(t, err)
	assert.NotNil(t, updatedOriginator)

	updatedUser := &users.User{}
	err = crudClient.Get(updatedOriginator, updatedUser, false)
	assert.NoError(t, err)

	results = nil
	_, err = crudClient.ListWithPagination(&results, "", 0)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, len(results), 2)
	assert.Equal(t, results[0], updatedUser)
	assert.Equal(t, results[1], user2)

	results = nil
	nextPage, err = crudClient.ListWithPagination(&results, "", 1)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0], updatedUser)

	results = nil
	nextPage, err = crudClient.ListWithPagination(&results, nextPage, 1)
	assert.NoError(t, err)
	assert.NotNil(t, results)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0], user2)

	// it will get the user1 again because of the pagination since the events
	// are like user1.Created user2.Created user1.Updated
	results = nil
	_, err = crudClient.ListWithPagination(&results, nextPage, 1)
	assert.NoError(t, err)
	assert.Equal(t, len(results), 1)
	assert.Equal(t, results[0], updatedUser)
}

func TestCrudStoreSvcProvider_RegisterType(tm *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name string
		spec *crudstore.CrudEntitySpec
		err  bool
		code codes.Code
		cb   func(client crudstore.CrudStoreServiceClient, spec *crudstore.CrudEntitySpec) error
	}{

		{
			name: "valid spec",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "com.makkalot.Project4",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "http://example.com/product.schema.json",
  "title": "Product",
  "description": "A product from Acme's catalog",
  "type": "object",
  "properties": {
    "productId": {
      "description": "The unique identifier for a product",
      "type": "integer"
    }
  },
  "required": [ "productId" ]
}`,
				},
			},
		},
		{
			name: "invalid json spec",
			err:  true,
			code: codes.InvalidArgument,
			spec: &crudstore.CrudEntitySpec{
				EntityType: "com.makkalot.Project5",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema:    `{2132}`,
				},
			},
		},
		{
			name: "empty spec",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "com.makkalot.Project3",
				SchemaSpec: &crudstore.SchemaSpec{},
			},
		},
		{
			name: "duplicate spec",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "com.makkalot.Duplicate",
				SchemaSpec: &crudstore.SchemaSpec{
					JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#"}`,
					SchemaVersion: 1,
				},
			},
			cb: func(client crudstore.CrudStoreServiceClient, spec *crudstore.CrudEntitySpec) error {
				_, err := client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: spec,
				})
				if err != nil {
					return err
				}
				_, err = client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: spec,
				})
				return err
			},
			err:  true,
			code: codes.AlreadyExists,
		},
		{
			name: "duplicate spec allowed",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "com.makkalot.Duplicate",
				SchemaSpec: &crudstore.SchemaSpec{
					JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#"}`,
					SchemaVersion: 1,
				},
			},
			cb: func(client crudstore.CrudStoreServiceClient, spec *crudstore.CrudEntitySpec) error {
				_, err := client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: spec,
				})
				if err != nil {
					return err
				}
				_, err = client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec:          spec,
					SkipDuplicate: true,
				})
				return err
			},
		},

		{
			name: "save different",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "com.makkalot.Duplicate",
				SchemaSpec: &crudstore.SchemaSpec{
					JsonSchema: `{
		"$schema": "http://json-schema.org/draft-07/schema#"}`,
					SchemaVersion: 1,
				},
			},
			cb: func(client crudstore.CrudStoreServiceClient, spec *crudstore.CrudEntitySpec) error {
				_, err := client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: spec,
				})
				if err != nil {
					return err
				}

				spec.EntityType = "com.makkalot.Duplicate2"
				_, err = client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: spec,
				})
				return err
			},
		},
	}

	for _, tc := range testCases {
		tm.Run(tc.name, func(tt *testing.T) {
			var client crudstore.CrudStoreServiceClient
			unitC := newUnitTestClient(tt)
			client = unitC
			defer unitC.cleanup()

			crudClient, err := clients.NewCrudStoreWithActiveConn(
				ctx, client,
			)
			assert.NoError(tt, err)
			assert.NotNil(tt, crudClient)

			if tc.cb == nil {
				_, err = crudClient.CrudGRPC.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: tc.spec,
				})
			} else {
				err = tc.cb(crudClient.CrudGRPC, tc.spec)
			}

			if !tc.err {
				assert.NoError(tt, err)
			} else {
				assert.Error(tt, err)
				assert.NoError(tt, util.AssertGrpcCodeErr(err, tc.code))
			}
		})
	}

}

func TestCrudStoreSvcProvider_GetType(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name string
		err  bool
		code codes.Code
		cb   func(tt *testing.T, client crudstore.CrudStoreServiceClient) error
	}{
		{
			name: "retrieve created",
			cb: func(tt *testing.T, client crudstore.CrudStoreServiceClient) error {
				entitySpec := &crudstore.CrudEntitySpec{
					EntityType: "com.makkalot.Found",
					SchemaSpec: &crudstore.SchemaSpec{
						JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#"}`,
						SchemaVersion: 1,
					},
				}

				_, err := client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: entitySpec,
				})
				assert.NoError(tt, err)

				resp, err := client.GetType(ctx, &crudstore.GetTypeRequest{
					EntityType: "com.makkalot.Found",
				})
				if err != nil {
					return err
				}

				assert.Equal(tt, entitySpec, resp.Spec)
				return nil
			},
		},
		{
			name: "does not exist",
			err:  true,
			code: codes.NotFound,
			cb: func(tt *testing.T, client crudstore.CrudStoreServiceClient) error {

				resp, err := client.GetType(ctx, &crudstore.GetTypeRequest{
					EntityType: "com.makkalot.NotFound",
				})
				assert.Nil(tt, resp)
				return err
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var client crudstore.CrudStoreServiceClient
			unitC := newUnitTestClient(tt)
			client = unitC
			defer unitC.cleanup()

			crudClient, err := clients.NewCrudStoreWithActiveConn(
				ctx, client,
			)
			assert.NoError(tt, err)
			assert.NotNil(tt, crudClient)

			err = tc.cb(tt, crudClient.CrudGRPC)

			if !tc.err {
				assert.NoError(tt, err)
			} else {
				assert.Error(tt, err)
				assert.NoError(tt, util.AssertGrpcCodeErr(err, tc.code))
			}
		})
	}
}

func TestCrudStoreSvcProvider_UpdateType(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name string
		err  bool
		code codes.Code
		cb   func(tt *testing.T, client crudstore.CrudStoreServiceClient) error
	}{
		{
			name: "update success",
			cb: func(tt *testing.T, client crudstore.CrudStoreServiceClient) error {
				entitySpec := &crudstore.CrudEntitySpec{
					EntityType: "com.makkalot.Update",
					SchemaSpec: &crudstore.SchemaSpec{
						JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#"}`,
						SchemaVersion: 1,
					},
				}

				_, err := client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: entitySpec,
				})
				assert.NoError(tt, err)

				resp, err := client.GetType(ctx, &crudstore.GetTypeRequest{
					EntityType: "com.makkalot.Update",
				})
				if err != nil {
					return err
				}

				assert.Equal(tt, entitySpec, resp.Spec)

				updateSpec := &crudstore.CrudEntitySpec{
					EntityType: "com.makkalot.Update",
					SchemaSpec: &crudstore.SchemaSpec{
						JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
"$id": "http://example.com/product.schema.json",
  "title": "Product",
  "description": "A product from Acme's catalog",
  "type": "object"}`,
						SchemaVersion: 2,
					},
				}

				_, err = client.UpdateType(ctx, &crudstore.UpdateTypeRequest{
					Spec: updateSpec,
				})
				if err != nil {
					return err
				}

				resp, err = client.GetType(ctx, &crudstore.GetTypeRequest{
					EntityType: "com.makkalot.Update",
				})
				if err != nil {
					return err
				}

				assert.Equal(tt, updateSpec, resp.Spec)

				return nil
			},
		},
		{
			name: "update non found",
			err:  true,
			code: codes.NotFound,
			cb: func(tt *testing.T, client crudstore.CrudStoreServiceClient) error {
				updateSpec := &crudstore.CrudEntitySpec{
					EntityType: "com.makkalot.NotFound",
					SchemaSpec: &crudstore.SchemaSpec{
						JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
"$id": "http://example.com/product.schema.json",
  "title": "Product",
  "description": "A product from Acme's catalog",
  "type": "object"}`,
						SchemaVersion: 2,
					},
				}

				_, err := client.UpdateType(ctx, &crudstore.UpdateTypeRequest{
					Spec: updateSpec,
				})
				return err

			},
		},
		{
			name: "update invalid schema",
			err:  true,
			code: codes.NotFound,
			cb: func(tt *testing.T, client crudstore.CrudStoreServiceClient) error {
				updateSpec := &crudstore.CrudEntitySpec{
					EntityType: "com.makkalot.NotFound",
					SchemaSpec: &crudstore.SchemaSpec{
						JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
"$id": "http://example.com/product.schema.json",
  "title": "Product",
  "description": "A product from Acme's catalog",
  "type": "object"}`,
						SchemaVersion: 2,
					},
				}

				_, err := client.UpdateType(ctx, &crudstore.UpdateTypeRequest{
					Spec: updateSpec,
				})
				return err

			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var client crudstore.CrudStoreServiceClient
			unitC := newUnitTestClient(tt)
			client = unitC
			defer unitC.cleanup()

			crudClient, err := clients.NewCrudStoreWithActiveConn(
				ctx, client,
			)
			assert.NoError(tt, err)
			assert.NotNil(tt, crudClient)

			err = tc.cb(tt, crudClient.CrudGRPC)

			if !tc.err {
				assert.NoError(tt, err)
			} else {
				assert.Error(tt, err)
				assert.NoError(tt, util.AssertGrpcCodeErr(err, tc.code))
			}
		})
	}
}

func TestCrudStoreSvcProvider_Create_CheckSchema(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name string
		spec *crudstore.CrudEntitySpec
		user *users.User
		err  bool
		code codes.Code
		cb   func(tt *testing.T, client crudstore.CrudStoreServiceClient) error
	}{
		{
			name: "no schema",
			user: &users.User{},
		},
		{
			name: "empty schema",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
			},
			user: &users.User{},
		},
		{
			name: "missing required",
			err:  true,
			code: codes.InvalidArgument,
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "http://example.com/product.schema.json",
  "title": "User",
  "type": "object",
  "properties": {
    "email": {
      "type": "string"
    },
	"firstName": {
      "type": "string"
    }
  },
  "required": [ "email", "firstName" ]
}`,
				},
			},
			user: &users.User{},
		},
		{
			name: "missing required fixed",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "http://example.com/product.schema.json",
  "title": "User",
  "type": "object",
  "properties": {
    "email": {
      "type": "string"
    },
	"firstName": {
      "type": "string"
    }
  },
  "required": [ "email", "firstName" ]
}`,
				},
			},
			user: &users.User{
				Email:     "linux@linux.com",
				FirstName: "Linus",
			},
		},
		{
			name: "min length",
			err:  true,
			code: codes.InvalidArgument,
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "http://example.com/product.schema.json",
  "title": "User",
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "minLength": 3
    },
	"firstName": {
      "type": "string",
      "minLength": 3
    }
  },
  "required": [ "email", "firstName" ]
}`,
				},
			},
			user: &users.User{
				Email:     "worm@gmail.com",
				FirstName: "Wo",
			},
		},
		{
			name: "min length fixed",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
  "$schema": "http://json-schema.org/draft-07/schema#",
  "$id": "http://example.com/product.schema.json",
  "title": "User",
  "type": "object",
  "properties": {
    "email": {
      "type": "string",
      "minLength": 3
    },
	"firstName": {
      "type": "string",
      "minLength": 3
    }
  },
  "required": [ "email", "firstName" ]
}`,
				},
			},
			user: &users.User{
				Email:     "worm@gmail.com",
				FirstName: "Worm",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var client crudstore.CrudStoreServiceClient
			unitC := newUnitTestClient(tt)
			client = unitC
			defer unitC.cleanup()

			crudClient, err := clients.NewCrudStoreWithActiveConn(
				ctx, client,
			)
			assert.NoError(tt, err)
			assert.NotNil(tt, crudClient)

			if tc.spec != nil {
				_, err := client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
					Spec: tc.spec,
				})
				assert.NoError(tt, err)
			}

			var userOriginator *common.Originator
			if tc.cb != nil {
				err = tc.cb(tt, crudClient.CrudGRPC)
			} else {
				_, err = crudClient.Create(tc.user)
			}

			if !tc.err {
				assert.NoError(tt, err)
				if userOriginator != nil {
					retrievedUser := &users.User{}
					err := crudClient.Get(userOriginator, retrievedUser, false)
					assert.NoError(tt, err)
					assert.Equal(tt, tc.user, retrievedUser)
				}
			} else {
				assert.Error(tt, err)
				if err != nil {
					assert.NoError(tt, util.AssertGrpcCodeErr(err, tc.code))
				}
			}
		})
	}
}

func TestCrudStoreSvcProvider_Update_CheckSchema(t *testing.T) {
	ctx := context.Background()

	testCases := []struct {
		name string
		spec *crudstore.CrudEntitySpec
		user *users.User
		err  bool
		code codes.Code
		cb   func(tt *testing.T, client crudstore.CrudStoreServiceClient) error
	}{
		{
			name: "no schema",
			user: &users.User{
				FirstName: "SomeName",
			},
		},
		{
			name: "missing required",
			err:  true,
			code: codes.InvalidArgument,
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "User",
		 "type": "object",
		 "properties": {
		   "email": {
		     "type": "string"
		   },
			"firstName": {
		     "type": "string"
		   }
		 },
		 "required": [ "email", "firstName" ]
		}`,
				},
			},
			user: &users.User{},
		},
		{
			name: "missing required fixed",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "User",
		 "type": "object",
		 "properties": {
		   "email": {
		     "type": "string"
		   },
			"firstName": {
		     "type": "string"
		   }
		 },
		 "required": [ "email", "firstName" ]
		}`,
				},
			},
			user: &users.User{
				Email:     "linux@linux.com",
				FirstName: "Linus",
			},
		},
		{
			name: "min length",
			err:  true,
			code: codes.InvalidArgument,
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "User",
		 "type": "object",
		 "properties": {
		   "email": {
		     "type": "string",
		     "minLength": 3
		   },
			"firstName": {
		     "type": "string",
		     "minLength": 3
		   }
		 },
		 "required": [ "email", "firstName" ]
		}`,
				},
			},
			user: &users.User{
				Email:     "worm@gmail.com",
				FirstName: "Wo",
			},
		},
		{
			name: "min length fixed",
			spec: &crudstore.CrudEntitySpec{
				EntityType: "contracts.users.User",
				SchemaSpec: &crudstore.SchemaSpec{
					SchemaVersion: 1,
					JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "User",
		 "type": "object",
		 "properties": {
		   "email": {
		     "type": "string",
		     "minLength": 3
		   },
			"firstName": {
		     "type": "string",
		     "minLength": 3
		   }
		 },
		 "required": [ "email", "firstName" ]
		}`,
				},
			},
			user: &users.User{
				Email:     "worm@gmail.com",
				FirstName: "Worm",
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(tt *testing.T) {
			var client crudstore.CrudStoreServiceClient
			unitC := newUnitTestClient(tt)
			client = unitC
			defer unitC.cleanup()

			crudClient, err := clients.NewCrudStoreWithActiveConn(
				ctx, client,
			)
			assert.NoError(tt, err)
			assert.NotNil(tt, crudClient)

			// add the initial schema
			_, err = client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
				Spec: &crudstore.CrudEntitySpec{
					EntityType: "contracts.users.User",
				},
			})
			assert.NoError(tt, err)

			// add an empty user
			createdOriginator, err := crudClient.Create(&users.User{})
			assert.NoError(tt, err)

			// first update the schema if any
			if tc.spec != nil {
				_, err := crudClient.CrudGRPC.UpdateType(ctx, &crudstore.UpdateTypeRequest{
					Spec: tc.spec,
				})
				assert.NoError(tt, err)
			}

			// now try to update the object
			tc.user.Originator = createdOriginator
			updatedOriginator, err := crudClient.Update(tc.user)

			if !tc.err {
				assert.NoError(tt, err)
				if updatedOriginator != nil {
					retrievedUser := &users.User{}
					err := crudClient.Get(updatedOriginator, retrievedUser, false)
					assert.NoError(tt, err)
					tc.user.Originator = updatedOriginator
					assert.True(tt, proto.Equal(tc.user, retrievedUser))
					if !proto.Equal(tc.user, retrievedUser) {
						tt.Logf("The diff is : %v", deep.Equal(tc.user, retrievedUser))
					}
				}
			} else {
				assert.Error(tt, err)
				if err != nil {
					assert.NoError(tt, util.AssertGrpcCodeErr(err, tc.code))
				}
			}
		})
	}
}

func TestCrudStoreSvcProvider_ListTypes(t *testing.T) {
	ctx := context.Background()

	var client crudstore.CrudStoreServiceClient
	unitC := newUnitTestClient(t)
	client = unitC
	defer unitC.cleanup()

	crudClient, err := clients.NewCrudStoreWithActiveConn(
		ctx, client,
	)
	assert.NoError(t, err)
	assert.NotNil(t, crudClient)

	resp, err := client.ListTypes(ctx, &crudstore.ListTypesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.GetResults(), 0)

	spec1 := &crudstore.CrudEntitySpec{
		EntityType: "contracts.users.User",
		SchemaSpec: &crudstore.SchemaSpec{
			SchemaVersion: 1,
			JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "User",
		 "type": "object",
		 "properties": {
		   "email": {
		     "type": "string",
		     "minLength": 3
		   },
			"firstName": {
		     "type": "string",
		     "minLength": 3
		   }
		 },
		 "required": [ "email", "firstName" ]
		}`,
		},
	}

	_, err = client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
		Spec: spec1,
	})
	assert.NoError(t, err)

	resp, err = client.ListTypes(ctx, &crudstore.ListTypesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.GetResults(), 1)
	assert.Equal(t, resp.Results[0], spec1)

	spec2 := &crudstore.CrudEntitySpec{
		EntityType: "contracts.users.User2",
		SchemaSpec: &crudstore.SchemaSpec{
			SchemaVersion: 1,
			JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "User",
		 "type": "object",
		 "properties": {
		   "email": {
		     "type": "string",
		     "minLength": 3
		   }
		 },
		 "required": [ "email"]
		}`,
		},
	}

	_, err = client.RegisterType(ctx, &crudstore.RegisterTypeRequest{
		Spec: spec2,
	})
	assert.NoError(t, err)

	resp, err = client.ListTypes(ctx, &crudstore.ListTypesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.GetResults(), 2)
	assert.Equal(t, resp.Results[0], spec1)
	assert.Equal(t, resp.Results[1], spec2)

	spec1Updated := &crudstore.CrudEntitySpec{
		EntityType: "contracts.users.User",
		SchemaSpec: &crudstore.SchemaSpec{
			SchemaVersion: 2,
			JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "User",
		 "type": "object",
		 "properties": {}
		}`,
		},
	}
	_, err = crudClient.CrudGRPC.UpdateType(ctx, &crudstore.UpdateTypeRequest{
		Spec: spec1Updated,
	})
	assert.NoError(t, err)

	resp, err = client.ListTypes(ctx, &crudstore.ListTypesRequest{})
	assert.NoError(t, err)
	assert.Len(t, resp.GetResults(), 2)
	assert.Equal(t, resp.Results[0], spec1Updated)
	assert.Equal(t, resp.Results[1], spec2)

}

type unitClient struct {
	crudstore.CrudStoreServiceClient

	t *testing.T
}

func newUnitTestClient(t *testing.T) *unitClient {
	eventStore := eventstore2.NewInMemoryStore()
	//eventStore, err := eventstore.NewInMemoryStore("sqlite3", "estore.db")
	//assert.NoError(t, err)
	assert.NotNil(t, eventStore)

	//err = eventStore.Cleanup()
	//assert.NoError(t, err)

	ctx := context.Background()
	crudStore, err := crudstore2.NewCrudStoreProvider(ctx, eventStore)
	assert.NoError(t, err)

	api, err := NewCrudStoreApiProvider(crudStore)
	assert.NoError(t, err)
	assert.NotNil(t, api)

	client := clients.NewCrudStoreClientWithNoNetworking(api)
	c := &unitClient{
		CrudStoreServiceClient: client,
		t:                      t,
	}
	return c
}

func (u *unitClient) cleanup() {
	if _, err := os.Stat("estore.db"); err == nil {
		assert.NoError(u.t, os.Remove("estore.db"))
	}
}
