package tests

import (
	"context"
	"fmt"
	"testing"
	"time"
)

// Simulated classifier check
func runClassifier(ctx context.Context, input string, resultChan chan<- bool) {
	// Simulate processing time
	time.Sleep(500 * time.Millisecond)

	// Let's say it’s invalid input
	isValid := true

	select {
	case resultChan <- isValid:
	case <-ctx.Done():
		// If context is canceled, just return
		return
	}
}

// Simulated streaming API call
func runStreamingAPI(ctx context.Context, input string, doneChan chan<- string) {
	for i := 0; i < 5; i++ {
		select {
		case <-ctx.Done():
			fmt.Println("Streaming cancelled early!")
			return
		default:
			time.Sleep(300 * time.Millisecond)
			fmt.Printf("Streaming chunk %d...\n", i+1)
		}
	}

	select {
	case doneChan <- "Streaming complete!":
	case <-ctx.Done():
	}
}

func TestSpeedup(t *testing.T) {
	input := "some user input"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	classifierChan := make(chan bool, 1)
	streamingChan := make(chan string, 1)

	go runClassifier(ctx, input, classifierChan)
	go runStreamingAPI(ctx, input, streamingChan)

	var classifierOK bool
	var streamDone bool

	for !streamDone {
		select {
		case valid := <-classifierChan:
			classifierOK = valid
			if !valid {
				fmt.Println("Classifier rejected input. Cancelling stream...")
				cancel() // kills the streaming request
				return
			} else {
				fmt.Println("Classifier accepted input. Continuing streaming...")
			}
		case streamResp := <-streamingChan:
			streamDone = true
			if classifierOK {
				fmt.Println("✅ Success:", streamResp)
			} else {
				fmt.Println("⚠️ Streaming finished, but classifier was false. This shouldn't happen!")
			}
		}
	}
}
