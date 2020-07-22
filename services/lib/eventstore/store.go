package eventstore

import (
	"fmt"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/makkalot/eskit/generated/grpc/go/common"
	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
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

type ErrDuplicate struct {
	msg string
}

func (e *ErrDuplicate) Error() string {
	return fmt.Sprintf("duplicate error : %s", e.msg)
}

type Store interface {
	Append(event *store.Event) error
	Get(originator *common.Originator, fromVersion bool) ([]*store.Event, error)
	Logs(fromID uint64, size uint32, pipelineID string) ([]*store.AppLogEntry, error)
}
