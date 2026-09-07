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

package core_test

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
)

// savgolReference is the output of testdata/validation/generate_savgol_reference.py.
type savgolReference struct {
	ScipyMode string                     `json:"scipy_mode"`
	Delta     float64                    `json:"delta"`
	Configs   []savgolRefConfig          `json:"configs"`
	Synthetic map[string]savgolRefSignal `json:"synthetic"`
	Spectra   struct {
		Wavelengths []string          `json:"wavelengths"`
		Samples     []savgolRefSignal `json:"samples"`
	} `json:"spectra"`
}

type savgolRefConfig struct {
	WindowLength int `json:"window_length"`
	PolyOrder    int `json:"polyorder"`
	Deriv        int `json:"deriv"`
}

type savgolRefSignal struct {
	Input   []float64            `json:"input"`
	Outputs map[string][]float64 `json:"outputs"`
}

// TestValidateSavGolAgainstScipy compares the Go filter with
// scipy.signal.savgol_filter, signal by signal and variable by variable.
//
// The tolerance is tight on purpose. Both sides evaluate the same closed-form
// linear filter -- there is no optimiser here whose search path could
// legitimately differ, as there is for the power transforms. So any
// disagreement beyond floating-point noise is a disagreement about the filter,
// and loosening the tolerance would only hide it.
//
// The name begins with TestValidate so the CI validation filter catches it.
func TestValidateSavGolAgainstScipy(t *testing.T) {
	path := filepath.Join("..", "..", "testdata", "validation", "reference_results", "savgol_reference.json")
	requireReferences(t, "Savitzky-Golay", path)

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("reading %s: %v", path, err)
	}
	var ref savgolReference
	if err := json.Unmarshal(raw, &ref); err != nil {
		t.Fatalf("parsing %s: %v", path, err)
	}

	// The Go implementation has no spacing parameter and fits the polynomial to
	// the terminal window rather than padding. If the generator were ever
	// changed to a different scipy mode or spacing, every comparison below would
	// still run and would still mostly pass -- edge values would drift while the
	// interior stayed put, which reads as a small numerical discrepancy rather
	// than as two different filters. Check the premise instead of assuming it.
	if ref.ScipyMode != "interp" {
		t.Fatalf("the reference was generated with scipy mode %q, but the Go filter implements "+
			"mode \"interp\"; these are different filters at the edges", ref.ScipyMode)
	}
	if ref.Delta != 1.0 {
		t.Fatalf("the reference was generated with delta=%v, but the Go filter assumes unit spacing", ref.Delta)
	}

	compared := 0

	check := func(t *testing.T, label string, sig savgolRefSignal) {
		t.Helper()
		if len(sig.Outputs) == 0 {
			t.Fatalf("%s carries no filtered outputs, so this case would check nothing", label)
		}
		for _, cfg := range ref.Configs {
			key := fmt.Sprintf("w%d_p%d_d%d", cfg.WindowLength, cfg.PolyOrder, cfg.Deriv)
			want, ok := sig.Outputs[key]
			if !ok {
				continue // not every configuration applies to every signal length
			}

			filter, err := core.NewSavGol(core.SavGolConfig{
				WindowLength: cfg.WindowLength,
				PolyOrder:    cfg.PolyOrder,
				Deriv:        cfg.Deriv,
			}, len(sig.Input))
			if err != nil {
				t.Fatalf("%s/%s: building the filter failed: %v", label, key, err)
			}
			got, err := filter.Apply(sig.Input)
			if err != nil {
				t.Fatalf("%s/%s: applying the filter failed: %v", label, key, err)
			}
			if len(got) != len(want) {
				t.Fatalf("%s/%s: Go returned %d values, scipy %d", label, key, len(got), len(want))
			}

			// Scale the tolerance by the size of the signal the filter acted on,
			// not by each individual output. A derivative output can pass
			// through zero, and a purely relative test there demands agreement
			// far beyond what double precision offers.
			//
			// The worst measured disagreement across every case here is 6.4e-13,
			// so this leaves roughly two orders of magnitude of headroom for
			// LAPACK differences between platforms. That is still far tighter
			// than any real defect: a wrong coefficient shifts values in the
			// third decimal or worse, not the twelfth.
			var scale float64
			for _, v := range want {
				scale = math.Max(scale, math.Abs(v))
			}
			tol := 1e-10 * math.Max(1, scale)

			worst, worstAt := 0.0, -1
			for i := range got {
				if d := math.Abs(got[i] - want[i]); d > worst {
					worst, worstAt = d, i
				}
			}
			if worst > tol {
				t.Errorf("%s/%s: largest disagreement with scipy is %.3g at variable %d "+
					"(Go %.12g, scipy %.12g), tolerance %.3g",
					label, key, worst, worstAt, got[worstAt], want[worstAt], tol)
			}
			compared++
		}
	}

	t.Run("synthetic", func(t *testing.T) {
		if len(ref.Synthetic) == 0 {
			t.Fatal("the reference carries no synthetic signals")
		}
		for name, sig := range ref.Synthetic {
			sig := sig
			t.Run(name, func(t *testing.T) { check(t, "synthetic/"+name, sig) })
		}
	})

	t.Run("corn_spectra", func(t *testing.T) {
		if len(ref.Spectra.Samples) == 0 {
			t.Fatal("the reference carries no real spectra")
		}
		for i, sig := range ref.Spectra.Samples {
			sig := sig
			t.Run(fmt.Sprintf("sample_%d", i), func(t *testing.T) {
				check(t, fmt.Sprintf("corn/sample_%d", i), sig)
			})
		}
	})

	// A comparison that ran zero cases passes silently and proves nothing. The
	// reference has been reshaped twice already while this was being written.
	if compared == 0 {
		t.Fatal("no configurations were compared: the reference keys and the Go configurations do not line up")
	}
	t.Logf("compared %d filtered signals against scipy", compared)
}
