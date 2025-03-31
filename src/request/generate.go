package request

import (
	"errors"
	"netrunner/src/register"
	"netrunner/src/types"
	"netrunner/src/user"

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
	Model       user.Model
	IsStreaming bool
	MaxTokens   *types.MaxTokens   // Now a pointer to make it optional
	Temperature *types.Temperature // Now a pointer to make it optional
	Messages    []types.Message
}

func ParseGenerate(c *gin.Context, registry *register.Registry) (Generate, error) {

	var rg rawGenerate
	if err := c.ShouldBindJSON(&rg); err != nil {
		return Generate{}, err
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

	// Initialize the payload with required fields
	payload := Generate{
		Model:       modelInfo,
		IsStreaming: rg.IsStreaming,
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

	// Validate messages array
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
	payload.Messages = messagesArray

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
