package store

import (
	"context"
	"github.com/golang/protobuf/proto"
	common3 "github.com/makkalot/eskit/services/lib/common"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/satori/go.uuid"
	"google.golang.org/grpc"

	"fmt"
	"github.com/davecgh/go-spew/spew"
	"github.com/go-test/deep"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/crudstore"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"github.com/makkalot/eskit/services/clients"
	"github.com/makkalot/eskit/tests/integration/util"
	"google.golang.org/grpc/codes"
)

var _ = Describe("Crud Store", func() {
	var crudStoreClient store.CrudStoreServiceClient
	var conn *grpc.ClientConn
	var ctx context.Context

	BeforeEach(func() {
		conn, err := grpc.Dial(crudStoreEndpoint, grpc.WithInsecure())
		Expect(err).To(BeNil())

		crudStoreClient = store.NewCrudStoreServiceClient(conn)
		Expect(crudStoreClient).NotTo(BeNil())

		ctx = context.Background()
		resp, err := crudStoreClient.Healtz(ctx, &store.HealthRequest{})
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())

	})

	AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
	})

	Context("Type Registration", func() {

		It("Should be able to register a new type", func() {
			waterEntity := "com.makkalot.types.Water"
			_, err := crudStoreClient.RegisterType(ctx, &store.RegisterTypeRequest{
				Spec: &store.CrudEntitySpec{
					EntityType: waterEntity,
				},
			})
			Expect(err).To(BeNil())

			bottleEntity := "com.makkalot.types.Bottle"
			_, err = crudStoreClient.RegisterType(ctx, &store.RegisterTypeRequest{
				Spec: &store.CrudEntitySpec{
					EntityType: bottleEntity,
				},
			})
			Expect(err).To(BeNil())

			err = common3.RetryNormal(func() error {
				types, err := crudStoreClient.ListTypes(ctx, &store.ListTypesRequest{})
				if err != nil {
					return err
				}

				if len(types.Results) < 2 {
					return fmt.Errorf("expected 2 got %d", len(types.Results))
				}

				expectedEntities := map[string]bool{
					waterEntity:  true,
					bottleEntity: true,
				}

				gotEntities := map[string]bool{}

				for _, spec := range types.Results {
					gotEntities[spec.EntityType] = true
				}

				for k, v := range expectedEntities {
					if gotEntities[k] != v {
						Fail(fmt.Sprintf("entity type : %s was not found in response", k))
					}
				}

				return nil
			})
			Expect(err).To(BeNil())

			// try to register the same one to get a duplicate error
			_, err = crudStoreClient.RegisterType(ctx, &store.RegisterTypeRequest{
				Spec: &store.CrudEntitySpec{
					EntityType: bottleEntity,
				},
			})
			Expect(err).NotTo(BeNil())

			_, err = crudStoreClient.RegisterType(ctx, &store.RegisterTypeRequest{
				Spec: &store.CrudEntitySpec{
					EntityType: bottleEntity,
				},
				SkipDuplicate: true,
			})
			Expect(err).To(BeNil())

		})

		It("Should be able to update an existing type", func() {

			waterEntity := "com.makkalot.types.UpdateWater"
			_, err := crudStoreClient.RegisterType(ctx, &store.RegisterTypeRequest{
				Spec: &store.CrudEntitySpec{
					EntityType: waterEntity,
				},
			})
			Expect(err).To(BeNil())

			err = common3.RetryNormal(func() error {
				types, err := crudStoreClient.ListTypes(ctx, &store.ListTypesRequest{})
				if err != nil {
					return err
				}

				if len(types.Results) < 2 {
					return fmt.Errorf("expected 2 got %d", len(types.Results))
				}

				expectedEntities := map[string]bool{
					waterEntity: true,
				}

				gotEntities := map[string]bool{}

				for _, spec := range types.Results {
					gotEntities[spec.EntityType] = true
				}

				GinkgoT().Logf("All registered entities so far are : %v", gotEntities)

				for k, v := range expectedEntities {
					if gotEntities[k] != v {
						return fmt.Errorf("entity type : %s was not found in response", k)
					}
				}

				return nil
			})
			Expect(err).To(BeNil())

			crudSpec := &store.CrudEntitySpec{
				EntityType: waterEntity,
				SchemaSpec: &store.SchemaSpec{
					SchemaVersion: 2,
					JsonSchema: `{
		 "$schema": "http://json-schema.org/draft-07/schema#",
		 "$id": "http://example.com/product.schema.json",
		 "title": "Water",
		 "type": "object",
		 "properties": {}
		}`,
				},
			}
			_, err = crudStoreClient.UpdateType(ctx, &store.UpdateTypeRequest{
				Spec: crudSpec,
			})

			err = common3.RetryNormal(func() error {
				GinkgoT().Logf("pulling the types for updated indexes")
				types, err := crudStoreClient.ListTypes(ctx, &store.ListTypesRequest{})
				if err != nil {
					return err
				}

				if len(types.Results) < 2 {
					return fmt.Errorf("expected 2 got %d", len(types.Results))
				}

				expectedEntities := map[string]bool{
					waterEntity: true,
				}

				gotEntities := map[string]bool{}

				var waterSpec *store.CrudEntitySpec
				for _, spec := range types.Results {
					gotEntities[spec.EntityType] = true
					if spec.EntityType == waterEntity {
						waterSpec = spec
					}
				}

				for k, v := range expectedEntities {
					if gotEntities[k] != v {
						Fail(fmt.Sprintf("entity type : %s was not found in response", k))
					}
				}

				if !proto.Equal(waterSpec, crudSpec) {
					if diff := deep.Equal(waterSpec, crudSpec); diff != nil {
						return fmt.Errorf("schema was not updated : %v", diff)
					}
				}

				return nil
			})
			Expect(err).To(BeNil())

		})

	})

	Context("When create a new entity via crud store", func() {

		var user *users.User
		var originator *common.Originator
		var ctx context.Context
		var crudClient *clients.CrudStoreClient

		BeforeEach(func() {
			ctx = context.Background()

			if crudClient == nil {
				var err error
				crudClient, err = clients.NewCrudStoreWithActiveConn(
					ctx, crudStoreClient,
				)
				Expect(err).To(BeNil())
			}

			u1 := uuid.Must(uuid.NewV4())
			user = &users.User{
				Originator: &common.Originator{
					Id: u1.String(),
				},
				Email:     "delidumrul@gmail.com",
				FirstName: "Deli",
				LastName:  "Dumrul",
				Active:    true,
			}

			var err error
			originator, err = crudClient.Create(user)
			Expect(err).To(BeNil())
			user.Originator = originator
			GinkgoT().Logf("originator of the created user is : %s", spew.Sdump(originator))

		})

		It("Should be able to retrieve the created item via .Get method and originator", func() {
			retrievedUser := &users.User{}
			err := crudClient.Get(originator, retrievedUser, false)
			Expect(err).To(BeNil())
			Expect(proto.Equal(user, retrievedUser)).To(BeTrue(), "user : %s, retrievedUser : %s", user, retrievedUser)

		})

		It("Should be able to update it via .Update", func() {
			retrievedUser := &users.User{}
			err := crudClient.Get(originator, retrievedUser, false)
			Expect(err).To(BeNil())
			retrievedUser.Email = "chnaged@gmail.com"

			//GinkgoT().Logf("The retrieved user is : %s", spew.Sdump(retrievedUser))

			updatedOriginator, err := crudClient.Update(retrievedUser)
			Expect(err).To(BeNil())

			updatedUser := &users.User{}
			err = crudClient.Get(updatedOriginator, updatedUser, false)
			Expect(err).To(BeNil())
			Expect(retrievedUser.Email).To(Equal(updatedUser.Email))

			// check if can retrieve the prev version as well
			prevUser := &users.User{}
			err = crudClient.Get(originator, prevUser, false)
			Expect(err).To(BeNil())
			Expect(prevUser.Email).To(Equal(user.Email))

			// when query without the version should retrieve the latest one
			versionlessOriginator := &common.Originator{
				Id: originator.Id,
			}

			latestUser := &users.User{}
			err = crudClient.Get(versionlessOriginator, latestUser, false)
			Expect(err).To(BeNil())
			Expect(proto.Equal(updatedUser, latestUser)).To(BeTrue(), "updatedUser : %s, latestUser : %s", updatedUser, latestUser)

		})

		It("Should be able to delete it via .Delete", func() {
			deleteOriginator, err := crudClient.Delete(originator, &users.User{})
			Expect(err).To(BeNil())
			Expect(deleteOriginator).NotTo(BeNil())

			versionlessOriginator := &common.Originator{
				Id: originator.Id,
			}

			err = crudClient.Get(deleteOriginator, &users.User{}, false)
			Expect(err).NotTo(BeNil())
			util.AssertGrpcCode(err, codes.NotFound)

			err = crudClient.Get(versionlessOriginator, &users.User{}, false)
			Expect(err).NotTo(BeNil())
			util.AssertGrpcCode(err, codes.NotFound)

			// try to get em via delete = true
			err = crudClient.Get(deleteOriginator, &users.User{}, true)
			Expect(err).To(BeNil())

			err = crudClient.Get(versionlessOriginator, &users.User{}, true)
			Expect(err).To(BeNil())
		})
	})
})
