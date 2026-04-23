// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"github.com/bitjungle/gopca/pkg/utils"
)

// missingValueIndicators is the set of strings treated as missing values by GoCSV.
// It extends the shared default set with additional common indicators.
var missingValueIndicators = append(utils.DefaultMissingValues(), "-", "?", "none", "missing")

// isMissingValue reports whether value should be treated as a missing value.
func isMissingValue(value string) bool {
	return utils.IsMissingValue(value, missingValueIndicators)
}
