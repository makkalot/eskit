package main

import (
	"context"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"github.com/makkalot/eskit/services/lib/crudstore"
	"github.com/makkalot/eskit/services/users/provider"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
	"net/http"
)

type UserStoreConfig struct {
	ListenAddr string `json:"listenAddr" mapstructure:"listenAddr"`
	DbUri      string `json:"dbUri" mapstructure:"dbUri"`
}

func (c UserStoreConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ListenAddr, validation.Required),
		validation.Field(&c.DbUri, validation.Required),
	)
}

func main() {

	viper.SetDefault("listenAddr", ":9090")
	viper.BindEnv("dbUri", "DB_URI")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/userstore")
	viper.AddConfigPath(".")

	var config UserStoreConfig

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

	s := grpc.NewServer()

	crudStoreClient, err := crudstore.NewClient(context.Background(), config.DbUri)
	if err != nil {
		log.Fatalf("creating crudstore client failed : %v", err)
	}

	userProvider, err := provider.NewUserServiceProvider(crudStoreClient)
	if err != nil {
		log.Fatalf("user provider failed initializing : %v", err)
	}

	grpc_prometheus.Register(s)
	http.Handle("/metrics", promhttp.Handler())
	go func() {
		http.ListenAndServe(":8888", nil)
	}()

	reflection.Register(s)
	users.RegisterUserServiceServer(s, userProvider)
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
