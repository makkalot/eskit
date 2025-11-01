package eventstore

import (
	"errors"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/makkalot/eskit/lib/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	lastStreamID = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "eskit_events_stream_last_id",
			Help: "LastID in the stream",
		}, []string{
			"application_id", "partition_id",
		})

	streamCounter = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eskit_events_stream_count",
			Help: "Stream Count",
		}, []string{
			"application_id", "partition_id",
		})
)

var ErrDuplicate = errors.New("duplicate")

type Store interface {
	Append(event *types.Event) error
	Get(originator *types.Originator, fromVersion bool) ([]*types.Event, error)
	Logs(fromID uint64, size uint32, pipelineID string) ([]*types.AppLogEntry, error)
}

// StoreWithCleanup has the same methods as Store but also has Cleanup method
// it's useful when working in tests with the store
type StoreWithCleanup interface {
	Store
	Cleanup() error
}
