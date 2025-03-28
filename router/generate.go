package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"netrunner/register"
	"netrunner/request"
	"netrunner/user"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Generate(c *gin.Context, registry *register.Registry, httpClient *http.Client, hook func(*gin.Context, *user.GeneratePayload) (int, error)) {

	// ========================= Request Metrics =========================

	metrics := request.Metrics{
		StartTime: time.Now(),
	}

	// Defer function to log metrics
	defer func() {
		metrics.TotalProcessTime = time.Since(metrics.StartTime)

		logData, _ := json.Marshal(map[string]interface{}{
			"timestamp":       time.Now().Format(time.RFC3339),
			"name":            metrics.Name.String(),
			"model":           metrics.Model.String(),
			"status":          metrics.StatusCode,
			"lookup_ms":       metrics.ModelLookupTime.Milliseconds(),
			"body_process_ms": metrics.RequestBodyTime.Milliseconds(),
			"upstream_ms":     metrics.UpstreamLatency.Milliseconds(),
			"total_ms":        metrics.TotalProcessTime.Milliseconds(),
			"streaming":       metrics.StreamingResponse,
			"path":            c.Param("path"),
		})

		log.Printf("request_metrics: %s\n", logData)
		fmt.Println()
	}()

	// ========================= Read Request =========================
	fmt.Println()
	log.Printf("reading request made to %s 🚀\n", c.Param("path"))
	fmt.Println()

	var generateRequest request.GenerateRequest
	if err := c.ShouldBindJSON(&generateRequest); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ========================= Parse Request =========================
	log.Printf("parsing request 🔍\n")
	fmt.Println()

	modelLookupStart := time.Now()

	generatePayload, err := request.ParseGenerateRequest(generateRequest, registry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metrics.ModelLookupTime = time.Since(modelLookupStart)
	metrics.Name = generatePayload.Model.Name
	metrics.Model = generatePayload.Model.Model

	// ========================= Run Hook ===========================

	if hook != nil {
		log.Println("entering hook function ✅")
		fmt.Println()
		if status, err := hook(c, &generatePayload); err != nil {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
	} else {
		log.Println("no hook function provided ❌")
		fmt.Println()
	}

	// ========================= Build Request =========================
	log.Printf("building request 🏗️\n")
	fmt.Println()

	bodyProcessStart := time.Now()
	requestData := generatePayload.ToMap()
	modifiedRequestBody, err := json.Marshal(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request to json"})
		return
	}
	metrics.RequestBodyTime = time.Since(bodyProcessStart)

	// Build target URL
	u, err := url.Parse(generatePayload.Model.ApiUrl.String())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "server configuration error"})
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
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create request"})
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
	log.Printf("making request to %s 🚀\n\n", targetURL)
	upstreamStart := time.Now()
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream service unavailable"})
		return
	}
	metrics.UpstreamLatency = time.Since(upstreamStart)
	metrics.StatusCode = resp.StatusCode
	metrics.StreamingResponse = generatePayload.IsStreaming

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// Set the status code
	c.Writer.WriteHeader(resp.StatusCode)

	// Stream or copy the response body
	if generatePayload.IsStreaming {
		// For streaming responses, we need to flush after each write
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			log.Println("streaming requested but responsewriter doesn't support flush")
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
}
