package request

import (
	"errors"
	"netrunner/register"
	"netrunner/types"
	"netrunner/user"
)

// GenerateRequest represents the incoming JSON request
type GenerateRequest struct {
	Name        string        `json:"model" binding:"required"`
	IsStreaming bool          `json:"stream"`
	MaxTokens   *int          `json:"max_tokens"`  // Pointer to make it optional
	Temperature *float32      `json:"temperature"` // Pointer to make it optional
	Messages    []interface{} `json:"messages" binding:"required"`
}

func ParseGenerateRequest(generateRequest GenerateRequest, registry *register.Registry) (user.GeneratePayload, error) {
	// Look for model in the parsed data
	name, err := types.NewName(generateRequest.Name)
	if err != nil {
		return user.GeneratePayload{}, err
	}

	// Look up model info
	modelInfo, exists := registry.GetInfo(name.String())
	if !exists {
		return user.GeneratePayload{}, errors.New("model not found")
	}

	// Initialize the payload with required fields
	payload := user.GeneratePayload{
		Model:       modelInfo,
		IsStreaming: generateRequest.IsStreaming,
	}

	// Handle optional parameters
	if generateRequest.MaxTokens != nil {
		maxTokens, err := types.NewMaxTokens(*generateRequest.MaxTokens)
		if err != nil {
			return user.GeneratePayload{}, err
		}
		payload.MaxTokens = &maxTokens
	}

	if generateRequest.Temperature != nil {
		temp, err := types.NewTemperature(*generateRequest.Temperature)
		if err != nil {
			return user.GeneratePayload{}, err
		}
		payload.Temperature = &temp
	}

	// Validate messages array
	if len(generateRequest.Messages) == 0 {
		return user.GeneratePayload{}, errors.New("messages must be a non-empty array")
	}

	messagesArray := []types.Message{}
	// Check each message format
	for _, msg := range generateRequest.Messages {
		message, err := types.NewMessageFromJson(msg)
		if err != nil {
			return user.GeneratePayload{}, err
		}
		messagesArray = append(messagesArray, message)
	}
	payload.Messages = messagesArray

	return payload, nil
}
