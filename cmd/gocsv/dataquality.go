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
	"fmt"

	"github.com/bitjungle/gopca/pkg/dataquality"
)

// FillMissingValuesRequest represents a request to fill missing values.
type FillMissingValuesRequest struct {
	Strategy string `json:"strategy"` // "mean", "median", "mode", "forward", "backward", "custom"
	Column   string `json:"column"`   // Column name, or empty for all columns
	Value    string `json:"value"`    // Custom value for "custom" strategy
}

// AnalyzeMissingValues analyzes missing value patterns in the data.
// Delegates to the dataquality package.
func (a *App) AnalyzeMissingValues(data *FileData) *dataquality.MissingValueStats {
	if data == nil || len(data.Data) == 0 {
		return &dataquality.MissingValueStats{
			ColumnStats: make(map[string]*dataquality.ColumnMissing),
			RowStats:    make(map[int]*dataquality.RowMissing),
		}
	}
	return dataquality.AnalyzeMissing(data.Data, data.Headers)
}

// FillMissingValues fills missing values in the data according to the specified strategy.
// Delegates to the dataquality package.
func (a *App) FillMissingValues(data *FileData, request FillMissingValuesRequest) (*FileData, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to process")
	}

	req := dataquality.FillRequest{
		Strategy: request.Strategy,
		Column:   request.Column,
		Value:    request.Value,
	}

	newData, err := dataquality.Fill(data.Data, data.Headers, data.ColumnTypes, req)
	if err != nil {
		return nil, err
	}

	result := &FileData{
		Headers:              data.Headers,
		RowNames:             data.RowNames,
		Data:                 newData,
		Rows:                 data.Rows,
		Columns:              data.Columns,
		CategoricalColumns:   data.CategoricalColumns,
		NumericTargetColumns: data.NumericTargetColumns,
		ColumnTypes:          data.ColumnTypes,
	}

	a.markDirty()
	return result, nil
}

// AnalyzeDataQuality performs comprehensive data quality analysis on the given
// file data. Delegates to the dataquality package.
func (a *App) AnalyzeDataQuality(data *FileData) (*dataquality.DataQualityReport, error) {
	if data == nil || len(data.Data) == 0 {
		return nil, fmt.Errorf("no data to analyze")
	}

	in := dataquality.AnalysisInput{
		Data:        data.Data,
		Headers:     data.Headers,
		ColumnTypes: data.ColumnTypes,
		RowNames:    data.RowNames,
		Rows:        data.Rows,
		Columns:     data.Columns,
	}

	return dataquality.AnalyzeDataQuality(in)
}
