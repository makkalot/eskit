package main

import (
	"net"
	"log"
	"google.golang.org/grpc"
	"github.com/makkalot/eskit/services/users/provider"
	"github.com/makkalot/eskit/generated/grpc/go/users"
	"google.golang.org/grpc/reflection"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/spf13/viper"
	"github.com/grpc-ecosystem/go-grpc-prometheus"
	"net/http"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type UserStoreConfig struct {
	ListenAddr        string `json:"listenAddr" mapstructure:"listenAddr"`
	CrudStoreEndpoint string `json:"crudStoreEndpoint" mapstructure:"crudStoreEndpoint"`
}

func (c UserStoreConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.CrudStoreEndpoint, validation.Required),
	)
}

func main() {

	viper.SetDefault("listenAddr", ":9090")
	viper.BindEnv("crudStoreEndpoint", "CRUDSTORE_ENDPOINT")

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
	userProvider, err := provider.NewUserServiceProvider(config.CrudStoreEndpoint)
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
