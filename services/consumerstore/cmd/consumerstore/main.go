package main

import (
	"net"
	"log"
	"google.golang.org/grpc"

	"github.com/makkalot/eskit/services/consumerstore/provider"
	"github.com/makkalot/eskit/generated/grpc/go/consumerstore"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"google.golang.org/grpc/reflection"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
)

type ConsumerStoreConfig struct {
	ListenAddr string `json:"listenAddr" mapstructure:"listenAddr"`
	DbUri      string `json:"dbUri" mapstructure:"dbUri"`
	DbPassword string `json:"dbPassword" mapstructure:"dbPassword"`
}

func (c ConsumerStoreConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.DbUri, validation.Required),
	)
}

func main() {

	viper.BindEnv("dbURI", "DB_URI")
	viper.BindEnv("dbPassword", "DB_PASSWORD")
	viper.SetDefault("listenAddr", ":9090")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/consumerstore")
	viper.AddConfigPath(".")

	var config ConsumerStoreConfig

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

	var dbUri string

	if config.DbPassword != "" {
		dbUri = config.DbUri + " password=" + config.DbPassword
	} else {
		dbUri = config.DbUri
	}

	consumerProvider, err := provider.NewConsumerApiProvider(dbUri)
	if err != nil {
		log.Fatalf("consumerapi provider failed initializing : %v", err)
	}
	consumerstore.RegisterConsumerServiceServer(s, consumerProvider)

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
