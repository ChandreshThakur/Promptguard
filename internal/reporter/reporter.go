package reporter

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"html/template"
	"os"
	"path/filepath"
	"strings"
	"time"

	"promptguard/internal/runner"
	"promptguard/internal/diff"
)

// Reporter interface for different output formats
type Reporter interface {
	Generate(results *runner.Results, outputFile string) error
}

// New creates a new reporter for the specified format
func New(format string) Reporter {
	switch format {
	case "json":
		return &JSONReporter{}
	case "junit":
		return &JUnitReporter{}
	case "html":
		return &HTMLReporter{}
	case "markdown":
		return &MarkdownReporter{}
	case "console":
		return &ConsoleReporter{}
	default:
		return &ConsoleReporter{}
	}
}

// JSONReporter outputs results in JSON format
type JSONReporter struct{}

func (r *JSONReporter) Generate(results *runner.Results, outputFile string) error {
	data, err := json.MarshalIndent(results, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal JSON: %w", err)
	}

	if outputFile == "" {
		fmt.Println(string(data))
		return nil
	}

	return os.WriteFile(outputFile, data, 0644)
}

// JUnitReporter outputs results in JUnit XML format
type JUnitReporter struct{}

type JUnitTestSuite struct {
	XMLName   xml.Name        `xml:"testsuite"`
	Name      string          `xml:"name,attr"`
	Tests     int             `xml:"tests,attr"`
	Failures  int             `xml:"failures,attr"`
	Errors    int             `xml:"errors,attr"`
	Time      string          `xml:"time,attr"`
	TestCases []JUnitTestCase `xml:"testcase"`
}

type JUnitTestCase struct {
	Name      string           `xml:"name,attr"`
	ClassName string           `xml:"classname,attr"`
	Time      string           `xml:"time,attr"`
	Failure   *JUnitFailure    `xml:"failure,omitempty"`
	SystemOut string           `xml:"system-out,omitempty"`
}

type JUnitFailure struct {
	Message string `xml:"message,attr"`
	Text    string `xml:",chardata"`
}

func (r *JUnitReporter) Generate(results *runner.Results, outputFile string) error {
	testSuite := JUnitTestSuite{
		Name:     "PromptGuard Tests",
		Tests:    results.Total,
		Failures: results.Failed,
		Errors:   0,
		Time:     fmt.Sprintf("%.3f", results.Duration.Seconds()),
	}

	for _, testResult := range results.TestResults {
		testCase := JUnitTestCase{
			Name:      testResult.Name,
			ClassName: testResult.PromptFile,
			Time:      fmt.Sprintf("%.3f", testResult.Duration.Seconds()),
			SystemOut: fmt.Sprintf("Provider: %s\nCost: $%.4f\nResponse: %s", 
				testResult.Provider, testResult.Cost, testResult.Response),
		}

		if testResult.Status == "failed" {
			failureMessages := []string{}
			for _, assertion := range testResult.Assertions {
				if !assertion.Passed {
					failureMessages = append(failureMessages, assertion.Message)
				}
			}
			
			if len(failureMessages) > 0 || testResult.Error != "" {
				message := strings.Join(failureMessages, "; ")
				if testResult.Error != "" {
					message = testResult.Error
				}
				
				testCase.Failure = &JUnitFailure{
					Message: message,
					Text:    message,
				}
			}
		}

		testSuite.TestCases = append(testSuite.TestCases, testCase)
	}

	data, err := xml.MarshalIndent(testSuite, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal XML: %w", err)
	}

	xmlData := append([]byte(xml.Header), data...)

	if outputFile == "" {
		fmt.Println(string(xmlData))
		return nil
	}

	return os.WriteFile(outputFile, xmlData, 0644)
}

// HTMLReporter generates an interactive HTML report
type HTMLReporter struct{}

func (r *HTMLReporter) Generate(results *runner.Results, outputFile string) error {
	htmlTemplate := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>PromptGuard Report</title>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, 'Segoe UI', sans-serif; margin: 0; padding: 20px; background: #f5f5f5; }
        .container { max-width: 1200px; margin: 0 auto; background: white; border-radius: 8px; box-shadow: 0 2px 10px rgba(0,0,0,0.1); overflow: hidden; }
        .header { background: linear-gradient(135deg, #667eea 0%, #764ba2 100%); color: white; padding: 30px; text-align: center; }
        .header h1 { margin: 0; font-size: 2.5em; }
        .header .subtitle { opacity: 0.9; margin-top: 10px; }
        .summary { display: grid; grid-template-columns: repeat(auto-fit, minmax(200px, 1fr)); gap: 20px; padding: 30px; background: #f8f9fa; }
        .metric { text-align: center; }
        .metric-value { font-size: 2em; font-weight: bold; margin-bottom: 5px; }
        .metric-label { color: #666; text-transform: uppercase; font-size: 0.9em; letter-spacing: 1px; }
        .passed { color: #28a745; }
        .failed { color: #dc3545; }
        .cost { color: #ffc107; }
        .tests { padding: 30px; }
        .test-item { border: 1px solid #e9ecef; border-radius: 6px; margin-bottom: 20px; overflow: hidden; }
        .test-header { padding: 15px 20px; background: #f8f9fa; border-bottom: 1px solid #e9ecef; cursor: pointer; }
        .test-header:hover { background: #e9ecef; }
        .test-content { padding: 20px; display: none; }
        .test-content.show { display: block; }
        .status-badge { padding: 4px 12px; border-radius: 20px; font-size: 0.8em; font-weight: bold; text-transform: uppercase; }
        .badge-passed { background: #d4edda; color: #155724; }
        .badge-failed { background: #f8d7da; color: #721c24; }
        .assertion { margin: 10px 0; padding: 10px; border-left: 4px solid #ccc; background: #f8f9fa; }
        .assertion.passed { border-left-color: #28a745; }
        .assertion.failed { border-left-color: #dc3545; }
        .response { background: #f1f3f4; padding: 15px; border-radius: 4px; margin: 10px 0; white-space: pre-wrap; font-family: monospace; }
    </style>
</head>
<body>
    <div class="container">
        <div class="header">
            <h1>PromptGuard Report</h1>
            <div class="subtitle">{{.Metadata.Timestamp}}</div>
            {{if .Metadata.CommitSHA}}<div class="subtitle">Commit: {{.Metadata.CommitSHA}}</div>{{end}}
        </div>
        
        <div class="summary">
            <div class="metric">
                <div class="metric-value passed">{{.Passed}}</div>
                <div class="metric-label">Passed</div>
            </div>
            <div class="metric">
                <div class="metric-value failed">{{.Failed}}</div>
                <div class="metric-label">Failed</div>
            </div>
            <div class="metric">
                <div class="metric-value">{{.Total}}</div>
                <div class="metric-label">Total</div>
            </div>
            <div class="metric">
                <div class="metric-value cost">${{printf "%.4f" .TotalCost}}</div>
                <div class="metric-label">Cost</div>
            </div>
        </div>

        <div class="tests">
            <h2>Test Results</h2>
            {{range $index, $test := .TestResults}}
            <div class="test-item">
                <div class="test-header" onclick="toggleTest({{$index}})">
                    <span style="font-weight: bold;">{{$test.Name}}</span>
                    <span class="status-badge badge-{{$test.Status}}">{{$test.Status}}</span>
                    <span style="float: right;">{{$test.Provider}} • ${{printf "%.4f" $test.Cost}}</span>
                </div>
                <div id="test-{{$index}}" class="test-content">
                    {{if $test.Error}}
                    <div class="assertion failed">
                        <strong>Error:</strong> {{$test.Error}}
                    </div>
                    {{end}}
                    
                    {{range $test.Assertions}}
                    <div class="assertion {{if .Passed}}passed{{else}}failed{{end}}">
                        <strong>{{.Type}}:</strong> {{.Message}}
                        {{if .Score}}<br><em>Score: {{printf "%.2f" .Score}}</em>{{end}}
                    </div>
                    {{end}}
                    
                    <div class="response">{{$test.Response}}</div>
                </div>
            </div>
            {{end}}
        </div>
    </div>

    <script>
        function toggleTest(index) {
            const content = document.getElementById('test-' + index);
            content.classList.toggle('show');
        }
    </script>
</body>
</html>`

	tmpl, err := template.New("html").Parse(htmlTemplate)
	if err != nil {
		return fmt.Errorf("failed to parse HTML template: %w", err)
	}

	if outputFile == "" {
		return tmpl.Execute(os.Stdout, results)
	}

	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(outputFile), 0755); err != nil {
		return fmt.Errorf("failed to create directory: %w", err)
	}

	file, err := os.Create(outputFile)
	if err != nil {
		return fmt.Errorf("failed to create file: %w", err)
	}
	defer file.Close()

	return tmpl.Execute(file, results)
}

// MarkdownReporter generates a markdown report
type MarkdownReporter struct{}

func (r *MarkdownReporter) Generate(results *runner.Results, outputFile string) error {
	var sb strings.Builder

	// If there are failures, generate detailed diff analysis
	if results.HasFailures() {
		differ := &diff.MarkdownDiffer{}
		diffContent := differ.GenerateFailureDiff(results)
		sb.WriteString(diffContent)
		sb.WriteString("\n---\n\n")
	}

	// Standard report content
	sb.WriteString(fmt.Sprintf("# PromptGuard Report\n\n"))
	sb.WriteString(fmt.Sprintf("**Generated:** %s\n", results.Metadata.Timestamp))
	
	if results.Metadata.CommitSHA != "" {
		sb.WriteString(fmt.Sprintf("**Commit:** %s\n", results.Metadata.CommitSHA))
	}
	
	sb.WriteString("\n## Summary\n\n")
	sb.WriteString("| Metric | Value |\n")
	sb.WriteString("|--------|-------|\n")
	sb.WriteString(fmt.Sprintf("| Tests | %d |\n", results.Total))
	sb.WriteString(fmt.Sprintf("| Passed | %d |\n", results.Passed))
	sb.WriteString(fmt.Sprintf("| Failed | %d |\n", results.Failed))
	sb.WriteString(fmt.Sprintf("| Cost | $%.4f |\n", results.TotalCost))
	sb.WriteString(fmt.Sprintf("| Duration | %v |\n", results.Duration))

	sb.WriteString("\n## Test Results\n\n")
	
	for _, test := range results.TestResults {
		status := "✅"
		if test.Status == "failed" {
			status = "❌"
		}
		
		sb.WriteString(fmt.Sprintf("### %s %s\n\n", status, test.Name))
		sb.WriteString(fmt.Sprintf("- **Provider:** %s\n", test.Provider))
		sb.WriteString(fmt.Sprintf("- **Cost:** $%.4f\n", test.Cost))
		sb.WriteString(fmt.Sprintf("- **Duration:** %v\n", test.Duration))
		
		if test.Error != "" {
			sb.WriteString(fmt.Sprintf("- **Error:** %s\n", test.Error))
		}
		
		sb.WriteString("\n**Assertions:**\n\n")
		for _, assertion := range test.Assertions {
			assertionStatus := "✅"
			if !assertion.Passed {
				assertionStatus = "❌"
			}
			sb.WriteString(fmt.Sprintf("- %s **%s:** %s\n", assertionStatus, assertion.Type, assertion.Message))
		}
		
		sb.WriteString("\n")
	}

	content := sb.String()

	if outputFile == "" {
		fmt.Print(content)
		return nil
	}

	return os.WriteFile(outputFile, []byte(content), 0644)
}

// ConsoleReporter outputs results to the console
type ConsoleReporter struct{}

func (r *ConsoleReporter) Generate(results *runner.Results, outputFile string) error {
	fmt.Printf("\n=== PromptGuard Test Results ===\n")
	fmt.Printf("Generated: %s\n", results.Metadata.Timestamp)
	
	if results.Metadata.CommitSHA != "" {
		fmt.Printf("Commit: %s\n", results.Metadata.CommitSHA)
	}
	
	fmt.Printf("\nSummary:\n")
	fmt.Printf("  Tests: %d\n", results.Total)
	fmt.Printf("  Passed: %d\n", results.Passed)
	fmt.Printf("  Failed: %d\n", results.Failed)
	fmt.Printf("  Cost: $%.4f\n", results.TotalCost)
	fmt.Printf("  Duration: %v\n", results.Duration)

	if results.Failed > 0 {
		fmt.Printf("\nFailures:\n")
		for _, test := range results.TestResults {
			if test.Status == "failed" {
				fmt.Printf("  ❌ %s\n", test.Name)
				if test.Error != "" {
					fmt.Printf("     Error: %s\n", test.Error)
				}
				for _, assertion := range test.Assertions {
					if !assertion.Passed {
						fmt.Printf("     %s: %s\n", assertion.Type, assertion.Message)
					}
				}
			}
		}
	}

	return nil
}
