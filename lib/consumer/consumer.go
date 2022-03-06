package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/makkalot/eskit/generated/grpc/go/eventstore"
	"github.com/makkalot/eskit/lib/common"
	"github.com/makkalot/eskit/lib/consumerstore"
	"github.com/makkalot/eskit/lib/crudstore"
	eventstore2 "github.com/makkalot/eskit/lib/eventstore"
	"io"
	"log"
	"strconv"
	"time"
)

// TODO: (Future) maybe add some RAFT for high availability !!!

type LogOffset int

type AppLogConsumer struct {
	name          string
	offset        LogOffset
	consumerStore consumerstore.Store
	storeClient   eventstore2.Store
	selector      string
}

const (
	FromBeginning LogOffset = 1
	FromSaved               = 2
)

var (
	// FatalConsumerError is raised from internal loop of consumer
	FatalConsumerError = errors.New("fatal consumer error")
	// StopConsumerError is raised from CB so the consumer can stop otherwise the error is ignored
	StopConsumerError = errors.New("stop consumer error")
)

type ConsumeCB func(entry *eventstore.AppLogEntry) error
type ConsumeCrudCb func(entityType string, oldMessage, newMessage interface{})

func NewAppLogConsumer(storeClient eventstore2.Store, consumerStore consumerstore.Store, name string, offset LogOffset, selector string) (*AppLogConsumer, error) {
	return &AppLogConsumer{
		name:          name,
		offset:        offset,
		consumerStore: consumerStore,
		storeClient:   storeClient,
		selector:      selector,
	}, nil
}

// Consume starts consuming entries on cb
// success the offset is saved to the server so on crash continues
func (consumer *AppLogConsumer) Consume(ctx context.Context, cb ConsumeCB) error {
	ch, chErr, err := consumer.Stream(ctx)
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

			if err := common.RetryShort(func() error {
				return consumer.SaveProgress(ctx, entry.Id)
			}); err != nil {
				return err
			}

		case err := <-chErr:
			if errors.Is(err, context.Canceled){
				return nil
			}

			if errors.Is(err, FatalConsumerError) || errors.Is(err, StopConsumerError){
				return err
			}

			// do nothing here just continue
		}
	}
}

func (consumer *AppLogConsumer) Stream(ctx context.Context) (chan *eventstore.AppLogEntry, chan error, error) {
	req := &eventstore.AppLogRequest{}
	if consumer.offset == FromBeginning {
		req.FromId = "1"
	} else if consumer.offset == FromSaved {
		resp, err := consumer.consumerStore.GetLogConsume(
			ctx,
			consumer.name,
		)
		if err != nil {
			if !errors.Is(err, crudstore.RecordNotFound) {
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

	offsetInt, err := strconv.ParseInt(req.FromId, 10, 64)
	if err != nil {
		return nil, nil, err
	}

	ch := make(chan *eventstore.AppLogEntry)
	chErr := make(chan error)
	lastIDInt := offsetInt

	go func() {
		defer func() {
			close(ch)
			close(chErr)
		}()

		for {

			// check if it was done
			select {
			case <- ctx.Done():
				chErr <- ctx.Err()
			default:

			}

			results, err := consumer.storeClient.Logs(uint64(lastIDInt), 10, "")
			if err != nil {
				chErr <- fmt.Errorf("fetch logs : %v", err)
				return
			}

			if results == nil || len(results) == 0 {
				time.Sleep(time.Millisecond * 100)
				continue
			}

			for _, r := range results {
				if common.IsEventCompliant(r.Event, consumer.selector) {
					ch <- r
				}

				nextID := results[len(results)-1].Id
				nextIDInt, err := strconv.ParseUint(nextID, 10, 64)
				if err != nil {
					chErr <- fmt.Errorf("invalid fromID : %v", err)
				}

				nextIDInt++
				lastIDInt = int64(nextIDInt)
			}
		}

	}()

	return ch, chErr, nil
}

func (consumer *AppLogConsumer) SaveProgress(ctx context.Context, offset string) error {
	err := consumer.consumerStore.LogConsume(ctx, &consumerstore.AppLogConsumeProgress{
		ConsumerId: consumer.name,
		Offset:     offset,
	})

	if err != nil {
		return err
	}

	return nil
}
