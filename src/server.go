package server

import (
	"context"
	"covalence/src/db/postgres"
	"covalence/src/firewall"
	"covalence/src/internal"
	"covalence/src/register"
	"covalence/src/router"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func Start() {
	ctx := context.Background()

	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Create model registry
	registry := register.NewModelRegistry()

	// Load Model Providers
	modelProviders, err := register.ReadModelProviders()
	if err != nil {
		log.Fatalf("failed to load model providers: %v", err)
		return
	}

	// Load Internal Models
	internal.LoadModels("models.yaml")

	// Load Firewall Config
	firewallConfig, err := firewall.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load firewall config: %v", err)
		return
	}

	// Load Audit DB
	// Connect to database
	db, err := postgres.New(ctx, "user=alialh dbname=covalence_dev sslmode=disable")
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Create a custom HTTP client with connection pooling
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 60 * time.Second, // Longer timeout for streaming responses
	}

	// Model registration endpoint
	r.POST("/model/register", func(c *gin.Context) {
		c.Set("registry", registry)
		router.RegisterModel(c)
	})

	// List registered models endpoint
	r.GET("/model/list", func(c *gin.Context) {
		c.Set("registry", registry)
		router.ListRegisteredModels(c)
	})

	// List registered models endpoint
	r.GET("/model/list/providers", func(c *gin.Context) {
		c.Set("providers", modelProviders)
		router.ListModelProviders(c)
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		router.Health(c)
	})

	// Proxy endpoint - catch all requests
	r.Any("/v1/*path", func(c *gin.Context) {
		c.Set("registry", registry)
		c.Set("httpClient", httpClient)
		c.Set("db", db)

		router.Generate(c, &firewallConfig, firewall.HookFirewalls)
	})

	port := 8080

	// Start server
	log.Printf("starting ai model proxy server on :%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
