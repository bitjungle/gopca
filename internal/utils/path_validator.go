// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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
