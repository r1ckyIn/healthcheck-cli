// Root command configuration
// Defines the CLI root command and global flags
package cmd

import (
	"errors"
	"os"

	"github.com/r1ckyIn/healthcheck-cli/internal/checker"
	"github.com/spf13/cobra"
)

// Custom error types for exit code handling
var (
	// ErrConfig indicates a configuration error (exit code 2)
	ErrConfig = errors.New("configuration error")
	// ErrUnhealthy indicates unhealthy endpoint(s) (exit code 1)
	ErrUnhealthy = errors.New("unhealthy endpoint")
)

// Global variables
var (
	noColor bool
)

// rootCmd is the CLI root command
var rootCmd = &cobra.Command{
	Use:   "healthcheck",
	Short: "A CLI tool for HTTP endpoint health checking",
	Long: `Health Check CLI is a command-line tool for batch HTTP endpoint health checking.
Supports concurrent checks, multiple output formats, and configuration file management.

Example usage:
  healthcheck check https://api.example.com/health
  healthcheck run -c endpoints.yaml
  healthcheck config init > endpoints.yaml`,
}

// Execute executes the root command and handles exit codes
func Execute() {
	if err := rootCmd.Execute(); err != nil {
		// Cobra already prints the error, so we just need to set exit code
		if errors.Is(err, ErrConfig) {
			os.Exit(2)
		}
		os.Exit(1)
	}
}

func init() {
	// Set checker version for User-Agent
	checker.Version = Version

	// Global flags
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Support NO_COLOR environment variable (https://no-color.org/)
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}

	// Detect if running in non-TTY environment, auto-disable colors
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		noColor = true
	}
}

// IsNoColor returns whether colors are disabled
func IsNoColor() bool {
	return noColor
}
