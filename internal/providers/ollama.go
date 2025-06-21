package providers

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"promptguard/internal/config"
)

// OllamaClient implements the Ollama provider for local models
type OllamaClient struct {
	baseURL string
	model   string
	config  map[string]interface{}
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(model string, config map[string]interface{}) (*OllamaClient, error) {
	baseURL := "http://localhost:11434" // Default Ollama URL
	if url, ok := config["base_url"].(string); ok {
		baseURL = url
	}

	return &OllamaClient{
		baseURL: baseURL,
		model:   model,
		config:  config,
	}, nil
}

// Complete executes a prompt completion using Ollama
func (c *OllamaClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	// Get temperature from config
	temperature := 0.0
	if temp, ok := c.config["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			temperature = tempFloat
		}
	}

	// Prepare request body for Ollama API
	requestBody := map[string]interface{}{
		"model":  c.model,
		"prompt": prompt,
		"options": map[string]interface{}{
			"temperature": temperature,
		},
		"stream": false,
	}

	jsonBody, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	// Make HTTP request to Ollama
	resp, err := http.Post(
		fmt.Sprintf("%s/api/generate", c.baseURL),
		"application/json",
		strings.NewReader(string(jsonBody)),
	)
	if err != nil {
		return nil, fmt.Errorf("Ollama API request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("Ollama API returned status %d", resp.StatusCode)
	}

	// Parse response
	var ollamaResp struct {
		Response string `json:"response"`
		Done     bool   `json:"done"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&ollamaResp); err != nil {
		return nil, fmt.Errorf("failed to decode Ollama response: %w", err)
	}

	// Ollama is free/local, so cost is 0
	return &Response{
		Text:     ollamaResp.Response,
		Cost:     0.0, // Local models are free
		Tokens:   len(strings.Fields(ollamaResp.Response)), // Approximate
		Provider: "ollama",
		Model:    c.model,
	}, nil
}

func (c *OllamaClient) GetName() string {
	return "ollama"
}

func (c *OllamaClient) GetModel() string {
	return c.model
}
