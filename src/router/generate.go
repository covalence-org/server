package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"netrunner/src/firewall"
	"netrunner/src/register"
	"netrunner/src/request"
	"netrunner/src/utils"
	"path"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Generate(c *gin.Context, registry *register.Registry, httpClient *http.Client, firewallConfig *firewall.Config, hook func(*gin.Context, *request.Generate, *firewall.Config) (int, error)) {

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

		utils.BoxLog(fmt.Sprintf("request_metrics: %s", logData))
	}()

	// ========================= Read Request =========================
	utils.BoxLog(fmt.Sprintf("reading request made to %s 🚀", c.Param("path")))

	var generateRequestRaw request.RawGenerate
	if err := c.ShouldBindJSON(&generateRequestRaw); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ========================= Parse Request =========================
	utils.BoxLog("parsing request 🔍")

	modelLookupStart := time.Now()

	generateRequest, err := generateRequestRaw.Parse(registry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metrics.ModelLookupTime = time.Since(modelLookupStart)
	metrics.Name = generateRequest.Model.Name
	metrics.Model = generateRequest.Model.Model

	// ========================= Run Hook ===========================

	if hook != nil {
		utils.BoxLog("entering hook function ✅")
		if status, err := hook(c, &generateRequest, firewallConfig); err != nil {
			c.JSON(status, gin.H{"error": err.Error()})
			return
		}
	} else {
		utils.BoxLog("no hook function provided ❌")
	}

	// ========================= Build Request =========================
	utils.BoxLog("building request 🏗️")

	bodyProcessStart := time.Now()
	requestData := generateRequest.ToMap()
	modifiedRequestBody, err := json.Marshal(requestData)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to process request to json"})
		return
	}
	metrics.RequestBodyTime = time.Since(bodyProcessStart)

	// Build target URL
	u, err := url.Parse(generateRequest.Model.APIURL.String())
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
	utils.BoxLog(fmt.Sprintf("making request to %s 🚀", targetURL))
	upstreamStart := time.Now()
	resp, err := httpClient.Do(proxyReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "upstream service unavailable"})
		return
	}
	metrics.UpstreamLatency = time.Since(upstreamStart)
	metrics.StatusCode = resp.StatusCode
	metrics.StreamingResponse = generateRequest.IsStreaming

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Writer.Header().Add(key, value)
		}
	}

	// Set the status code
	c.Writer.WriteHeader(resp.StatusCode)

	// Stream or copy the response body
	if generateRequest.IsStreaming {
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
