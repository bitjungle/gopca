package csv

import (
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// TestIssue787_SampleNamesMatchScoreRows guards an inconsistency in exported
// models. Temporal PCA returns one score row per sliding window, T-L+1 of them,
// while RowNames still holds one entry per input row — so the export paired
// 14,976 names with 14,945 scores on the EEG dataset. Plotly zips elementwise
// and so displayed the right labels anyway, but a downstream consumer reading
// the JSON has no way to know which end of the longer array to discard.
func TestIssue787_SampleNamesMatchScoreRows(t *testing.T) {
	tests := []struct {
		name      string
		rowNames  []string
		scoreRows int
		wantNames int
	}{
		{"temporal PCA drops the trailing L-1 names", []string{"a", "b", "c", "d", "e"}, 3, 3},
		{"standard PCA is unaffected", []string{"a", "b", "c"}, 3, 3},
		{"fewer names than scores is left alone", []string{"a", "b"}, 3, 2},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			scores := make(types.Matrix, tt.scoreRows)
			for i := range scores {
				scores[i] = []float64{float64(i), float64(i) * 2}
			}
			result := &types.PCAResult{
				Scores:            scores,
				Loadings:          types.Matrix{{1, 0}, {0, 1}},
				ExplainedVar:      []float64{1, 0.5},
				ExplainedVarRatio: []float64{66.7, 33.3},
				CumulativeVar:     []float64{66.7, 100},
				ComponentLabels:   []string{"PC1", "PC2"},
				Method:            "temporal",
			}
			data := &Data{
				Headers:  []string{"x", "y"},
				RowNames: tt.rowNames,
				Matrix:   types.Matrix{{1, 2}, {3, 4}},
				Rows:     len(tt.rowNames),
				Columns:  2,
			}
			out := ConvertToPCAOutputData(result, data, nil, false, types.PCAConfig{Method: "temporal"}, nil, nil, nil)
			if got := len(out.Results.Samples.Names); got != tt.wantNames {
				t.Errorf("names = %d, want %d (scores = %d)", got, tt.wantNames, tt.scoreRows)
			}
			// The labels kept must be the leading ones: window i starts at input row i.
			for i, n := range out.Results.Samples.Names {
				if n != tt.rowNames[i] {
					t.Errorf("name[%d] = %q, want %q", i, n, tt.rowNames[i])
				}
			}
		})
	}
}
