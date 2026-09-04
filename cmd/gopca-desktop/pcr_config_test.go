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
	"math"
	"strings"
	"testing"
)

// frontendDefaultMethod is what DEFAULT_PCA_CONFIG in
// frontend/src/hooks/usePCAConfig.ts stores for the method. It is the display
// spelling, not the engine's, and passing it through unchanged is what made the
// first regression attempt fail with `unknown PCA method "SVD"`.
const frontendDefaultMethod = "SVD"

// TestRunPCRAcceptsTheMethodAsTheInterfaceSpellsIt is the regression test for
// that failure.
//
// The engine compares method names against lowercase constants, while the
// configuration panel stores them as they are displayed. Every path from the
// interface to the engine therefore has to fold the case, and this one did not.
// Nothing caught it, because the engine tests construct their own configuration
// and never see the interface's spelling.
func TestRunPCRAcceptsTheMethodAsTheInterfaceSpellsIt(t *testing.T) {
	app := &App{}

	response := app.RunPCR(PCRRequest{
		PCA: PCARequest{
			Data:            wellConditionedMatrix(24, 5),
			Headers:         []string{"a", "b", "c", "d", "e"},
			Method:          frontendDefaultMethod,
			MeanCenter:      true,
			MissingStrategy: "error",
		},
		Response:       "y#target",
		ResponseValues: linearResponse(24),
		Components:     2,
	})

	if !response.Success {
		t.Fatalf("the interface's own default method was rejected: %s", response.Error)
	}
	if response.Result == nil {
		t.Fatal("no result returned")
	}
	if response.Result.Components != 2 {
		t.Errorf("retained %d components, want 2", response.Result.Components)
	}
}

// TestRunPCRFoldsCaseOnEveryStringFromTheInterface checks the whole family
// rather than the one field that happened to break, since any of them could be
// respelled by a later change to the panel.
func TestRunPCRFoldsCaseOnEveryStringFromTheInterface(t *testing.T) {
	cases := []struct {
		name    string
		mutate  func(*PCRRequest)
		wantErr bool
	}{
		{"method in upper case", func(r *PCRRequest) { r.PCA.Method = "SVD" }, false},
		{"method in mixed case", func(r *PCRRequest) { r.PCA.Method = "NiPaLs" }, false},
		{"scheme in upper case", func(r *PCRRequest) {
			r.Components = 0
			r.MaxComponents = 3
			r.CVFolds = 4
			r.CVScheme = "RANDOM"
		}, false},
		{"metric in upper case", func(r *PCRRequest) {
			r.Components = 0
			r.MaxComponents = 3
			r.CVFolds = 4
			r.Metric = "RMSE"
		}, false},
		{"rule in upper case", func(r *PCRRequest) {
			r.Components = 0
			r.MaxComponents = 3
			r.CVFolds = 4
			r.SelectRule = "ONE-SE"
		}, false},
		{"a method that does not exist is still refused", func(r *PCRRequest) {
			r.PCA.Method = "Quantum"
		}, true},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := PCRRequest{
				PCA: PCARequest{
					Data:            wellConditionedMatrix(24, 5),
					Headers:         []string{"a", "b", "c", "d", "e"},
					Method:          "svd",
					MeanCenter:      true,
					MissingStrategy: "error",
				},
				Response:       "y#target",
				ResponseValues: linearResponse(24),
				Components:     2,
			}
			tc.mutate(&request)

			response := (&App{}).RunPCR(request)
			if tc.wantErr {
				if response.Success {
					t.Error("expected the request to be refused")
				}
				return
			}
			if !response.Success {
				t.Errorf("request was refused: %s", response.Error)
			}
		})
	}
}

// TestRunPCRExcludedRowsFilterTheResponseToo guards the alignment hazard: the
// predictors and the response are indexed by the same rows, so excluding a row
// from one without the other pairs every later sample with the wrong response.
//
// The check is made at full rank deliberately. With every component retained,
// PCR is ordinary least squares, so a response that is an exact linear function
// of the predictors must be reproduced exactly whichever rows survive. That is
// true of the aligned data and impossible for misaligned data, which makes it a
// real test of the filtering.
//
// A truncated fit would not do. Removing rows rotates the leading components, so
// R2C at k below full rank moves for reasons that have nothing to do with
// alignment; an earlier version of this test asserted against that movement and
// failed on correct code.
func TestRunPCRExcludedRowsFilterTheResponseToo(t *testing.T) {
	const rows, predictors = 24, 4
	data := wellConditionedMatrix(rows, predictors)
	y := linearResponse(rows)

	fit := func(excluded []int) *PCRResultJSON {
		t.Helper()
		response := (&App{}).RunPCR(PCRRequest{
			PCA: PCARequest{
				Data: data, Headers: []string{"a", "b", "c", "d"},
				Method: "svd", MeanCenter: true, MissingStrategy: "error",
				ExcludedRows: excluded,
			},
			Response: "y#target", ResponseValues: y, Components: predictors,
		})
		if !response.Success {
			t.Fatalf("fit failed: %s", response.Error)
		}
		return response.Result
	}

	baseline := fit(nil)
	if 1-float64(baseline.R2C) > 1e-9 {
		t.Fatalf("baseline R2C = %.6f; the response should be exactly linear in the "+
			"predictors, so this test cannot detect anything", float64(baseline.R2C))
	}

	excluded := fit([]int{0, 1, 2})
	if got := len(excluded.Fitted); got != rows-3 {
		t.Errorf("fitted %d rows, want %d after excluding 3", got, rows-3)
	}
	if 1-float64(excluded.R2C) > 1e-9 {
		t.Errorf("R2C = %.6f after excluding three rows. At full rank an exactly "+
			"linear response must still be reproduced exactly, so the response is no "+
			"longer aligned with its predictors", float64(excluded.R2C))
	}

	// Excluding from the far end as well, so an off-by-one in the filter that
	// happens to be harmless at the start of the data is still caught.
	tail := fit([]int{rows - 1, rows - 2})
	if 1-float64(tail.R2C) > 1e-9 {
		t.Errorf("R2C = %.6f after excluding the last two rows", float64(tail.R2C))
	}
}

func TestRunPCRRejectsMalformedRequests(t *testing.T) {
	base := func() PCRRequest {
		return PCRRequest{
			PCA: PCARequest{
				Data: wellConditionedMatrix(24, 4), Headers: []string{"a", "b", "c", "d"},
				Method: "svd", MeanCenter: true, MissingStrategy: "error",
			},
			Response: "y#target", ResponseValues: linearResponse(24), Components: 2,
		}
	}

	cases := []struct {
		name   string
		mutate func(*PCRRequest)
	}{
		{"no data", func(r *PCRRequest) { r.PCA.Data = nil }},
		{"no response chosen", func(r *PCRRequest) { r.Response = "" }},
		{"response length mismatch", func(r *PCRRequest) { r.ResponseValues = []float64{1, 2} }},
		{"unknown scheme", func(r *PCRRequest) {
			r.Components = 0
			r.MaxComponents = 3
			r.CVFolds = 4
			r.CVScheme = "sideways"
		}},
		{"grouping labels of the wrong length", func(r *PCRRequest) {
			r.Components = 0
			r.MaxComponents = 3
			r.CVFolds = 4
			r.CVGroupColumn = "batch"
			r.CVGroupLabels = []string{"a", "b"}
		}},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			request := base()
			tc.mutate(&request)
			if response := (&App{}).RunPCR(request); response.Success {
				t.Error("expected the request to be refused")
			}
		})
	}
}

// TestListResponsesSeparatesNumericFromCategorical checks that a column marked as
// a target but holding categories is named rather than silently omitted, so the
// interface can explain why it is not offered.
func TestListResponsesSeparatesNumericFromCategorical(t *testing.T) {
	options := (&App{}).ListResponses(&FileData{
		NumericTargetColumns: map[string][]float64{
			"Oil#target":      {1, 2},
			"Moisture#target": {3, 4},
		},
		CategoricalColumns: map[string][]string{
			"species#target": {"a", "b"},
			"batch":          {"x", "y"},
		},
	})

	want := []string{"Moisture#target", "Oil#target"}
	if len(options.Numeric) != len(want) {
		t.Fatalf("numeric responses = %v, want %v", options.Numeric, want)
	}
	for i, name := range want {
		if options.Numeric[i] != name {
			t.Errorf("numeric[%d] = %q, want %q (the list must be sorted so a picker "+
				"does not reshuffle between loads)", i, options.Numeric[i], name)
		}
	}

	if len(options.Categorical) != 1 || options.Categorical[0] != "species#target" {
		t.Errorf("categorical targets = %v, want only species#target; a plain "+
			"categorical column is not a target and should not be listed",
			options.Categorical)
	}
}

func TestListResponsesHandlesNil(t *testing.T) {
	options := (&App{}).ListResponses(nil)
	if options == nil || len(options.Numeric) != 0 || len(options.Categorical) != 0 {
		t.Errorf("expected empty options for nil data, got %+v", options)
	}
}

func TestIsTargetName(t *testing.T) {
	cases := map[string]bool{
		"Moisture#target": true,
		"MOISTURE#TARGET": true,
		"x #target":       true,
		"target":          false,
		"#targetish":      false,
		"":                false,
		"#tar":            false,
	}
	for name, want := range cases {
		if got := isTargetName(name); got != want {
			t.Errorf("isTargetName(%q) = %v, want %v", name, got, want)
		}
	}
}

// wellConditionedMatrix builds a deterministic matrix with a low condition
// number, following the guidance in CLAUDE.md that platform-sensitive numerical
// tests should not rest on ill-conditioned data.
func wellConditionedMatrix(rows, columns int) [][]float64 {
	data := make([][]float64, rows)
	for i := 0; i < rows; i++ {
		data[i] = make([]float64, columns)
		for j := 0; j < columns; j++ {
			angle := float64(i+1) * float64(j+1)
			data[i][j] = math.Sin(angle) + 0.5*math.Cos(angle/3) + float64(j)
		}
	}
	return data
}

// linearResponse is a clean function of the first two predictors, so a correct
// fit explains nearly all of it and a misaligned one does not.
func linearResponse(rows int) []float64 {
	data := wellConditionedMatrix(rows, 4)
	y := make([]float64, rows)
	for i := 0; i < rows; i++ {
		y[i] = 2*data[i][0] - 1.5*data[i][1] + 3
	}
	return y
}

// TestFrontendDefaultMethodIsStillSpeltThisWay documents the coupling this file
// depends on, so that a change to the panel's spelling is noticed here rather
// than by a user clicking Fit regression.
func TestFrontendDefaultMethodIsStillSpeltThisWay(t *testing.T) {
	if strings.ToLower(frontendDefaultMethod) != "svd" {
		t.Fatalf("frontendDefaultMethod %q no longer folds to a method the engine knows",
			frontendDefaultMethod)
	}
}

// matrixWithGaps returns a well-conditioned matrix with missing values punched
// into the rows named, mimicking a file where a few samples are incomplete.
func matrixWithGaps(rows, columns int, gapRows ...int) [][]float64 {
	data := wellConditionedMatrix(rows, columns)
	for _, row := range gapRows {
		data[row][columns-1] = math.NaN()
	}
	return data
}

// TestRunPCRAppliesTheMissingValueStrategy is the regression test for a strategy
// that was declared and ignored.
//
// The configuration panel offers a missing-value strategy, and the regression
// path passed the matrix to the engine untouched. The engine refuses incomplete
// predictors, so any file with a single gap could not be fitted at all, whatever
// the user chose. A control that accepts a choice and acts on none of them is
// worse than an absent one, because the user believes the gaps were handled.
func TestRunPCRAppliesTheMissingValueStrategy(t *testing.T) {
	const rows, columns = 30, 4
	data := matrixWithGaps(rows, columns, 2, 7, 19)
	y := linearResponse(rows)

	fit := func(strategy string) PCRResponse {
		t.Helper()
		return (&App{}).RunPCR(PCRRequest{
			PCA: PCARequest{
				Data: data, Headers: []string{"a", "b", "c", "d"},
				Method: "SVD", MeanCenter: true, MissingStrategy: strategy,
			},
			Response: "y#target", ResponseValues: y, Components: 2,
		})
	}

	t.Run("drop removes the incomplete rows", func(t *testing.T) {
		response := fit("drop")
		if !response.Success {
			t.Fatalf("drop was refused: %s", response.Error)
		}
		if got := len(response.Result.Fitted); got != rows-3 {
			t.Errorf("fitted %d rows, want %d after dropping 3 incomplete ones", got, rows-3)
		}
	})

	t.Run("zero keeps every row", func(t *testing.T) {
		response := fit("zero")
		if !response.Success {
			t.Fatalf("zero was refused: %s", response.Error)
		}
		if got := len(response.Result.Fitted); got != rows {
			t.Errorf("fitted %d rows, want all %d", got, rows)
		}
	})

	t.Run("an unset strategy explains what to choose", func(t *testing.T) {
		response := fit("error")
		if response.Success {
			t.Fatal("expected incomplete data with no strategy to be refused")
		}
		// The message has to be actionable: it must say what is wrong, how much of
		// the data is affected, and which choices resolve it.
		for _, expected := range []string{"missing", "3 rows", "drop", "zero"} {
			if !strings.Contains(response.Error, expected) {
				t.Errorf("the message does not mention %q: %s", expected, response.Error)
			}
		}
	})

	t.Run("learned imputation is refused", func(t *testing.T) {
		for _, strategy := range []string{"mean", "median"} {
			response := fit(strategy)
			if response.Success {
				t.Errorf("%s imputation should be refused before cross-validation", strategy)
				continue
			}
			if !strings.Contains(response.Error, "cross-validation") {
				t.Errorf("the refusal should say why: %s", response.Error)
			}
		}
	})
}

// TestRunPCRDropKeepsTheResponseAligned checks the alignment hazard on the
// missing-value path, which has the same shape as the exclusion path: dropping a
// row from the predictors without dropping it from the response pairs every later
// sample with the wrong measurement.
//
// Asserted at full rank, where an exactly linear response must be reproduced
// exactly whichever rows survive.
func TestRunPCRDropKeepsTheResponseAligned(t *testing.T) {
	const rows, columns = 30, 4
	data := matrixWithGaps(rows, columns, 1, 14, 28)
	y := linearResponse(rows)

	response := (&App{}).RunPCR(PCRRequest{
		PCA: PCARequest{
			Data: data, Headers: []string{"a", "b", "c", "d"},
			Method: "svd", MeanCenter: true, MissingStrategy: "drop",
		},
		Response: "y#target", ResponseValues: y, Components: columns,
	})
	if !response.Success {
		t.Fatalf("fit failed: %s", response.Error)
	}
	if 1-float64(response.Result.R2C) > 1e-9 {
		t.Errorf("R2C = %.6f after dropping incomplete rows. At full rank an exactly "+
			"linear response must still be reproduced exactly, so the response is no "+
			"longer aligned with its predictors", float64(response.Result.R2C))
	}
}

// TestRunPCRReportsRowsWithoutAResponse checks that rows whose response was never
// measured are counted and reported, and that they still reach the decomposition.
//
// That combination is the whole point of the semi-supervised path: a sample with
// predictors but no measurement still carries structure, and in calibration data
// such samples are often the majority. Silently discarding them would be a real
// loss with nothing on screen to show for it.
func TestRunPCRReportsRowsWithoutAResponse(t *testing.T) {
	const rows, columns = 30, 4
	data := wellConditionedMatrix(rows, columns)
	y := linearResponse(rows)

	unmeasured := []int{3, 8, 15, 22}
	for _, row := range unmeasured {
		y[row] = math.NaN()
	}

	response := (&App{}).RunPCR(PCRRequest{
		PCA: PCARequest{
			Data: data, Headers: []string{"a", "b", "c", "d"},
			Method: "svd", MeanCenter: true, MissingStrategy: "error",
		},
		Response: "y#target", ResponseValues: y, Components: 2,
	})
	if !response.Success {
		t.Fatalf("fit failed: %s", response.Error)
	}

	if got := len(response.Result.ExcludedRows); got != len(unmeasured) {
		t.Errorf("reported %d rows without a response, want %d; the interface reads this "+
			"field to tell the user", got, len(unmeasured))
	}
	if got := len(response.Result.LabelledRows); got != rows-len(unmeasured) {
		t.Errorf("regressed on %d rows, want %d", got, rows-len(unmeasured))
	}
	if got := len(response.Result.PCA.Scores); got != rows {
		t.Errorf("the decomposition used %d rows, want all %d: rows without a measured "+
			"response still carry predictor structure", got, rows)
	}
}

// TestRunPCRRestoresGapsFromTheMasks is the regression test for a corruption that
// happened entirely in transit.
//
// NaN cannot be represented in JSON. The engine marshals it as null, the
// interface holds it as null, and unmarshalling null back into a float64 yields
// zero. A response that was never measured therefore arrived as a genuine
// measurement of zero: every row counted as observed, and the model was fitted
// against numbers nobody recorded.
//
// On testdata/bronir2/bronir2.csv that produced a model of nought components
// reporting 880 observed responses when only 457 are measured, with an RMSEC of
// 472 against a response whose standard deviation is about 8. The engine was
// behaving correctly; it was being handed a different response from the one
// chosen. The predictor matrix has carried a mask for this reason since before
// regression existed, and this path ignored it.
//
// The request below is built the way the interface builds it, with gaps sent as
// zero alongside a mask, because that is the shape in which the bug occurs.
func TestRunPCRRestoresGapsFromTheMasks(t *testing.T) {
	const rows, columns = 30, 4
	data := wellConditionedMatrix(rows, columns)
	y := linearResponse(rows)

	unmeasured := map[int]bool{4: true, 11: true, 17: true, 25: true}

	// Serialise the way the round trip does: gaps become zero, and the mask
	// carries the knowledge that they were gaps.
	sentResponse := make([]float64, rows)
	responseMissing := make([]bool, rows)
	for i := 0; i < rows; i++ {
		if unmeasured[i] {
			sentResponse[i] = 0
			responseMissing[i] = true
			continue
		}
		sentResponse[i] = y[i]
	}

	// Two predictor cells are missing as well, sent the same way.
	sentData := make([][]float64, rows)
	predictorMask := make([][]bool, rows)
	for i := range data {
		sentData[i] = make([]float64, columns)
		copy(sentData[i], data[i])
		predictorMask[i] = make([]bool, columns)
	}
	sentData[2][1], predictorMask[2][1] = 0, true
	sentData[19][3], predictorMask[19][3] = 0, true

	response := (&App{}).RunPCR(PCRRequest{
		PCA: PCARequest{
			Data: sentData, MissingMask: predictorMask,
			Headers: []string{"a", "b", "c", "d"},
			Method:  "SVD", MeanCenter: true, MissingStrategy: "drop",
		},
		Response:        "y#target",
		ResponseValues:  sentResponse,
		ResponseMissing: responseMissing,
		Components:      2,
	})
	if !response.Success {
		t.Fatalf("fit failed: %s", response.Error)
	}

	// Two rows are dropped for incomplete predictors. Of the 28 that remain, the
	// four unmeasured responses leave the regression but stay in the decomposition.
	const droppedForPredictors = 2
	remaining := rows - droppedForPredictors
	wantUnmeasured := len(unmeasured)
	for row := range unmeasured {
		if row == 2 || row == 19 {
			wantUnmeasured-- // already removed for a missing predictor
		}
	}

	if got := len(response.Result.ExcludedRows); got != wantUnmeasured {
		t.Errorf("reported %d rows without a response, want %d. Zero is a legitimate "+
			"measurement, so a gap that arrives as zero is indistinguishable from one",
			got, wantUnmeasured)
	}
	if got := len(response.Result.LabelledRows); got != remaining-wantUnmeasured {
		t.Errorf("regressed on %d rows, want %d", got, remaining-wantUnmeasured)
	}
	if got := len(response.Result.PCA.Scores); got != remaining {
		t.Errorf("the decomposition used %d rows, want %d", got, remaining)
	}
}

// TestRunPCRWithoutMasksTreatsEveryValueAsMeasured pins the other half of the
// contract: absent a mask, the numbers are taken at face value. A caller that
// genuinely has complete data should not have to send an all-false mask.
func TestRunPCRWithoutMasksTreatsEveryValueAsMeasured(t *testing.T) {
	const rows, columns = 24, 4
	response := (&App{}).RunPCR(PCRRequest{
		PCA: PCARequest{
			Data: wellConditionedMatrix(rows, columns), Headers: []string{"a", "b", "c", "d"},
			Method: "svd", MeanCenter: true, MissingStrategy: "error",
		},
		Response: "y#target", ResponseValues: linearResponse(rows), Components: 2,
	})
	if !response.Success {
		t.Fatalf("fit failed: %s", response.Error)
	}
	if got := len(response.Result.ExcludedRows); got != 0 {
		t.Errorf("excluded %d rows from complete data, want none", got)
	}
	if got := len(response.Result.LabelledRows); got != rows {
		t.Errorf("regressed on %d rows, want all %d", got, rows)
	}
}
