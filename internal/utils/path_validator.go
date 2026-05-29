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

package utils

import (
	"github.com/bitjungle/gopca/pkg/security"
)

// ValidateFilePath checks if a file path is safe to use
// It prevents directory traversal attacks and ensures the path is clean
// This function now delegates to the enhanced security module
func ValidateFilePath(path string) error {
	return security.ValidateInputPath(path)
}

// ValidateOutputPath ensures an output path is safe to write to
// This function now delegates to the enhanced security module which includes
// comprehensive checks for system directories, path traversal, and write permissions
func ValidateOutputPath(path string) error {
	return security.ValidateOutputPath(path)
}
