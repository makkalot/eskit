package main

import (
	"log"
	"net"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/makkalot/eskit/generated/grpc/go/crudstore"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc"
	"github.com/makkalot/eskit/services/crudstore/provider"
	"github.com/makkalot/eskit/services/clients"
	"context"
	"google.golang.org/grpc/reflection"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
)

type CrudStoreConfig struct {
	ListenAddr         string `json:"listenAddr" mapstructure:"listenAddr"`
	EventStoreEndpoint string `json:"eventStoreEndpoint" mapstructure:"eventStoreEndpoint"`
}

func (c CrudStoreConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.EventStoreEndpoint, validation.Required),
	)
}

func main() {

	viper.SetDefault("listenAddr", ":9090")
	viper.BindEnv("eventStoreEndpoint", "EVENT_STORE_ENDPOINT")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/crudstore")
	viper.AddConfigPath(".")

	var config CrudStoreConfig

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

	log.Println("Going to listen on : ", config.ListenAddr)
	lis, err := net.Listen("tcp", config.ListenAddr)
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	grpc_prometheus.EnableHandlingTimeHistogram()

	ctx := context.Background()
	eventStoreClient, err := clients.NewStoreClientWithWait(ctx, config.EventStoreEndpoint)
	if err != nil {
		log.Fatalf("initializing eventstore client failed : %v", err)
	}

	crudStore, err := provider.NewCrudStoreProvider(ctx, eventStoreClient)
	if err != nil {
		log.Fatalf("initializing crud crudstore failed : %v", err)
	}

	api, err := provider.NewCrudStoreApiProvider(crudStore)
	if err != nil {
		log.Fatalf("initializing crud crudstore api failed : %v", err)
	}

	crudstore.RegisterCrudStoreServiceServer(s, api)

	grpc_prometheus.Register(s)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":8888", nil)
	}()

	reflection.Register(s)

	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
