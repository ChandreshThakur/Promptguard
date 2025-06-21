package config

import (
	"fmt"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v3"
)

// Config represents the main configuration structure
type Config struct {
	Description string     `yaml:"description"`
	Prompts     []string   `yaml:"prompts"`
	Providers   []Provider `yaml:"providers"`
	Tests       []Test     `yaml:"tests"`
	Settings    Settings   `yaml:"settings,omitempty"`
}

// Provider represents an LLM provider configuration
type Provider struct {
	ID     string                 `yaml:"id"`
	Config map[string]interface{} `yaml:"config,omitempty"`
}

// Test represents a test case configuration
type Test struct {
	Name        string                 `yaml:"name,omitempty"`
	Description string                 `yaml:"description,omitempty"`
	Variables   map[string]interface{} `yaml:"vars"`
	Assert      []Assertion            `yaml:"assert"`
	Provider    string                 `yaml:"provider,omitempty"`
}

// Assertion represents a test assertion
type Assertion struct {
	Type      string      `yaml:"type"`
	Value     interface{} `yaml:"value,omitempty"`
	Threshold float64     `yaml:"threshold,omitempty"`
	Required  bool        `yaml:"required,omitempty"`
}

// Settings represents global settings
type Settings struct {
	CostBudget   float64 `yaml:"costBudget,omitempty"`
	Timeout      int     `yaml:"timeout,omitempty"`
	MaxRetries   int     `yaml:"maxRetries,omitempty"`
	CacheResults bool    `yaml:"cacheResults,omitempty"`
}

// Load loads configuration from promptguard.yaml
func Load() (*Config, error) {
	configPaths := []string{
		"promptguard.yaml",
		"promptguard.yml",
		".promptguard/config.yaml",
		".promptguard/config.yml",
	}

	var configFile string
	for _, path := range configPaths {
		if _, err := os.Stat(path); err == nil {
			configFile = path
			break
		}
	}

	if configFile == "" {
		return nil, fmt.Errorf("no configuration file found. Create promptguard.yaml in your project root")
	}

	return LoadFromFile(configFile)
}

// LoadFromFile loads configuration from a specific file
func LoadFromFile(filename string) (*Config, error) {
	data, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file %s: %w", filename, err)
	}

	var config Config
	if err := yaml.Unmarshal(data, &config); err != nil {
		return nil, fmt.Errorf("failed to parse config file %s: %w", filename, err)
	}

	// Validate configuration
	if err := config.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	// Expand prompt file paths
	if err := config.expandPromptPaths(); err != nil {
		return nil, fmt.Errorf("failed to expand prompt paths: %w", err)
	}

	return &config, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if len(c.Prompts) == 0 {
		return fmt.Errorf("no prompt files specified")
	}

	if len(c.Providers) == 0 {
		return fmt.Errorf("no providers specified")
	}

	if len(c.Tests) == 0 {
		return fmt.Errorf("no tests specified")
	}

	// Validate provider IDs
	providerIDs := make(map[string]bool)
	for _, provider := range c.Providers {
		if provider.ID == "" {
			return fmt.Errorf("provider missing ID")
		}
		if providerIDs[provider.ID] {
			return fmt.Errorf("duplicate provider ID: %s", provider.ID)
		}
		providerIDs[provider.ID] = true
	}

	// Validate test assertions
	for i, test := range c.Tests {
		if len(test.Assert) == 0 {
			return fmt.Errorf("test %d has no assertions", i)
		}

		for j, assertion := range test.Assert {
			if err := assertion.Validate(); err != nil {
				return fmt.Errorf("test %d, assertion %d: %w", i, j, err)
			}
		}
	}

	return nil
}

// Validate validates an assertion
func (a *Assertion) Validate() error {
	validTypes := map[string]bool{
		"answer-relevance": true,
		"contains-json":    true,
		"cost":            true,
		"llm-rubric":      true,
		"closed-qa":       true,
		"toxicity":        true,
		"jailbreak":       true,
	}

	if !validTypes[a.Type] {
		return fmt.Errorf("invalid assertion type: %s", a.Type)
	}

	// Type-specific validation
	switch a.Type {
	case "cost":
		if a.Threshold <= 0 {
			return fmt.Errorf("cost assertion requires positive threshold")
		}
	case "answer-relevance":
		if a.Threshold < 0 || a.Threshold > 1 {
			return fmt.Errorf("answer-relevance threshold must be between 0 and 1")
		}
	}

	return nil
}

// expandPromptPaths expands glob patterns in prompt paths
func (c *Config) expandPromptPaths() error {
	var expandedPaths []string

	for _, pattern := range c.Prompts {
		matches, err := filepath.Glob(pattern)
		if err != nil {
			return fmt.Errorf("invalid glob pattern %s: %w", pattern, err)
		}

		if len(matches) == 0 {
			return fmt.Errorf("no files match pattern: %s", pattern)
		}

		expandedPaths = append(expandedPaths, matches...)
	}

	c.Prompts = expandedPaths
	return nil
}

// GetProvider returns a provider by ID
func (c *Config) GetProvider(id string) (*Provider, error) {
	for _, provider := range c.Providers {
		if provider.ID == id {
			return &provider, nil
		}
	}
	return nil, fmt.Errorf("provider not found: %s", id)
}
