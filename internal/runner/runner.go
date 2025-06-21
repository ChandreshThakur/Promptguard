package runner

import (
	"context"
	"fmt"
	"sync"	"time"

	"promptgaurd/internal/config"
	"promptgaurd/internal/prompts"
	"promptgaurd/internal/providers"
	"promptgaurd/internal/assertions"
	"promptgaurd/internal/metrics"
)

// Runner orchestrates prompt testing
type Runner struct {
	config  *config.Config
	options Options
	metrics *metrics.Store
}

// Options configures the test runner
type Options struct {
	Parallel        int
	UpdateBaseline  bool
	Filters         []string
	Verbose         bool
	CIMode          bool
	BaselinePath    string
	CommitSHA       string
	PRNumber        string
}

// Results contains test execution results
type Results struct {
	Total       int           `json:"total"`
	Passed      int           `json:"passed"`
	Failed      int           `json:"failed"`
	Skipped     int           `json:"skipped"`
	TotalCost   float64       `json:"totalCost"`
	Duration    time.Duration `json:"duration"`
	TestResults []TestResult  `json:"testResults"`
	Metadata    Metadata      `json:"metadata"`
}

// TestResult represents a single test result
type TestResult struct {
	Name         string                 `json:"name"`
	PromptFile   string                 `json:"promptFile"`
	Provider     string                 `json:"provider"`
	Variables    map[string]interface{} `json:"variables"`
	Response     string                 `json:"response"`
	Assertions   []AssertionResult      `json:"assertions"`
	Cost         float64                `json:"cost"`
	Duration     time.Duration          `json:"duration"`
	Status       string                 `json:"status"` // passed, failed, skipped
	Error        string                 `json:"error,omitempty"`
}

// AssertionResult represents a single assertion result
type AssertionResult struct {
	Type     string      `json:"type"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
	Passed   bool        `json:"passed"`
	Score    float64     `json:"score,omitempty"`
	Message  string      `json:"message,omitempty"`
}

// Metadata contains test run metadata
type Metadata struct {
	Timestamp string `json:"timestamp"`
	CommitSHA string `json:"commitSha,omitempty"`
	PRNumber  string `json:"prNumber,omitempty"`
	Branch    string `json:"branch,omitempty"`
	Version   string `json:"version"`
}

// New creates a new test runner
func New(cfg *config.Config, options Options) *Runner {
	return &Runner{
		config:  cfg,
		options: options,
		metrics: metrics.NewStore(),
	}
}

// Run executes all tests
func (r *Runner) Run() (*Results, error) {
	startTime := time.Now()

	results := &Results{
		TestResults: make([]TestResult, 0),
		Metadata: Metadata{
			Timestamp: startTime.Format(time.RFC3339),
			CommitSHA: r.options.CommitSHA,
			PRNumber:  r.options.PRNumber,
			Version:   "0.1.0",
		},
	}

	// Load prompts
	promptFiles, err := r.loadPrompts()
	if err != nil {
		return nil, fmt.Errorf("failed to load prompts: %w", err)
	}

	// Generate test cases
	testCases := r.generateTestCases(promptFiles)

	// Filter test cases if needed
	if len(r.options.Filters) > 0 {
		testCases = r.filterTestCases(testCases)
	}

	results.Total = len(testCases)

	// Run tests with parallelization
	testResults := make(chan TestResult, len(testCases))
	
	// Create worker pool
	var wg sync.WaitGroup
	semaphore := make(chan struct{}, r.options.Parallel)

	for _, testCase := range testCases {
		wg.Add(1)
		go func(tc TestCase) {
			defer wg.Done()
			semaphore <- struct{}{} // Acquire
			defer func() { <-semaphore }() // Release

			result := r.runSingleTest(tc)
			testResults <- result
		}(testCase)
	}

	// Wait for all tests to complete
	go func() {
		wg.Wait()
		close(testResults)
	}()

	// Collect results
	for result := range testResults {
		results.TestResults = append(results.TestResults, result)
		results.TotalCost += result.Cost

		switch result.Status {
		case "passed":
			results.Passed++
		case "failed":
			results.Failed++
		case "skipped":
			results.Skipped++
		}
	}

	results.Duration = time.Since(startTime)

	// Store metrics
	if err := r.metrics.Store(results); err != nil {
		fmt.Printf("Warning: failed to store metrics: %v\n", err)
	}

	return results, nil
}

// TestCase represents a single test execution
type TestCase struct {
	Name       string
	PromptFile string
	Provider   string
	Variables  map[string]interface{}
	Test       config.Test
}

func (r *Runner) loadPrompts() (map[string]*prompts.Prompt, error) {
	promptFiles := make(map[string]*prompts.Prompt)

	for _, file := range r.config.Prompts {
		prompt, err := prompts.LoadFromFile(file)
		if err != nil {
			return nil, fmt.Errorf("failed to load prompt %s: %w", file, err)
		}
		promptFiles[file] = prompt
	}

	return promptFiles, nil
}

func (r *Runner) generateTestCases(promptFiles map[string]*prompts.Prompt) []TestCase {
	var testCases []TestCase

	for promptFile, prompt := range promptFiles {
		for i, test := range r.config.Tests {
			// Determine provider
			provider := test.Provider
			if provider == "" && len(r.config.Providers) > 0 {
				provider = r.config.Providers[0].ID
			}

			testName := test.Name
			if testName == "" {
				testName = fmt.Sprintf("%s_test_%d", promptFile, i)
			}

			testCases = append(testCases, TestCase{
				Name:       testName,
				PromptFile: promptFile,
				Provider:   provider,
				Variables:  test.Variables,
				Test:       test,
			})
		}
	}

	return testCases
}

func (r *Runner) filterTestCases(testCases []TestCase) []TestCase {
	// TODO: Implement test filtering based on r.options.Filters
	return testCases
}

func (r *Runner) runSingleTest(testCase TestCase) TestResult {
	startTime := time.Now()

	result := TestResult{
		Name:       testCase.Name,
		PromptFile: testCase.PromptFile,
		Provider:   testCase.Provider,
		Variables:  testCase.Variables,
		Duration:   0,
		Status:     "failed",
		Assertions: make([]AssertionResult, 0),
	}

	// Load prompt
	prompt, err := prompts.LoadFromFile(testCase.PromptFile)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to load prompt: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Render prompt with variables
	renderedPrompt, err := prompt.Render(testCase.Variables)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to render prompt: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Get provider
	providerConfig, err := r.config.GetProvider(testCase.Provider)
	if err != nil {
		result.Error = fmt.Sprintf("Provider not found: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Create provider client
	client, err := providers.NewClient(providerConfig)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to create provider client: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	// Execute prompt
	ctx := context.Background()
	response, err := client.Complete(ctx, renderedPrompt)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to execute prompt: %v", err)
		result.Duration = time.Since(startTime)
		return result
	}

	result.Response = response.Text
	result.Cost = response.Cost

	// Run assertions
	allPassed := true
	for _, assertion := range testCase.Test.Assert {
		assertionResult := r.runAssertion(assertion, response)
		result.Assertions = append(result.Assertions, assertionResult)
		
		if !assertionResult.Passed {
			allPassed = false
		}
	}

	if allPassed {
		result.Status = "passed"
	}

	result.Duration = time.Since(startTime)
	return result
}

func (r *Runner) runAssertion(assertion config.Assertion, response *providers.Response) AssertionResult {
	evaluator := assertions.NewEvaluator(assertion.Type)
	
	result, err := evaluator.Evaluate(assertion, response)
	if err != nil {
		return AssertionResult{
			Type:    assertion.Type,
			Passed:  false,
			Message: fmt.Sprintf("Evaluation error: %v", err),
		}
	}

	return result
}

// HasFailures returns true if any tests failed
func (r *Results) HasFailures() bool {
	return r.Failed > 0
}
