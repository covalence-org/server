package textClassification

import (
	"bytes"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"netrunner/src/internal"
	"netrunner/src/types"
)

var (
	API_URL = "http://localhost:8000/api/v1/models/text/classification"
)

// GeneratePayload stores information about a generation request
type Request struct {
	Model       internal.Model
	Text        string
	Temperature *types.Temperature // Now a pointer to make it optional
}

type Response struct {
	Probabilities []float32 `json:"probabilities"`
	Labels        []string  `json:"labels"`
	ModelID       string    `json:"model_id"`
}

func NewRequest(model internal.Model, text string) (Request, error) {
	return Request{
		Model: model,
		Text:  text,
	}, nil
}

func (m Request) ToMap() map[string]interface{} {
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

func (m Request) Run() (Response, error) {
	// Start with required parameters
	requestMap := m.ToMap()
	url := API_URL

	// Marshal the requestMap into JSON
	jsonData, err := json.Marshal(requestMap)
	if err != nil {
		return Response{}, errors.New("failed to marshal request map: " + err.Error())
	}

	log.Printf("sending request to %s", url)

	// Create a new HTTP POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return Response{}, errors.New("failed to create HTTP request: " + err.Error())
	}

	// Set the appropriate headers
	req.Header.Set("Content-Type", "application/json")

	// Execute the HTTP request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		// Log the response for debugging purposes
		return Response{}, errors.New("failed to execute HTTP request: " + err.Error())
	}
	defer resp.Body.Close()

	// Handle the response (optional, depending on your use case)
	if resp.StatusCode != http.StatusOK {
		return Response{}, errors.New("received non-OK HTTP status: " + resp.Status)
	}

	// Decode the response body into a TextClassificationResponse struct
	var response Response
	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return Response{}, errors.New("failed to decode response body: " + err.Error())
	}

	return response, nil
}
