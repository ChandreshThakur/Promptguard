package assertions

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"promptguard/internal/config"
	"promptguard/internal/providers"
	"promptguard/internal/runner"
)

// Evaluator interface for different assertion types
type Evaluator interface {
	Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error)
}

// NewEvaluator creates a new evaluator for the given assertion type
func NewEvaluator(assertionType string) Evaluator {
	switch assertionType {
	case "answer-relevance":
		return &AnswerRelevanceEvaluator{}
	case "contains-json":
		return &ContainsJSONEvaluator{}
	case "cost":
		return &CostEvaluator{}
	case "llm-rubric":
		return &LLMRubricEvaluator{}
	case "closed-qa":
		return &ClosedQAEvaluator{}
	case "toxicity":
		return &ToxicityEvaluator{}
	case "jailbreak":
		return &JailbreakEvaluator{}
	default:
		return &UnsupportedEvaluator{Type: assertionType}
	}
}

// AnswerRelevanceEvaluator evaluates answer relevance
type AnswerRelevanceEvaluator struct{}

func (e *AnswerRelevanceEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	expectedValue, ok := assertion.Value.(string)
	if !ok {
		return runner.AssertionResult{}, fmt.Errorf("answer-relevance assertion value must be a string")
	}

	// Simple keyword-based relevance check (in real implementation, would use embeddings/LLM)
	score := calculateRelevanceScore(response.Text, expectedValue)
	threshold := assertion.Threshold
	if threshold == 0 {
		threshold = 0.7 // Default threshold
	}

	passed := score >= threshold

	return runner.AssertionResult{
		Type:     "answer-relevance",
		Expected: expectedValue,
		Actual:   response.Text,
		Passed:   passed,
		Score:    score,
		Message:  fmt.Sprintf("Relevance score: %.2f (threshold: %.2f)", score, threshold),
	}, nil
}

// ContainsJSONEvaluator checks if response contains valid JSON
type ContainsJSONEvaluator struct{}

func (e *ContainsJSONEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	// Extract JSON from response
	jsonStr := extractJSON(response.Text)
	
	result := runner.AssertionResult{
		Type:     "contains-json",
		Expected: assertion.Value,
		Actual:   jsonStr,
	}

	if jsonStr == "" {
		result.Passed = false
		result.Message = "No JSON found in response"
		return result, nil
	}

	// Parse JSON to validate structure
	var parsed interface{}
	if err := json.Unmarshal([]byte(jsonStr), &parsed); err != nil {
		result.Passed = false
		result.Message = fmt.Sprintf("Invalid JSON: %v", err)
		return result, nil
	}

	// Check if expected schema is provided
	if assertion.Value != nil {
		expectedSchema, ok := assertion.Value.(map[string]interface{})
		if ok {
			if err := validateJSONSchema(parsed, expectedSchema); err != nil {
				result.Passed = false
				result.Message = fmt.Sprintf("Schema validation failed: %v", err)
				return result, nil
			}
		}
	}

	result.Passed = true
	result.Message = "Valid JSON found"
	return result, nil
}

// CostEvaluator checks if the cost is within threshold
type CostEvaluator struct{}

func (e *CostEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	threshold := assertion.Threshold
	passed := response.Cost <= threshold

	return runner.AssertionResult{
		Type:     "cost",
		Expected: threshold,
		Actual:   response.Cost,
		Passed:   passed,
		Message:  fmt.Sprintf("Cost: $%.4f (threshold: $%.4f)", response.Cost, threshold),
	}, nil
}

// LLMRubricEvaluator uses an LLM to grade the response
type LLMRubricEvaluator struct{}

func (e *LLMRubricEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	// TODO: Implement LLM-based rubric grading
	return runner.AssertionResult{
		Type:    "llm-rubric",
		Passed:  false,
		Message: "LLM rubric evaluation not yet implemented",
	}, nil
}

// ClosedQAEvaluator evaluates closed-ended question answers
type ClosedQAEvaluator struct{}

func (e *ClosedQAEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	// TODO: Implement closed-QA evaluation
	return runner.AssertionResult{
		Type:    "closed-qa",
		Passed:  false,
		Message: "Closed-QA evaluation not yet implemented",
	}, nil
}

// ToxicityEvaluator checks for toxic content
type ToxicityEvaluator struct{}

func (e *ToxicityEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	// Simple toxicity check using keyword detection
	toxicKeywords := []string{"hate", "violence", "offensive", "toxic"}
	
	text := strings.ToLower(response.Text)
	for _, keyword := range toxicKeywords {
		if strings.Contains(text, keyword) {
			return runner.AssertionResult{
				Type:    "toxicity",
				Passed:  false,
				Message: fmt.Sprintf("Toxic content detected: %s", keyword),
			}, nil
		}
	}

	return runner.AssertionResult{
		Type:    "toxicity",
		Passed:  true,
		Message: "No toxic content detected",
	}, nil
}

// JailbreakEvaluator checks for jailbreak attempts
type JailbreakEvaluator struct{}

func (e *JailbreakEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	// TODO: Implement jailbreak detection
	return runner.AssertionResult{
		Type:    "jailbreak",
		Passed:  true,
		Message: "Jailbreak detection not yet implemented",
	}, nil
}

// UnsupportedEvaluator handles unsupported assertion types
type UnsupportedEvaluator struct {
	Type string
}

func (e *UnsupportedEvaluator) Evaluate(assertion config.Assertion, response *providers.Response) (runner.AssertionResult, error) {
	return runner.AssertionResult{}, fmt.Errorf("unsupported assertion type: %s", e.Type)
}

// Helper functions

func calculateRelevanceScore(text, expectedContent string) float64 {
	// Simple keyword-based relevance scoring
	// In a real implementation, this would use embeddings or LLM-based evaluation
	
	text = strings.ToLower(text)
	expectedContent = strings.ToLower(expectedContent)
	
	words := strings.Fields(expectedContent)
	matches := 0
	
	for _, word := range words {
		if strings.Contains(text, word) {
			matches++
		}
	}
	
	if len(words) == 0 {
		return 0
	}
	
	return float64(matches) / float64(len(words))
}

func extractJSON(text string) string {
	// Extract JSON from text using regex
	jsonRegex := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	matches := jsonRegex.FindAllString(text, -1)
	
	for _, match := range matches {
		// Try to parse each potential JSON
		var parsed interface{}
		if err := json.Unmarshal([]byte(match), &parsed); err == nil {
			return match
		}
	}
	
	return ""
}

func validateJSONSchema(data interface{}, schema map[string]interface{}) error {
	// Basic JSON schema validation
	// In a real implementation, would use a proper JSON schema validator
	
	if required, ok := schema["required"].([]interface{}); ok {
		dataMap, ok := data.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object, got %T", data)
		}
		
		for _, field := range required {
			fieldName, ok := field.(string)
			if !ok {
				continue
			}
			
			if _, exists := dataMap[fieldName]; !exists {
				return fmt.Errorf("required field missing: %s", fieldName)
			}
		}
	}
	
	return nil
}
