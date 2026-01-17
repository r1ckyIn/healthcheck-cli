// Config command group
// Configuration file management commands
package cmd

import (
	"fmt"

	"github.com/r1ckyIn/healthcheck-cli/internal/config"
	"github.com/spf13/cobra"
)

// Config command flags
var (
	configInitFull     bool
	configValidatePath string
)

// configCmd is the config command group
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration file management",
	Long: `Commands for managing healthcheck configuration files.

Available subcommands:
  init      - Generate a sample configuration file
  validate  - Validate an existing configuration file`,
}

// configInitCmd is the config init subcommand
var configInitCmd = &cobra.Command{
	Use:   "init",
	Short: "Generate a sample configuration file",
	Long: `Generate a sample configuration file that can be used as a starting point.

The output is written to stdout. Redirect to a file to save:
  healthcheck config init > endpoints.yaml

Examples:
  # Generate basic configuration
  healthcheck config init > endpoints.yaml

  # Generate full configuration with all options
  healthcheck config init --full > endpoints.yaml`,
	RunE: runConfigInit,
}

// configValidateCmd is the config validate subcommand
var configValidateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Validate a configuration file",
	Long: `Validate the syntax and content of a configuration file.

Checks for:
  - Valid YAML syntax
  - Required fields (url for each endpoint)
  - Valid URL format
  - Valid timeout format
  - Valid status code range

Examples:
  healthcheck config validate
  healthcheck config validate -c endpoints.yaml
  healthcheck config validate -c /path/to/config.yaml`,
	RunE: runConfigValidate,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)

	// config init flags
	configInitCmd.Flags().BoolVar(&configInitFull, "full", false,
		"Generate full configuration with all available options")

	// config validate flags
	configValidateCmd.Flags().StringVarP(&configValidatePath, "config", "c", "endpoints.yaml",
		"Path to configuration file to validate")
}

// runConfigInit executes the config init command
func runConfigInit(cmd *cobra.Command, args []string) error {
	sample := config.GenerateSampleConfig(configInitFull)
	fmt.Print(sample)
	return nil
}

// runConfigValidate executes the config validate command
func runConfigValidate(cmd *cobra.Command, args []string) error {
	// Load config file
	cfg, err := config.Load(configValidatePath)
	if err != nil {
		return fmt.Errorf("%w: %s", ErrConfig, err)
	}

	// Validate config
	configErrors := config.ValidateConfig(cfg)

	if len(configErrors) > 0 {
		errMsg := "configuration validation failed:"
		for _, e := range configErrors {
			errMsg += "\n  - " + e
		}
		return fmt.Errorf("%w: %s", ErrConfig, errMsg)
	}

	// Try converting to endpoints to check parsing
	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		return fmt.Errorf("%w: %s", ErrConfig, err)
	}

	fmt.Printf("Configuration is valid.\n")
	fmt.Printf("  Endpoints: %d\n", len(endpoints))

	// Show summary info
	if len(endpoints) > 0 {
		fmt.Printf("  Names:\n")
		for _, ep := range endpoints {
			fmt.Printf("    - %s\n", ep.Name)
		}
	}

	return nil
}
