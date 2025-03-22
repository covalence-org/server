package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// ModelInfo stores information about a registered model
type ModelInfo struct {
	CustomName  string // User-provided name
	ActualModel string // Real model name to use with API
	APIURL      string // URL to forward the request to
	CreatedAt   time.Time
}

// ModelRegistry stores registered models
type ModelRegistry struct {
	mu     sync.RWMutex
	models map[string]ModelInfo
}

// NewModelRegistry creates a new model registry
func NewModelRegistry() *ModelRegistry {
	return &ModelRegistry{
		models: make(map[string]ModelInfo),
	}
}

// RegisterModel adds or updates model information
func (r *ModelRegistry) RegisterModel(customName, actualModel, apiURL string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.models[customName] = ModelInfo{
		CustomName:  customName,
		ActualModel: actualModel,
		APIURL:      apiURL,
		CreatedAt:   time.Now(),
	}
}

// GetModelInfo retrieves model information by custom name
func (r *ModelRegistry) GetModelInfo(customName string) (ModelInfo, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	info, exists := r.models[customName]
	return info, exists
}

// RequestMetrics collects metrics about the request
type RequestMetrics struct {
	StartTime         time.Time
	ModelLookupTime   time.Duration
	RequestBodyTime   time.Duration
	UpstreamLatency   time.Duration
	TotalProcessTime  time.Duration
	StatusCode        int
	CustomModel       string
	ActualModel       string
	StreamingResponse bool
}

func main() {
	// Set Gin to release mode for production
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()

	// Create model registry
	registry := NewModelRegistry()

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
	r.POST("/register-model", func(c *gin.Context) {
		var req struct {
			Name        string `json:"name" binding:"required"`
			ActualModel string `json:"model" binding:"required"`
			APIURL      string `json:"api_url" binding:"required"`
		}

		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request format"})
			return
		}

		// Validate model names
		if !isValidModelName(req.Name) || !isValidModelName(req.ActualModel) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid model name format"})
			return
		}

		// Validate URL format and schema
		parsedURL, err := url.Parse(req.APIURL)
		if err != nil || (parsedURL.Scheme != "http" && parsedURL.Scheme != "https") {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid API URL format or schema"})
			return
		}

		registry.RegisterModel(req.Name, req.ActualModel, req.APIURL)
		log.Printf("Model registered: %s -> %s at %s", req.Name, req.ActualModel, req.APIURL)

		c.JSON(http.StatusOK, gin.H{"status": "Model registered", "name": req.Name})
	})

	// List registered models endpoint
	r.GET("/models", func(c *gin.Context) {
		registry.mu.RLock()
		defer registry.mu.RUnlock()

		models := make([]map[string]string, 0, len(registry.models))
		for _, info := range registry.models {
			models = append(models, map[string]string{
				"custom_name":   info.CustomName,
				"actual_model":  info.ActualModel,
				"registered_at": info.CreatedAt.Format(time.RFC3339),
			})
		}

		c.JSON(http.StatusOK, gin.H{"models": models})
	})

	// Health check endpoint
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy"})
	})

	// Proxy endpoint - catch all requests
	r.Any("/v1/*path", func(c *gin.Context) {
		metrics := RequestMetrics{
			StartTime: time.Now(),
		}

		defer func() {
			metrics.TotalProcessTime = time.Since(metrics.StartTime)

			// Log metrics in JSON format for easier parsing
			logData, _ := json.Marshal(map[string]interface{}{
				"timestamp":       time.Now().Format(time.RFC3339),
				"custom_model":    metrics.CustomModel,
				"actual_model":    metrics.ActualModel,
				"status":          metrics.StatusCode,
				"lookup_ms":       metrics.ModelLookupTime.Milliseconds(),
				"body_process_ms": metrics.RequestBodyTime.Milliseconds(),
				"upstream_ms":     metrics.UpstreamLatency.Milliseconds(),
				"total_ms":        metrics.TotalProcessTime.Milliseconds(),
				"streaming":       metrics.StreamingResponse,
				"path":            c.Param("path"),
			})

			log.Printf("REQUEST_METRICS: %s", logData)
		}()

		// Check for streaming request
		isStreaming := false
		if streamParam := c.Request.URL.Query().Get("stream"); streamParam == "true" {
			isStreaming = true
		}

		// Handle JSON request body
		requestBody, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Failed to read request"})
			return
		}
		c.Request.Body.Close()

		// Parse the request to extract model name
		var requestData map[string]interface{}
		if err := json.Unmarshal(requestBody, &requestData); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON format"})
			return
		}

		// Look for model in the parsed data
		modelValue, modelExists := requestData["model"]
		modelLookupStart := time.Now()

		if !modelExists {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Model field required"})
			return
		}

		customModelName, ok := modelValue.(string)
		if !ok || customModelName == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Model must be a non-empty string"})
			return
		}

		// Input validation for model parameters
		if maxTokens, exists := requestData["max_tokens"]; exists {
			if tokens, ok := maxTokens.(float64); !ok || tokens <= 0 || tokens > 32000 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid max_tokens value"})
				return
			}
		}

		if temperature, exists := requestData["temperature"]; exists {
			if temp, ok := temperature.(float64); !ok || temp < 0 || temp > 2 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid temperature value (must be between 0 and 2)"})
				return
			}
		}

		// Validate messages array if it exists
		if messages, exists := requestData["messages"]; exists {
			messagesArray, ok := messages.([]interface{})
			if !ok || len(messagesArray) == 0 {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Messages must be a non-empty array"})
				return
			}

			// Check each message format
			for i, msg := range messagesArray {
				msgObj, ok := msg.(map[string]interface{})
				if !ok {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid message format at index %d", i)})
					return
				}

				role, hasRole := msgObj["role"].(string)
				content, hasContent := msgObj["content"]

				log.Printf("CONTENT: %s", content)

				if !hasRole || (role != "user" && role != "assistant" && role != "system") {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Invalid or missing role at message index %d", i)})
					return
				}

				if !hasContent {
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("Missing content at message index %d", i)})
					return
				}
			}
		}

		// Look up model info
		modelInfo, exists := registry.GetModelInfo(customModelName)
		metrics.ModelLookupTime = time.Since(modelLookupStart)
		metrics.CustomModel = customModelName

		if !exists {
			c.JSON(http.StatusNotFound, gin.H{"error": "Model not registered"})
			return
		}

		metrics.ActualModel = modelInfo.ActualModel

		// Replace the model name with the actual model
		requestData["model"] = modelInfo.ActualModel

		// Re-marshal the modified request
		bodyProcessStart := time.Now()
		modifiedRequestBody, err := json.Marshal(requestData)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request"})
			return
		}
		metrics.RequestBodyTime = time.Since(bodyProcessStart)

		// Build target URL
		u, err := url.Parse(modelInfo.APIURL)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Server configuration error"})
			return
		}

		u.Path = path.Join(u.Path, c.Param("path"))

		// Preserve query parameters
		u.RawQuery = c.Request.URL.RawQuery
		targetURL := u.String()

		// Create context for the request
		ctx, cancel := context.WithTimeout(c.Request.Context(), 55*time.Second)
		defer cancel()

		// Create and send the proxied request
		proxyReq, err := http.NewRequestWithContext(ctx, c.Request.Method, targetURL, strings.NewReader(string(modifiedRequestBody)))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create request"})
			return
		}

		// Copy important headers
		safeHeaders := []string{
			"Authorization", "Content-Type", "Accept", "User-Agent",
			"OpenAI-Organization", "Anthropic-Version", "X-Request-ID",
		}

		for _, header := range safeHeaders {
			if value := c.GetHeader(header); value != "" {
				proxyReq.Header.Set(header, value)
			}
		}

		// Ensure proper content type
		if proxyReq.Header.Get("Content-Type") == "" {
			proxyReq.Header.Set("Content-Type", "application/json")
		}

		// Make the upstream request
		upstreamStart := time.Now()
		resp, err := httpClient.Do(proxyReq)
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": "Upstream service unavailable"})
			return
		}
		metrics.UpstreamLatency = time.Since(upstreamStart)
		metrics.StatusCode = resp.StatusCode
		metrics.StreamingResponse = isStreaming

		// Copy response headers
		for key, values := range resp.Header {
			for _, value := range values {
				c.Writer.Header().Add(key, value)
			}
		}

		// Set the status code
		c.Writer.WriteHeader(resp.StatusCode)

		// Stream or copy the response body
		if isStreaming {
			// For streaming responses, we need to flush after each write
			flusher, ok := c.Writer.(http.Flusher)
			if !ok {
				log.Println("Streaming requested but ResponseWriter doesn't support Flush")
				io.Copy(c.Writer, resp.Body)
			} else {
				// Create a buffer for efficient reading
				buf := make([]byte, 1024)
				for {
					n, err := resp.Body.Read(buf)
					if n > 0 {
						c.Writer.Write(buf[:n])
						flusher.Flush()
					}

					if err != nil {
						break
					}
				}
			}
		} else {
			// For non-streaming, just copy the entire response
			io.Copy(c.Writer, resp.Body)
		}

		resp.Body.Close()
	})

	// Start server
	log.Println("Starting AI Model Proxy server on :8080")
	if err := r.Run(":8080"); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// isValidModelName checks if a model name contains only valid characters
func isValidModelName(name string) bool {
	if len(name) < 1 || len(name) > 64 {
		return false
	}

	// Allow alphanumeric, dash, underscore, and dot
	for _, r := range name {
		if !(('a' <= r && r <= 'z') || ('A' <= r && r <= 'Z') || ('0' <= r && r <= '9') || r == '-' || r == '_' || r == '.') {
			return false
		}
	}

	return true
}
