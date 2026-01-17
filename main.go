// Health Check CLI entry point
// Program entry point for the health check command-line tool
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
