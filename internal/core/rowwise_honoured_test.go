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
	"math"
	"math/rand"
	"strings"
	"testing"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
)

// A preprocessing setting must either change the result or be refused. What it
// must never do is be accepted and quietly dropped.
//
// That is what happened with --snv under temporal PCA (#889): the flag was
// taken, the temporal engine built its own preprocessor with row-wise handling
// hard-coded off, and the analysis reported success. A user preprocessing a time
// series was told their data had been scatter-corrected when nothing of the sort
// had happened, and no output said otherwise.
//
// The specific bug is easy to fix once found. The reason it went unfound is that
// nothing connected "the configuration offers this" to "some method honours it":
// each engine was tested against configurations built for that engine. This test
// is the missing connection, and it is written over the whole grid rather than
// the one case that broke, so a method added later inherits the requirement.

// rowWiseSettings are the options that describe an operation along the variable
// axis. Add to this list when a new one appears.
var rowWiseSettings = []struct {
	name  string
	apply func(*types.PCAConfig)
}{
	{"SNV", func(c *types.PCAConfig) { c.SNV = true }},
	{"vector normalization", func(c *types.PCAConfig) { c.VectorNorm = true }},
	{"Savitzky-Golay smoothing", func(c *types.PCAConfig) {
		c.SavGolWindow, c.SavGolPolyOrder, c.SavGolDeriv = 9, 2, 0
	}},
	{"Savitzky-Golay derivative", func(c *types.PCAConfig) {
		c.SavGolWindow, c.SavGolPolyOrder, c.SavGolDeriv = 9, 2, 1
	}},
}

// TestEveryMethodHonoursOrRefusesRowWiseSettings walks every method against
// every row-wise setting.
func TestEveryMethodHonoursOrRefusesRowWiseSettings(t *testing.T) {
	data := smoothSeries(40, 24)

	methods := []struct {
		name      string
		configure func(*types.PCAConfig)
	}{
		{"svd", func(c *types.PCAConfig) { c.Method = "svd" }},
		{"nipals", func(c *types.PCAConfig) { c.Method = "nipals" }},
		{"kernel", func(c *types.PCAConfig) {
			c.Method = "kernel"
			c.KernelType = "rbf"
			c.KernelGamma = 0.1
			// Kernel PCA centers in kernel space, so it refuses column
			// centering; scale-only is the preprocessing it does accept.
			c.MeanCenter = false
			c.ScaleOnly = true
		}},
		{"temporal", func(c *types.PCAConfig) {
			c.Method = "temporal"
			c.TemporalLags = 4
		}},
	}

	for _, method := range methods {
		for _, setting := range rowWiseSettings {
			t.Run(method.name+"/"+setting.name, func(t *testing.T) {
				base := types.PCAConfig{Components: 2, MeanCenter: true}
				method.configure(&base)

				with := base
				setting.apply(&with)

				engine := core.NewPCAEngineForMethod(base.Method)
				got, err := engine.Fit(data, with)
				if err != nil {
					// Refused. That is a legitimate answer, provided it names
					// the setting so the user knows what to change.
					if !mentionsSetting(err.Error(), setting.name) {
						t.Errorf("%s refused %s but the message does not identify it: %v",
							method.name, setting.name, err)
					}
					return
				}

				// Accepted. Then it has to have done something: run the same
				// configuration without the setting and require a different
				// answer.
				baseline, err := core.NewPCAEngineForMethod(base.Method).Fit(data, base)
				if err != nil {
					t.Fatalf("%s failed without %s, so the comparison is impossible: %v",
						method.name, setting.name, err)
				}
				if scoresEqual(got.Scores, baseline.Scores) {
					t.Errorf("%s accepted %s and produced identical scores to running without it; "+
						"the setting was taken and then discarded, which tells the user their data "+
						"was preprocessed when it was not",
						method.name, setting.name)
				}
			})
		}
	}
}

// mentionsSetting checks the refusal names what was refused, allowing for the
// wording each message uses.
func mentionsSetting(message, setting string) bool {
	message = strings.ToLower(message)
	switch {
	case strings.Contains(setting, "SNV"):
		return strings.Contains(message, "snv")
	case strings.Contains(setting, "vector"):
		return strings.Contains(message, "vector norm")
	default:
		return strings.Contains(message, "savitzky-golay")
	}
}

func scoresEqual(a, b types.Matrix) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if len(a[i]) != len(b[i]) {
			return false
		}
		for j := range a[i] {
			// Sign is arbitrary in a decomposition, so compare magnitudes:
			// otherwise a flipped component would read as a difference and let
			// a dropped setting pass.
			if math.Abs(math.Abs(a[i][j])-math.Abs(b[i][j])) > 1e-9 {
				return false
			}
		}
	}
	return true
}

// smoothSeries builds data that every method can accept: enough rows for lags,
// smooth enough along the variables that Savitzky-Golay is meaningful, and
// varied enough that preprocessing visibly changes the decomposition.
func smoothSeries(rows, cols int) types.Matrix {
	rng := rand.New(rand.NewSource(20260909))
	data := make(types.Matrix, rows)
	for i := range data {
		data[i] = make([]float64, cols)
		scale := 1.0 + 0.4*float64(i%5)
		offset := 0.2 * float64(i%3)
		for j := range data[i] {
			t := float64(j)
			data[i][j] = offset + scale*(0.5+0.02*t+
				2.0*math.Exp(-((t-8.0)*(t-8.0))/6.0)+
				1.2*math.Exp(-((t-17.0)*(t-17.0))/4.0)) +
				rng.NormFloat64()*0.005
		}
	}
	return data
}
