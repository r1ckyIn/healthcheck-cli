// Health Check CLI 入口
// 健康检查命令行工具的程序入口点
package main

import (
	"os"

	"github.com/r1ckyIn/healthcheck-cli/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
