// Run command
// Implements batch checking multiple endpoints from config file
package cmd

import (
	"fmt"
	"os"
	"time"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
	"github.com/r1ckyIn/healthcheck-cli/internal/config"
	"github.com/r1ckyIn/healthcheck-cli/internal/output"
	"github.com/spf13/cobra"
)

// Run command flags
var (
	runConfigPath  string
	runTimeout     time.Duration
	runConcurrency int
	runOutput      string
	runQuiet       bool
	runInsecure    bool
)

// runCmd is the run subcommand
var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run batch health checks from config file",
	Long: `Run health checks on multiple endpoints defined in a configuration file.

Endpoints are checked concurrently for faster results. The configuration file
uses YAML format and supports global defaults and per-endpoint settings.

Examples:
  # Basic usage
  healthcheck run -c endpoints.yaml

  # Override timeout for all endpoints
  healthcheck run -c endpoints.yaml --timeout 10s

  # Increase concurrency
  healthcheck run -c endpoints.yaml --concurrency 20

  # JSON output for CI/CD
  healthcheck run -c endpoints.yaml -o json

  # Quiet mode (exit code only)
  healthcheck run -c endpoints.yaml -q`,
	RunE: runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)

	// Define flags
	runCmd.Flags().StringVarP(&runConfigPath, "config", "c", "endpoints.yaml",
		"Path to configuration file")
	runCmd.Flags().DurationVarP(&runTimeout, "timeout", "t", 0,
		"Override timeout for all endpoints (e.g., 5s, 10s)")
	runCmd.Flags().IntVarP(&runConcurrency, "concurrency", "n", 10,
		"Maximum concurrent checks")
	runCmd.Flags().StringVarP(&runOutput, "output", "o", "table",
		"Output format (table/json)")
	runCmd.Flags().BoolVarP(&runQuiet, "quiet", "q", false,
		"Quiet mode (no output, exit code only)")
	runCmd.Flags().BoolVarP(&runInsecure, "insecure", "k", false,
		"Skip SSL certificate verification for all endpoints")
}

// runRun executes the run command
func runRun(cmd *cobra.Command, args []string) error {
	// Load config file
	cfg, err := config.Load(runConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	// Validate config
	if errors := config.ValidateConfig(cfg); len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Configuration errors:\n")
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(2)
	}

	// Convert to checker.Endpoint
	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	// Apply command line override flags
	if runTimeout > 0 {
		for i := range endpoints {
			endpoints[i].Timeout = runTimeout
		}
	}

	if runInsecure {
		for i := range endpoints {
			endpoints[i].Insecure = true
		}
	}

	// Create checker and execute
	c := checker.New(checker.WithConcurrency(runConcurrency))
	result := c.CheckAll(endpoints)

	// Output results
	if !runQuiet {
		formatter := output.NewFormatter(
			output.OutputFormat(runOutput),
			os.Stdout,
			IsNoColor(),
		)

		if err := formatter.FormatBatch(result); err != nil {
			return fmt.Errorf("failed to format output: %w", err)
		}
	}

	// Set exit code based on result
	if result.Summary.Unhealthy > 0 {
		os.Exit(1)
	}

	return nil
}
