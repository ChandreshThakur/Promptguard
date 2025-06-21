package prompts

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
)

// Prompt represents a prompt template
type Prompt struct {
	Content  string            `json:"content"`
	Metadata map[string]string `json:"metadata"`
	Template *template.Template
}

// LoadFromFile loads a prompt from a file
func LoadFromFile(filename string) (*Prompt, error) {
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("failed to read prompt file %s: %w", filename, err)
	}

	prompt := &Prompt{
		Content:  string(content),
		Metadata: make(map[string]string),
	}

	// Parse metadata from frontmatter if present
	if err := prompt.parseFrontmatter(); err != nil {
		return nil, fmt.Errorf("failed to parse frontmatter in %s: %w", filename, err)
	}

	// Create template
	tmpl, err := template.New(filepath.Base(filename)).Parse(prompt.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template in %s: %w", filename, err)
	}

	prompt.Template = tmpl
	return prompt, nil
}

// Render renders the prompt with given variables
func (p *Prompt) Render(variables map[string]interface{}) (string, error) {
	var buf strings.Builder
	
	if err := p.Template.Execute(&buf, variables); err != nil {
		return "", fmt.Errorf("failed to render prompt: %w", err)
	}

	return buf.String(), nil
}

// parseFrontmatter extracts YAML frontmatter from the prompt content
func (p *Prompt) parseFrontmatter() error {
	// Check for YAML frontmatter
	frontmatterRegex := regexp.MustCompile(`^---\s*\n(.*?)\n---\s*\n(.*)`)
	matches := frontmatterRegex.FindStringSubmatch(p.Content)
	
	if len(matches) == 3 {
		// TODO: Parse YAML frontmatter and extract metadata
		// For now, just use the content without frontmatter
		p.Content = matches[2]
	}

	return nil
}

// GetVariables extracts variable names from the prompt template
func (p *Prompt) GetVariables() []string {
	// Simple regex to find {{.Variable}} patterns
	varRegex := regexp.MustCompile(`\{\{\s*\.(\w+)\s*\}\}`)
	matches := varRegex.FindAllStringSubmatch(p.Content, -1)
	
	var variables []string
	seen := make(map[string]bool)
	
	for _, match := range matches {
		if len(match) > 1 && !seen[match[1]] {
			variables = append(variables, match[1])
			seen[match[1]] = true
		}
	}
	
	return variables
}

// Validate checks if the prompt is valid
func (p *Prompt) Validate() error {
	if strings.TrimSpace(p.Content) == "" {
		return fmt.Errorf("prompt content is empty")
	}

	// Try to parse as template
	_, err := template.New("test").Parse(p.Content)
	if err != nil {
		return fmt.Errorf("invalid template syntax: %w", err)
	}

	return nil
}
