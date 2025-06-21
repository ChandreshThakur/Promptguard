package cmd

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"runtime"
	"github.com/spf13/cobra"
	"promptgaurd/internal/viewer"
)

var (
	port    int
	viewCmd = &cobra.Command{
		Use:   "view",
		Short: "Launch interactive diff viewer",
		Long: `Launch an interactive HTML viewer to explore test results,
compare baseline vs current runs, and analyze prompt diffs.

The viewer provides:
- Side-by-side diff comparison
- Historical metrics charts
- Interactive "what-if" analysis
- Cost vs relevance tracking`,
		RunE: runView,
	}
)

func init() {
	rootCmd.AddCommand(viewCmd)

	viewCmd.Flags().IntVarP(&port, "port", "p", 8080, "Port for the web server")
	viewCmd.Flags().String("results-file", "artifacts/results.json", "Path to results file")
	viewCmd.Flags().Bool("open-browser", true, "Automatically open browser")
}

func runView(cmd *cobra.Command, args []string) error {
	resultsFile := getStringFlag(cmd, "results-file")
	openBrowser := getBoolFlag(cmd, "open-browser")

	// Check if results file exists
	if _, err := os.Stat(resultsFile); os.IsNotExist(err) {
		fmt.Printf("Results file not found: %s\n", resultsFile)
		fmt.Println("Run 'pg test' or 'pg ci' first to generate results.")
		return nil
	}

	// Create and start the viewer server
	server := viewer.NewServer(resultsFile)
	
	// Start server in background
	go func() {
		fmt.Printf("Starting PromptGuard viewer on http://localhost:%d\n", port)
		if err := http.ListenAndServe(fmt.Sprintf(":%d", port), server); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}()

	// Open browser if requested
	if openBrowser {
		url := fmt.Sprintf("http://localhost:%d", port)
		if err := openBrowserURL(url); err != nil {
			fmt.Printf("Failed to open browser: %v\n", err)
			fmt.Printf("Please visit: %s\n", url)
		}
	}

	fmt.Printf("PromptGuard viewer running on http://localhost:%d\n", port)
	fmt.Println("Press Ctrl+C to stop")

	// Keep the server running
	select {}
}

func openBrowserURL(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "rundll32"
		args = []string{"url.dll,FileProtocolHandler", url}
	case "darwin":
		cmd = "open"
		args = []string{url}
	default: // linux and others
		cmd = "xdg-open"
		args = []string{url}
	}

	return exec.Command(cmd, args...).Start()
}
