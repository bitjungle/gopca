// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package cobra

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

// NewVersionCommand creates the version subcommand
func NewVersionCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Show version information",
		Long:  `Display version, build time, and platform information.`,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Printf("GoPCA CLI version %s\n", Version)
			fmt.Printf("  Built:    %s\n", BuildTime)
			fmt.Printf("  Commit:   %s\n", Commit)
			fmt.Printf("  Platform: %s/%s\n", runtime.GOOS, runtime.GOARCH)
			fmt.Printf("  Go:       %s\n", runtime.Version())
		},
	}

	return cmd
}
