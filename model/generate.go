package model

import "netrunner/types"

// GeneratePayload stores information about a generation request
type GeneratePayload struct {
	Model       Model
	IsStreaming bool
	MaxTokens   *types.MaxTokens   // Now a pointer to make it optional
	Temperature *types.Temperature // Now a pointer to make it optional
	Messages    []types.Message
}

func (m GeneratePayload) ToMap() map[string]interface{} {
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
