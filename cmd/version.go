// Version command
// Displays version information
package cmd

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// Version info variables, injected via ldflags at build time
var (
	Version   = "dev"
	BuildTime = "unknown"
	GitCommit = "unknown"
)

// versionCmd is the version subcommand
var versionCmd = &cobra.Command{
	Use:   "version",
	Short: "Show version information",
	Long:  `Display the version, build time, git commit, and Go version of healthcheck.`,
	Run:   runVersion,
}

func init() {
	rootCmd.AddCommand(versionCmd)
}

// runVersion executes the version command
func runVersion(cmd *cobra.Command, args []string) {
	fmt.Printf("healthcheck %s\n", Version)
	fmt.Printf("  Built:    %s\n", BuildTime)
	fmt.Printf("  Commit:   %s\n", GitCommit)
	fmt.Printf("  Go:       %s\n", runtime.Version())
}
