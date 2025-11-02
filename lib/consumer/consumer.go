package consumer

import (
	"context"
	"errors"
	"fmt"
	"github.com/makkalot/eskit/lib/common"
	"github.com/makkalot/eskit/lib/consumerstore"
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/eventstore"
	"github.com/makkalot/eskit/lib/types"
	"io"
	"log"
	"time"
)

// TODO: (Future) maybe add some RAFT for high availability !!!

type LogOffset int

type AppLogConsumer struct {
	name          string
	offset        LogOffset
	consumerStore consumerstore.Store
	storeClient   eventstore.Store
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

type ConsumeCB func(entry *types.AppLogEntry) error
type ConsumeCrudCb func(entityType string, oldMessage, newMessage interface{})

func NewAppLogConsumer(storeClient eventstore.Store, consumerStore consumerstore.Store, name string, offset LogOffset, selector string) (*AppLogConsumer, error) {
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
				return consumer.SaveProgress(ctx, entry.ID)
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

func (consumer *AppLogConsumer) Stream(ctx context.Context) (chan *types.AppLogEntry, chan error, error) {
	var fromID uint64
	if consumer.offset == FromBeginning {
		fromID = 1
	} else if consumer.offset == FromSaved {
		resp, err := consumer.consumerStore.GetLogConsume(
			ctx,
			consumer.name,
		)
		if err != nil {
			if !errors.Is(err, crudstore.RecordNotFound) {
				return nil, nil, err
			}
			fromID = 1
		} else {
			fromID = resp.Offset + 1
		}

		log.Println("starting the consuming from offset : ", fromID)
	} else {
		return nil, nil, fmt.Errorf("invalid offset supplied")
	}

	ch := make(chan *types.AppLogEntry)
	chErr := make(chan error)
	lastIDInt := fromID

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

			results, err := consumer.storeClient.Logs(lastIDInt, 10, "")
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

				nextID := results[len(results)-1].ID
				lastIDInt = nextID + 1
			}
		}

	}()

	return ch, chErr, nil
}

func (consumer *AppLogConsumer) SaveProgress(ctx context.Context, offset uint64) error {
	err := consumer.consumerStore.LogConsume(ctx, &consumerstore.AppLogConsumeProgress{
		ConsumerId: consumer.name,
		Offset:     offset,
	})

	if err != nil {
		return err
	}

	return nil
}
