package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
	"promptguard/internal/runner"
	"promptguard/internal/diff"
	"encoding/json"
)

var (
	baselineFile string
	currentFile  string
	diffCmd      = &cobra.Command{
		Use:   "diff",
		Short: "Generate markdown diff for failed tests",
		Long: `Generate a detailed markdown diff analysis for test failures.
Compares current results with baseline and shows red/green diffs
for failed assertions.`,
		RunE: runDiff,
	}
)

func init() {
	rootCmd.AddCommand(diffCmd)

	diffCmd.Flags().StringVar(&baselineFile, "baseline", ".promptguard/baseline.json", "Baseline results file")
	diffCmd.Flags().StringVar(&currentFile, "current", "artifacts/results.json", "Current results file")
	diffCmd.Flags().StringVar(&outputFile, "output", "", "Output file for diff (default: stdout)")
}

func runDiff(cmd *cobra.Command, args []string) error {
	// Load current results
	var currentResults runner.Results
	if err := loadResults(currentFile, &currentResults); err != nil {
		return fmt.Errorf("failed to load current results: %w", err)
	}

	differ := &diff.MarkdownDiffer{}

	// Generate failure diff
	failureDiff := differ.GenerateFailureDiff(&currentResults)

	// If baseline exists, also generate baseline comparison
	var baselineComparison string
	if _, err := os.Stat(baselineFile); err == nil {
		var baselineResults runner.Results
		if err := loadResults(baselineFile, &baselineResults); err == nil {
			baselineComparison = differ.GenerateBaselineComparison(&currentResults, &baselineResults)
		}
	}

	// Combine outputs
	output := failureDiff
	if baselineComparison != "" {
		output += "\n" + baselineComparison
	}

	// Write output
	if outputFile == "" {
		fmt.Print(output)
	} else {
		if err := os.WriteFile(outputFile, []byte(output), 0644); err != nil {
			return fmt.Errorf("failed to write output file: %w", err)
		}
		fmt.Printf("Diff analysis written to: %s\n", outputFile)
	}

	return nil
}

func loadResults(filename string, results *runner.Results) error {
	data, err := os.ReadFile(filename)
	if err != nil {
		return err
	}

	return json.Unmarshal(data, results)
}
