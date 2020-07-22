package clients

import (
	"context"
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/consumerstore"
	"github.com/makkalot/eskit/generated/grpc/go/eventstore"
	common2 "github.com/makkalot/eskit/services/lib/common"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"io"
	"log"
	"strconv"
	"strings"
)

// TODO: (Future) maybe add some RAFT for high availability !!!

type LogOffset int

type AppLogConsumer struct {
	name         string
	offset       LogOffset
	consumerGRPC consumerstore.ConsumerServiceClient
	storeClient  eventstore.EventstoreServiceClient
	ctx          context.Context
	selector     string
}

const (
	FromBeginning LogOffset = 1
	FromSaved               = 2
)

type ConsumeCB func(entry *eventstore.AppLogEntry) error

func NewAppLogConsumer(ctx context.Context, storeClientGRPC eventstore.EventstoreServiceClient, consumerGRPC consumerstore.ConsumerServiceClient, name string, offset LogOffset, selector string) (*AppLogConsumer, error) {
	return &AppLogConsumer{
		name:         name,
		offset:       offset,
		consumerGRPC: consumerGRPC,
		storeClient:  storeClientGRPC,
		ctx:          ctx,
		selector:     selector,
	}, nil
}

// Consume starts consuming entries on cb
// success the offset is saved to the server so on crash continues
func (consumer *AppLogConsumer) Consume(cb ConsumeCB) error {

	ch, chErr, err := consumer.Stream()
	if err != nil {
		return err
	}

	for {
		select {
		case entry, ok := <-ch:
			if !ok {
				return io.EOF
			}
			if err := cb(entry); err != nil {
				return err
			}

			eventType := entry.Event.EventType
			parts := strings.Split(eventType, ".")
			if strings.Join(parts[:len(parts)-1], ".") != "LogConsumer" {
				//log.Println("Saving progress for : ", entry.Event.EventType)
				//log.Println("Saving progress for : ", entry.Event.Originator)
				//log.Println("Saving progress for : ", entry.Id)

				if err := common2.RetryShort(func() error {
					return consumer.SaveProgress(entry.Id)
				}); err != nil {
					return err
				}

			}

		case err := <-chErr:
			return err
		}
	}
	return nil
}

func (consumer *AppLogConsumer) Stream() (chan *eventstore.AppLogEntry, chan error, error) {
	req := &eventstore.AppLogRequest{}
	if consumer.offset == FromBeginning {
		req.FromId = "1"
	} else if consumer.offset == FromSaved {
		resp, err := consumer.consumerGRPC.GetLogConsume(
			consumer.ctx,
			&consumerstore.GetAppLogConsumeRequest{
				ConsumerId: consumer.name,
			},
		)
		if err != nil {
			if status.Code(err) != codes.NotFound {
				return nil, nil, err
			}
			req.FromId = "1"
		} else {
			offsetInt, err := strconv.ParseInt(resp.Offset, 10, 64)
			if err != nil {
				return nil, nil, err
			}

			// we want to start from the next id to skip the last saved so
			// we don't process duplicates again
			offsetInt += 1
			req.FromId = strconv.Itoa(int(offsetInt))
		}

		log.Println("starting the consuming from offset : ", req.FromId)
	} else {
		return nil, nil, fmt.Errorf("invalid offset supplied")
	}

	if consumer.selector != "*" && consumer.selector != "" {
		req.Selector = consumer.selector
	}

	stream, err := consumer.storeClient.LogsPoll(consumer.ctx, req)
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan *eventstore.AppLogEntry)
	chErr := make(chan error)

	go func() {
		defer func() {
			close(ch)
			close(chErr)
		}()

		for {
			entry, err := stream.Recv()
			if err != nil {
				if err == io.EOF || status.Code(err) == codes.Canceled {
					return
				}

				log.Println("sending error to errCH : ", err)
				chErr <- err
				return
			}

			//log.Println("retrieved a new entry in the consumer : ", spew.Sdump(entry))
			ch <- entry
		}
	}()

	return ch, chErr, nil
}

func (consumer *AppLogConsumer) SaveProgress(offset string) error {
	_, err := consumer.consumerGRPC.LogConsume(consumer.ctx, &consumerstore.AppLogConsumeRequest{
		ConsumerId: consumer.name,
		Offset:     offset,
	})

	if err != nil {
		return err
	}

	return nil
}

func NewConsumerStoreGRPCClient(ctx context.Context, storeEndpoint string) (consumerstore.ConsumerServiceClient, error) {
	conn, err := grpc.Dial(storeEndpoint, grpc.WithInsecure())
	if err != nil {
		return nil, err
	}

	return consumerstore.NewConsumerServiceClient(conn), nil

}

func NewConsumerStoreGrpcClientWithWait(ctx context.Context, storeEndpoint string) (consumerstore.ConsumerServiceClient, error) {
	var conn *grpc.ClientConn
	var storeClient consumerstore.ConsumerServiceClient

	err := common2.RetryNormal(func() error {
		var err error
		conn, err = grpc.Dial(storeEndpoint, grpc.WithInsecure())
		if err != nil {
			return err
		}

		storeClient = consumerstore.NewConsumerServiceClient(conn)
		_, err = storeClient.Healtz(ctx, &consumerstore.HealthRequest{})
		if err != nil {
			return err
		}

		return nil
	})
	return storeClient, err
}
