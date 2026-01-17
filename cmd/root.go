// Root command configuration
// Defines the CLI root command and global flags
package cmd

import (
	"os"

	"github.com/spf13/cobra"
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

// Execute executes the root command
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// Global flags
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// Detect if running in non-TTY environment, auto-disable colors
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		noColor = true
	}
}

// IsNoColor returns whether colors are disabled
func IsNoColor() bool {
	return noColor
}
