package server

import (
	"context"
	"covalence/src/db/postgres"
	"covalence/src/firewall"
	"covalence/src/internal"
	"covalence/src/middleware" // Import middleware package
	"covalence/src/register"
	"covalence/src/router"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv" // Optional: for loading .env file
)

func Start() {
	// Optional: Load .env file for local development
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) { // Only log real errors, not "file not found"
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Configure Gin mode
	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
		log.Println("Running in Debug Mode")
	}

	// Initialize shared resources (like Redis) *before* setting up middleware that depends on them
	middleware.InitRedis() // Initialize Redis connection

	// --- Gin Engine Setup ---
	server := gin.New() // Use New for explicit middleware control

	// Configure trusted proxies
	// For development, you might want to trust only localhost
	if gin.Mode() == gin.DebugMode {
		// In debug mode, only trust localhost
		server.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	} else {
		// In production, you should set this to your known proxy IPs
		// This could be your load balancer, reverse proxy IPs, etc.
		// Example for AWS ELB: server.SetTrustedProxies([]string{"10.0.0.0/8"})
		// Or if not using a proxy: server.SetTrustedProxies(nil)

		// Read from environment variable if available
		trustedProxies := os.Getenv("TRUSTED_PROXIES")
		if trustedProxies != "" {
			server.SetTrustedProxies(strings.Split(trustedProxies, ","))
		} else {
			// If not specified, don't trust any proxies in production
			server.SetTrustedProxies(nil)
		}
	}

	// --- Global Middleware (Order Matters!) ---
	// 1. Logger & Recovery (Essential)
	server.Use(gin.Logger())
	server.Use(gin.Recovery())

	// 2. CORS (Handles browser pre-flight OPTIONS requests early)
	server.Use(middleware.CorsHandler()) // Reads CORS_ALLOWED_ORIGINS env var

	// 3. Rate Limiter (Apply early to protect resources)
	server.Use(middleware.RateLimiter()) // Reads RATE_LIMIT env var, uses Redis

	// 4. Session Management (Required before routes that use sessions)
	server.Use(middleware.SessionHandler()) // Reads SESSION_SECRET_KEY env var, uses Redis

	// 5. Security Headers (Applied to outgoing responses)
	server.Use(middleware.SecureHeaders()) // Reads GIN_MODE for dev/prod settings

	// --- Application Specific Setup ---
	ctx := context.Background()
	registry := register.NewModelRegistry()
	modelProviders, err := register.ReadModelProviders()
	if err != nil {
		log.Fatalf("failed to load model providers: %v", err)
	}
	internal.LoadModels("models.yaml")
	firewallConfig, err := firewall.LoadConfig("config.yaml")
	if err != nil {
		log.Fatalf("failed to load firewall config: %v", err)
	}

	// Database Connection (using Env Var)
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable not set!")
	}
	db, err := postgres.New(ctx, dbURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// HTTP Client
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 60 * time.Second,
	}

	// --- Routes ---
	server.POST("/model/register", func(c *gin.Context) {
		c.Set("registry", registry)
		router.RegisterModel(c)
	})

	server.GET("/model/list", func(c *gin.Context) {
		c.Set("registry", registry)
		router.ListRegisteredModels(c)
	})

	server.GET("/model/list/providers", func(c *gin.Context) {
		c.Set("providers", modelProviders)
		router.ListModelProviders(c)
	})

	// Health check - Note: This is now also rate-limited.
	// If you need it excluded, define it *before* server.Use(middleware.RateLimiter())
	// or use more advanced rate limiter configurations (e.g., path exclusion).
	server.GET("/health", func(c *gin.Context) {
		router.Health(c)
	})

	// API Group
	v1 := server.Group("/v1")
	// Add authentication middleware here if needed:
	// v1.Use(middleware.RequireAuth())
	{
		v1.Any("/*path", func(c *gin.Context) {
			// Access session data if needed:
			// session := sessions.Default(c)
			// userID := session.Get("userID")

			c.Set("registry", registry)
			c.Set("httpClient", httpClient)
			c.Set("db", db)
			router.Generate(c, &firewallConfig, firewall.HookFirewalls)
		})
	}

	// --- Start Server ---
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080" // Default port
	}
	log.Printf("Starting server on port %s (Mode: %s)", port, gin.Mode())
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
