package hook

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	internal "netrunner/internal/text_classification"
	"netrunner/types"
	"netrunner/user"

	"github.com/gin-gonic/gin"
)

func contains(slice []string, item string) bool {
	for _, v := range slice {
		if v == item {
			return true
		}
	}
	return false
}

var (
	model      = "meta-llama/Prompt-Guard-86M"
	safeLabels = []string{"safe", "neutral", "benign"}
)

func isMessageSafe(message types.Message, config *FirewallConfig) (bool, error) {
	content := message.Content

	textClassificationRequest, err := internal.NewTextClassificationRequest(model, content)
	if err != nil {
		log.Printf("Error creating text classification request: %v", err)
		return false, err
	}

	response, err := textClassificationRequest.Run()
	if err != nil {
		log.Printf("Error running text classification request: %v", err)
		return false, err
	}

	log.Printf("Text classification response: %v", response)

	// Now, check if the response values are above the threshold. if they are, block the request
	for i, label := range response.Labels {
		if contains(safeLabels, label) {
			continue
		}
		probability := response.Probabilities[i]
		if probability > config.BlockingThreshold {
			log.Printf("Blocking request due to high confidence label: %v", label)
			return false, nil
		}
	}

	return true, nil
}

func Firewall(c *gin.Context, payload *user.GeneratePayload, config *FirewallConfig) (int, error) {
	log.Printf("Firewall hook called with payload")
	fmt.Println()

	// Check latest message
	if config.Enabled {
		res, err := isMessageSafe(payload.Messages[len(payload.Messages)-1], config)
		if err != nil {
			return http.StatusInternalServerError, err
		}
		if !res {
			return http.StatusForbidden, errors.New("request rejected: blocked by firewall")
		}
	}

	return http.StatusOK, nil
}
