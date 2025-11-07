package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/go-ozzo/ozzo-validation"
	"github.com/jinzhu/gorm"
	_ "github.com/jinzhu/gorm/dialects/postgres"
	_ "github.com/jinzhu/gorm/dialects/sqlite"
	"github.com/makkalot/eskit/lib/common"
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

	// Create event store
	var estore eventstore.Store
	var db *gorm.DB

	if config.DbUri == "inmemory://" {
		estore = eventstore.NewInMemoryStore()
		log.Println("Using in-memory event store")
		log.Println("WARNING: In-memory mode has limited functionality for admin interface")
		// For in-memory mode, we'll create a temporary SQLite DB for admin queries
		config.DbDialect = "sqlite3"
		config.DbUri = ":memory:"
	} else {
		var err error
		estore, err = eventstore.NewSqlStore(config.DbDialect, config.DbUri)
		if err != nil {
			log.Fatalf("failed to create event store: %v", err)
		}
		log.Printf("Using %s event store", config.DbDialect)
	}

	// Create separate DB connection for admin queries
	// This allows us to run custom queries without modifying the library
	err = common.RetryNormal(func() error {
		var err error
		db, err = gorm.Open(config.DbDialect, config.DbUri)
		if err != nil {
			return fmt.Errorf("connecting to db for admin queries: %v", err)
		}
		return nil
	})

	if err != nil {
		log.Fatalf("failed to create DB connection: %v", err)
	}

	log.Println("Database connection established")

	// Initialize HTML templates
	if err := provider.InitTemplates("./templates"); err != nil {
		log.Fatalf("Failed to load templates: %v", err)
	}
	log.Println("Templates loaded successfully")

	// Create admin provider
	adminProvider := provider.NewAdminProvider(estore)

	// Setup HTTP routes
	mux := http.NewServeMux()
	adminProvider.SetupRoutes(db, mux)

	log.Printf("Starting server on %s", config.ListenAddr)
	log.Printf("Web UI available at http://localhost%s/", config.ListenAddr)
	log.Printf("  - Raw Events: http://localhost%s/events", config.ListenAddr)
	log.Printf("  - Application Log: http://localhost%s/applog", config.ListenAddr)
	log.Printf("  - CRUD Entities: http://localhost%s/crud", config.ListenAddr)

	if err := http.ListenAndServe(config.ListenAddr, mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
