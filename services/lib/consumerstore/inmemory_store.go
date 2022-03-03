package consumerstore

import (
	"context"
	"fmt"
	"github.com/makkalot/eskit/services/lib/crudstore"
)

type InMemoryConsumerApiProvider struct {
	progress map[string]string
}

func NewInMemoryConsumerApiProvider() *InMemoryConsumerApiProvider {
	return &InMemoryConsumerApiProvider{
		progress: map[string]string{},
	}
}

func (consumer *InMemoryConsumerApiProvider) Cleanup() {
	consumer.progress = map[string]string{}
}

func (consumer *InMemoryConsumerApiProvider) LogConsume(ctx context.Context, request *AppLogConsumeProgress) error {
	if request.ConsumerId == "" {
		return fmt.Errorf("missing consumer id")
	}

	if request.Offset == "" {
		return fmt.Errorf("missing offset")
	}

	consumer.progress[request.ConsumerId] = request.Offset

	return nil
}

func (consumer *InMemoryConsumerApiProvider) GetLogConsume(ctx context.Context, consumerID string) (*AppLogConsumeProgress, error) {
	if consumerID == "" {
		return nil, fmt.Errorf("missing consumer id")
	}

	if consumer.progress[consumerID] == "" {
		return nil, crudstore.RecordNotFound
	}

	return &AppLogConsumeProgress{
		ConsumerId: consumerID,
		Offset:     consumer.progress[consumerID],
	}, nil
}

func (consumer *InMemoryConsumerApiProvider) List(ctx context.Context) ([]*AppLogConsumeProgress, error) {

	var results []*AppLogConsumeProgress
	for consumerID, offset := range consumer.progress {
		results = append(results, &AppLogConsumeProgress{
			ConsumerId: consumerID,
			Offset:     offset,
		})
	}

	return results, nil
}
