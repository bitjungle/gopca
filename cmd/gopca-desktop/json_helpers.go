// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"math"

	"github.com/bitjungle/gopca/pkg/types"
)

// FileDataJSON is a JSON-safe version of FileData
type FileDataJSON struct {
	Headers              []string                       `json:"headers"`
	RowNames             []string                       `json:"rowNames"`
	Data                 [][]types.JSONFloat64          `json:"data"`
	MissingMask          [][]bool                       `json:"missingMask,omitempty"`
	CategoricalColumns   map[string][]string            `json:"categoricalColumns,omitempty"`
	NumericTargetColumns map[string][]types.JSONFloat64 `json:"numericTargetColumns,omitempty"`
}

// ToJSONSafe converts FileData to a JSON-safe version
func (fd *FileData) ToJSONSafe() *FileDataJSON {
	if fd == nil {
		return nil
	}

	// Convert float64 data to types.JSONFloat64 and build missing mask
	jsonData := make([][]types.JSONFloat64, len(fd.Data))
	missingMask := make([][]bool, len(fd.Data))
	hasMissing := false

	for i, row := range fd.Data {
		jsonData[i] = make([]types.JSONFloat64, len(row))
		missingMask[i] = make([]bool, len(row))
		for j, val := range row {
			jsonData[i][j] = types.JSONFloat64(val)
			if math.IsNaN(val) {
				missingMask[i][j] = true
				hasMissing = true
			}
		}
	}

	result := &FileDataJSON{
		Headers:            fd.Headers,
		RowNames:           fd.RowNames,
		Data:               jsonData,
		CategoricalColumns: fd.CategoricalColumns,
	}

	// Only include missing mask if there are missing values
	if hasMissing {
		result.MissingMask = missingMask
	}

	// Convert numeric target columns to JSON-safe format
	if len(fd.NumericTargetColumns) > 0 {
		result.NumericTargetColumns = make(map[string][]types.JSONFloat64)
		for colName, values := range fd.NumericTargetColumns {
			jsonValues := make([]types.JSONFloat64, len(values))
			for i, val := range values {
				jsonValues[i] = types.JSONFloat64(val)
			}
			result.NumericTargetColumns[colName] = jsonValues
		}
	}

	return result
}

// PCAResultJSON is a JSON-safe version of types.PCAResult
type PCAResultJSON struct {
	Scores               [][]types.JSONFloat64       `json:"scores"`
	Loadings             [][]types.JSONFloat64       `json:"loadings"`
	ExplainedVar         []types.JSONFloat64         `json:"explained_variance"`
	ExplainedVarRatio    []types.JSONFloat64         `json:"explained_variance_ratio"`
	CumulativeVar        []types.JSONFloat64         `json:"cumulative_variance"`
	ComponentLabels      []string                    `json:"component_labels"`
	VariableLabels       []string                    `json:"variable_labels,omitempty"`
	ComponentsComputed   int                         `json:"components_computed"`
	Method               string                      `json:"method"`
	PreprocessingApplied bool                        `json:"preprocessing_applied"`
	Means                []types.JSONFloat64         `json:"means,omitempty"`
	StdDevs              []types.JSONFloat64         `json:"stddevs,omitempty"`
	Metrics              []SampleMetricsJSON         `json:"metrics,omitempty"`
	T2Limit95            types.JSONFloat64           `json:"t2_limit_95,omitempty"`
	T2Limit99            types.JSONFloat64           `json:"t2_limit_99,omitempty"`
	QLimit95             types.JSONFloat64           `json:"q_limit_95,omitempty"`
	QLimit99             types.JSONFloat64           `json:"q_limit_99,omitempty"`
	Eigencorrelations    *EigencorrelationResultJSON `json:"eigencorrelations,omitempty"`
	AllEigenvalues       []types.JSONFloat64         `json:"all_eigenvalues,omitempty"`
}

// EigencorrelationResultJSON is a JSON-safe version of types.EigencorrelationResult
type EigencorrelationResultJSON struct {
	Correlations map[string][]types.JSONFloat64 `json:"correlations"`
	PValues      map[string][]types.JSONFloat64 `json:"pValues"`
	Variables    []string                       `json:"variables"`
	Components   []string                       `json:"components"`
	Method       string                         `json:"method"`
}

// SampleMetricsJSON is a JSON-safe version of types.SampleMetrics
type SampleMetricsJSON struct {
	HotellingT2 types.JSONFloat64 `json:"hotelling_t2"`
	Mahalanobis types.JSONFloat64 `json:"mahalanobis"`
	RSS         types.JSONFloat64 `json:"rss"`
	IsOutlier   bool              `json:"is_outlier"`
}

// ConvertPCAResultToJSON converts types.PCAResult to a JSON-safe version
func ConvertPCAResultToJSON(result *types.PCAResult) *PCAResultJSON {
	if result == nil {
		return nil
	}

	// Convert scores
	scores := make([][]types.JSONFloat64, len(result.Scores))
	for i, row := range result.Scores {
		scores[i] = make([]types.JSONFloat64, len(row))
		for j, val := range row {
			scores[i][j] = types.JSONFloat64(val)
		}
	}

	// Convert loadings
	loadings := make([][]types.JSONFloat64, len(result.Loadings))
	for i, row := range result.Loadings {
		loadings[i] = make([]types.JSONFloat64, len(row))
		for j, val := range row {
			loadings[i][j] = types.JSONFloat64(val)
		}
	}

	// Convert explained variance arrays
	explainedVar := make([]types.JSONFloat64, len(result.ExplainedVar))
	for i, val := range result.ExplainedVar {
		explainedVar[i] = types.JSONFloat64(val)
	}

	explainedVarRatio := make([]types.JSONFloat64, len(result.ExplainedVarRatio))
	for i, val := range result.ExplainedVarRatio {
		explainedVarRatio[i] = types.JSONFloat64(val)
	}

	cumulativeVar := make([]types.JSONFloat64, len(result.CumulativeVar))
	for i, val := range result.CumulativeVar {
		cumulativeVar[i] = types.JSONFloat64(val)
	}

	// Convert means and stddevs
	means := make([]types.JSONFloat64, len(result.Means))
	for i, val := range result.Means {
		means[i] = types.JSONFloat64(val)
	}

	stdDevs := make([]types.JSONFloat64, len(result.StdDevs))
	for i, val := range result.StdDevs {
		stdDevs[i] = types.JSONFloat64(val)
	}

	// Convert metrics
	metrics := make([]SampleMetricsJSON, len(result.Metrics))
	for i, m := range result.Metrics {
		metrics[i] = SampleMetricsJSON{
			HotellingT2: types.JSONFloat64(m.HotellingT2),
			Mahalanobis: types.JSONFloat64(m.Mahalanobis),
			RSS:         types.JSONFloat64(m.RSS),
			IsOutlier:   m.IsOutlier,
		}
	}

	// Convert eigencorrelations if present
	var eigencorrelations *EigencorrelationResultJSON
	if result.Eigencorrelations != nil {
		eigencorrelations = &EigencorrelationResultJSON{
			Variables:  result.Eigencorrelations.Variables,
			Components: result.Eigencorrelations.Components,
			Method:     result.Eigencorrelations.Method,
		}

		// Convert correlations map
		eigencorrelations.Correlations = make(map[string][]types.JSONFloat64)
		for variable, values := range result.Eigencorrelations.Correlations {
			jsonValues := make([]types.JSONFloat64, len(values))
			for i, val := range values {
				jsonValues[i] = types.JSONFloat64(val)
			}
			eigencorrelations.Correlations[variable] = jsonValues
		}

		// Convert p-values map
		eigencorrelations.PValues = make(map[string][]types.JSONFloat64)
		for variable, values := range result.Eigencorrelations.PValues {
			jsonValues := make([]types.JSONFloat64, len(values))
			for i, val := range values {
				jsonValues[i] = types.JSONFloat64(val)
			}
			eigencorrelations.PValues[variable] = jsonValues
		}
	}

	// Convert all eigenvalues
	var allEigenvalues []types.JSONFloat64
	if result.AllEigenvalues != nil {
		allEigenvalues = make([]types.JSONFloat64, len(result.AllEigenvalues))
		for i, val := range result.AllEigenvalues {
			allEigenvalues[i] = types.JSONFloat64(val)
		}
	}

	return &PCAResultJSON{
		Scores:               scores,
		Loadings:             loadings,
		ExplainedVar:         explainedVar,
		ExplainedVarRatio:    explainedVarRatio,
		CumulativeVar:        cumulativeVar,
		ComponentLabels:      result.ComponentLabels,
		VariableLabels:       result.VariableLabels,
		ComponentsComputed:   result.ComponentsComputed,
		Method:               result.Method,
		PreprocessingApplied: result.PreprocessingApplied,
		Means:                means,
		StdDevs:              stdDevs,
		Metrics:              metrics,
		T2Limit95:            types.JSONFloat64(result.T2Limit95),
		T2Limit99:            types.JSONFloat64(result.T2Limit99),
		QLimit95:             types.JSONFloat64(result.QLimit95),
		QLimit99:             types.JSONFloat64(result.QLimit99),
		Eigencorrelations:    eigencorrelations,
		AllEigenvalues:       allEigenvalues,
	}
}
