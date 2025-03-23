package hook

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"netrunner/model"
	"netrunner/types"

	"github.com/gin-gonic/gin"
)

func checkMessage(message types.Message) bool {
	return false
}

func Firewall(c *gin.Context, payload *model.GeneratePayload) (int, error) {
	log.Printf("Firewall hook called with payload")
	fmt.Println()

	// Check latest message
	if !checkMessage(payload.Messages[len(payload.Messages)-1]) {
		return http.StatusForbidden, errors.New("request rejected: blocked by firewall")
	}

	return http.StatusOK, nil
}
