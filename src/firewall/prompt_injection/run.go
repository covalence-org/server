package promptInjection

import (
	"log"
	"netrunner/src/internal"
	textClassification "netrunner/src/internal/text_classification"
	"netrunner/src/types"
	"netrunner/src/utils"
	"strings"
)

var (
	safeLabels = []string{"safe", "neutral", "benign"}
)

func Run(message types.Message, model internal.Model, blockingThreshold float32) (bool, error) {
	content := message.Content

	textClassificationRequest, err := textClassification.NewRequest(model, content)
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
		if utils.Contains(safeLabels, strings.ToLower(label)) {
			log.Printf("Skipping safe label: %v", label)
			continue // Skip safe labels (we only care about unsafe labels)
		}
		probability := response.Probabilities[i]
		if probability > blockingThreshold {
			log.Printf("Blocking request due to high confidence label: %v. Confidence: %v > %v", label, probability, blockingThreshold)
			return false, nil
		}
	}

	return true, nil
}
