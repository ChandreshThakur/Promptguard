package providers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"github.com/sashabaranov/go-openai"
	"promptgaurd/internal/config"
)

// Response represents a provider response
type Response struct {
	Text     string  `json:"text"`
	Cost     float64 `json:"cost"`
	Tokens   int     `json:"tokens"`
	Provider string  `json:"provider"`
	Model    string  `json:"model"`
}

// Client interface for LLM providers
type Client interface {
	Complete(ctx context.Context, prompt string) (*Response, error)
	GetName() string
	GetModel() string
}

// NewClient creates a new provider client
func NewClient(provider *config.Provider) (Client, error) {
	parts := strings.SplitN(provider.ID, ":", 2)
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid provider ID format: %s (expected provider:model)", provider.ID)
	}

	providerName := parts[0]
	model := parts[1]

	switch providerName {
	case "openai":
		return NewOpenAIClient(model, provider.Config)
	case "anthropic":
		return NewAnthropicClient(model, provider.Config)
	case "mistral":
		return NewMistralClient(model, provider.Config)
	case "ollama":
		return NewOllamaClient(model, provider.Config)
	default:
		return nil, fmt.Errorf("unsupported provider: %s", providerName)
	}
}

// OpenAIClient implements the OpenAI provider
type OpenAIClient struct {
	client *openai.Client
	model  string
	config map[string]interface{}
}

// NewOpenAIClient creates a new OpenAI client
func NewOpenAIClient(model string, config map[string]interface{}) (*OpenAIClient, error) {
	apiKey := os.Getenv("OPENAI_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OPENAI_API_KEY environment variable not set")
	}

	client := openai.NewClient(apiKey)

	return &OpenAIClient{
		client: client,
		model:  model,
		config: config,
	}, nil
}

// Complete executes a prompt completion
func (c *OpenAIClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	// Get temperature from config, default to 0
	temperature := float32(0)
	if temp, ok := c.config["temperature"]; ok {
		if tempFloat, ok := temp.(float64); ok {
			temperature = float32(tempFloat)
		}
	}

	// Get max tokens from config
	maxTokens := 1000
	if tokens, ok := c.config["max_tokens"]; ok {
		if tokensInt, ok := tokens.(int); ok {
			maxTokens = tokensInt
		}
	}

	req := openai.ChatCompletionRequest{
		Model:       c.model,
		Temperature: &temperature,
		MaxTokens:   maxTokens,
		Messages: []openai.ChatCompletionMessage{
			{
				Role:    openai.ChatMessageRoleUser,
				Content: prompt,
			},
		},
	}

	resp, err := c.client.CreateChatCompletion(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("OpenAI API error: %w", err)
	}

	if len(resp.Choices) == 0 {
		return nil, fmt.Errorf("no completion choices returned")
	}

	// Calculate cost (simplified - would need actual pricing)
	cost := calculateOpenAICost(c.model, resp.Usage.PromptTokens, resp.Usage.CompletionTokens)

	return &Response{
		Text:     resp.Choices[0].Message.Content,
		Cost:     cost,
		Tokens:   resp.Usage.TotalTokens,
		Provider: "openai",
		Model:    c.model,
	}, nil
}

func (c *OpenAIClient) GetName() string {
	return "openai"
}

func (c *OpenAIClient) GetModel() string {
	return c.model
}

// AnthropicClient implements the Anthropic provider
type AnthropicClient struct {
	model  string
	config map[string]interface{}
}

// NewAnthropicClient creates a new Anthropic client
func NewAnthropicClient(model string, config map[string]interface{}) (*AnthropicClient, error) {
	apiKey := os.Getenv("ANTHROPIC_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("ANTHROPIC_API_KEY environment variable not set")
	}

	return &AnthropicClient{
		model:  model,
		config: config,
	}, nil
}

func (c *AnthropicClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	// TODO: Implement Anthropic API integration
	return nil, fmt.Errorf("Anthropic provider not yet implemented")
}

func (c *AnthropicClient) GetName() string {
	return "anthropic"
}

func (c *AnthropicClient) GetModel() string {
	return c.model
}

// MistralClient implements the Mistral provider
type MistralClient struct {
	model  string
	config map[string]interface{}
}

// NewMistralClient creates a new Mistral client
func NewMistralClient(model string, config map[string]interface{}) (*MistralClient, error) {
	apiKey := os.Getenv("MISTRAL_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("MISTRAL_API_KEY environment variable not set")
	}

	return &MistralClient{
		model:  model,
		config: config,
	}, nil
}

func (c *MistralClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	// TODO: Implement Mistral API integration
	return nil, fmt.Errorf("Mistral provider not yet implemented")
}

func (c *MistralClient) GetName() string {
	return "mistral"
}

func (c *MistralClient) GetModel() string {
	return c.model
}

// OllamaClient implements the Ollama provider
type OllamaClient struct {
	model  string
	config map[string]interface{}
}

// NewOllamaClient creates a new Ollama client
func NewOllamaClient(model string, config map[string]interface{}) (*OllamaClient, error) {
	apiKey := os.Getenv("OLLAMA_API_KEY")
	if apiKey == "" {
		return nil, fmt.Errorf("OLLAMA_API_KEY environment variable not set")
	}

	return &OllamaClient{
		model:  model,
		config: config,
	}, nil
}

func (c *OllamaClient) Complete(ctx context.Context, prompt string) (*Response, error) {
	// TODO: Implement Ollama API integration
	return nil, fmt.Errorf("Ollama provider not yet implemented")
}

func (c *OllamaClient) GetName() string {
	return "ollama"
}

func (c *OllamaClient) GetModel() string {
	return c.model
}

// calculateOpenAICost calculates the cost for OpenAI API usage
func calculateOpenAICost(model string, promptTokens, completionTokens int) float64 {
	// Simplified cost calculation - real implementation would use current pricing
	var promptCost, completionCost float64

	switch model {
	case "gpt-4o":
		promptCost = 0.005 / 1000     // $0.005 per 1K prompt tokens
		completionCost = 0.015 / 1000 // $0.015 per 1K completion tokens
	case "gpt-4":
		promptCost = 0.03 / 1000      // $0.03 per 1K prompt tokens
		completionCost = 0.06 / 1000  // $0.06 per 1K completion tokens
	case "gpt-3.5-turbo":
		promptCost = 0.0005 / 1000    // $0.0005 per 1K prompt tokens
		completionCost = 0.0015 / 1000 // $0.0015 per 1K completion tokens
	default:
		// Default to GPT-3.5-turbo pricing
		promptCost = 0.0005 / 1000
		completionCost = 0.0015 / 1000
	}

	return (float64(promptTokens) * promptCost) + (float64(completionTokens) * completionCost)
}
