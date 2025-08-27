// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

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

	//go:embed met.csv.gz
	MetCSVGZ []byte
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
	case "met.csv":
		compressedData = MetCSVGZ
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
