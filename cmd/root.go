// 根命令配置
// 定义 CLI 的根命令和全局参数
package cmd

import (
	"os"

	"github.com/spf13/cobra"
)

// 全局变量
var (
	noColor bool
)

// rootCmd 是 CLI 的根命令
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

// Execute 执行根命令
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	// 全局参数
	rootCmd.PersistentFlags().BoolVar(&noColor, "no-color", false, "Disable colored output")

	// 检测是否在非 TTY 环境下运行，自动禁用颜色
	if fileInfo, _ := os.Stdout.Stat(); (fileInfo.Mode() & os.ModeCharDevice) == 0 {
		noColor = true
	}
}

// IsNoColor 返回是否禁用颜色
func IsNoColor() bool {
	return noColor
}
