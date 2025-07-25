package main

import (
	"encoding/json"
	"math"

	"github.com/bitjungle/gopca/pkg/types"
)

// JSONFloat64 is a float64 that marshals NaN and Inf values as null
type JSONFloat64 float64

func (f JSONFloat64) MarshalJSON() ([]byte, error) {
	if math.IsNaN(float64(f)) || math.IsInf(float64(f), 0) {
		return []byte("null"), nil
	}
	return json.Marshal(float64(f))
}

// FileDataJSON is a JSON-safe version of FileData
type FileDataJSON struct {
	Headers            []string            `json:"headers"`
	RowNames           []string            `json:"rowNames"`
	Data               [][]JSONFloat64     `json:"data"`
	MissingMask        [][]bool            `json:"missingMask,omitempty"`
	CategoricalColumns map[string][]string `json:"categoricalColumns,omitempty"`
}

// ToJSONSafe converts FileData to a JSON-safe version
func (fd *FileData) ToJSONSafe() *FileDataJSON {
	if fd == nil {
		return nil
	}

	// Convert float64 data to JSONFloat64 and build missing mask
	jsonData := make([][]JSONFloat64, len(fd.Data))
	missingMask := make([][]bool, len(fd.Data))
	hasMissing := false

	for i, row := range fd.Data {
		jsonData[i] = make([]JSONFloat64, len(row))
		missingMask[i] = make([]bool, len(row))
		for j, val := range row {
			jsonData[i][j] = JSONFloat64(val)
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

	return result
}

// PCAResultJSON is a JSON-safe version of types.PCAResult
type PCAResultJSON struct {
	Scores               [][]JSONFloat64       `json:"scores"`
	Loadings             [][]JSONFloat64       `json:"loadings"`
	ExplainedVar         []JSONFloat64         `json:"explained_variance"`
	ExplainedVarRatio    []JSONFloat64         `json:"explained_variance_ratio"`
	CumulativeVar        []JSONFloat64         `json:"cumulative_variance"`
	ComponentLabels      []string              `json:"component_labels"`
	VariableLabels       []string              `json:"variable_labels,omitempty"`
	ComponentsComputed   int                   `json:"components_computed"`
	Method               string                `json:"method"`
	PreprocessingApplied bool                  `json:"preprocessing_applied"`
	Means                []JSONFloat64         `json:"means,omitempty"`
	StdDevs              []JSONFloat64         `json:"stddevs,omitempty"`
	Metrics              []SampleMetricsJSON   `json:"metrics,omitempty"`
	T2Limit95            JSONFloat64           `json:"t2_limit_95,omitempty"`
	T2Limit99            JSONFloat64           `json:"t2_limit_99,omitempty"`
	QLimit95             JSONFloat64           `json:"q_limit_95,omitempty"`
	QLimit99             JSONFloat64           `json:"q_limit_99,omitempty"`
}

// SampleMetricsJSON is a JSON-safe version of types.SampleMetrics
type SampleMetricsJSON struct {
	HotellingT2 JSONFloat64 `json:"hotelling_t2"`
	Mahalanobis JSONFloat64 `json:"mahalanobis"`
	RSS         JSONFloat64 `json:"rss"`
	IsOutlier   bool        `json:"is_outlier"`
}

// ConvertPCAResultToJSON converts types.PCAResult to a JSON-safe version
func ConvertPCAResultToJSON(result *types.PCAResult) *PCAResultJSON {
	if result == nil {
		return nil
	}

	// Convert scores
	scores := make([][]JSONFloat64, len(result.Scores))
	for i, row := range result.Scores {
		scores[i] = make([]JSONFloat64, len(row))
		for j, val := range row {
			scores[i][j] = JSONFloat64(val)
		}
	}

	// Convert loadings
	loadings := make([][]JSONFloat64, len(result.Loadings))
	for i, row := range result.Loadings {
		loadings[i] = make([]JSONFloat64, len(row))
		for j, val := range row {
			loadings[i][j] = JSONFloat64(val)
		}
	}

	// Convert explained variance arrays
	explainedVar := make([]JSONFloat64, len(result.ExplainedVar))
	for i, val := range result.ExplainedVar {
		explainedVar[i] = JSONFloat64(val)
	}

	explainedVarRatio := make([]JSONFloat64, len(result.ExplainedVarRatio))
	for i, val := range result.ExplainedVarRatio {
		explainedVarRatio[i] = JSONFloat64(val)
	}

	cumulativeVar := make([]JSONFloat64, len(result.CumulativeVar))
	for i, val := range result.CumulativeVar {
		cumulativeVar[i] = JSONFloat64(val)
	}

	// Convert means and stddevs
	means := make([]JSONFloat64, len(result.Means))
	for i, val := range result.Means {
		means[i] = JSONFloat64(val)
	}

	stdDevs := make([]JSONFloat64, len(result.StdDevs))
	for i, val := range result.StdDevs {
		stdDevs[i] = JSONFloat64(val)
	}

	// Convert metrics
	metrics := make([]SampleMetricsJSON, len(result.Metrics))
	for i, m := range result.Metrics {
		metrics[i] = SampleMetricsJSON{
			HotellingT2: JSONFloat64(m.HotellingT2),
			Mahalanobis: JSONFloat64(m.Mahalanobis),
			RSS:         JSONFloat64(m.RSS),
			IsOutlier:   m.IsOutlier,
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
		T2Limit95:            JSONFloat64(result.T2Limit95),
		T2Limit99:            JSONFloat64(result.T2Limit99),
		QLimit95:             JSONFloat64(result.QLimit95),
		QLimit99:             JSONFloat64(result.QLimit99),
	}
}
