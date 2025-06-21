package diff

import (
	"fmt"
	"strings"

	"github.com/sergi/go-diff/diffmatchpatch"
	"promptguard/internal/runner"
)

// MarkdownDiffer generates markdown-formatted diffs for failed assertions
type MarkdownDiffer struct{}

// GenerateFailureDiff creates a markdown diff view for test failures
func (d *MarkdownDiffer) GenerateFailureDiff(results *runner.Results) string {
	var md strings.Builder

	md.WriteString("# 🔍 PromptGuard Failure Analysis\n\n")

	if results.Failed == 0 {
		md.WriteString("✅ **All tests passed!** No failures to analyze.\n")
		return md.String()
	}

	md.WriteString(fmt.Sprintf("❌ **%d test(s) failed** - Analysis below:\n\n", results.Failed))

	for _, test := range results.TestResults {
		if test.Status == "failed" {
			md.WriteString(d.generateTestFailureDiff(test))
			md.WriteString("\n---\n\n")
		}
	}

	md.WriteString("## 📊 Summary\n\n")
	md.WriteString(fmt.Sprintf("- **Total Tests:** %d\n", results.Total))
	md.WriteString(fmt.Sprintf("- **✅ Passed:** %d\n", results.Passed))
	md.WriteString(fmt.Sprintf("- **❌ Failed:** %d\n", results.Failed))
	md.WriteString(fmt.Sprintf("- **💰 Total Cost:** $%.4f\n", results.TotalCost))

	return md.String()
}

func (d *MarkdownDiffer) generateTestFailureDiff(test runner.TestResult) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("## ❌ `%s`\n\n", test.Name))
	md.WriteString(fmt.Sprintf("**📁 File:** `%s`  \n", test.PromptFile))
	md.WriteString(fmt.Sprintf("**🤖 Provider:** `%s`  \n", test.Provider))
	md.WriteString(fmt.Sprintf("**💰 Cost:** $%.4f  \n", test.Cost))

	if test.Error != "" {
		md.WriteString(fmt.Sprintf("\n**🚨 Error:**\n```\n%s\n```\n\n", test.Error))
	}

	// Show failed assertions
	md.WriteString("### 🔬 Failed Assertions\n\n")
	for _, assertion := range test.Assertions {
		if !assertion.Passed {
			md.WriteString(d.generateAssertionDiff(assertion))
		}
	}

	// Show actual response
	md.WriteString("### 📄 Actual Response\n\n")
	md.WriteString("```json\n")
	md.WriteString(test.Response)
	md.WriteString("\n```\n\n")

	return md.String()
}

func (d *MarkdownDiffer) generateAssertionDiff(assertion runner.AssertionResult) string {
	var md strings.Builder

	md.WriteString(fmt.Sprintf("#### ❌ `%s`\n\n", assertion.Type))
	md.WriteString(fmt.Sprintf("**Message:** %s\n\n", assertion.Message))

	switch assertion.Type {
	case "answer-relevance":
		md.WriteString("**Expected Keywords/Concepts:**\n")
		md.WriteString(fmt.Sprintf("```\n%v\n```\n\n", assertion.Expected))
		
		if assertion.Score > 0 {
			md.WriteString(fmt.Sprintf("**Relevance Score:** %.2f ❌\n\n", assertion.Score))
		}

	case "contains-json":
		md.WriteString("**Expected JSON Structure:**\n")
		md.WriteString(fmt.Sprintf("```json\n%v\n```\n\n", assertion.Expected))
		
		md.WriteString("**Actual Response:**\n")
		md.WriteString(fmt.Sprintf("```json\n%v\n```\n\n", assertion.Actual))

		// Generate diff if both are strings
		if expectedStr, ok := assertion.Expected.(string); ok {
			if actualStr, ok := assertion.Actual.(string); ok {
				md.WriteString("**Diff:**\n")
				md.WriteString(d.generateStringDiff(expectedStr, actualStr))
			}
		}

	case "cost":
		expected := assertion.Expected.(float64)
		actual := assertion.Actual.(float64)
		
		md.WriteString("| Metric | Expected | Actual | Status |\n")
		md.WriteString("|--------|----------|--------|---------|\n")
		md.WriteString(fmt.Sprintf("| Cost | ≤ $%.4f | $%.4f | ❌ Over budget |\n\n", expected, actual))
		
		overagePercent := ((actual - expected) / expected) * 100
		md.WriteString(fmt.Sprintf("**💸 Cost overage:** %.1f%% over threshold\n\n", overagePercent))

	default:
		md.WriteString(fmt.Sprintf("**Expected:** `%v`\n", assertion.Expected))
		md.WriteString(fmt.Sprintf("**Actual:** `%v`\n\n", assertion.Actual))
	}

	return md.String()
}

func (d *MarkdownDiffer) generateStringDiff(expected, actual string) string {
	dmp := diffmatchpatch.New()
	diffs := dmp.DiffMain(expected, actual, false)
	diffs = dmp.DiffCleanupSemantic(diffs)

	var md strings.Builder
	md.WriteString("```diff\n")

	for _, diff := range diffs {
		switch diff.Type {
		case diffmatchpatch.DiffInsert:
			md.WriteString(fmt.Sprintf("+ %s\n", diff.Text))
		case diffmatchpatch.DiffDelete:
			md.WriteString(fmt.Sprintf("- %s\n", diff.Text))
		case diffmatchpatch.DiffEqual:
			// Show context lines (first/last few lines of equal text)
			lines := strings.Split(diff.Text, "\n")
			if len(lines) > 6 {
				for i, line := range lines[:3] {
					if i == 0 && line == "" {
						continue
					}
					md.WriteString(fmt.Sprintf("  %s\n", line))
				}
				md.WriteString("  ...\n")
				for _, line := range lines[len(lines)-3:] {
					if line == "" && len(lines) > 1 {
						continue
					}
					md.WriteString(fmt.Sprintf("  %s\n", line))
				}
			} else {
				for _, line := range lines {
					if line != "" || len(lines) == 1 {
						md.WriteString(fmt.Sprintf("  %s\n", line))
					}
				}
			}
		}
	}

	md.WriteString("```\n\n")
	return md.String()
}

// GenerateBaselineComparison compares current results with baseline
func (d *MarkdownDiffer) GenerateBaselineComparison(current, baseline *runner.Results) string {
	var md strings.Builder

	md.WriteString("# 📊 Baseline Comparison Report\n\n")

	// Summary comparison
	md.WriteString("## 📈 Summary Changes\n\n")
	md.WriteString("| Metric | Baseline | Current | Change |\n")
	md.WriteString("|--------|----------|---------|--------|\n")
	
	passedChange := current.Passed - baseline.Passed
	failedChange := current.Failed - baseline.Failed
	costChange := current.TotalCost - baseline.TotalCost
	
	md.WriteString(fmt.Sprintf("| Passed | %d | %d | %s |\n", 
		baseline.Passed, current.Passed, formatChange(passedChange)))
	md.WriteString(fmt.Sprintf("| Failed | %d | %d | %s |\n", 
		baseline.Failed, current.Failed, formatChange(failedChange)))
	md.WriteString(fmt.Sprintf("| Cost | $%.4f | $%.4f | %s |\n", 
		baseline.TotalCost, current.TotalCost, formatCostChange(costChange)))

	// Regression detection
	if current.Failed > baseline.Failed {
		md.WriteString("\n🚨 **REGRESSION DETECTED** - More tests failing than baseline!\n\n")
	} else if current.Failed < baseline.Failed {
		md.WriteString("\n✅ **IMPROVEMENT** - Fewer test failures than baseline!\n\n")
	}

	if costChange > 0.001 { // Significant cost increase
		md.WriteString(fmt.Sprintf("💸 **COST ALERT** - Cost increased by $%.4f (%.1f%%)\n\n", 
			costChange, (costChange/baseline.TotalCost)*100))
	}

	return md.String()
}

func formatChange(change int) string {
	if change > 0 {
		return fmt.Sprintf("🔺 +%d", change)
	} else if change < 0 {
		return fmt.Sprintf("🔽 %d", change)
	}
	return "➖ 0"
}

func formatCostChange(change float64) string {
	if change > 0.0001 {
		return fmt.Sprintf("🔺 +$%.4f", change)
	} else if change < -0.0001 {
		return fmt.Sprintf("🔽 -$%.4f", -change)
	}
	return "➖ $0.0000"
}
