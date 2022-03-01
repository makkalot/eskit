package provider

import (
	"context"
	"fmt"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	"github.com/makkalot/eskit/generated/grpc/go/consumerstore"
	common2 "github.com/makkalot/eskit/services/lib/common"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
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

type ConsumerApiProvider struct {
	db    *gorm.DB
	dbURI string
}

func NewConsumerApiProvider(dbURI string) (consumerstore.ConsumerServiceServer, error) {
	var db *gorm.DB

	err := common2.RetryNormal(func() error {
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

	return &ConsumerApiProvider{
		db:    db,
		dbURI: dbURI,
	}, nil
}

func (consumer *ConsumerApiProvider) Healtz(ctx context.Context, request *consumerstore.HealthRequest) (*consumerstore.HealthResponse, error) {
	return &consumerstore.HealthResponse{}, nil
}

func (consumer *ConsumerApiProvider) LogConsume(ctx context.Context, request *consumerstore.AppLogConsumeRequest) (*consumerstore.AppLogConsumeResponse, error) {
	if request.ConsumerId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing consumer id")
	}

	if request.Offset == "" {
		return nil, status.Error(codes.InvalidArgument, "missing offset")
	}

	entry := &ConsumerEntry{}
	if result := consumer.db.Where("id = ?", request.ConsumerId).First(&entry); result.Error != nil {
		if result.RecordNotFound() {
			if result := consumer.db.Create(&ConsumerEntry{
				ID:     request.ConsumerId,
				Offset: request.Offset,
			}); result.Error != nil {
				return nil, status.Errorf(codes.Internal, "updating record failed : %v", result.Error)
			}

			return &consumerstore.AppLogConsumeResponse{}, nil
		}
		return nil, status.Errorf(codes.Internal, "fetching record failed : %v", result.Error)
	}

	offsetFloat, err := strconv.ParseFloat(entry.Offset, 64)
	if err != nil {
		return nil, status.Errorf(codes.InvalidArgument, "invalid offset : %v", err)
	}

	entry.Offset = request.Offset
	if result := consumer.db.Save(entry); result.Error != nil {
		return nil, status.Errorf(codes.Internal, "updating record failed : %v", result.Error)
	}

	consumerProgress.With(prometheus.Labels{"consumer_name": entry.ID}).Set(offsetFloat)
	consumedCount.With(prometheus.Labels{"consumer_name": entry.ID}).Inc()
	consumerLastSeen.With(prometheus.Labels{"consumer_name": entry.ID}).SetToCurrentTime()

	return &consumerstore.AppLogConsumeResponse{}, nil
}

func (consumer *ConsumerApiProvider) GetLogConsume(ctx context.Context, request *consumerstore.GetAppLogConsumeRequest) (*consumerstore.GetAppLogConsumeResponse, error) {
	if request.ConsumerId == "" {
		return nil, status.Error(codes.InvalidArgument, "missing consumer id")
	}

	entry := &ConsumerEntry{}
	if result := consumer.db.Where("id = ?", request.ConsumerId).First(&entry); result.Error != nil {
		if result.RecordNotFound() {
			return nil, status.Error(codes.NotFound, "consumer not found")
		}
		return nil, status.Error(codes.Internal, "fetching failed")
	}

	return &consumerstore.GetAppLogConsumeResponse{
		ConsumerId: entry.ID,
		Offset:     entry.Offset,
	}, nil
}

func (consumer *ConsumerApiProvider) List(ctx context.Context, request *consumerstore.ListConsumersRequest) (*consumerstore.ListConsumersResponse, error) {

	entries := []*ConsumerEntry{}
	if result := consumer.db.Find(entries); result.Error != nil {
		return nil, status.Errorf(codes.NotFound, "consumer not found %v", result.Error)
	}

	var results []*consumerstore.GetAppLogConsumeResponse
	for _, e := range entries {
		results = append(results, &consumerstore.GetAppLogConsumeResponse{
			ConsumerId: e.ID,
			Offset:     e.Offset,
		})
	}

	return &consumerstore.ListConsumersResponse{
		Consumers: results,
	}, nil
}
