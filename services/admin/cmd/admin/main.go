package main

import (
	"context"
	"log"
	"net/http"

	"github.com/go-ozzo/ozzo-validation"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/eventstore"
	"github.com/makkalot/eskit/services/admin/provider"
	"github.com/spf13/viper"
)

type AdminServiceConfig struct {
	ListenAddr string `json:"listenAddr" mapstructure:"listenAddr"`
	DbUri      string `json:"dbUri" mapstructure:"dbUri"`
	DbDialect  string `json:"dbDialect" mapstructure:"dbDialect"`
}

func (c AdminServiceConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ListenAddr, validation.Required),
		validation.Field(&c.DbUri, validation.Required),
		validation.Field(&c.DbDialect, validation.Required),
	)
}

func main() {
	// Set defaults
	viper.SetDefault("listenAddr", ":8082")
	viper.SetDefault("dbUri", "inmemory://")
	viper.SetDefault("dbDialect", "postgres")

	// Bind environment variables
	viper.BindEnv("dbUri", "DB_URI")
	viper.BindEnv("dbDialect", "DB_DIALECT")
	viper.BindEnv("listenAddr", "LISTEN_ADDR")

	// Config file paths
	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/eskit-admin")
	viper.AddConfigPath(".")

	var config AdminServiceConfig

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("Warning: Error reading config file: %s, using defaults", err)
	}

	err := viper.Unmarshal(&config)
	if err != nil {
		log.Fatalf("unable to decode into struct, %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("config validation : %v", err)
	}

	log.Println("Starting ESKit Admin Service")
	log.Println("Listen Address:", config.ListenAddr)
	log.Println("DB URI:", config.DbUri)
	log.Println("DB Dialect:", config.DbDialect)

	// Create event store using library API
	var estore eventstore.Store

	if config.DbUri == "inmemory://" {
		estore = eventstore.NewInMemoryStore()
		log.Println("Using in-memory event store")
	} else {
		var err error
		estore, err = eventstore.NewSqlStore(config.DbDialect, config.DbUri)
		if err != nil {
			log.Fatalf("failed to create event store: %v", err)
		}
		log.Printf("Using %s event store", config.DbDialect)
	}

	// Create CRUD store using library API
	ctx := context.Background()
	crudStore, err := crudstore.NewCrudStoreProvider(ctx, estore)
	if err != nil {
		log.Fatalf("failed to create crud store: %v", err)
	}
	log.Println("CRUD store initialized")

	// Initialize HTML templates
	if err := provider.InitTemplates("./templates"); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	log.Println("Templates loaded successfully")

	// Create admin provider with library stores
	adminProvider := provider.NewAdminProvider(estore, crudStore)

	// Setup HTTP routes (no DB parameter - using library APIs only)
	mux := http.NewServeMux()
	adminProvider.SetupRoutes(mux)

	log.Printf("Starting server on %s", config.ListenAddr)
	log.Printf("Web UI available at http://localhost%s/", config.ListenAddr)
	log.Printf("  - Raw Events: http://localhost%s/events", config.ListenAddr)
	log.Printf("  - Application Log: http://localhost%s/applog", config.ListenAddr)
	log.Printf("  - CRUD Entities: http://localhost%s/crud", config.ListenAddr)

	if err := http.ListenAndServe(config.ListenAddr, mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
