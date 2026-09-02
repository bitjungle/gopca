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
	"math"
	"sort"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
)

// PCRRequest is a principal component regression request from the frontend.
//
// The predictor side reuses PCARequest verbatim, so every preprocessing and
// method option the Explore mode offers is available here without a second set
// of controls that could drift out of step with the first.
type PCRRequest struct {
	PCA PCARequest `json:"pca"`

	// Response names the numeric target column to predict. The frontend takes it
	// from FileData.NumericTargetColumns, which the parser has already separated
	// from the predictor matrix.
	Response string `json:"response"`

	// ResponseValues carries the response itself, indexed by original row. Sending
	// the values rather than looking them up again keeps the backend from having
	// to re-parse the file, and keeps one definition of which column was chosen.
	ResponseValues []float64 `json:"responseValues"`

	// Components fixes the retained count. Zero means choose it by
	// cross-validation, bounded by MaxComponents.
	Components    int `json:"components"`
	MaxComponents int `json:"maxComponents"`

	// CVFolds is the number of folds; zero means one fold per group, which is
	// leave-one-out under the default grouping.
	CVFolds  int    `json:"cvFolds"`
	CVScheme string `json:"cvScheme"`
	CVSeed   int64  `json:"cvSeed"`

	// CVGroupColumn and CVGroupLabels keep replicates of one object in the same
	// fold. The labels are sent alongside the name so the backend does not need
	// the whole categorical map.
	CVGroupColumn string   `json:"cvGroupColumn,omitempty"`
	CVGroupLabels []string `json:"cvGroupLabels,omitempty"`

	SelectRule string  `json:"selectRule"`
	Metric     string  `json:"metric"`
	Tolerance  float64 `json:"tolerance,omitempty"`
	WoldR      float64 `json:"woldR,omitempty"`
}

// PCRResponse carries the regression result back to the frontend.
type PCRResponse struct {
	Success bool           `json:"success"`
	Error   string         `json:"error,omitempty"`
	Result  *PCRResultJSON `json:"result,omitempty"`
	Info    string         `json:"info,omitempty"`
}

// RunPCR fits a principal component regression model.
//
// Row exclusions are applied to the predictors and the response together. They
// are indexed by the same original rows, so filtering one without the other
// would pair each surviving sample with a different sample's response, and every
// number downstream would be wrong while looking entirely ordinary.
func (a *App) RunPCR(request PCRRequest) (response PCRResponse) {
	defer func() {
		if r := recover(); r != nil {
			response = PCRResponse{
				Success: false,
				Error:   fmt.Sprintf("regression failed unexpectedly: %v", r),
			}
		}
	}()

	if len(request.PCA.Data) == 0 || len(request.PCA.Data[0]) == 0 {
		return PCRResponse{Success: false, Error: "no data to analyse"}
	}
	if request.Response == "" {
		return PCRResponse{Success: false, Error: "no response column selected"}
	}
	if len(request.ResponseValues) != len(request.PCA.Data) {
		return PCRResponse{
			Success: false,
			Error: fmt.Sprintf("the response has %d values but the data has %d rows",
				len(request.ResponseValues), len(request.PCA.Data)),
		}
	}

	data := request.PCA.Data
	y := request.ResponseValues
	groupLabels := request.CVGroupLabels

	if len(request.PCA.ExcludedRows) > 0 || len(request.PCA.ExcludedColumns) > 0 {
		filtered, err := utils.FilterMatrix(data, request.PCA.ExcludedRows, request.PCA.ExcludedColumns)
		if err != nil {
			return PCRResponse{Success: false, Error: fmt.Sprintf("failed to apply exclusions: %v", err)}
		}
		data = filtered

		if len(request.PCA.ExcludedRows) > 0 {
			y = filterFloatsByExcludedRows(y, request.PCA.ExcludedRows)
			if groupLabels != nil {
				groupLabels = filterStringsByExcludedRows(groupLabels, request.PCA.ExcludedRows)
			}
		}
	}

	config, err := buildDesktopPCRConfig(request, groupLabels, len(data))
	if err != nil {
		return PCRResponse{Success: false, Error: err.Error()}
	}

	engine := core.NewPCREngine()
	result, err := engine.Fit(data, y, config)
	if err != nil {
		return PCRResponse{Success: false, Error: err.Error()}
	}

	return PCRResponse{
		Success: true,
		Result:  ConvertPCRResultToJSON(result),
		Info:    describePCRFit(result),
	}
}

// ListResponses reports which columns of a loaded file can serve as a regression
// response, and which were marked as targets but cannot.
//
// Categorical targets are returned separately rather than omitted. A user who
// marked a column with #target and cannot find it in the picker is owed the
// reason, and "predicting a category is classification" is a better answer than
// an unexplained absence.
func (a *App) ListResponses(data *FileData) *ResponseOptions {
	options := &ResponseOptions{
		Numeric:     []string{},
		Categorical: []string{},
	}
	if data == nil {
		return options
	}

	for name := range data.NumericTargetColumns {
		options.Numeric = append(options.Numeric, name)
	}
	sort.Strings(options.Numeric)

	for name := range data.CategoricalColumns {
		if isTargetName(name) {
			options.Categorical = append(options.Categorical, name)
		}
	}
	sort.Strings(options.Categorical)

	return options
}

// ResponseOptions lists the columns available as a regression response.
type ResponseOptions struct {
	// Numeric columns can be predicted.
	Numeric []string `json:"numeric"`
	// Categorical columns were marked as targets but cannot be predicted by
	// regression; naming them lets the interface explain why.
	Categorical []string `json:"categorical"`
}

// isTargetName reports whether a column name carries the target marker.
func isTargetName(name string) bool {
	const marker = "#target"
	if len(name) < len(marker) {
		return false
	}
	lower := make([]byte, len(name))
	for i := 0; i < len(name); i++ {
		c := name[i]
		if c >= 'A' && c <= 'Z' {
			c += 'a' - 'A'
		}
		lower[i] = c
	}
	return string(lower[len(lower)-len(marker):]) == marker
}

// buildDesktopPCRConfig turns a frontend request into an engine configuration.
func buildDesktopPCRConfig(request PCRRequest, groupLabels []string, rows int) (types.PCRConfig, error) {
	// The configuration panel stores the method as it is displayed, so it arrives
	// as "SVD" rather than "svd" and the engine, which compares against lowercase
	// names, rejects it. RunPCA normalises the same field for the same reason.
	// Everything the frontend can send is folded here rather than trusted to
	// arrive in the right case, because the shared panel is free to change how it
	// labels its options without knowing this path exists.
	pca := types.PCAConfig{
		Method:          strings.ToLower(request.PCA.Method),
		MeanCenter:      request.PCA.MeanCenter,
		StandardScale:   request.PCA.StandardScale,
		RobustScale:     request.PCA.RobustScale,
		ScaleOnly:       request.PCA.ScaleOnly,
		SNV:             request.PCA.SNV,
		VectorNorm:      request.PCA.VectorNorm,
		MissingStrategy: types.MissingValueStrategy(strings.ToLower(request.PCA.MissingStrategy)),
	}

	config := types.PCRConfig{PCA: pca, Response: request.Response}

	if request.Components > 0 {
		config.PCA.Components = request.Components
		config.Selection = types.SelectionConfig{
			Mode:   "fixed",
			Fixed:  request.Components,
			Metric: metricOrDefault(request.Metric),
		}
		return config, nil
	}

	maxComponents := request.MaxComponents
	if maxComponents <= 0 {
		maxComponents = 20
	}
	config.PCA.Components = maxComponents

	scheme := strings.ToLower(request.CVScheme)
	switch scheme {
	case types.CVRandom, types.CVContiguous, types.CVForwardChaining:
	case "":
		scheme = types.CVRandom
	default:
		return config, fmt.Errorf("unknown validation scheme %q", request.CVScheme)
	}

	cv := types.CVConfig{
		Scheme: scheme,
		Folds:  request.CVFolds,
		Seed:   request.CVSeed,
	}

	if request.CVGroupColumn != "" {
		if len(groupLabels) != rows {
			return config, fmt.Errorf(
				"grouping column %q has %d values but the data has %d rows",
				request.CVGroupColumn, len(groupLabels), rows)
		}
		cv.GroupBy = request.CVGroupColumn
		cv.Groups = encodeDesktopGroups(groupLabels)
	}

	config.Selection = types.SelectionConfig{
		Mode:      "cv",
		Metric:    metricOrDefault(request.Metric),
		Rule:      ruleOrDefault(request.SelectRule),
		Tolerance: request.Tolerance,
		WoldR:     request.WoldR,
		CV:        cv,
	}
	return config, nil
}

func metricOrDefault(metric string) string {
	metric = strings.ToLower(metric)
	if metric == "" {
		return "rmse"
	}
	return metric
}

func ruleOrDefault(rule string) string {
	rule = strings.ToLower(rule)
	if rule == "" {
		return types.SelectOneSE
	}
	return rule
}

// encodeDesktopGroups maps categorical levels to dense identifiers, ordered by
// first appearance so the same file always produces the same design.
func encodeDesktopGroups(labels []string) []int {
	ids := make(map[string]int, len(labels))
	groups := make([]int, len(labels))
	for i, label := range labels {
		id, seen := ids[label]
		if !seen {
			id = len(ids)
			ids[label] = id
		}
		groups[i] = id
	}
	return groups
}

// filterFloatsByExcludedRows drops the excluded rows from a per-row vector.
func filterFloatsByExcludedRows(values []float64, excluded []int) []float64 {
	drop := make(map[int]bool, len(excluded))
	for _, row := range excluded {
		drop[row] = true
	}
	kept := make([]float64, 0, len(values))
	for i, v := range values {
		if !drop[i] {
			kept = append(kept, v)
		}
	}
	return kept
}

// filterStringsByExcludedRows drops the excluded rows from a per-row label slice.
func filterStringsByExcludedRows(values []string, excluded []int) []string {
	drop := make(map[int]bool, len(excluded))
	for _, row := range excluded {
		drop[row] = true
	}
	kept := make([]string, 0, len(values))
	for i, v := range values {
		if !drop[i] {
			kept = append(kept, v)
		}
	}
	return kept
}

// describePCRFit summarises the fit for the status line, naming what each error
// figure is so that RMSEC is not read as a performance estimate.
func describePCRFit(result *types.PCRResult) string {
	observed := len(result.LabelledRows)
	summary := fmt.Sprintf("%d components, %d rows with an observed response, RMSEC %.4g",
		result.Components, observed, result.RMSEC)

	if len(result.ExcludedRows) > 0 {
		summary += fmt.Sprintf(" (%d rows had no response and informed only the decomposition)",
			len(result.ExcludedRows))
	}
	if result.CV != nil {
		for i, k := range result.CV.Candidates {
			if k == result.Components && !math.IsNaN(result.CV.RMSECV[i]) {
				summary += fmt.Sprintf(", RMSECV %.4g", result.CV.RMSECV[i])
				break
			}
		}
	}
	return summary
}
