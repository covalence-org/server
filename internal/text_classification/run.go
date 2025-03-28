package internal

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"netrunner/internal"
	"netrunner/types"
)

// GeneratePayload stores information about a generation request
type TextClassificationRequest struct {
	Model       internal.Model
	Text        string
	Temperature *types.Temperature // Now a pointer to make it optional
}

type TextClassificationResponse struct {
	Probabilities []float32 `json:"probabilities"`
	Labels        []string  `json:"labels"`
	ModelId       string    `json:"model_id"`
}

func NewTextClassificationRequest(modelRaw string, text string) (TextClassificationRequest, error) {
	model, err := internal.GetModel(modelRaw)
	if err != nil {
		return TextClassificationRequest{}, err
	}

	return TextClassificationRequest{
		Model: model,
		Text:  text,
	}, nil
}

func (m TextClassificationRequest) ToMap() map[string]interface{} {
	// Start with required parameters
	requestMap := map[string]interface{}{
		"model": m.Model.Model.String(),
		"text":  m.Text,
	}

	if m.Temperature != nil {
		requestMap["temperature"] = m.Temperature.Float32()
	}

	return requestMap
}

func (m TextClassificationRequest) Run() (TextClassificationResponse, error) {
	// Start with required parameters
	requestMap := m.ToMap()
	url := m.Model.ApiUrl.String()

	// Marshal the requestMap into JSON
	jsonData, err := json.Marshal(requestMap)
	if err != nil {
		return TextClassificationResponse{}, errors.New("failed to marshal request map: " + err.Error())
	}

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return TextClassificationResponse{}, errors.New("failed to create HTTP request: " + err.Error())
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return TextClassificationResponse{}, errors.New("failed to execute HTTP request: " + err.Error())
	}
	defer resp.Body.Close()

	// Handle the response (optional, depending on your use case)
	if resp.StatusCode != http.StatusOK {
		return TextClassificationResponse{}, errors.New("received non-OK HTTP status: " + resp.Status)
	}

	// Decode the response body into a TextClassificationResponse struct
	var response TextClassificationResponse
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return TextClassificationResponse{}, errors.New("failed to decode response body: " + err.Error())
	}

	return response, nil
}
