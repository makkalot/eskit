package main

import (
	"context"
	"github.com/go-ozzo/ozzo-validation"
	"github.com/makkalot/eskit/lib/crudstore"
	"github.com/makkalot/eskit/lib/eventstore"
	"github.com/makkalot/eskit/services/camconfig/provider"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
	"log"
	"net/http"
	"path/filepath"
)

type CamConfigServiceConfig struct {
	ListenAddr  string `json:"listenAddr" mapstructure:"listenAddr"`
	DbUri       string `json:"dbUri" mapstructure:"dbUri"`
	TemplateDir string `json:"templateDir" mapstructure:"templateDir"`
}

func (c CamConfigServiceConfig) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.ListenAddr, validation.Required),
		validation.Field(&c.DbUri, validation.Required),
	)
}

func main() {
	viper.SetDefault("listenAddr", ":8081")
	viper.SetDefault("dbUri", "inmemory://")
	viper.SetDefault("templateDir", "./web/templates")
	viper.BindEnv("dbUri", "DB_URI")

	viper.SetConfigName("config")
	viper.AddConfigPath("/etc/camconfig")
	viper.AddConfigPath(".")

	var config CamConfigServiceConfig

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

	log.Println("Starting CamConfig Service")
	log.Println("Listen Address: ", config.ListenAddr)
	log.Println("DB URI: ", config.DbUri)

	// Create event store (in-memory for this example)
	var estore eventstore.Store
	if config.DbUri == "inmemory://" {
		estore = eventstore.NewInMemoryStore()
		log.Println("Using in-memory event store")
	} else {
		var err error
		estore, err = eventstore.NewSqlStore("postgres", config.DbUri)
		if err != nil {
			log.Fatalf("failed to create event store: %v", err)
		}
		log.Println("Using PostgreSQL event store")
	}

	// Create CRUD store client
	crudStoreClient, err := crudstore.NewClient(context.Background(), config.DbUri)
	if err != nil {
		log.Fatalf("creating crudstore client failed : %v", err)
	}

	// Create service provider
	camConfigProvider, err := provider.NewCamConfigServiceProvider(crudStoreClient, estore)
	if err != nil {
		log.Fatalf("camconfig provider failed initializing : %v", err)
	}

	// Initialize HTML templates
	templatePath := filepath.Join(config.TemplateDir)
	if err := provider.InitTemplates(templatePath); err != nil {
		log.Printf("Warning: Failed to load templates from %s: %v", templatePath, err)
		log.Println("Web interface may not work correctly")
	} else {
		log.Println("Templates loaded from:", templatePath)
	}

	// Setup REST API routes
	mux := http.NewServeMux()

	// Health endpoint
	mux.HandleFunc("/v1/health", camConfigProvider.HealthHandler)

	// JSON API CRUD endpoints
	mux.HandleFunc("/v1/camconfigs", func(w http.ResponseWriter, r *http.Request) {
		switch r.Method {
		case http.MethodPost:
			camConfigProvider.CreateCamConfigHandler(w, r)
		case http.MethodGet:
			camConfigProvider.GetCamConfigHandler(w, r)
		case http.MethodPut:
			camConfigProvider.UpdateCamConfigHandler(w, r)
		case http.MethodDelete:
			camConfigProvider.DeleteCamConfigHandler(w, r)
		default:
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		}
	})

	// Web interface endpoints
	mux.HandleFunc("/web/", camConfigProvider.WebIndexHandler)
	mux.HandleFunc("/web/create", camConfigProvider.WebCreateHandler)
	mux.HandleFunc("/web/edit", camConfigProvider.WebEditHandler)
	mux.HandleFunc("/web/delete", camConfigProvider.WebDeleteHandler)
	mux.HandleFunc("/web/audit", camConfigProvider.WebAuditLogHandler)

	// Serve static files for HTMX (we'll include it inline in templates)
	// Root redirect
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/" {
			http.Redirect(w, r, "/web/", http.StatusSeeOther)
		} else {
			http.NotFound(w, r)
		}
	})

	// Prometheus metrics endpoint
	mux.Handle("/metrics", promhttp.Handler())

	log.Printf("Starting server on %s", config.ListenAddr)
	log.Printf("Web UI available at http://localhost%s/web/", config.ListenAddr)
	log.Printf("API available at http://localhost%s/v1/camconfigs", config.ListenAddr)
	log.Printf("Audit log at http://localhost%s/web/audit", config.ListenAddr)

	if err := http.ListenAndServe(config.ListenAddr, mux); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
