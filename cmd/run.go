// run 命令
// 实现从配置文件批量检查多个端点
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

// run 命令的参数
var (
	runConfigPath  string
	runTimeout     time.Duration
	runConcurrency int
	runOutput      string
	runQuiet       bool
	runInsecure    bool
)

// runCmd 是 run 子命令
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

	// 定义参数
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

// runRun 执行 run 命令
func runRun(cmd *cobra.Command, args []string) error {
	// 加载配置文件
	cfg, err := config.Load(runConfigPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	// 验证配置
	if errors := config.ValidateConfig(cfg); len(errors) > 0 {
		fmt.Fprintf(os.Stderr, "Configuration errors:\n")
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "  - %s\n", e)
		}
		os.Exit(2)
	}

	// 转换为 checker.Endpoint
	endpoints, err := cfg.ToCheckerEndpoints()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(2)
	}

	// 应用命令行覆盖参数
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

	// 创建检查器并执行
	c := checker.New(checker.WithConcurrency(runConcurrency))
	result := c.CheckAll(endpoints)

	// 输出结果
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

	// 根据结果设置退出码
	if result.Summary.Unhealthy > 0 {
		os.Exit(1)
	}

	return nil
}
