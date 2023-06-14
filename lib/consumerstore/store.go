package consumerstore

import (
	"context"
)

// AppLogConsumeProgress represents the progress of the app log consumer
type AppLogConsumeProgress struct {
	ConsumerId string `json:"consumer_id"`
	Offset     string `json:"offset"`
}

// Store is the interface for the consumer store
type Store interface {
	LogConsume(ctx context.Context, request *AppLogConsumeProgress) error
	GetLogConsume(ctx context.Context, consumerID string) (*AppLogConsumeProgress, error)
	List(ctx context.Context) ([]*AppLogConsumeProgress, error)
}
