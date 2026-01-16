// version 命令
// 显示版本信息
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// 版本信息变量，通过 ldflags 在编译时注入
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// versionCmd 是 version 子命令
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version, build time, git commit, and Go version of healthcheck.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// runVersion 执行 version 命令
func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("healthcheck %s\n", Version)
	fmt.Printf("  Built:    %s\n", BuildTime)
	fmt.Printf("  Commit:   %s\n", GitCommit)
	fmt.Printf("  Go:       %s\n", runtime.Version())
}
