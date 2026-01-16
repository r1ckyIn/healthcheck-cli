// config 命令组
// 配置文件管理相关命令
package cmd

import (
	"fmt"
	"os"

	"github.com/r1ckyIn/healthcheck-cli/internal/config"
	"github.com/spf13/cobra"
)

// config 命令参数
var (
	configInitFull     bool
	configValidatePath string
)

// configCmd 是 config 命令组
var configCmd = &cobra.Command{
	Use:   "config",
	Short: "Configuration file management",
	Long: `Commands for managing healthcheck configuration files.

Available subcommands:
  init      - Generate a sample configuration file
  validate  - Validate an existing configuration file`,
}

// configInitCmd 是 config init 子命令
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
	Run: runConfigInit,
}

// configValidateCmd 是 config validate 子命令
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
	Run: runConfigValidate,
}

func init() {
	rootCmd.AddCommand(configCmd)
	configCmd.AddCommand(configInitCmd)
	configCmd.AddCommand(configValidateCmd)

	// config init 参数
	configInitCmd.Flags().BoolVar(&configInitFull, "full", false,
		"Generate full configuration with all available options")

	// config validate 参数
	configValidateCmd.Flags().StringVarP(&configValidatePath, "config", "c", "endpoints.yaml",
		"Path to configuration file to validate")
}

// runConfigInit 执行 config init 命令
func runConfigInit(cmd *cobra.Command, args []string) {
	sample := config.GenerateSampleConfig(configInitFull)
	fmt.Print(sample)
}

// runConfigValidate 执行 config validate 命令
func runConfigValidate(cmd *cobra.Command, args []string) {
	// 加载配置文件
	cfg, err := config.Load(configValidatePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	// 验证配置
	errors := config.ValidateConfig(cfg)

	if len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Configuration validation failed:\n")
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(2)
	}

	// 尝试转换为 endpoints，检查是否能正常解析
	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	fmt.Printf("Configuration is valid.\n")
	fmt.Printf("  Endpoints: %d\n", len(endpoints))

	// 显示概要信息
	if len(endpoints) > 0 {
		fmt.Printf("  Names:\n")
		for _, ep := range endpoints {
			fmt.Printf("    - %s\n", ep.Name)
		}
	}
}
