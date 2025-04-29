package router

import (
	"context"
	"covalence/src/request"
	"covalence/src/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func PassThrough(c *gin.Context) {

	httpClient := c.MustGet("httpClient").(*http.Client)

	// ========================= Request Metrics =========================

	metrics := request.Metrics{
		StartTime: time.Now(),
	}

	// Defer function to log metrics
	defer func() {
		metrics.TotalProcessTime = time.Since(metrics.StartTime)

		logData, _ := json.Marshal(map[string]interface{}{
			"timestamp":              time.Now().Format(time.RFC3339),
			"name":                   metrics.Name.String(),
			"model":                  metrics.Model.String(),
			"status":                 metrics.StatusCode,
			"request_preparation_ms": metrics.RequestPreparationTime.Milliseconds(),
			"hook_time_ms":           metrics.HookTime.Milliseconds(),
			"body_process_ms":        metrics.RequestBodyTime.Milliseconds(),
			"upstream_ms":            metrics.UpstreamLatency.Milliseconds(),
			"total_ms":               metrics.TotalProcessTime.Milliseconds(),
			"streaming":              metrics.StreamingResponse,
			"path":                   c.Param("path"),
		})

		utils.BoxLog(fmt.Sprintf("request_metrics: %s", logData))
	}()

	// ========================= Read & Parse Request =========================

	utils.BoxLog(fmt.Sprintf("reading & parsing request made to %s 🚀", c.Param("path")))

	// Create context for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 55*time.Second)
	defer cancel()

	targetURL, err := url.JoinPath("https://api.openai.com/v1", c.Param("path"))
	if err != nil {
		log.Printf("Error joining URL path: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to construct target URL"})
		return
	}

	// Create and send the proxied request
	// Read the request body
	requestBodyBytes, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read request body"})
		return
	}
	c.Request.Body.Close() // Close the original body

	// Create a new reader for the proxy request
	bodyReader := strings.NewReader(string(requestBodyBytes))

	// Log the request body for debugging purposes
	utils.BoxLog(fmt.Sprintf("request body: %s", string(requestBodyBytes)))

	// Create and send the proxied request
	proxyReq, err := http.NewRequestWithContext(ctx, c.Request.Method, targetURL, bodyReader)
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

	utils.BoxLog(fmt.Sprintf("making request to %s 🚀", targetURL))
	upstreamStart := time.Now()
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream service unavailable", "message": err.Error()})
		return
	}
	metrics.UpstreamLatency = time.Since(upstreamStart)
	metrics.StatusCode = resp.StatusCode

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// Set the status code
	c.Writer.WriteHeader(resp.StatusCode)

	responseBody, _ := io.ReadAll(resp.Body)

	// Write to body
	_, err = c.Writer.Write(responseBody)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to write response"})
		return
	}
	// Flush the response writer to ensure all data is sent
	c.Writer.Flush()

	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "response couldn't be parsed"})
		return
	}

	// Log the response body for debugging purposes
	utils.BoxLog(fmt.Sprintf("response body: %v", response))

	resp.Body.Close()
}
