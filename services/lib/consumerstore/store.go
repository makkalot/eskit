package consumerstore

import (
	"context"
)

type AppLogConsumeProgress struct {
	ConsumerId string `json:"consumer_id"`
	Offset     string `json:"offset"`
}

type Store interface {
	LogConsume(ctx context.Context, request *AppLogConsumeProgress) error
	GetLogConsume(ctx context.Context, consumerID string) (*AppLogConsumeProgress, error)
	List(ctx context.Context) ([]*AppLogConsumeProgress, error)
}
