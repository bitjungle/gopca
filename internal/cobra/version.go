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
