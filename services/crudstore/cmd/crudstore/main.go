package main

import (
	"context"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/makkalot/eskit/generated/grpc/go/crudstore"
	"github.com/makkalot/eskit/services/crudstore/provider"
	crudstore2 "github.com/makkalot/eskit/services/lib/crudstore"
	"github.com/makkalot/eskit/services/lib/eventstore"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
)

type CrudStoreConfig struct {
	DbUri              string `json:"dbUri" mapstructure:"dbUri"`
	DbPassword         string `json:"dbPassword" mapstructure:"dbPassword"`
	ListenAddr         string `json:"listenAddr" mapstructure:"listenAddr"`
	EventStoreEndpoint string `json:"eventStoreEndpoint" mapstructure:"eventStoreEndpoint"`
}

func (c CrudStoreConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DbUri, validation.Required),
	)
}

func main() {

	viper.SetDefault("listenAddr", ":9090")
	_ = viper.BindEnv("eventStoreEndpoint", "EVENT_STORE_ENDPOINT")
	_ = viper.BindEnv("dbUri", "DB_URI")

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

	var estore eventstore.Store
	var dbUri string

	if config.DbPassword != "" {
		dbUri = config.DbUri + " password=" + config.DbPassword
	} else {
		dbUri = config.DbUri
	}

	if dbUri == "inmemory://" {
		estore = eventstore.NewInMemoryStore()
	} else {
		estore, err = eventstore.NewSqlStore("postgres", dbUri)
		if err != nil {
			log.Fatalf("failed to create event store : %v", err)
		}
	}

	s := grpc.NewServer(
		grpc.StreamInterceptor(grpc_prometheus.StreamServerInterceptor),
		grpc.UnaryInterceptor(grpc_prometheus.UnaryServerInterceptor),
	)
	grpc_prometheus.EnableHandlingTimeHistogram()

	ctx := context.Background()
	crudStore, err := crudstore2.NewCrudStoreProvider(ctx, estore)
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
		if err := http.ListenAndServe(":8888", nil); err != nil {
			log.Fatalf("starting metrics server failed : %v", err)
		}
	}()

	reflection.Register(s)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
