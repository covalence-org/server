package firewall

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	custom "netrunner/firewall/custom"
	hallucinationRisk "netrunner/firewall/hallucination_risk"
	maliciousIntent "netrunner/firewall/malicious_intent"
	obfuscation "netrunner/firewall/obfuscation"
	policyViolation "netrunner/firewall/policy_violation"
	promptInjection "netrunner/firewall/prompt_injection"
	sensitiveData "netrunner/firewall/sensitive_data"
	spam "netrunner/firewall/spam"

	"netrunner/types"
	"netrunner/user"

	"github.com/gin-gonic/gin"
)

func (f Firewall) Apply(message types.Message) (bool, error) {
	if f.Enabled {
		switch f.Type.String() {
		case "prompt-injection":
			return promptInjection.Run(message, f.Model, f.BlockingThreshold)
		case "malicious-intent":
			return maliciousIntent.Run(message, f.Model, f.BlockingThreshold)
		case "custom":
			return custom.Run(message, f.Model, f.BlockingThreshold)
		case "policy-violation":
			return policyViolation.Run(message, f.Model, f.BlockingThreshold)
		case "sensitive-data":
			return sensitiveData.Run(message, f.Model, f.BlockingThreshold)
		case "hallucination-risk":
			return hallucinationRisk.Run(message, f.Model, f.BlockingThreshold)
		case "spam":
			return spam.Run(message, f.Model, f.BlockingThreshold)
		case "obfuscation":
			return obfuscation.Run(message, f.Model, f.BlockingThreshold)
		default:
			return true, nil
		}
	}

	return true, nil
}

func HookFirewalls(c *gin.Context, payload *user.GeneratePayload, config *Config) (int, error) {
	log.Printf("Firewall hook called with payload")
	fmt.Println()

	// Check latest message
	for _, firewall := range config.Firewalls {
		res, err := firewall.Apply(payload.Messages[len(payload.Messages)-1])
		if err != nil {
			return http.StatusInternalServerError, err
		}

		if !res {
			return http.StatusForbidden, errors.New("request rejected: blocked by firewall")
		}
	}

	return http.StatusOK, nil
}
