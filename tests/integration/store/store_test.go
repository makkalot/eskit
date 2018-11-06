package store

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	"github.com/makkalot/eskit/generated/grpc/go/crudstore"
	"google.golang.org/grpc"
	"context"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	"github.com/satori/go.uuid"
	"github.com/davecgh/go-spew/spew"
	"io"
	"fmt"
	"time"
	"strings"
	"github.com/golang/protobuf/proto"
	"github.com/makkalot/eskit/tests/integration/util"
)

var _ = Describe("Event Store", func() {

	var storeClient store.EventstoreServiceClient
	var crudStoreClient crudstore.CrudStoreServiceClient
	var conn *grpc.ClientConn

	BeforeEach(func() {
		conn, err := grpc.Dial(storeEndpoint, grpc.WithInsecure())
		Expect(err).To(BeNil())

		storeClient = store.NewEventstoreServiceClient(conn)
		Expect(storeClient).NotTo(BeNil())

		crudStoreClient = crudstore.NewCrudStoreServiceClient(conn)
		Expect(crudStoreClient).NotTo(BeNil())

		resp, err := storeClient.Healtz(context.Background(), &store.HealthRequest{})
		Expect(err).To(BeNil())
		Expect(resp).NotTo(BeNil())

	})

	AfterEach(func() {
		if conn != nil {
			conn.Close()
		}
	})

	Context("When Append a new event", func() {
		It("Should be added to the list of events", func() {
			entityID, err := uuid.NewV4()
			Expect(err).To(BeNil())

			event := &store.Event{
				Originator: &common.Originator{
					Id:      entityID.String(),
					Version: "1",
				},
				EventType: "User.Created",
				Payload:   "{}",
			}
			resp, err := storeClient.Append(context.Background(), &store.AppendEventRequest{
				Event: event,
			})
			Expect(err).To(BeNil())
			Expect(resp).NotTo(BeNil())

			events, err := storeClient.GetEvents(context.Background(), &store.GetEventsRequest{
				Originator: &common.Originator{
					Id: entityID.String(),
				},
				EntityType: "User",
			})

			Expect(err).To(BeNil())
			Expect(events).NotTo(BeNil())
			util.AssertContainsEvent(event, events.Events)

			logEvents, err := storeClient.Logs(context.Background(), &store.AppLogRequest{
				Selector: "User.*",
			})
			Expect(err).To(BeNil())
			Expect(logEvents.Results).NotTo(BeNil())
			util.AssertContainsEventLogEntry(event, logEvents.Results)
		})
	})

	Context("When want to poll for incoming events from the beginning", func() {
		var stream store.EventstoreService_LogsPollClient
		var recvChan chan *store.AppLogEntry
		var quit chan struct{}
		var eventEntityID string

		BeforeEach(func() {
			if stream == nil {

				entityID, err := uuid.NewV4()
				Expect(err).To(BeNil())
				eventEntityID = entityID.String()

				recvChan = make(chan *store.AppLogEntry)

				stream, err = storeClient.LogsPoll(context.Background(), &store.AppLogRequest{
					Selector: "UserLog.*",
				})
				Expect(err).To(BeNil())
				Expect(stream).NotTo(BeNil())

				go func() {
					defer func() {
						close(recvChan)
					}()
					defer GinkgoRecover()

					for {
						select {
						case <-quit:
							GinkgoT().Logf("quit was called")
							return
						default:
						}

						GinkgoT().Logf("waiting on the stream")
						entry, err := stream.Recv()
						if err == io.EOF {
							GinkgoT().Logf("EOF encountered quiting")
							return
						}

						if err != nil {
							Fail(fmt.Sprintf("Poll Failed : %v", err))
						}

						if strings.Split(entry.Event.EventType, ".")[0] != "UserLog" {
							GinkgoT().Logf("encountered non UserLog event skipping : %s", spew.Sdump(entry.Event))
							continue
						}

						GinkgoT().Logf("sending entry to the recvChan : %s", spew.Sdump(entry))
						recvChan <- entry
					}
				}()
			}
		})

		Context("When add the first event", func() {
			var event *store.Event
			BeforeEach(func() {
				event = &store.Event{
					Originator: &common.Originator{
						Id:      eventEntityID,
						Version: "1",
					},
					EventType: "UserLog.Created",
					Payload:   "{}",
				}
				resp, err := storeClient.Append(context.Background(), &store.AppendEventRequest{
					Event: event,
				})
				Expect(err).To(BeNil())
				Expect(resp).NotTo(BeNil())

				events, err := storeClient.GetEvents(context.Background(), &store.GetEventsRequest{
					Originator: &common.Originator{
						Id: eventEntityID,
					},
					EntityType: "UserLog",
				})

				Expect(err).To(BeNil())
				Expect(events).NotTo(BeNil())
				Expect(len(events.Events)).To(Equal(1))
				Expect(proto.Equal(events.Events[0], event)).To(BeTrue(), "getevents : %v, event : %v", events.Events[0], event)

			})
			It("Should appear in the stream for first event", func(done Done) {
				select {
				case res := <-recvChan:
					GinkgoT().Logf("received first entry from chan : %s", res)

					Expect(proto.Equal(res.Event, event)).To(BeTrue(), "resEvent : %s, event : %s", res.Event, event)
				case <-time.After(5 * time.Second):
					Fail("timeout on first event")
				}

				close(done)
			}, 5)
		})

		Context("When add the second event", func() {
			var event *store.Event

			BeforeEach(func() {
				event = &store.Event{
					Originator: &common.Originator{
						Id:      eventEntityID,
						Version: "2",
					},
					EventType: "UserLog.Updated",
					Payload:   "{}",
				}
				resp, err := storeClient.Append(context.Background(), &store.AppendEventRequest{
					Event: event,
				})
				Expect(err).To(BeNil())
				Expect(resp).NotTo(BeNil())

				events, err := storeClient.GetEvents(context.Background(), &store.GetEventsRequest{
					Originator: &common.Originator{
						Id: eventEntityID,
					},
					EntityType: "UserLog",
				})

				Expect(err).To(BeNil())
				Expect(events).NotTo(BeNil())
				Expect(len(events.Events)).To(Equal(2))
				Expect(proto.Equal(events.Events[1], event)).To(BeTrue(), "getevents[1] : %s, event : %s", events.Events[1], event)
			})

			It("Should appear in the stream for the second event", func(done Done) {
				select {
				case res := <-recvChan:
					GinkgoT().Logf("received second entry from chan : %s", res)
					Expect(proto.Equal(res.Event, event)).To(BeTrue(), "res.Event : %s, event : %s", res.Event, event)
				case <-time.After(5 * time.Second):
					Fail("timeout on second event")
				}

				go func() {
					quit <- struct{}{}
				}()
				close(done)
			}, 5)
		})
	})

})
