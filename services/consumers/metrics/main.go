package main

import (
	"github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
	"log"
	"context"
	"github.com/makkalot/eskit/services/clients"
	metrics "github.com/makkalot/eskit/services/consumers/metrics/provider"
)

type ConsumerConfig struct {
	ConsumerName          string `json:"consumerName" mapstructure:"consumerName"`
	CrudStoreEndpoint     string `json:"crudStoreEndpoint" mapstructure:"crudStoreEndpoint"`
	ConsumerStoreEndpoint string `json:"consumerStoreEndpoint" mapstructure:"consumerStoreEndpoint"`
	EventStoreEndpoint    string `json:"eventStoreEndpoint" mapstructure:"eventStoreEndpoint"`
}

func (c ConsumerConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ConsumerName, validation.Required),
		validation.Field(&c.ConsumerStoreEndpoint, validation.Required),
		validation.Field(&c.EventStoreEndpoint, validation.Required),
		validation.Field(&c.CrudStoreEndpoint, validation.Required),
	)
}

func main() {
	viper.BindEnv("consumerName", "CONSUMER_NAME")
	viper.BindEnv("consumerStoreEndpoint", "CONSUMERSTORE_ENDPOINT")
	viper.BindEnv("eventStoreEndpoint", "EVENT_STORE_ENDPOINT")
	viper.BindEnv("crudStoreEndpoint", "CRUDSTORE_ENDPOINT")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/metrics")
	viper.AddConfigPath(".")

	var config ConsumerConfig

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("config validation : %v", err)
	}

	ctx := context.Background()

	crudStoreClient, err := clients.NewCrudStoreGrpcClientWithWait(ctx, config.CrudStoreEndpoint)
	if err != nil {
		log.Fatalf("crud store client initialization failed : %v", err)
	}

	consumerStoreClient, err := clients.NewConsumerStoreGrpcClientWithWait(ctx, config.ConsumerStoreEndpoint)
	if err != nil {
		log.Fatalf("consumer store client initialization failed : %v", err)
	}

	eventStoreClient, err := clients.NewStoreClientWithWait(ctx, config.EventStoreEndpoint)
	if err != nil {
		log.Fatalf("event store client initialization failed : %v", err)
	}

	metricsConsumer := metrics.NewPrometheusMetricsConsumer(ctx, crudStoreClient)

	appLogConsumer, err := clients.NewAppLogConsumer(ctx, eventStoreClient, consumerStoreClient, config.ConsumerName, clients.FromSaved, "*")
	if err != nil {
		log.Fatalf("applog consumer initialization failed : %v", err)
	}

	if err := appLogConsumer.Consume(metricsConsumer.ConsumerCB); err != nil {
		log.Fatalf("consuming failed : %v", err)
	}
}
