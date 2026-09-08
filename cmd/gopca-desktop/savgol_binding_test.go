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
	"os"
	"path/filepath"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// The Savitzky-Golay settings travel from a TypeScript object to a Go struct by
// name, through Wails' JSON encoding. Nothing checks that the two agree: the
// frontend spreads its configuration into the request, so an extra field is
// silently carried and a misspelt one is silently dropped, and the Go side
// simply sees zero. A zero window reads as "no filter", which is a perfectly
// ordinary thing for it to be -- so the run succeeds, returns scores of the
// right shape, and is quietly not what was asked for.
//
// These tests compare the two sides directly rather than trusting that a
// rename would be noticed.

// savGolJSONTags returns the json tag names for the three fields of a struct.
func savGolJSONTags(t *testing.T, v interface{}) map[string]string {
	t.Helper()
	tags := map[string]string{}
	rt := reflect.TypeOf(v)
	for i := 0; i < rt.NumField(); i++ {
		field := rt.Field(i)
		if !strings.HasPrefix(field.Name, "SavGol") {
			continue
		}
		tag := field.Tag.Get("json")
		name, _, _ := strings.Cut(tag, ",")
		if name == "" || name == "-" {
			t.Fatalf("%s.%s has no usable json tag (%q); the frontend addresses it by name",
				rt.Name(), field.Name, tag)
		}
		tags[field.Name] = name
	}
	if len(tags) != 3 {
		t.Fatalf("%s declares %d Savitzky-Golay fields, want 3", rt.Name(), len(tags))
	}
	return tags
}

func readFrontendFile(t *testing.T, relative string) string {
	t.Helper()
	path := filepath.Join("frontend", "src", relative)
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	return string(content)
}

// TestSavGolFieldNamesMatchTheFrontend checks that every name the Go structs
// expect is a name the TypeScript actually sends.
func TestSavGolFieldNamesMatchTheFrontend(t *testing.T) {
	// Both request structs must agree with each other first: one is what the
	// analysis and regression paths receive, the other what the model export
	// receives, and the panel fills both from the same configuration object.
	requestTags := savGolJSONTags(t, PCARequest{})
	configTags := savGolJSONTags(t, PCAConfig{})
	if !reflect.DeepEqual(requestTags, configTags) {
		t.Fatalf("PCARequest and PCAConfig disagree about the field names:\n  PCARequest: %v\n  PCAConfig:  %v",
			requestTags, configTags)
	}

	sources := map[string]string{
		// Where the panel stores them.
		"hooks/usePCAConfig.ts": readFrontendFile(t, "hooks/usePCAConfig.ts"),
		// The interface describing what RunPCA receives.
		"types/index.ts": readFrontendFile(t, "types/index.ts"),
	}

	for field, tag := range requestTags {
		// A TypeScript property may be declared optional, so the colon can be
		// preceded by "?". Matching the bare name would also match it inside a
		// comment or a longer identifier, which is why the punctuation is part
		// of the pattern.
		declaration := regexp.MustCompile(`\b` + regexp.QuoteMeta(tag) + `\??\s*:`)
		for name, content := range sources {
			if !declaration.MatchString(content) {
				t.Errorf("Go field %s is sent as %q, but %s never mentions it; "+
					"the value would arrive as zero, which reads as \"no filter\"",
					field, tag, name)
			}
		}
	}
}

// TestSavGolDefaultsAreOffAndValid pins the frontend's starting point. A
// non-zero default would apply a filter nobody asked for, and a default order
// of zero would make the first derivative unavailable the moment it was chosen.
func TestSavGolDefaultsAreOffAndValid(t *testing.T) {
	content := readFrontendFile(t, "hooks/usePCAConfig.ts")
	for _, want := range []string{
		"savgolWindow: 0",
		"savgolPolyOrder: 2",
		"savgolDeriv: 0",
	} {
		if !strings.Contains(content, want) {
			t.Errorf("DEFAULT_PCA_CONFIG does not contain %q", want)
		}
	}
}

// TestApplySavGolSettings covers the shared conversion every path uses.
func TestApplySavGolSettings(t *testing.T) {
	tests := []struct {
		name       string
		method     string
		window     int
		polyOrder  int
		deriv      int
		wantErr    string
		wantWindow int
	}{
		{"disabled leaves the config alone", "svd", 0, 2, 1, "", 0},
		{"applied", "svd", 11, 2, 1, "", 11},
		{"applied for kernel too", "kernel", 11, 2, 1, "", 11},
		{"even window refused", "svd", 10, 2, 1, "must be odd", 0},
		// Only zero disables. A negative window swallowed here would leave the
		// run unfiltered and silent, which is the failure this path guards.
		{"negative window refused", "svd", -5, 2, 1, "positive odd number", 0},
		{"order not below window", "svd", 5, 5, 0, "less than the window length", 0},
		{"deriv above order", "svd", 7, 2, 3, "zero everywhere", 0},
		{"temporal refused", "temporal", 11, 2, 1, "not supported with Temporal PCA", 0},
		// Without a filter there is nothing to refuse, so temporal is fine.
		{"temporal without a filter", "temporal", 0, 2, 0, "", 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := types.PCAConfig{Method: tt.method}
			err := applySavGolSettings(&config, tt.window, tt.polyOrder, tt.deriv)
			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("expected an error mentioning %q, got none", tt.wantErr)
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not mention %q", err.Error(), tt.wantErr)
				}
				if config.SavGolWindow != 0 {
					t.Errorf("a refused configuration still wrote window=%d", config.SavGolWindow)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if config.SavGolWindow != tt.wantWindow {
				t.Errorf("window = %d, want %d", config.SavGolWindow, tt.wantWindow)
			}
			if tt.wantWindow > 0 {
				if config.SavGolPolyOrder != tt.polyOrder || config.SavGolDeriv != tt.deriv {
					t.Errorf("order/deriv = %d/%d, want %d/%d",
						config.SavGolPolyOrder, config.SavGolDeriv, tt.polyOrder, tt.deriv)
				}
			}
		})
	}
}

// TestRunPCAAppliesSavGol checks the analysis path behaviourally, not
// structurally. The names could agree and the field still never be read.
func TestRunPCAAppliesSavGol(t *testing.T) {
	data := savGolSpectra(20, 40)
	headers := make([]string, 40)
	for i := range headers {
		headers[i] = "v" + strings.Repeat("x", i%3)
	}

	base := PCARequest{
		Data:            data,
		Headers:         headers,
		Components:      2,
		MeanCenter:      true,
		Method:          "SVD",
		MissingStrategy: "error",
	}

	plain := (&App{}).RunPCA(base)
	if !plain.Success {
		t.Fatalf("unfiltered run failed: %s", plain.Error)
	}

	filtered := base
	filtered.SavGolWindow = 9
	filtered.SavGolPolyOrder = 2
	filtered.SavGolDeriv = 1
	got := (&App{}).RunPCA(filtered)
	if !got.Success {
		t.Fatalf("filtered run failed: %s", got.Error)
	}

	// A first derivative of a peak on a sloping baseline is a different matrix
	// entirely, so the loadings cannot coincide. Comparing loadings rather than
	// scores avoids the sign ambiguity of the decomposition being mistaken for
	// a difference.
	same := true
	for j := range got.Result.Loadings {
		for k := range got.Result.Loadings[j] {
			a := plain.Result.Loadings[j][k]
			b := got.Result.Loadings[j][k]
			if diff := a - b; diff > 1e-9 || diff < -1e-9 {
				same = false
			}
		}
	}
	if same {
		t.Error("the filtered and unfiltered runs produced identical loadings; " +
			"the Savitzky-Golay settings did not reach the engine")
	}
}

// TestRunPCARefusesSavGolWithTemporal checks that the refusal reaches the user
// rather than the setting being silently discarded, which is what --snv does on
// this path today (#889).
func TestRunPCARefusesSavGolWithTemporal(t *testing.T) {
	response := (&App{}).RunPCA(PCARequest{
		Data:            savGolSpectra(30, 20),
		Headers:         make([]string, 20),
		Components:      2,
		MeanCenter:      true,
		Method:          "temporal",
		TemporalLags:    3,
		MissingStrategy: "error",
		SavGolWindow:    9,
		SavGolPolyOrder: 2,
		SavGolDeriv:     1,
	})
	if response.Success {
		t.Fatal("Temporal PCA accepted a Savitzky-Golay filter it would never apply")
	}
	if !strings.Contains(response.Error, "Temporal PCA") {
		t.Errorf("error does not explain the combination: %s", response.Error)
	}
}

// TestExportedModelCarriesSavGol covers the third path, and the one whose
// failure would be hardest to notice: an exported model missing its filter
// still loads and still transforms.
func TestExportedModelCarriesSavGol(t *testing.T) {
	request := ExportPCAModelRequest{
		Config: PCAConfig{
			Components:      3,
			MeanCenter:      true,
			Method:          "svd",
			SNV:             true,
			SavGolWindow:    11,
			SavGolPolyOrder: 2,
			SavGolDeriv:     1,
		},
	}
	config := request.toEngineConfig()
	if config.SavGolWindow != 11 || config.SavGolPolyOrder != 2 || config.SavGolDeriv != 1 {
		t.Errorf("the exported configuration carries window=%d order=%d deriv=%d, want 11/2/1",
			config.SavGolWindow, config.SavGolPolyOrder, config.SavGolDeriv)
	}
}

// savGolSpectra builds smooth peaks on a sloping baseline: something a
// derivative visibly changes, unlike noise, where a filtered and an unfiltered
// decomposition can happen to look alike.
func savGolSpectra(rows, cols int) [][]float64 {
	data := make([][]float64, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		for j := range data[i] {
			t := float64(j)
			centre := 8.0 + float64(i%7)
			data[i][j] = 0.5 + 0.02*t + 3.0*math.Exp(-((t-centre)*(t-centre))/6.0)
		}
	}
	return data
}
