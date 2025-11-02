package main

import (
	"context"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/services/users/provider"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"log"
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

	viper.SetDefault("listenAddr", ":8080")
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

	crudStoreClient, err := crudstore.NewClient(context.Background(), config.DbUri)
	if err != nil {
		log.Fatalf("creating crudstore client failed : %v", err)
	}

	userProvider, err := provider.NewUserServiceProvider(crudStoreClient)
	if err != nil {
		log.Fatalf("user provider failed initializing : %v", err)
	}

	// Setup REST API routes
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/v1/health", userProvider.HealthHandler)

	// User CRUD endpoints
	mux.HandleFunc("/v1/users", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			userProvider.CreateUserHandler(w, r)
		case http.MethodGet:
			userProvider.GetUserHandler(w, r)
		case http.MethodPut:
			userProvider.UpdateUserHandler(w, r)
		case http.MethodDelete:
			userProvider.DeleteUserHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting REST API server on %s", config.ListenAddr)
	if err := http.ListenAndServe(config.ListenAddr, mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
