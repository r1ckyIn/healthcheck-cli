// Completion command / completion 命令
// Generate shell auto-completion scripts / 生成 Shell 自动补全脚本
package cmd

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// completionCmd is the completion subcommand / completion 子命令
var completionCmd = &cobra.Command{
	Use:   "completion [bash|zsh|fish|powershell]",
	Short: "Generate shell completion scripts",
	Long: `Generate shell completion scripts for healthcheck.

To load completions:

Bash:
  $ source <(healthcheck completion bash)

  # To load completions for each session, execute once:
  # Linux:
  $ healthcheck completion bash > /etc/bash_completion.d/healthcheck
  # macOS:
  $ healthcheck completion bash > $(brew --prefix)/etc/bash_completion.d/healthcheck

Zsh:
  # If shell completion is not already enabled in your environment,
  # you will need to enable it. You can execute the following once:
  $ echo "autoload -U compinit; compinit" >> ~/.zshrc

  # To load completions for each session, execute once:
  $ healthcheck completion zsh > "${fpath[1]}/_healthcheck"

  # You will need to start a new shell for this setup to take effect.

Fish:
  $ healthcheck completion fish | source

  # To load completions for each session, execute once:
  $ healthcheck completion fish > ~/.config/fish/completions/healthcheck.fish

PowerShell:
  PS> healthcheck completion powershell | Out-String | Invoke-Expression

  # To load completions for every new session, run:
  PS> healthcheck completion powershell > healthcheck.ps1
  # and source this file from your PowerShell profile.
`,
	DisableFlagsInUseLine: true,
	ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
	Args:                  cobra.MatchAll(cobra.ExactArgs(1), cobra.OnlyValidArgs),
	RunE:                  runCompletion,
}

func init() {
	rootCmd.AddCommand(completionCmd)
}

// runCompletion executes completion command / 执行 completion 命令
func runCompletion(cmd *cobra.Command, args []string) error {
	var err error

	switch args[0] {
	case "bash":
		err = cmd.Root().GenBashCompletion(os.Stdout)
	case "zsh":
		err = cmd.Root().GenZshCompletion(os.Stdout)
	case "fish":
		err = cmd.Root().GenFishCompletion(os.Stdout, true)
	case "powershell":
		err = cmd.Root().GenPowerShellCompletionWithDesc(os.Stdout)
	}

	if err != nil {
		return fmt.Errorf("failed to generate %s completion: %w", args[0], err)
	}

	return nil
}
