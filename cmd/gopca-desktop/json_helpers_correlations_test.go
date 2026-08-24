package main

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
)

// TestIssue795_VariableCorrelationsReachTheFrontend covers a boundary that two
// earlier checks both missed. #793 added VariableCorrelations to
// types.PCAResult, verified the engine fills it, and confirmed the frontend
// compiles. Neither touched this file — and the desktop does not send
// types.PCAResult, it sends PCAResultJSON. The field was dropped in conversion,
// the plot menu (correctly) hid an entry it could not draw, and the Circle of
// Correlations disappeared from a working build.
func TestIssue795_VariableCorrelationsReachTheFrontend(t *testing.T) {
	data := types.Matrix{
		{5.1, 3.5, 1.4, 0.2}, {4.9, 3.0, 1.4, 0.2}, {6.2, 3.4, 5.4, 2.3},
		{5.9, 3.0, 5.1, 1.8}, {7.0, 3.2, 4.7, 1.4}, {6.4, 3.2, 4.5, 1.5},
		{5.5, 2.3, 4.0, 1.3}, {6.3, 3.3, 6.0, 2.5},
	}
	result, err := core.RunPCAWithDiagnostics(data, types.PCAConfig{
		Method: "svd", Components: 3, MeanCenter: true, StandardScale: true,
	})
	if err != nil {
		t.Fatalf("fit failed: %v", err)
	}
	if len(result.VariableCorrelations) == 0 {
		t.Fatal("the engine produced no correlations; this test can prove nothing")
	}

	jsonResult := ConvertPCAResultToJSON(result)
	if len(jsonResult.VariableCorrelations) != len(result.VariableCorrelations) {
		t.Fatalf("conversion dropped the correlations: %d rows in, %d out",
			len(result.VariableCorrelations), len(jsonResult.VariableCorrelations))
	}

	// The frontend reads the marshalled form, so check that, not the struct.
	encoded, err := json.Marshal(jsonResult)
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if !strings.Contains(string(encoded), `"variable_correlations"`) {
		t.Error(`"variable_correlations" is absent from the JSON the frontend receives`)
	}

	var decoded struct {
		VariableCorrelations [][]float64 `json:"variable_correlations"`
	}
	if err := json.Unmarshal(encoded, &decoded); err != nil {
		t.Fatalf("unmarshal failed: %v", err)
	}
	// Check the shape before indexing into it. A conversion that truncated a row
	// is exactly the bug this test guards, so it must report that plainly rather
	// than panic with an index-out-of-range stack trace. The row count is checked
	// again here because the assertion above covers conversion only — this one
	// covers the marshal/unmarshal leg, which is a separate opportunity to lose data.
	if len(decoded.VariableCorrelations) != len(result.VariableCorrelations) {
		t.Fatalf("the round trip changed the row count: %d in, %d out",
			len(result.VariableCorrelations), len(decoded.VariableCorrelations))
	}
	for j := range result.VariableCorrelations {
		if len(decoded.VariableCorrelations[j]) != len(result.VariableCorrelations[j]) {
			t.Fatalf("row %d changed width: %d columns in, %d out",
				j, len(result.VariableCorrelations[j]), len(decoded.VariableCorrelations[j]))
		}
		for k := range result.VariableCorrelations[j] {
			if got, want := decoded.VariableCorrelations[j][k], result.VariableCorrelations[j][k]; got != want {
				t.Errorf("value [%d][%d] survived as %v, want %v", j, k, got, want)
			}
		}
	}
}

// TestIssue795_AbsentCorrelationsStayAbsent confirms the omitempty path: where
// the engine cannot produce correlations the key must not appear at all, since
// the plot menu keys off its presence.
func TestIssue795_AbsentCorrelationsStayAbsent(t *testing.T) {
	encoded, err := json.Marshal(ConvertPCAResultToJSON(&types.PCAResult{
		Scores:          types.Matrix{{1, 2}},
		Loadings:        types.Matrix{{1, 0}, {0, 1}},
		ComponentLabels: []string{"PC1", "PC2"},
		Method:          "kernel",
	}))
	if err != nil {
		t.Fatalf("marshal failed: %v", err)
	}
	if strings.Contains(string(encoded), `"variable_correlations"`) {
		t.Error("empty correlations were serialised; the menu would offer a plot that cannot draw")
	}
}
