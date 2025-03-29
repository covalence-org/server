package spam

import (
	"log"
	"netrunner/src/internal"
	"netrunner/src/types"
)

func Run(message types.Message, model internal.Model, blockingThreshold float32) (bool, error) {
	content := message.Content

	log.Printf("Running custom firewall with content: %v", content)

	return true, nil
}
