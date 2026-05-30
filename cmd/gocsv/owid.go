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

package main

import (
	_ "embed"
	"encoding/json"
	"fmt"
	"os"
	"strings"
)

//go:embed owid_catalog.json
var owidCatalogJSON []byte

const owidBaseURL = "https://catalog.ourworldindata.org/"

// OWIDDataset represents one curated entry in the OWID catalog.
type OWIDDataset struct {
	Namespace   string `json:"namespace"`
	Dataset     string `json:"dataset"`
	Version     string `json:"version"`
	Table       string `json:"table"`
	Title       string `json:"title"`
	Description string `json:"description"`
	Path        string `json:"path"`
}

var owidCatalog []OWIDDataset

func init() {
	if err := json.Unmarshal(owidCatalogJSON, &owidCatalog); err != nil {
		panic("owid_catalog.json is invalid: " + err.Error())
	}
}

// SearchOWIDDatasets returns catalog entries whose title, namespace, dataset, or
// table contains query (case-insensitive). Returns all entries if query is empty.
func (a *App) SearchOWIDDatasets(query string) []OWIDDataset {
	q := strings.ToLower(query)
	if q == "" {
		return owidCatalog
	}
	var results []OWIDDataset
	for _, d := range owidCatalog {
		if strings.Contains(strings.ToLower(d.Title), q) ||
			strings.Contains(strings.ToLower(d.Namespace), q) ||
			strings.Contains(strings.ToLower(d.Dataset), q) ||
			strings.Contains(strings.ToLower(d.Table), q) ||
			strings.Contains(strings.ToLower(d.Description), q) {
			results = append(results, d)
		}
	}
	return results
}

// LoadOWIDDataset fetches the Parquet file for the given catalog path and loads
// it into a FileData struct ready for display in the grid.
// path is the catalog path field, e.g. "garden/energy/2024-06-20/energy_mix/energy_mix".
func (a *App) LoadOWIDDataset(path string) (*FileData, error) {
	url := owidBaseURL + path + ".parquet"
	a.logInfo(fmt.Sprintf("Loading OWID dataset from: %s", url))

	tmpPath, err := fetchRemoteFile(url)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch OWID dataset: %w", err)
	}
	defer os.Remove(tmpPath)

	return a.loadParquet(tmpPath)
}
