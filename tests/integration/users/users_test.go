package users_test

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"google.golang.org/grpc"
	"context"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/tests/integration/util"
	"google.golang.org/grpc/codes"
)

var _ = Describe("Users", func() {

	var userClient users.UserServiceClient
	var conn *grpc.ClientConn

	BeforeEach(func() {
		conn, err := grpc.Dial(userEndpoint, grpc.WithInsecure())
		Expect(err).To(BeNil())

		userClient = users.NewUserServiceClient(conn)
		Expect(userClient).NotTo(BeNil())
	})

	AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
	})

	Context("When User is Created", func() {
		It("Should be returned via its ID", func() {

			req := &users.CreateRequest{
				Email:     "ahmet@ahmet.com",
				FirstName: "SameName",
				LastName:  "Eskit",
			}
			resp, err := userClient.Create(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

			getResp, err := userClient.Get(context.Background(), &users.GetRequest{
				Originator: &common.Originator{
					Id: resp.User.Originator.Id,
				},
			})
			Expect(err).To(BeNil())
			Expect(getResp.User).NotTo(BeNil())

			Expect(getResp.User.Email).To(Equal(req.Email))
			Expect(getResp.User.FirstName).To(Equal(req.FirstName))
			Expect(getResp.User.LastName).To(Equal(req.LastName))

			req2 := &users.CreateRequest{
				Email:     "osman@osman.com",
				FirstName: "SameName",
				LastName:  "Eskit",
			}
			resp2, err := userClient.Create(context.Background(), req2)
			Expect(err).To(BeNil())
			Expect(resp2).NotTo(BeNil())

			getResp, err = userClient.Get(context.Background(), &users.GetRequest{
				Originator: &common.Originator{
					Id: resp2.User.Originator.Id,
				},
			})
			Expect(err).To(BeNil())
			Expect(getResp.User).NotTo(BeNil())

			Expect(getResp.User.Email).To(Equal(req2.Email))
			Expect(getResp.User.FirstName).To(Equal(req2.FirstName))
			Expect(getResp.User.LastName).To(Equal(req2.LastName))

		})
	})

	Context("When User is Deleted", func() {
		var deletedUserID string
		var runOnce bool

		BeforeEach(func() {
			if runOnce {
				return
			}
			req := &users.CreateRequest{
				Email:     "eskit@gmail.com",
				FirstName: "Ahmet",
				LastName:  "Abi",
			}
			resp, err := userClient.Create(context.Background(), req)

			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())
			deletedUserID = resp.User.Originator.Id

			userClient.Delete(context.Background(), &users.DeleteRequest{
				Originator: &common.Originator{
					Id: deletedUserID,
				},
			})
			runOnce = true
		})

		It("Should not be returned via its ID", func() {
			_, err := userClient.Get(context.Background(), &users.GetRequest{
				Originator: &common.Originator{
					Id: deletedUserID,
				},
			})
			Expect(err).NotTo(BeNil())
			util.AssertGrpcCode(err, codes.NotFound)

		})
	})
})
