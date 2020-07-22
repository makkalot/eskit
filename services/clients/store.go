package clients

import (
	"context"
	"fmt"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	common2 "github.com/makkalot/eskit/services/lib/common"
	"google.golang.org/grpc"
)

// Creates a new EventstoreServiceClient but first waits for health endpoint to become ready
func NewStoreClientWithWait(ctx context.Context, storeEndpoint string) (store.EventstoreServiceClient, error) {
	var conn *grpc.ClientConn
	var storeClient store.EventstoreServiceClient

	err := common2.RetryNormal(func() error {
		var err error
		conn, err = grpc.Dial(storeEndpoint, grpc.WithInsecure())
		if err != nil {
			return err
		}

		storeClient = store.NewEventstoreServiceClient(conn)
		_, err = storeClient.Healtz(ctx, &store.HealthRequest{})
		if err != nil {
			return err
		}

		return nil
	})
	return storeClient, err
}

type eventStoreClientWithNoNetworking struct {
	server store.EventstoreServiceServer
}

// NewEventStoreServiceClientWithNoNetworking is useful for unittesting where you can embed the server directly inside the
// client and can do all kinds of tests without having to worry about the networking bit and spinning up servers
func NewEventStoreServiceClientWithNoNetworking(server store.EventstoreServiceServer) store.EventstoreServiceClient {
	return &eventStoreClientWithNoNetworking{
		server: server,
	}
}

func (c *eventStoreClientWithNoNetworking) Healtz(ctx context.Context, in *store.HealthRequest, opts ...grpc.CallOption) (*store.HealthResponse, error) {
	return c.server.Healtz(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) Append(ctx context.Context, in *store.AppendEventRequest, opts ...grpc.CallOption) (*store.AppendEventResponse, error) {
	return c.server.Append(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) GetEvents(ctx context.Context, in *store.GetEventsRequest, opts ...grpc.CallOption) (*store.GetEventsResponse, error) {
	return c.server.GetEvents(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) Logs(ctx context.Context, in *store.AppLogRequest, opts ...grpc.CallOption) (*store.AppLogResponse, error) {
	return c.server.Logs(ctx, in)
}

func (c *eventStoreClientWithNoNetworking) LogsPoll(ctx context.Context, in *store.AppLogRequest, opts ...grpc.CallOption) (store.EventstoreService_LogsPollClient, error) {
	return nil, fmt.Errorf("not implemented")
}
