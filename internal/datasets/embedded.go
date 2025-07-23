package datasets

import (
	_ "embed"
)

// Embed the sample dataset files
var (
	//go:embed bronir2.csv
	BroNIR2CSV string

	//go:embed corn.csv
	CornCSV string

	//go:embed iris.csv
	IrisCSV string

	//go:embed wine.csv
	WineCSV string
)

// GetDataset returns the embedded dataset content by filename
func GetDataset(filename string) (string, bool) {
	switch filename {
	case "bronir2.csv":
		return BroNIR2CSV, true
	case "corn.csv":
		return CornCSV, true
	case "iris.csv":
		return IrisCSV, true
	case "wine.csv":
		return WineCSV, true
	default:
		return "", false
	}
}