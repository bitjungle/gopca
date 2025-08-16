// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cobra

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"
)

// NewCompletionCommand creates the completion subcommand
func NewCompletionCommand(rootCmd *cobra.Command) *cobra.Command {
	cmd := &cobra.Command{
		Use:   "completion [bash|zsh|fish|powershell]",
		Short: "Generate shell completion scripts",
		Long: `Generate shell completion scripts for GoPCA CLI.

To enable completions:

Bash:
  $ source <(pca completion bash)
  # To load completions for every session, add to ~/.bashrc:
  $ echo 'source <(pca completion bash)' >> ~/.bashrc

Zsh:
  $ source <(pca completion zsh)
  # To load completions for every session, add to ~/.zshrc:
  $ echo 'source <(pca completion zsh)' >> ~/.zshrc

Fish:
  $ pca completion fish | source
  # To load completions for every session, add to ~/.config/fish/config.fish:
  $ pca completion fish > ~/.config/fish/completions/pca.fish

PowerShell:
  PS> pca completion powershell | Out-String | Invoke-Expression
  # To load completions for every session, add to $PROFILE:
  PS> pca completion powershell >> $PROFILE`,
		ValidArgs:             []string{"bash", "zsh", "fish", "powershell"},
		DisableFlagsInUseLine: true,
		Args:                  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletion(os.Stdout)
			case "zsh":
				return rootCmd.GenZshCompletion(os.Stdout)
			case "fish":
				return rootCmd.GenFishCompletion(os.Stdout, true)
			case "powershell":
				return rootCmd.GenPowerShellCompletionWithDesc(os.Stdout)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}

	return cmd
}
