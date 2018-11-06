package main

import (
	"log"
	"google.golang.org/grpc"

	store "github.com/makkalot/eskit/generated/grpc/go/eventstore"
	provider2 "github.com/makkalot/eskit/services/eventstore/provider"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/reflection"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
	"net"
	"io/ioutil"
	"fmt"
	"github.com/davecgh/go-spew/spew"
)

type EventStoreConfig struct {
	ListenAddr string `json:"listenAddr" mapstructure:"listenAddr"`
	DbUri      string `json:"dbUri" mapstructure:"dbUri"`
	DbPassword string `json:"dbPassword" mapstructure:"dbPassword"`
}

func (c EventStoreConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DbUri, validation.Required),
	)
}

func main() {

	viper.BindEnv("dbUri", "DB_URI")
	viper.BindEnv("dbPassword", "DB_PASSWORD")
	viper.SetDefault("listenAddr", ":9090")

	viper.SetConfigName("config")
	viper.SetConfigType("yaml")

	viper.AddConfigPath("/etc/eventstore")
	viper.AddConfigPath(".")

	var config EventStoreConfig

	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Error reading config file, %s", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	b, err := ioutil.ReadFile("/etc/eventstore/config.yaml") // just pass the file name
	if err != nil {
		fmt.Print(err)
	}

	log.Println("Satrting the service with config : ", string(b))
	log.Println("Parsed viper config : ", spew.Sdump(config))

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

	var estore provider2.Store

	var dbUri string

	if config.DbPassword != "" {
		dbUri = config.DbUri + " password=" + config.DbPassword
	} else {
		dbUri = config.DbUri
	}

	if dbUri == "inmemory://" {
		estore = provider2.NewInMemoryStore()
	} else {
		estore, err = provider2.NewSqlStore("postgres", dbUri)
		if err != nil {
			log.Fatalf("failed to create event store : %v", err)
		}
	}

	eventStoreProvider, err := provider2.NewEventStoreApiProvider(estore)
	if err != nil {
		log.Fatalf("consumerapi provider failed initializing : %v", err)
	}
	store.RegisterEventstoreServiceServer(s, eventStoreProvider)

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
