package consumer

import (
	"context"
	"fmt"
	"time"
	"github.com/davecgh/go-spew/spew"
	"github.com/makkalot/eskit/lib/types"
	"github.com/makkalot/eskit/lib/consumerstore"
	"github.com/makkalot/eskit/lib/eventstore"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewAppLogConsumer(t *testing.T) {
	consumerStore := consumerstore.NewInMemoryConsumerApiProvider()
	estore := eventstore.NewInMemoryStore()

	e1 := &types.Event{
		Originator: &types.Originator{
			ID:      "originator1",
			Version: "1",
		},
		EventType:  "User.Created",
		Payload:    "{}",
		OccurredOn: time.Now().UTC(),
	}

	err := estore.Append(e1)
	assert.NoError(t, err)

	ctx, cancel := context.WithCancel(context.Background())
	consumer, err := NewAppLogConsumer(estore, consumerStore, "users-consumer", FromBeginning, "*")
	assert.NoError(t, err)
	err = consumer.Consume(ctx, func(entry *types.AppLogEntry) error {
		t.Logf("inside consume : %s", spew.Sdump(entry))
		cancel()
		return nil
	})
	assert.NoError(t, err)
}

func TestNewAppLogConsumerStop(t *testing.T) {
	consumerStore := consumerstore.NewInMemoryConsumerApiProvider()
	estore := eventstore.NewInMemoryStore()

	e1 := &types.Event{
		Originator: &types.Originator{
			ID:      "originator1",
			Version: "1",
		},
		EventType:  "User.Created",
		Payload:    "{}",
		OccurredOn: time.Now().UTC(),
	}

	err := estore.Append(e1)
	assert.NoError(t, err)

	ctx := context.Background()
	consumer, err := NewAppLogConsumer(estore, consumerStore, "users-consumer", FromBeginning, "*")
	assert.NoError(t, err)
	err = consumer.Consume(ctx, func(entry *types.AppLogEntry) error {
		t.Logf("inside consumer pre exit : %s", spew.Sdump(entry))
		return fmt.Errorf("exit %w", StopConsumerError)
	})
	assert.ErrorIs(t, err, StopConsumerError)
}

func TestNewAppLogConsumerProgress(t *testing.T) {
	consumerStore := consumerstore.NewInMemoryConsumerApiProvider()
	estore := eventstore.NewInMemoryStore()

	e1 := &types.Event{
		Originator: &types.Originator{
			ID:      "originator1",
			Version: "1",
		},
		EventType:  "User.Created",
		Payload:    "{}",
		OccurredOn: time.Now().UTC(),
	}

	err := estore.Append(e1)
	assert.NoError(t, err)


	ctx, cancel := context.WithCancel(context.Background())
	consumer, err := NewAppLogConsumer(estore, consumerStore, "users-consumer", FromSaved, "*")
	assert.NoError(t, err)
	err = consumer.Consume(ctx, func(entry *types.AppLogEntry) error {
		t.Logf("inside consumer pre exit : %s", spew.Sdump(entry))
		cancel()
		return nil
	})
	assert.NoError(t, err)

	e2 := &types.Event{
		Originator: &types.Originator{
			ID:      "originator1",
			Version: "2",
		},
		EventType:  "User.Updated",
		Payload:    `{"name":"makkalot"}`,
		OccurredOn: time.Now().UTC(),
	}

	err = estore.Append(e2)
	assert.NoError(t, err)


	ctx, cancel = context.WithCancel(context.Background())
	err = consumer.Consume(ctx, func(entry *types.AppLogEntry) error {
		t.Logf("inside consumer pre exit : %s", spew.Sdump(entry))
		cancel()
		return nil
	})
	assert.NoError(t, err)
}
