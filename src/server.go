package server

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"covalence/src/db/postgres"
	"covalence/src/firewall"
	"covalence/src/internal"
	"covalence/src/middleware"
	"covalence/src/register"
	"covalence/src/router"

	"github.com/gin-gonic/gin"
	"github.com/go-openapi/spec"
	"github.com/joho/godotenv"
	"github.com/supabase-community/supabase-go"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// Start initializes and runs the Gin server
func Start() {
	// Load environment variables
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		log.Printf("Warning: Could not load .env file: %v", err)
	}

	// Set Gin mode
	if os.Getenv("GIN_MODE") == "release" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
		log.Println("Running in Debug Mode")
	}

	// Initialize Redis, middleware, etc.
	middleware.InitRedis()
	server := gin.New()
	configureProxies(server)
	server.Use(
		gin.Logger(),
		gin.Recovery(),
		middleware.CorsHandler(),
		middleware.RateLimiter(),
		middleware.SessionHandler(),
		middleware.SecureHeaders(),
	)

	// Initialize Supabase client
	supabaseURL := os.Getenv("SUPABASE_URL")
	serviceRoleKey := os.Getenv("SUPABASE_SERVICE_ROLE_KEY")
	anonKey := os.Getenv("SUPABASE_ANON_KEY")
	if serviceRoleKey == "" || anonKey == "" {
		log.Fatal("FATAL: SUPABASE_SERVICE_ROLE_KEY and SUPABASE_ANON_KEY must be set")
	}
	serviceClient, err := supabase.NewClient(supabaseURL, serviceRoleKey, &supabase.ClientOptions{})
	if err != nil {
		log.Fatalf("Failed to initialize Supabase client: %v", err)
	}

	// Load models & firewall config
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
		log.Fatalf("Database connection failed: %v", err)
	}
	defer db.Close()

	// HTTP client
	httpClient := &http.Client{
		Transport: &http.Transport{
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     90 * time.Second,
		},
		Timeout: 60 * time.Second,
	}

	// Register application routes
	registerRoutes(server, serviceClient, anonKey, supabaseURL, registry, modelProviders, db, httpClient, firewallConfig)

	// Register Swagger endpoints
	setupSwagger(server)

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

// configureProxies sets trusted proxies based on environment
func configureProxies(server *gin.Engine) {
	if gin.Mode() == gin.DebugMode {
		server.SetTrustedProxies([]string{"127.0.0.1", "::1"})
	} else if tp := os.Getenv("TRUSTED_PROXIES"); tp != "" {
		server.SetTrustedProxies(strings.Split(tp, ","))
	} else {
		server.SetTrustedProxies(nil)
	}
}

// registerRoutes defines all application endpoints
func registerRoutes(
	server *gin.Engine,
	serviceClient *supabase.Client,
	anonKey, supabaseURL string,
	registry *register.Registry,
	modelProviders []*register.ModelProvider,
	db *postgres.DB,
	httpClient *http.Client,
	firewallConfig firewall.Config,
) {
	// Model group
	model := server.Group("/model")
	model.Use(middleware.Auth(serviceClient, anonKey, supabaseURL))
	{
		model.POST("/register", func(c *gin.Context) {
			c.Set("registry", registry)
			router.RegisterModel(c)
		})
		model.GET("/list", func(c *gin.Context) {
			c.Set("registry", registry)
			router.ListRegisteredModels(c)
		})
		model.GET("/list/providers", func(c *gin.Context) {
			c.Set("providers", modelProviders)
			router.ListModelProviders(c)
		})
	}

	// Auth group
	auth := server.Group("/auth")
	{
		auth.POST("/login", func(c *gin.Context) {
			router.UserLogin(c, serviceClient)
		})
		auth.POST("/signup", func(c *gin.Context) {
			router.UserSignup(c, serviceClient)
		})
		auth.POST("/refresh", func(c *gin.Context) {
			router.UserRefreshToken(c, serviceClient)
		})
		auth.POST("/me", middleware.Auth(serviceClient, anonKey, supabaseURL), func(c *gin.Context) {
			router.UserMe(c)
		})
	}

	// Health endpoint
	server.GET("/health", router.Health)

	// v1 proxy
	v1 := server.Group("/v1")
	v1.Use(middleware.Auth(serviceClient, anonKey, supabaseURL))
	{
		v1.Any("/*path", func(c *gin.Context) {
			c.Set("registry", registry)
			c.Set("db", db)
			c.Set("httpClient", httpClient)
			router.Generate(c, &firewallConfig, firewall.HookFirewalls)
		})
	}

	dead := server.Group("/dead")
	{
		dead.Any("/*path", func(c *gin.Context) {
			c.Set("httpClient", httpClient)
			router.PassThrough(c)
		})
	}
}

// setupSwagger builds a Swagger 2.0 spec from the registered Gin routes,
// serves it at /swagger.json and mounts Swagger‑UI under /swagger/*any.
// A tiny override stylesheet at /swagger/index.css removes the default top‑bar.
func setupSwagger(server *gin.Engine) {
	/* ---------- build spec ---------- */
	paths := map[string]spec.PathItem{}
	for _, r := range server.Routes() {
		if strings.HasPrefix(r.Path, "/swagger") {
			continue // don’t document the docs
		}
		item := paths[r.Path]
		op := spec.NewOperation(r.Handler).
			WithID(r.Handler).
			WithSummary(r.Method+" "+r.Path).
			WithProduces("application/json").
			RespondsWith(http.StatusOK, spec.NewResponse().WithDescription("OK"))

		switch r.Method {
		case http.MethodGet:
			item.Get = op
		case http.MethodPost:
			item.Post = op
		case http.MethodPut:
			item.Put = op
		case http.MethodPatch:
			item.Patch = op
		case http.MethodDelete:
			item.Delete = op
		}
		paths[r.Path] = item
	}
	sw := spec.Swagger{
		SwaggerProps: spec.SwaggerProps{
			Swagger: "2.0",
			Info: &spec.Info{InfoProps: spec.InfoProps{
				Title:       "Covalence API",
				Version:     "1.0",
				Description: "Auto-generated from Gin routes",
			}},
			Paths: &spec.Paths{Paths: paths},
		},
	}
	specBytes, err := json.MarshalIndent(sw, "", "  ")
	if err != nil {
		log.Fatalf("marshal swagger spec: %v", err)
	}

	/* ---------- raw JSON ---------- */               // your helper
	server.GET("/swagger.json", func(c *gin.Context) { // raw JSON
		c.Data(http.StatusOK, "application/json", specBytes)
	})

	// ----- Swagger‑UI – ONE handler only -----
	custom := []byte(`.swagger-ui .topbar{display:none!important}`)
	ui := ginSwagger.WrapHandler(swaggerfiles.Handler, ginSwagger.URL("/swagger.json"))

	server.GET("/swagger/*any", func(c *gin.Context) {
		c.Writer.Header().Del("Content-Security-Policy") // inline CSS/JS
		if strings.HasSuffix(c.Param("any"), "index.css") {
			c.Data(http.StatusOK, "text/css; charset=utf-8", custom)
			return
		}
		ui(c) // everything else → default asset handler
	})

}
