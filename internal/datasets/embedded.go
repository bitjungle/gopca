// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package datasets

import (
	_ "embed"
)

// Embed the sample dataset files
var (
	//go:embed corn.csv
	CornCSV string

	//go:embed iris.csv
	IrisCSV string

	//go:embed wine.csv
	WineCSV string

	//go:embed swiss_roll.csv
	SwissRollCSV string
)

// GetDataset returns the embedded dataset content by filename
func GetDataset(filename string) (string, bool) {
	switch filename {
	case "corn.csv":
		return CornCSV, true
	case "iris.csv":
		return IrisCSV, true
	case "wine.csv":
		return WineCSV, true
	case "swiss_roll.csv":
		return SwissRollCSV, true
	default:
		return "", false
	}
}
