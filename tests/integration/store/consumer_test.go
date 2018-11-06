package store

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/makkalot/eskit/generated/grpc/go/consumerstore"
	"github.com/makkalot/eskit/generated/grpc/go/eventstore"
	"google.golang.org/grpc"
	"context"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/makkalot/eskit/services/clients"
	"time"
	"github.com/makkalot/eskit/tests/integration/util"
	"google.golang.org/grpc/codes"
	"io"
	"github.com/satori/go.uuid"
)

var _ = Describe("EventLog Consumer", func() {
	var storeClient eventstore.EventstoreServiceClient
	var consumerClient consumerstore.ConsumerServiceClient

	var conn *grpc.ClientConn
	var ctx context.Context

	BeforeEach(func() {
		ctx = context.Background()
		conn, err := grpc.Dial(storeEndpoint, grpc.WithInsecure())
		Expect(err).To(BeNil())

		storeClient = eventstore.NewEventstoreServiceClient(conn)
		Expect(storeClient).NotTo(BeNil())

		consumerConn, err := grpc.Dial(consumerEndpoint, grpc.WithInsecure())
		Expect(err).To(BeNil())

		consumerClient = consumerstore.NewConsumerServiceClient(consumerConn)
		Expect(consumerClient).NotTo(BeNil())
	})

	AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
	})

	Context("When start consuming from the beginning", func() {
		var firstConsumerID = uuid.Must(uuid.NewV4()).String()
		var secondConsumerID = uuid.Must(uuid.NewV4()).String()

		It("Should not have any entries for the consumer at the beginning", func() {
			_, err := consumerClient.GetLogConsume(ctx, &consumerstore.GetAppLogConsumeRequest{
				ConsumerId: firstConsumerID,
			})
			Expect(err).NotTo(BeNil())
			util.AssertGrpcCode(err, codes.NotFound)

			_, err = consumerClient.GetLogConsume(ctx, &consumerstore.GetAppLogConsumeRequest{
				ConsumerId: secondConsumerID,
			})
			Expect(err).NotTo(BeNil())
			util.AssertGrpcCode(err, codes.NotFound)
		})

		It("Should be able to save progress ", func() {

			By("saving the progress for the first consumer")
			_, err := consumerClient.LogConsume(ctx, &consumerstore.AppLogConsumeRequest{
				ConsumerId: firstConsumerID,
				Offset:     "2",
			})
			Expect(err).To(BeNil())

			By("saving the progress for the second consumer")
			_, err = consumerClient.LogConsume(ctx, &consumerstore.AppLogConsumeRequest{
				ConsumerId: secondConsumerID,
				Offset:     "3",
			})
			Expect(err).To(BeNil())

			By("retrieving the progress for the first consumer")
			resp, err := consumerClient.GetLogConsume(ctx, &consumerstore.GetAppLogConsumeRequest{
				ConsumerId: firstConsumerID,
			})
			Expect(err).To(BeNil())
			Expect(resp.Offset).To(Equal("2"))

			By("retrieving the progress for the second consumer")
			resp, err = consumerClient.GetLogConsume(ctx, &consumerstore.GetAppLogConsumeRequest{
				ConsumerId: secondConsumerID,
			})
			Expect(err).To(BeNil())
			Expect(resp.Offset).To(Equal("3"))

			By("updating the progress for the first consumer")
			_, err = consumerClient.LogConsume(ctx, &consumerstore.AppLogConsumeRequest{
				ConsumerId: firstConsumerID,
				Offset:     "10",
			})
			Expect(err).To(BeNil())

			By("updating the progress for the second consumer")
			_, err = consumerClient.LogConsume(ctx, &consumerstore.AppLogConsumeRequest{
				ConsumerId: secondConsumerID,
				Offset:     "6",
			})
			Expect(err).To(BeNil())

			By("retrieving the progress for the first consumer")
			resp, err = consumerClient.GetLogConsume(ctx, &consumerstore.GetAppLogConsumeRequest{
				ConsumerId: firstConsumerID,
			})
			Expect(err).To(BeNil())
			Expect(resp.Offset).To(Equal("10"))

			By("retrieving the progress for the second consumer")
			resp, err = consumerClient.GetLogConsume(ctx, &consumerstore.GetAppLogConsumeRequest{
				ConsumerId: secondConsumerID,
			})
			Expect(err).To(BeNil())
			Expect(resp.Offset).To(Equal("6"))

		})

	})

	Context("When starts consuming the event stream from the saved progress", func() {

		var initialized bool
		var consumerNameCB = uuid.Must(uuid.NewV4()).String()
		var consumerCtx context.Context
		var consumerCancel context.CancelFunc
		var consumer *clients.AppLogConsumer
		var entityID = uuid.Must(uuid.NewV4()).String()

		BeforeEach(func() {
			if initialized {
				return
			}

			event := &eventstore.Event{
				Originator: &common.Originator{
					Id:      entityID,
					Version: "1",
				},
				EventType: "ConsumerUser.Created",
				Payload:   "{}",
			}
			resp, err := storeClient.Append(context.Background(), &eventstore.AppendEventRequest{
				Event: event,
			})
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

			initialized = true
		})

		It("should be able to get the first event in the eventlog for the selector", func(done Done) {
			var err error
			consumerCtx, consumerCancel = context.WithCancel(ctx)
			consumer, err = clients.NewAppLogConsumer(consumerCtx, storeClient, consumerClient, consumerNameCB, clients.FromSaved, "ConsumerUser.*")
			Expect(err).To(BeNil())
			Expect(consumer).NotTo(BeNil())

			ch := make(chan struct{})
			f := func(entry *eventstore.AppLogEntry) error {

				Expect(entry.Event.EventType).To(Equal("ConsumerUser.Created"))
				Expect(entry.Event.Originator).To(Equal(&common.Originator{
					Id:      entityID,
					Version: "1",
				}, ))

				ch <- struct{}{}
				return nil
			}

			finished := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				err := consumer.Consume(f)
				GinkgoT().Logf("the consumer is stopped : %v", err)
				if err != nil && err != io.EOF {
					Expect(err).To(BeNil())
				}
				finished <- struct{}{}
			}()

			// wait for the entry to appear
			_ = <-ch
			// wait for consumer to save the progress first
			time.Sleep(time.Second * 2)

			GinkgoT().Logf("stopping the consumer now")
			consumerCancel()
			_ = <-finished
			close(done)
		}, 10)

		It("should only receive the new events", func(done Done) {
			var err error
			consumerCtx, consumerCancel = context.WithCancel(ctx)
			consumer, err = clients.NewAppLogConsumer(consumerCtx, storeClient, consumerClient, consumerNameCB, clients.FromSaved, "ConsumerUser.*")
			Expect(err).To(BeNil())
			Expect(consumer).NotTo(BeNil())

			event := &eventstore.Event{
				Originator: &common.Originator{
					Id:      entityID,
					Version: "2",
				},
				EventType: "ConsumerUser.Updated",
				Payload:   "{}",
			}
			resp, err := storeClient.Append(context.Background(), &eventstore.AppendEventRequest{
				Event: event,
			})
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

			ch := make(chan struct{})
			f := func(entry *eventstore.AppLogEntry) error {

				Expect(entry.Event.EventType).To(Equal("ConsumerUser.Updated"))
				Expect(entry.Event.Originator).To(Equal(&common.Originator{
					Id:      entityID,
					Version: "2",
				}, ))

				ch <- struct{}{}
				return nil
			}

			finished := make(chan struct{})
			go func() {
				defer GinkgoRecover()
				err := consumer.Consume(f)
				GinkgoT().Logf("the consumer is stopped : %v", err)
				if err != nil && err != io.EOF {
					Expect(err).To(BeNil())
				}
				finished <- struct{}{}
			}()

			// wait for the entry to appear
			_ = <-ch
			// wait for consumer to save the progress first
			time.Sleep(time.Second * 2)

			consumerCancel()
			_ = <-finished
			close(done)
		}, 10)
	})
})
