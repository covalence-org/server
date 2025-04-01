package router

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"netrunner/src/audit"
	"netrunner/src/db/postgres"
	"netrunner/src/firewall"
	"netrunner/src/register"
	"netrunner/src/request"
	"netrunner/src/utils"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

func Generate(c *gin.Context, firewallConfig *firewall.Config, hook func(*gin.Context, *request.Generate, *firewall.Config) (int, error)) {

	registry := c.MustGet("registry").(*register.Registry)
	httpClient := c.MustGet("httpClient").(*http.Client)
	db := c.MustGet("db").(*postgres.DB)

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

	// ========================= Read & Parse Request =========================

	utils.BoxLog(fmt.Sprintf("reading & parsing request made to %s 🚀", c.Param("path")))

	modelLookupStart := time.Now()

	generateRequest, err := request.ParseGenerate(c, registry)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// ========================= Audit: Log Request =========================

	utils.BoxLog("audit loggging: request 📝")

	auditRequest := generateRequest.ToAuditRequest()
	requestID, err := audit.LogRequest(c.Request.Context(), auditRequest, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log request"})
		return
	}

	// Set RequestID
	c.Set("requestID", requestID)

	// ========================= Init Metrics =========================

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

	// Create context for the request
	ctx, cancel := context.WithTimeout(c.Request.Context(), 55*time.Second)
	defer cancel()

	// Create and send the proxied request
	proxyReq, err := http.NewRequestWithContext(ctx, c.Request.Method, generateRequest.TargetURL.String(), strings.NewReader(string(modifiedRequestBody)))
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
	utils.BoxLog(fmt.Sprintf("making request to %s 🚀", generateRequest.TargetURL.String()))
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
	var responseBody []byte
	if generateRequest.IsStreaming {
		// For streaming responses, we need to flush after each write
		flusher, ok := c.Writer.(http.Flusher)
		if !ok {
			log.Println("streaming requested but responsewriter doesn't support flush")
			responseBody, _ = io.ReadAll(resp.Body)
			c.Writer.Write(responseBody)
		} else {
			// Create a buffer for efficient reading
			buf := make([]byte, 1024)
			for {
				n, err := resp.Body.Read(buf)
				if n > 0 {
					c.Writer.Write(buf[:n])
					flusher.Flush()
					responseBody = append(responseBody, buf[:n]...)
				}

				if err != nil {
					break
				}
			}
		}
	} else {
		// For non-streaming, just copy the entire response
		responseBody, _ = io.ReadAll(resp.Body)
		// Log the response body for debugging purposes
		utils.BoxLog(fmt.Sprintf("response body: %s", string(responseBody)))
	}

	var response map[string]interface{}
	err = json.Unmarshal(responseBody, &response)
	if err != nil {
		c.JSON(http.StatusBadGateway, gin.H{"error": "response couldn't be parsed"})
		return
	}

	// Log the response body for debugging purposes
	utils.BoxLog(fmt.Sprintf("response body: %v", response))

	// Audit log response
	utils.BoxLog("audit loggging: response 📝")
	auditResponse := audit.Response{
		RequestID: requestID,
		Response:  response,
		LatencyMs: metrics.UpstreamLatency.Milliseconds(),
	}
	err = audit.LogResponse(c.Request.Context(), auditResponse, db)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to log response"})
		return
	}

	resp.Body.Close()
}
