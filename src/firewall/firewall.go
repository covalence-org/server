package firewall

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"netrunner/src/audit"
	"netrunner/src/db/postgres"
	custom "netrunner/src/firewall/custom"
	hallucinationRisk "netrunner/src/firewall/hallucination_risk"
	maliciousIntent "netrunner/src/firewall/malicious_intent"
	obfuscation "netrunner/src/firewall/obfuscation"
	policyViolation "netrunner/src/firewall/policy_violation"
	promptInjection "netrunner/src/firewall/prompt_injection"
	sensitiveData "netrunner/src/firewall/sensitive_data"
	spam "netrunner/src/firewall/spam"
	"netrunner/src/request"
	"netrunner/src/utils"

	"netrunner/src/types"

	"github.com/gin-gonic/gin"
)

func (f Firewall) Apply(messages []types.Message) (bool, error) {
	message := messages[len(messages)-1]
	if f.Enabled {
		log.Printf(":::: running %s firewall ::::", f.Type.String())
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

func HookFirewalls(c *gin.Context, payload *request.Generate, config *Config) (int, error) {
	log.Printf("firewall hook called with payload")
	db := c.MustGet("db").(*postgres.DB)
	requestID := c.MustGet("requestID").(string)

	// Check latest message
	for _, firewall := range config.Firewalls {
		res, err := firewall.Apply(payload.Messages)
		if err != nil {
			return http.StatusInternalServerError, err
		}

		// Log the firewall event
		utils.BoxLog(fmt.Sprintf("audit loggging: firewall event %s 📝", firewall.Type.String()))

		fe := audit.FirewallEvent{
			RequestID:     requestID,
			FirewallID:    firewall.ID.String(),
			FirewallType:  firewall.Type.String(),
			Blocked:       !res,
			BlockedReason: "",
			RiskScore:     0.0,
		}

		audit.LogFirewallEvent(c, fe, db)

		if !res {
			return http.StatusForbidden, errors.New("request rejected: blocked by firewall")
		}
	}

	return http.StatusOK, nil
}
