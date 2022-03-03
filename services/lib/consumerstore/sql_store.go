package consumerstore

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	eskitcommon "github.com/makkalot/eskit/services/lib/common"
	"github.com/makkalot/eskit/services/lib/crudstore"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"strconv"
)

var (
	consumerProgress = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "eskit_consumers_progress",
			Help: "Consumer Progress",
		}, []string{
			"consumer_name",
		})

	consumerLastSeen = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "eskit_consumers_lastseen",
			Help: "Consumer Progress",
		}, []string{
			"consumer_name",
		})

	consumedCount = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "eskit_consumers_consumed_count",
			Help: "Consume count",
		}, []string{
			"consumer_name",
		})
)

type ConsumerEntry struct {
	ID     string
	Offset string `gorm:"type:varchar(100); not null"`
}

type SQLConsumerApiProvider struct {
	db    *gorm.DB
	dbURI string
}

func NewSQLConsumerApiProvider(dbURI string) (Store, error) {
	var db *gorm.DB

	err := eskitcommon.RetryNormal(func() error {
		var err error
		db, err = gorm.Open("postgres", dbURI)
		if err != nil {
			return fmt.Errorf("connecting to db : %v", err)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	if result := db.AutoMigrate(&ConsumerEntry{}); result.Error != nil {
		return nil, result.Error
	}

	return &SQLConsumerApiProvider{
		db:    db,
		dbURI: dbURI,
	}, nil
}

func (consumer *SQLConsumerApiProvider) LogConsume(ctx context.Context, request *AppLogConsumeProgress) error {
	if request.ConsumerId == "" {
		return fmt.Errorf("missing consumer id")
	}

	if request.Offset == "" {
		return fmt.Errorf("missing offset")
	}

	entry := &ConsumerEntry{}
	if result := consumer.db.Where("id = ?", request.ConsumerId).First(&entry); result.Error != nil {
		if result.RecordNotFound() {
			if result := consumer.db.Create(&ConsumerEntry{
				ID:     request.ConsumerId,
				Offset: request.Offset,
			}); result.Error != nil {
				return fmt.Errorf("updating record failed : %v", result.Error)
			}

			return nil
		}
		return fmt.Errorf("fetching record failed : %v", result.Error)
	}

	offsetFloat, err := strconv.ParseFloat(entry.Offset, 64)
	if err != nil {
		return fmt.Errorf("invalid offset : %v", err)
	}

	entry.Offset = request.Offset
	if result := consumer.db.Save(entry); result.Error != nil {
		return fmt.Errorf("updating record failed : %v", result.Error)
	}

	consumerProgress.With(prometheus.Labels{"consumer_name": entry.ID}).Set(offsetFloat)
	consumedCount.With(prometheus.Labels{"consumer_name": entry.ID}).Inc()
	consumerLastSeen.With(prometheus.Labels{"consumer_name": entry.ID}).SetToCurrentTime()

	return nil
}

func (consumer *SQLConsumerApiProvider) GetLogConsume(ctx context.Context, consumerID string) (*AppLogConsumeProgress, error) {
	if consumerID == "" {
		return nil, fmt.Errorf("missing consumer id")
	}

	entry := &ConsumerEntry{}
	if result := consumer.db.Where("id = ?", consumerID).First(&entry); result.Error != nil {
		if result.RecordNotFound() {
			return nil, crudstore.RecordNotFound
		}
		return nil, fmt.Errorf("fetching failed : %v", result.Error)
	}

	return &AppLogConsumeProgress{
		ConsumerId: entry.ID,
		Offset:     entry.Offset,
	}, nil
}

func (consumer *SQLConsumerApiProvider) List(ctx context.Context) ([]*AppLogConsumeProgress, error) {

	entries := []*ConsumerEntry{}
	if result := consumer.db.Find(entries); result.Error != nil {
		return nil, fmt.Errorf("fetching consumers progress failed : %v", result.Error)
	}

	var results []*AppLogConsumeProgress
	for _, e := range entries {
		results = append(results, &AppLogConsumeProgress{
			ConsumerId: e.ID,
			Offset:     e.Offset,
		})
	}

	return results, nil
}
