package server

import (
	"context"
	"covalence/src/db/postgres"
	"covalence/src/firewall"
	"covalence/src/internal"
	"covalence/src/middleware"
	"covalence/src/register"
	"covalence/src/router"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
)

func Start() {
	// Load .env (if present)
	err := godotenv.Load()
	if err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
		log.Println("Running in Debug Mode")
	}

	// Init Redis for rate limiter, sessions, etc.
	middleware.InitRedis()

	// Create Gin engine with global middleware
	server := gin.New()
	// Configure trusted proxies
	// For development, you might want to trust only localhost
	if gin.Mode() == gin.DebugMode {
		// In debug mode, only trust localhost
		server.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	} else {
		trustedProxies := os.Getenv("TRUSTED_PROXIES")
		if trustedProxies != "" {
			server.SetTrustedProxies(strings.Split(trustedProxies, ","))
		} else {
			// If not specified, don't trust any proxies in production
			server.SetTrustedProxies(nil)
		}
	}

	server.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.CorsHandler(),
		middleware.RateLimiter(),
		middleware.SessionHandler(),
		middleware.SecureHeaders(),
	)

	// Supabase setup (admin and anon keys)
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if serviceRoleKey == "" || anonKey == "" {
		log.Fatal("FATAL: SUPABASE_SERVICE_ROLE_KEY and SUPABASE_ANON_KEY must be set")
	}

	// Admin client for server-side operations
	serviceClient, err := supabase.NewClient(supabaseURL, serviceRoleKey, &supabase.ClientOptions{Headers: map[string]string{}})
	if err != nil {
		log.Fatalf("Failed to initialize Supabase service client: %v", err)
	}

	// Load configs and models
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

	// Database connection
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		log.Fatal("FATAL: DATABASE_URL environment variable not set!")
	}
	db, err := postgres.New(ctx, dbURL)
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// HTTP client for external calls
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 60 * time.Second,
	}

	// Routes (all protected)
	model := server.Group("/model")
	{
		model.POST("/register", middleware.Auth(serviceClient, anonKey, supabaseURL), func(c *gin.Context) {
			c.Set("registry", registry)
			router.RegisterModel(c)
		})
		model.GET("/list", middleware.Auth(serviceClient, anonKey, supabaseURL), func(c *gin.Context) {
			c.Set("registry", registry)
			router.ListRegisteredModels(c)
		})
		model.GET("/list/providers", middleware.Auth(serviceClient, anonKey, supabaseURL), func(c *gin.Context) {
			c.Set("providers", modelProviders)
			router.ListModelProviders(c)
		})
	}

	// auth API group
	auth := server.Group("/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			router.UserLogin(c, serviceClient)
		})
		auth.POST("/signup", func(c *gin.Context) {
			router.UserSignup(c, serviceClient)
		})
		auth.POST("/refresh", middleware.Auth(serviceClient, anonKey, supabaseURL), func(c *gin.Context) {
			router.UserRefreshToken(c)
		})
		auth.POST("/me", middleware.Auth(serviceClient, anonKey, supabaseURL), func(c *gin.Context) {
			router.UserMe(c)
		})
	}

	server.GET("/health", router.Health)

	// v1 API group
	v1 := server.Group("/v1")
	// Apply AuthMiddleware to all routes
	v1.Use(middleware.Auth(serviceClient, anonKey, supabaseURL))
	{
		v1.Any("/*path", func(c *gin.Context) {
			c.Set("registry", registry)
			c.Set("db", db)
			c.Set("httpClient", httpClient)
			router.Generate(c, &firewallConfig, firewall.HookFirewalls)
		})
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	log.Printf("Starting server on port %s (Mode: %s)", port, gin.Mode())
	if err := server.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}
