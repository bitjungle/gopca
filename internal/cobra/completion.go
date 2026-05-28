// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// GoPCA Suite is source-available software with free binary redistribution.
// Official compiled binary releases may be used and redistributed free of charge
// under the GoPCA Suite Source-Available Freeware License.
//
// The source code is provided for viewing, review, education, security analysis,
// research, interoperability analysis, and evaluation only.
//
// Modification, redistribution, publication, sublicensing, reuse, incorporation
// into another project, or creation of derivative works based on the source code
// is not permitted without prior written permission from the copyright holder.
//
// Usage Restriction: GoPCA Suite may not be used, directly or indirectly, for
// military, warfare, weapons, intelligence, surveillance, targeting, or
// law-enforcement surveillance applications.
//
// See LICENSE for the full license terms.

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
