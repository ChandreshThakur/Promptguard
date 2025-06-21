package cmd

import (
	"fmt"
	"os"
	"time"
	"github.com/spf13/cobra"
	"promptgaurd/internal/config"
	"promptgaurd/internal/runner"
	"promptgaurd/internal/reporter"
)

var (
	outputFormat string
	outputFile   string
	parallel     int
	testCmd      = &cobra.Command{
		Use:   "test",
		Short: "Run prompt tests locally",
		Long: `Run prompt tests against configured LLM providers with assertions.
This command is designed for local development and testing.`,
		RunE: runTest,
	}
)

func init() {
	rootCmd.AddCommand(testCmd)

	testCmd.Flags().StringVarP(&outputFormat, "output", "o", "console", "Output format (console, json, junit)")
	testCmd.Flags().StringVar(&outputFile, "output-file", "", "Output file path")
	testCmd.Flags().IntVarP(&parallel, "parallel", "p", 1, "Number of parallel test executions")
	testCmd.Flags().Bool("update-baseline", false, "Update baseline results")
	testCmd.Flags().StringSlice("filter", []string{}, "Filter tests by name pattern")
}

func runTest(cmd *cobra.Command, args []string) error {
	startTime := time.Now()

	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create test runner
	testRunner := runner.New(cfg, runner.Options{
		Parallel:        parallel,
		UpdateBaseline:  cmd.Flag("update-baseline").Changed,
		Filters:         getStringSliceFlag(cmd, "filter"),
		Verbose:         cmd.Flag("verbose").Changed,
	})

	// Run tests
	results, err := testRunner.Run()
	if err != nil {
		return fmt.Errorf("test execution failed: %w", err)
	}

	// Generate report
	reporter := reporter.New(outputFormat)
	if err := reporter.Generate(results, outputFile); err != nil {
		return fmt.Errorf("failed to generate report: %w", err)
	}

	// Print summary
	duration := time.Since(startTime)
	printTestSummary(results, duration)

	// Exit with non-zero code if tests failed
	if results.HasFailures() {
		os.Exit(1)
	}

	return nil
}

func printTestSummary(results *runner.Results, duration time.Duration) {
	fmt.Printf("\n=== Test Summary ===\n")
	fmt.Printf("Tests run: %d\n", results.Total)
	fmt.Printf("Passed: %d\n", results.Passed)
	fmt.Printf("Failed: %d\n", results.Failed)
	fmt.Printf("Skipped: %d\n", results.Skipped)
	fmt.Printf("Duration: %v\n", duration)
	fmt.Printf("Total cost: $%.4f\n", results.TotalCost)

	if results.HasFailures() {
		fmt.Printf("\n❌ Some tests failed. Run 'pg view' to see details.\n")
	} else {
		fmt.Printf("\n✅ All tests passed!\n")
	}
}

func getStringSliceFlag(cmd *cobra.Command, name string) []string {
	value, _ := cmd.Flags().GetStringSlice(name)
	return value
}
