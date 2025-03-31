package main

import (
	"context"
	"fmt"
	"log"

	"netrunner/src/audit"
	"netrunner/src/db/postgres"
)

func main() {
	ctx := context.Background()

	// Connect to database
	db, err := postgres.New(ctx, "user=alialh dbname=netrunner_dev sslmode=disable")
	if err != nil {
		log.Fatal("Database connection failed:", err)
	}
	defer db.Close()

	// Generate UUIDs
	userID := audit.NewUUID()
	apiKeyID := audit.NewUUID()

	request := audit.Request{
		UserID:     userID,
		APIKeyID:   apiKeyID,
		Model:      "gpt-4",
		TargetURL:  "https://api.openai.com/v1/chat/completions",
		Messages:   []map[string]interface{}{{"role": "user", "content": "Tell me something cool"}},
		Parameters: map[string]interface{}{"temperature": 0.7},
		ClientIP:   "127.0.0.1",
	}

	// Log a request
	requestID, err := audit.LogRequest(ctx, request, db)
	if err != nil {
		log.Fatal("Failed to log request:", err)
	}
	fmt.Println("Request logged:", requestID)

	response := audit.Response{
		RequestID:    requestID,
		Response:     "Here's something cool: Fire is hot.",
		LatencyMs:    150,
		InputTokens:  8,
		OutputTokens: 12,
		TotalTokens:  20,
	}
	// Log a response
	err = audit.LogResponse(ctx, response, db)
	if err != nil {
		log.Fatal("Failed to log response:", err)
	}
	fmt.Println("Response logged")

	firewallEvent := audit.FirewallEvent{
		RequestID:     requestID,
		FirewallID:    "NO_HATE_SPEECH",
		FirewallType:  "triggered",
		Blocked:       false,
		BlockedReason: "Hate speech detected.",
		RiskScore:     0.12,
	}

	// Log a firewall event
	err = audit.LogFirewallEvent(ctx, firewallEvent, db)
	if err != nil {
		log.Fatal("Failed to log firewall event:", err)
	}
	fmt.Println("Firewall event logged")

	// Get trace
	trace, err := audit.GetTrace(ctx, requestID, db)
	if err != nil {
		log.Fatal("Failed to get trace:", err)
	}

	// Print trace info
	fmt.Println("\nRequest Trace:")
	fmt.Printf("ID: %s\n", trace.RequestID)
	fmt.Printf("Model: %s\n", trace.Model)
	fmt.Printf("Messages: %v\n", trace.Messages)
	fmt.Printf("Response: %s\n", trace.Response)

	if len(trace.FirewallInfo) > 0 {
		fmt.Println("\nFirewall Events:")
		for i, event := range trace.FirewallInfo {
			fmt.Printf("  Event %d: %s - %v, %s \n",
				i+1, event.FirewallID, event.Blocked, event.BlockedReason)
		}
	}
}
