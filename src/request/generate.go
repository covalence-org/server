package request

import (
	"errors"
	"net/url"
	"netrunner/src/audit"
	"netrunner/src/register"
	"netrunner/src/types"
	"netrunner/src/user"
	"path"
	"strings"

	"github.com/gin-gonic/gin"
)

// GenerateRequest represents the incoming JSON request
type rawGenerate struct {
	Name        string        `json:"model" binding:"required"`
	IsStreaming bool          `json:"stream"`
	MaxTokens   *int          `json:"max_tokens"`  // Pointer to make it optional
	Temperature *float32      `json:"temperature"` // Pointer to make it optional
	Messages    []interface{} `json:"messages" binding:"required"`
}

// GeneratePayload stores information about a generation request
type Generate struct {
	User        user.User
	Model       user.Model
	TargetURL   url.URL
	IsStreaming bool
	MaxTokens   *types.MaxTokens   // Now a pointer to make it optional
	Temperature *types.Temperature // Now a pointer to make it optional
	Messages    []types.Message
	ClientIP    string
}

func ParseGenerate(c *gin.Context, registry *register.Registry) (Generate, error) {

	var rg rawGenerate
	if err := c.ShouldBindJSON(&rg); err != nil {
		return Generate{}, err
	}

	// Read API key from Authorization header
	authHeader := c.GetHeader("Authorization")
	// Expecting format: "Bearer <apikey>"
	if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
		return Generate{}, errors.New("missing or invalid Authorization header")
	}
	apiKey := strings.TrimPrefix(authHeader, "Bearer ")
	apiKey = strings.TrimSpace(apiKey)

	// Look up user by API key
	user, err := user.GetUserByAPIKey(apiKey)
	if err != nil {
		return Generate{}, errors.New("Invalid API key")
	}

	// Look for model in the parsed data
	name, err := types.NewName(rg.Name)
	if err != nil {
		return Generate{}, err
	}

	// Look up model info
	modelInfo, exists := registry.GetInfo(name.String())
	if !exists {
		return Generate{}, errors.New("model not found")
	}

	// Get the client IP address
	clientIP := c.RemoteIP()

	// Build target URL
	urlRaw := path.Join(modelInfo.APIURL.String(), c.Param("path"))
	targetURL, err := url.Parse(urlRaw)
	if err != nil {
		return Generate{}, err
	}

	// Build messages array
	if len(rg.Messages) == 0 {
		return Generate{}, errors.New("messages must be a non-empty array")
	}

	messagesArray := []types.Message{}
	// Check each message format
	for _, msg := range rg.Messages {
		message, err := types.NewMessageFromJson(msg)
		if err != nil {
			return Generate{}, err
		}
		messagesArray = append(messagesArray, message)
	}

	// Initialize the payload with required fields
	payload := Generate{
		Model:       modelInfo,
		IsStreaming: rg.IsStreaming,
		TargetURL:   *targetURL,
		ClientIP:    clientIP,
		Messages:    messagesArray,
		User:        user,
	}

	// Handle optional parameters
	if rg.MaxTokens != nil {
		maxTokens, err := types.NewMaxTokens(*rg.MaxTokens)
		if err != nil {
			return Generate{}, err
		}
		payload.MaxTokens = &maxTokens
	}

	if rg.Temperature != nil {
		temp, err := types.NewTemperature(*rg.Temperature)
		if err != nil {
			return Generate{}, err
		}
		payload.Temperature = &temp
	}

	return payload, nil
}

func (m Generate) ToMap() map[string]interface{} {
	// Start with required parameters
	requestMap := map[string]interface{}{
		"model":    m.Model.Model.String(),
		"messages": make([]map[string]string, len(m.Messages)),
		"stream":   m.IsStreaming,
	}

	// Convert messages
	for i, msg := range m.Messages {
		requestMap["messages"].([]map[string]string)[i] = msg.ToMap()
	}

	// Only add optional parameters if they were explicitly set
	if m.MaxTokens != nil {
		requestMap["max_tokens"] = m.MaxTokens.Int()
	}

	if m.Temperature != nil {
		requestMap["temperature"] = m.Temperature.Float32()
	}

	return requestMap
}

func (m Generate) ToAuditRequest() audit.Request {

	endpoint := "/v1/generate"

	parameters := map[string]interface{}{
		"stream":      m.IsStreaming,
		"max_tokens":  m.MaxTokens,
		"temperature": m.Temperature,
	}

	var messages []map[string]interface{}
	for _, message := range m.Messages {
		msgMap := message.ToMap()
		// Convert map[string]string to map[string]interface{}
		interfaceMap := make(map[string]interface{})
		for k, v := range msgMap {
			interfaceMap[k] = v
		}
		messages = append(messages, interfaceMap)
	}

	return audit.Request{
		UserID:     m.User.ID.String(),
		APIKeyID:   m.User.APIKeyID.String(),
		Model:      m.Model.Model.String(),
		TargetURL:  endpoint,
		Inputs:     messages,
		Parameters: parameters,
		ClientIP:   m.ClientIP,
	}
}
