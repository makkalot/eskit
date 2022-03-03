package clients

import (
	"context"
	"github.com/davecgh/go-spew/spew"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	"github.com/makkalot/eskit/services/lib/consumerstore"
	"github.com/makkalot/eskit/services/lib/eventstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAppLogConsumer(t *testing.T) {
	consumerStore := consumerstore.NewInMemoryConsumerApiProvider()
	estore := eventstore.NewInMemoryStore()

	e1 := &store.Event{
		Originator: &common.Originator{
			Id:      "originator1",
			Version: "1",
		},
		EventType: "User.Created",
		Payload:   "{}",
	}

	err := estore.Append(e1)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	consumer, err := NewAppLogConsumer(ctx, estore, consumerStore, "users-consumer", FromBeginning, "*")
	assert.NoError(t, err)
	err = consumer.Consume(func(entry *store.AppLogEntry) error {
		t.Logf("inside consume : %s", spew.Sdump(entry))
		cancel()
		return nil
	})
	assert.NoError(t, err)
}