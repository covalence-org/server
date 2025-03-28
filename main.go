package main

import (
	"fmt"
	"log"
	"net/http"
	"netrunner/firewall"
	"netrunner/internal"
	"netrunner/register"
	"netrunner/router"
	"time"

	"github.com/gin-gonic/gin"
)

func main() {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Create model registry
	registry := register.NewModelRegistry()

	// Load Internal Models
	filePath := "internal/models.yaml"
	internal.LoadModels(filePath)

	// Load Firewall Config
	firewallConfig, err := firewall.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load firewall config: %v", err)
		return
	}

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
		router.RegisterModel(c, registry)
	})

	// List registered models endpoint
	r.GET("/model/list", func(c *gin.Context) {
		router.ListRegisteredModels(c, registry)
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		router.Health(c)
	})

	// Proxy endpoint - catch all requests
	r.Any("/v1/*path", func(c *gin.Context) {
		router.Generate(c, registry, httpClient, &firewallConfig, firewall.HookFirewalls)
	})

	port := 8080

	// Start server
	log.Printf("starting ai model proxy server on :%d", port)
	if err := r.Run(fmt.Sprintf(":%d", port)); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}
