package cmd

import (
	"fmt"
	"os"
	"github.com/spf13/cobra"
	"promptgaurd/internal/config"
	"promptgaurd/internal/runner"
	"promptgaurd/internal/reporter"
	"promptgaurd/internal/github"
)

var (
	ciCmd = &cobra.Command{
		Use:   "ci",
		Short: "Run tests in CI environment",
		Long: `Run prompt tests in continuous integration environment.
This command is optimized for CI/CD pipelines and includes:
- GitHub annotations for failures
- Artifact generation
- Badge status updates
- Baseline comparison`,
		RunE: runCI,
	}
)

func init() {
	rootCmd.AddCommand(ciCmd)

	ciCmd.Flags().String("baseline-path", ".promptguard/baseline.json", "Path to baseline results")
	ciCmd.Flags().String("artifacts-dir", "artifacts", "Directory for CI artifacts")
	ciCmd.Flags().Bool("github-annotations", true, "Generate GitHub annotations")
	ciCmd.Flags().Bool("update-badge", true, "Update GitHub badge")
	ciCmd.Flags().String("commit-sha", "", "Git commit SHA")
	ciCmd.Flags().String("pr-number", "", "Pull request number")
}

func runCI(cmd *cobra.Command, args []string) error {
	// Load configuration
	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	// Create CI-optimized runner
	testRunner := runner.New(cfg, runner.Options{
		Parallel:     4, // Default to 4 parallel executions in CI
		CIMode:       true,
		BaselinePath: getStringFlag(cmd, "baseline-path"),
		CommitSHA:    getStringFlag(cmd, "commit-sha"),
		PRNumber:     getStringFlag(cmd, "pr-number"),
	})

	// Run tests
	results, err := testRunner.Run()
	if err != nil {
		return fmt.Errorf("CI test execution failed: %w", err)
	}

	// Generate CI artifacts
	artifactsDir := getStringFlag(cmd, "artifacts-dir")
	if err := os.MkdirAll(artifactsDir, 0755); err != nil {
		return fmt.Errorf("failed to create artifacts directory: %w", err)
	}

	// Generate multiple report formats for CI
	reporters := []struct {
		format string
		file   string
	}{
		{"json", fmt.Sprintf("%s/results.json", artifactsDir)},
		{"junit", fmt.Sprintf("%s/junit.xml", artifactsDir)},
		{"html", fmt.Sprintf("%s/promptguard.html", artifactsDir)},
		{"markdown", fmt.Sprintf("%s/report.md", artifactsDir)},
	}

	for _, r := range reporters {
		reporter := reporter.New(r.format)
		if err := reporter.Generate(results, r.file); err != nil {
			fmt.Printf("Warning: failed to generate %s report: %v\n", r.format, err)
		}
	}

	// Generate GitHub annotations if enabled
	if getBoolFlag(cmd, "github-annotations") {
		if err := github.GenerateAnnotations(results); err != nil {
			fmt.Printf("Warning: failed to generate GitHub annotations: %v\n", err)
		}
	}

	// Update badge if enabled
	if getBoolFlag(cmd, "update-badge") {
		if err := github.UpdateBadge(results); err != nil {
			fmt.Printf("Warning: failed to update badge: %v\n", err)
		}
	}

	// Print summary
	fmt.Printf("=== CI Test Summary ===\n")
	fmt.Printf("Tests: %d passed, %d failed, %d skipped\n", 
		results.Passed, results.Failed, results.Skipped)
	fmt.Printf("Cost: $%.4f\n", results.TotalCost)
	fmt.Printf("Artifacts: %s/\n", artifactsDir)

	if results.HasFailures() {
		fmt.Printf("\n❌ Tests failed - check artifacts for details\n")
		return fmt.Errorf("tests failed")
	}

	fmt.Printf("\n✅ All tests passed!\n")
	return nil
}

func getStringFlag(cmd *cobra.Command, name string) string {
	value, _ := cmd.Flags().GetString(name)
	return value
}

func getBoolFlag(cmd *cobra.Command, name string) bool {
	value, _ := cmd.Flags().GetBool(name)
	return value
}
