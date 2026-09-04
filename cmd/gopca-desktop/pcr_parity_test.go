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

// Differential tests: the same data and the same settings must give the same
// regression through `pca regress` and through the desktop's RunPCR.
//
// The two paths share one engine but reach it by separate code. The CLI parses
// the file, filters rows and builds a types.PCRConfig in internal/cobra; the
// desktop receives an already-parsed matrix over JSON, restores its gaps from
// masks, and builds its own config in pcr.go. Nothing in the engine's tests can
// see a disagreement between its two callers, and that is where every serious
// defect of this feature has actually lived:
//
//   - Unmeasured responses arrived as measurements of zero. NaN marshals to null
//     and unmarshals into a float64 as 0, so 441 labelled rows presented as 880
//     and the model was fitted against values nobody recorded. Both sides agreed
//     on the type and disagreed on the value.
//   - The method name "SVD" was rejected as unknown. RunPCA folds case; the PCR
//     path did not, and the engine's own tests all pass lowercase.
//   - The missing-value strategy was ignored outright: the field existed on the
//     request, was never read, and no test asked whether it was.
//
// A test that hands Go structs straight to RunPCR cannot see the first of those,
// because the corruption happens in the marshalling. So the desktop leg here
// starts at App.LoadCSVFile and crosses encoding/json twice, once in each
// direction, exactly as the running application does.
//
// What this does not cover, stated so it is not assumed: the TypeScript that
// builds the request. usePCRRunner is covered by vitest, and the generated Wails
// bindings by neither. The remaining exposure is a frontend that sends a
// well-formed request with the wrong contents.

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/internal/cobra"
	"github.com/bitjungle/gopca/pkg/types"
)

// parityTolerance is tight on purpose. Both legs run the same engine on the same
// numbers, so the only differences that may appear are from formatting a float
// into JSON and reading it back, which is exact for float64 in Go's encoder.
// Anything larger means the two paths genuinely disagree.
const parityTolerance = 1e-12

// repoRoot locates the repository from this package's directory, so testdata
// paths do not depend on where the test was started.
func repoRoot() string { return filepath.Join("..", "..") }

// parityCase is one dataset and one set of user-visible settings, expressed once
// and translated into each path's own vocabulary.
type parityCase struct {
	name     string
	file     string
	response string

	// Predictor-side settings, shared by both legs.
	scale           string // "none", "standard"
	snv             bool
	missingStrategy string

	// Selection settings.
	components    int // 0 means choose by cross-validation
	maxComponents int
	cvFolds       int
	metric        string
	selectRule    string

	// why records what this case is here to catch, so a future reader can tell
	// whether it is still doing its job.
	why string
}

func parityCases() []parityCase {
	return []parityCase{
		{
			name: "corn standardized ten-fold", file: "testdata/corn/corn.csv",
			response: "Moisture#target", scale: "standard", missingStrategy: "error",
			maxComponents: 8, cvFolds: 10, metric: types.MetricRMSE, selectRule: types.SelectOneSE,
			why: "the ordinary path: p >> n, no missing values, cross-validated selection",
		},
		{
			name: "corn fixed component count", file: "testdata/corn/corn.csv",
			response: "Oil#target", scale: "standard", missingStrategy: "error",
			components: 5,
			why:        "no cross-validation runs, so the two legs must agree on the fixed-k branch",
		},
		{
			name: "corn selecting on MAE", file: "testdata/corn/corn.csv",
			response: "Protein#target", scale: "standard", missingStrategy: "error",
			maxComponents: 6, cvFolds: 5, metric: types.MetricMAE, selectRule: types.SelectOneSE,
			why: "the metric must survive the crossing; a dropped one silently means RMSE",
		},
		{
			name: "corn mean-centred only", file: "testdata/corn/corn.csv",
			response: "Moisture#target", scale: "none", missingStrategy: "error",
			maxComponents: 5, cvFolds: 5, metric: types.MetricRMSE, selectRule: types.SelectMin,
			why: "scaling off, which changes the coefficients rather than merely rescaling them",
		},
		{
			name: "corn with SNV", file: "testdata/corn/corn.csv",
			response: "Moisture#target", scale: "standard", snv: true, missingStrategy: "error",
			maxComponents: 5, cvFolds: 5, metric: types.MetricRMSE, selectRule: types.SelectOneSE,
			why: "row-wise preprocessing, where OriginalScaleValid is false and coefficients are absent",
		},
		{
			// iris ships species#target holding 0, 1 and 2 for three species, so
			// this is the first thing a reader is likely to try and the case where
			// both paths must say the fit means nothing. It is here for the
			// advisory, not for the arithmetic.
			name: "iris responding to a class-coded column", file: "testdata/iris/iris.csv",
			response: "species#target", scale: "standard", missingStrategy: "error",
			maxComponents: 3, cvFolds: 5, metric: types.MetricRMSE, selectRule: types.SelectOneSE,
			why: "a numeric column that is really a class label; both paths must caution",
		},
		{
			name: "bronir2 half the responses unmeasured", file: "testdata/bronir2/bronir2.csv",
			response: "Dens#target", scale: "standard", missingStrategy: "drop",
			maxComponents: 10, cvFolds: 5, metric: types.MetricRMSE, selectRule: types.SelectFirstMin,
			why: "the case that broke: NaN responses, a missing-value strategy, and a semi-supervised fit",
		},
	}
}

// TestCLIAndDesktopAgree is the differential test.
func TestCLIAndDesktopAgree(t *testing.T) {
	for _, tc := range parityCases() {
		t.Run(tc.name, func(t *testing.T) {
			// Both fixtures are tracked in git, so an absent one means a broken
			// checkout rather than an environment this test cannot run in.
			// Skipping would quietly drop the case that matters most — bronir2 is
			// the only fixture with unmeasured responses — and a run reporting
			// "ok" with the important half skipped is worse than a failure.
			path := filepath.Join(repoRoot(), tc.file)
			if _, err := os.Stat(path); err != nil {
				t.Fatalf("%s is tracked in git but missing here: %v", tc.file, err)
			}

			fromCLI := runCLILeg(t, tc, path)
			fromDesktop := runDesktopLeg(t, tc, path)

			compareRegressions(t, tc, fromCLI, fromDesktop)
		})
	}
}

// regressionFacts is what both legs must agree on, reduced to the quantities a
// user reads or a downstream tool consumes.
//
// The row counts are here for the same reason the coefficients are. The bronir2
// defect reported 880 labelled rows where 441 were labelled, and every
// coefficient was internally consistent with that wrong count; only the count
// itself gave the corruption away.
type regressionFacts struct {
	// Advisories is what the path told the reader about the response, over and
	// above the numbers. On the CLI leg it is the captured stdout; on the desktop
	// leg it is PCRResponse.Advisories. They are compared by containment rather
	// than equality, because the CLI wraps its text for a terminal.
	Advisories []string
	RawOutput  string

	Components     int
	LabelledRows   int
	Coefficients   []float64
	Intercept      float64
	OriginalScale  bool
	RMSEC          float64
	R2C            float64
	HasCV          bool
	Metric         string
	Rule           string
	Candidates     []int
	RMSECV         []float64
	MAE            []float64
	Q2             []float64
	Selected       int
	LowestError    int
	OutOfFoldFirst []float64
}

// runCLILeg drives `pca regress` into a temporary directory and reads back the
// model it wrote, which is the artifact a scripted user would consume.
func runCLILeg(t *testing.T, tc parityCase, path string) regressionFacts {
	t.Helper()

	out := t.TempDir()
	args := []string{
		"--response", tc.response,
		"--scale", tc.scale,
		"--missing-strategy", tc.missingStrategy,
		"--output", out,
		"--format", "table",
	}
	if tc.snv {
		args = append(args, "--snv")
	}
	if tc.components > 0 {
		args = append(args, "--components", fmt.Sprint(tc.components))
	} else {
		args = append(args,
			"--max-components", fmt.Sprint(tc.maxComponents),
			"--cv", fmt.Sprint(tc.cvFolds),
			"--metric", tc.metric,
			"--select", tc.selectRule,
			// Contiguous folds so the design does not depend on a shuffle. The
			// desktop leg is given the same scheme; a differential test must not
			// turn on which random generator each side happens to use.
			"--cv-scheme", "contiguous",
		)
	}
	args = append(args, path)

	cmd := cobra.NewRegressCommand()
	cmd.SetArgs(args)
	cmd.SetOut(&bytes.Buffer{})
	cmd.SetErr(&bytes.Buffer{})

	// The advisories go to stdout with fmt.Printf, not through cobra's writers,
	// so they have to be captured at the file descriptor. They are part of what
	// the two paths must agree on: a caution the CLI gives and the desktop does
	// not is a divergence even when every number matches, which is exactly how
	// the class-coded-response warning came to exist on one side only.
	printed := captureStdout(t, func() {
		if err := cmd.Execute(); err != nil {
			t.Fatalf("pca regress %v: %v", args, err)
		}
	})

	raw, err := os.ReadFile(filepath.Join(out, "pcr_model.json"))
	if err != nil {
		t.Fatalf("reading the model the CLI wrote: %v", err)
	}
	var model types.PCAOutputData
	if err := json.Unmarshal(raw, &model); err != nil {
		t.Fatalf("parsing pcr_model.json: %v", err)
	}
	if model.Regression == nil {
		t.Fatal("the CLI wrote a model with no regression block")
	}

	r := model.Regression
	facts := regressionFacts{
		RawOutput:     printed,
		Components:    r.Components,
		Coefficients:  r.Coefficients,
		Intercept:     r.InterceptOriginal,
		OriginalScale: r.OriginalScaleValid,
		RMSEC:         r.RMSEC,
		R2C:           r.R2C,
	}
	if cv := r.Validation; cv != nil {
		facts.HasCV = true
		facts.LabelledRows = cv.NSamples
		facts.Metric = cv.Metric
		facts.Rule = cv.Rule
		facts.Candidates = cv.Candidates
		facts.RMSECV = cv.RMSECV
		facts.MAE = cv.MAE
		facts.Q2 = cv.Q2
		facts.Selected = cv.Selected
		facts.LowestError = cv.LowestError
		facts.OutOfFoldFirst = cv.OutOfFold
	}
	return facts
}

// runDesktopLeg drives RunPCR the way the running application does: load the
// file through the same method the frontend calls, hand the result across JSON,
// rebuild the request from what survived, and hand that across JSON too.
func runDesktopLeg(t *testing.T, tc parityCase, path string) regressionFacts {
	t.Helper()

	app := NewApp()
	loaded, err := app.LoadCSVFile(path)
	if err != nil {
		t.Fatalf("LoadCSVFile: %v", err)
	}

	// The first crossing. The frontend does not receive FileDataJSON; it receives
	// whatever JSON.parse makes of it, in which a NaN has become null and no Go
	// type remains to say what it used to be. Decoding into `any` reproduces that
	// exactly, and forces this test to make the same null-handling decisions the
	// frontend makes rather than letting Go's decoder quietly make them.
	asSeenByTheFrontend := decodeAsJavaScript(t, loaded)

	data, missingMask := matrixFromJS(t, asSeenByTheFrontend["data"])
	responseValues, responseMissing := responseFromJS(t, asSeenByTheFrontend, tc.response)

	request := PCRRequest{
		PCA: PCARequest{
			Data:            data,
			MissingMask:     missingMask,
			Headers:         stringsFromJS(asSeenByTheFrontend["headers"]),
			RowNames:        stringsFromJS(asSeenByTheFrontend["rowNames"]),
			Method:          "SVD", // as the frontend spells it; case folding is the backend's job
			MeanCenter:      true,
			StandardScale:   tc.scale == "standard",
			SNV:             tc.snv,
			MissingStrategy: tc.missingStrategy,
		},
		Response:        tc.response,
		ResponseValues:  responseValues,
		ResponseMissing: responseMissing,
		Components:      tc.components,
		MaxComponents:   tc.maxComponents,
		CVFolds:         tc.cvFolds,
		CVScheme:        "contiguous",
		SelectRule:      tc.selectRule,
		Metric:          tc.metric,
	}

	// The second crossing, in the other direction. ResponseValues is []float64,
	// so a null here would arrive as a real zero; ResponseMissing exists because
	// of that and must survive alongside it.
	var delivered PCRRequest
	roundTrip(t, request, &delivered)

	response := app.RunPCR(delivered)
	if !response.Success {
		t.Fatalf("RunPCR: %s", response.Error)
	}

	// And the result travels back out the same way before anything reads it.
	var shown PCRResultJSON
	roundTrip(t, response.Result, &shown)

	facts := regressionFacts{
		Advisories:    response.Advisories,
		Components:    shown.Components,
		Coefficients:  floats(shown.Coefficients),
		Intercept:     float64(shown.InterceptOriginal),
		OriginalScale: shown.OriginalScaleValid,
		RMSEC:         float64(shown.RMSEC),
		R2C:           float64(shown.R2C),
		LabelledRows:  len(shown.LabelledRows),
	}
	if cv := shown.CV; cv != nil {
		facts.HasCV = true
		facts.Metric = cv.Metric
		facts.Rule = cv.Rule
		facts.Candidates = cv.Candidates
		facts.RMSECV = floats(cv.RMSECV)
		facts.MAE = floats(cv.MAE)
		facts.Q2 = floats(cv.Q2)
		facts.Selected = cv.Selected
		facts.LowestError = cv.LowestError
		facts.OutOfFoldFirst = floats(cv.OutOfFold)
	}
	return facts
}

func compareRegressions(t *testing.T, tc parityCase, cli, desktop regressionFacts) {
	t.Helper()

	compareAdvisories(t, cli.RawOutput, desktop.Advisories)

	if cli.Components != desktop.Components {
		t.Errorf("components: CLI %d, desktop %d (%s)",
			cli.Components, desktop.Components, tc.why)
	}
	if cli.OriginalScale != desktop.OriginalScale {
		t.Errorf("original_scale_valid: CLI %v, desktop %v",
			cli.OriginalScale, desktop.OriginalScale)
	}

	// Row counts before coefficients: a wrong count produces coefficients that
	// are perfectly self-consistent, and only the count reveals the cause.
	if cli.HasCV && desktop.HasCV && cli.LabelledRows != desktop.LabelledRows {
		t.Errorf("labelled rows: CLI %d, desktop %d — the two paths disagree about "+
			"which rows carry an observed response",
			cli.LabelledRows, desktop.LabelledRows)
	}

	closeEnough(t, "intercept", []float64{cli.Intercept}, []float64{desktop.Intercept})
	closeEnough(t, "rmsec", []float64{cli.RMSEC}, []float64{desktop.RMSEC})
	closeEnough(t, "r2c", []float64{cli.R2C}, []float64{desktop.R2C})
	closeEnough(t, "coefficients", cli.Coefficients, desktop.Coefficients)

	if cli.HasCV != desktop.HasCV {
		t.Fatalf("cross-validation ran on one path only: CLI %v, desktop %v",
			cli.HasCV, desktop.HasCV)
	}
	if !cli.HasCV {
		return
	}

	if cli.Metric != desktop.Metric {
		t.Errorf("metric: CLI %q, desktop %q — the two report different selection curves",
			cli.Metric, desktop.Metric)
	}
	if cli.Rule != desktop.Rule {
		t.Errorf("rule: CLI %q, desktop %q", cli.Rule, desktop.Rule)
	}
	if cli.Selected != desktop.Selected {
		t.Errorf("selected: CLI %d, desktop %d", cli.Selected, desktop.Selected)
	}
	if cli.LowestError != desktop.LowestError {
		t.Errorf("lowest_error: CLI %d, desktop %d", cli.LowestError, desktop.LowestError)
	}
	if len(cli.Candidates) != len(desktop.Candidates) {
		t.Fatalf("candidate counts differ: CLI %d, desktop %d",
			len(cli.Candidates), len(desktop.Candidates))
	}
	for i := range cli.Candidates {
		if cli.Candidates[i] != desktop.Candidates[i] {
			t.Fatalf("candidate %d: CLI %d, desktop %d",
				i, cli.Candidates[i], desktop.Candidates[i])
		}
	}

	closeEnough(t, "rmsecv", cli.RMSECV, desktop.RMSECV)
	closeEnough(t, "mae", cli.MAE, desktop.MAE)
	closeEnough(t, "q2", cli.Q2, desktop.Q2)
	closeEnough(t, "out-of-fold predictions", cli.OutOfFoldFirst, desktop.OutOfFoldFirst)
}

// closeEnough compares two float slices, reporting the first and worst
// disagreement rather than only that one exists.
func closeEnough(t *testing.T, what string, a, b []float64) {
	t.Helper()
	if len(a) != len(b) {
		t.Errorf("%s: CLI has %d values, desktop has %d", what, len(a), len(b))
		return
	}
	worst, at := 0.0, -1
	for i := range a {
		// NaN is a legitimate value here (Q2 on a constant response, for one), so
		// two NaNs agree. A NaN facing a number does not.
		if math.IsNaN(a[i]) && math.IsNaN(b[i]) {
			continue
		}
		d := math.Abs(a[i] - b[i])
		if scale := math.Abs(a[i]); scale > 1 {
			d /= scale
		}
		if d > worst || math.IsNaN(d) {
			worst, at = d, i
		}
	}
	if at >= 0 && !(worst <= parityTolerance) {
		t.Errorf("%s: worst disagreement at index %d, CLI %.17g against desktop %.17g "+
			"(difference %.3g, tolerance %g)",
			what, at, a[at], b[at], worst, parityTolerance)
	}
}

// --- crossing the boundary -------------------------------------------------

// decodeAsJavaScript marshals a value and decodes it into the untyped shape a
// browser would hold: maps, slices, float64 and nil. Decoding back into the
// original Go type would restore information JSON never carried and would make
// this test agree with itself.
func decodeAsJavaScript(t *testing.T, v any) map[string]any {
	t.Helper()
	raw, err := json.Marshal(v)
	if err != nil {
		t.Fatalf("marshalling for the frontend: %v", err)
	}
	var out map[string]any
	if err := json.Unmarshal(raw, &out); err != nil {
		t.Fatalf("decoding as the frontend would: %v", err)
	}
	return out
}

func roundTrip(t *testing.T, in, out any) {
	t.Helper()
	raw, err := json.Marshal(in)
	if err != nil {
		t.Fatalf("marshalling: %v", err)
	}
	if err := json.Unmarshal(raw, out); err != nil {
		t.Fatalf("unmarshalling: %v", err)
	}
}

// matrixFromJS converts the predictor matrix as the frontend holds it into the
// values-plus-mask pair the request carries. A null becomes a zero with its mask
// bit set, which is what App.RunPCA already requires and what usePCRRunner does.
func matrixFromJS(t *testing.T, v any) ([][]float64, [][]bool) {
	t.Helper()
	rows, ok := v.([]any)
	if !ok {
		t.Fatalf("data is %T, not an array", v)
	}
	values := make([][]float64, len(rows))
	mask := make([][]bool, len(rows))
	for i, row := range rows {
		cells, ok := row.([]any)
		if !ok {
			t.Fatalf("row %d is %T, not an array", i, row)
		}
		values[i] = make([]float64, len(cells))
		mask[i] = make([]bool, len(cells))
		for j, cell := range cells {
			if cell == nil {
				mask[i][j] = true
				continue
			}
			f, ok := cell.(float64)
			if !ok {
				t.Fatalf("cell [%d][%d] is %T, not a number or null", i, j, cell)
			}
			values[i][j] = f
		}
	}
	return values, mask
}

// responseFromJS is the same conversion for the response column, and is the step
// the original defect skipped. Without the mask, an unmeasured response arrives
// as a measurement of exactly zero and the row counts as observed.
func responseFromJS(t *testing.T, doc map[string]any, response string) ([]float64, []bool) {
	t.Helper()
	targets, ok := doc["numericTargetColumns"].(map[string]any)
	if !ok {
		t.Fatalf("numericTargetColumns is %T, not an object", doc["numericTargetColumns"])
	}
	column, ok := targets[response].([]any)
	if !ok {
		t.Fatalf("%q is not among the numeric responses", response)
	}
	values := make([]float64, len(column))
	missing := make([]bool, len(column))
	for i, cell := range column {
		if cell == nil {
			missing[i] = true
			continue
		}
		f, ok := cell.(float64)
		if !ok {
			t.Fatalf("response value %d is %T, not a number or null", i, cell)
		}
		values[i] = f
	}
	return values, missing
}

func stringsFromJS(v any) []string {
	items, ok := v.([]any)
	if !ok {
		return nil
	}
	out := make([]string, 0, len(items))
	for _, item := range items {
		s, _ := item.(string)
		out = append(out, s)
	}
	return out
}

func floats(values []types.JSONFloat64) []float64 {
	if values == nil {
		return nil
	}
	out := make([]float64, len(values))
	for i, v := range values {
		out[i] = float64(v)
	}
	return out
}

// --- advisories ------------------------------------------------------------

// compareAdvisories checks that the two paths caution the reader about the same
// things.
//
// This exists because the first version of the parity test compared only
// numbers, and so could not see the divergence that mattered most on the day it
// was written: `pca regress` warned that a response looked like a class label
// encoded as a number, and GoPCA Desktop offered the same column in a dropdown
// and said nothing. Every figure agreed to 1e-12. A model that is arithmetically
// identical on both paths and meaningless on both is still meaningless, and only
// one path said so.
//
// Containment rather than equality: the CLI wraps its text to a terminal width
// and prefixes "Warning: ", so the line breaks differ by design. Both sides are
// reduced to a single space-separated string before comparison, which ignores
// the wrapping and nothing else.
func compareAdvisories(t *testing.T, cliOutput string, desktopAdvisories []string) {
	t.Helper()

	flat := strings.Join(strings.Fields(cliOutput), " ")

	for _, advisory := range desktopAdvisories {
		want := strings.Join(strings.Fields(advisory), " ")
		if !strings.Contains(flat, want) {
			t.Errorf("the desktop gives an advisory the CLI does not:\n  %s", advisory)
		}
	}

	// The other direction. Without this the check passes whenever the desktop
	// says nothing, which is precisely the state being guarded against.
	if len(desktopAdvisories) == 0 && strings.Contains(cliOutput, "Warning:") {
		t.Errorf("the CLI printed a warning and the desktop returned none:\n%s",
			firstWarning(cliOutput))
	}
}

// firstWarning extracts the warning block from CLI output, for a readable
// failure message rather than the whole report.
func firstWarning(output string) string {
	lines := strings.Split(output, "\n")
	for i, line := range lines {
		if !strings.HasPrefix(line, "Warning:") {
			continue
		}
		end := i + 1
		for end < len(lines) && strings.TrimSpace(lines[end]) != "" {
			end++
		}
		return strings.Join(lines[i:end], "\n")
	}
	return output
}

// captureStdout redirects os.Stdout for the duration of fn.
//
// The CLI writes its report and its warnings with fmt.Printf, which goes to the
// process's stdout rather than through cobra's configured writer, so setting
// cmd.SetOut does not collect them.
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()

	read, write, err := os.Pipe()
	if err != nil {
		t.Fatalf("creating a pipe: %v", err)
	}
	original := os.Stdout
	os.Stdout = write

	// Drain concurrently: the CLI writes more than a pipe buffer holds on these
	// datasets, and reading only after fn returns would deadlock once it filled.
	done := make(chan string, 1)
	go func() {
		var buf bytes.Buffer
		_, _ = io.Copy(&buf, read)
		done <- buf.String()
	}()

	defer func() {
		os.Stdout = original
		_ = write.Close()
	}()

	fn()

	os.Stdout = original
	_ = write.Close()
	return <-done
}
