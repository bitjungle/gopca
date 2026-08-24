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

package cobra

import (
	"math"
	"strings"
	"testing"
)

// TestIssue809_UnsupportedMethodsAreRefused covers the two methods whose stored
// loadings cannot be used to project new data. Before the fix these reached
// ProjectData and panicked with index-out-of-range.
func TestIssue809_UnsupportedMethodsAreRefused(t *testing.T) {
	for _, tt := range []struct {
		method     string
		wantErr    bool
		wantPhrase string
	}{
		{"kernel", true, "training data"},
		{"Kernel", true, "training data"},
		{"temporal", true, "re-embedded"},
		{"svd", false, ""},
		{"nipals", false, ""},
		{"", false, ""},
	} {
		t.Run(tt.method, func(t *testing.T) {
			err := checkTransformSupported(tt.method)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("method %q was accepted; it cannot project new data", tt.method)
				}
				if !strings.Contains(err.Error(), tt.wantPhrase) {
					t.Errorf("error does not explain why: %v", err)
				}
				if !strings.Contains(err.Error(), "pca analyze") {
					t.Errorf("error does not say what to do instead: %v", err)
				}
			} else if err != nil {
				t.Errorf("method %q was refused: %v", tt.method, err)
			}
		})
	}
}

// TestIssue809_ProjectDataRejectsBadShapes checks that a malformed model yields
// an error rather than a panic. Each case is a shape a real or hand-edited model
// file can have; the kernel and temporal cases are the ones observed in #809.
func TestIssue809_ProjectDataRejectsBadShapes(t *testing.T) {
	data := [][]float64{{1, 2, 3}, {4, 5, 6}}
	for _, tt := range []struct {
		name     string
		data     [][]float64
		loadings [][]float64
	}{
		{"no loadings at all (kernel PCA)", data, [][]float64{}},
		{"fewer loading rows than variables (temporal PCA)", data, [][]float64{{1, 0}}},
		{"more loading rows than variables", data, [][]float64{{1, 0}, {0, 1}, {0, 0}, {0, 0}}},
		{"ragged loadings", data, [][]float64{{1, 0}, {0}, {0, 1}}},
		{"loadings with no components", data, [][]float64{{}, {}, {}}},
		{"no data rows", [][]float64{}, [][]float64{{1, 0}}},
		{"data rows with no columns", [][]float64{{}}, [][]float64{{1, 0}}},
		{"ragged data", [][]float64{{1, 2, 3}, {4, 5}}, [][]float64{{1, 0}, {0, 1}, {0, 0}}},
	} {
		t.Run(tt.name, func(t *testing.T) {
			// A panic here fails the test rather than taking down the process.
			got, err := ProjectData(tt.data, tt.loadings)
			if err == nil {
				t.Fatalf("accepted a malformed model and returned %v", got)
			}
		})
	}
}

// TestIssue809_ProjectDataStillProjects guards against the validation above
// rejecting input it should accept: the arithmetic must be unchanged.
func TestIssue809_ProjectDataStillProjects(t *testing.T) {
	// Two samples, three variables, projected onto two components that simply
	// select variable 1 and variable 3, so the expected scores are obvious.
	data := [][]float64{{1, 2, 3}, {4, 5, 6}}
	loadings := [][]float64{{1, 0}, {0, 0}, {0, 1}}

	got, err := ProjectData(data, loadings)
	if err != nil {
		t.Fatalf("a well-formed projection was refused: %v", err)
	}
	want := [][]float64{{1, 3}, {4, 6}}
	if len(got) != len(want) {
		t.Fatalf("got %d rows, want %d", len(got), len(want))
	}
	for i := range want {
		for j := range want[i] {
			if math.Abs(got[i][j]-want[i][j]) > 1e-12 {
				t.Errorf("scores[%d][%d] = %v, want %v", i, j, got[i][j], want[i][j])
			}
		}
	}
}
