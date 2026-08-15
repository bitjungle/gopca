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

package datasets

import (
	"bytes"
	"compress/gzip"
	_ "embed"
	"io"
)

// Embed the sample dataset files (compressed to reduce binary size)
var (
	//go:embed corn.csv.gz
	CornCSVGZ []byte

	//go:embed iris.csv.gz
	IrisCSVGZ []byte

	//go:embed wine.csv.gz
	WineCSVGZ []byte

	//go:embed swiss_roll.csv.gz
	SwissRollCSVGZ []byte

	//go:embed eeg_eye_state.csv.gz
	EEGEyeStateCSVGZ []byte

	//go:embed cstr.csv.gz
	CSTRCSVGZ []byte

	//go:embed body_measures.csv.gz
	BodyMeasuresCSVGZ []byte
)

// GetDataset returns the embedded dataset content by filename
// All datasets are stored compressed and decompressed on-the-fly to reduce binary size
func GetDataset(filename string) (string, bool) {
	var compressedData []byte

	switch filename {
	case "corn.csv":
		compressedData = CornCSVGZ
	case "iris.csv":
		compressedData = IrisCSVGZ
	case "wine.csv":
		compressedData = WineCSVGZ
	case "swiss_roll.csv":
		compressedData = SwissRollCSVGZ
	case "eeg_eye_state.csv":
		compressedData = EEGEyeStateCSVGZ
	case "cstr.csv":
		compressedData = CSTRCSVGZ
	case "body_measures.csv":
		compressedData = BodyMeasuresCSVGZ
	default:
		return "", false
	}

	// Decompress the dataset
	decompressed, err := decompressGzip(compressedData)
	if err != nil {
		return "", false
	}
	return decompressed, true
}

// decompressGzip decompresses gzip-compressed data
func decompressGzip(data []byte) (string, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer func() { _ = reader.Close() }()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}
