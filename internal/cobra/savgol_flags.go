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
	"fmt"
	"io"
	"strings"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/spf13/cobra"
)

// Savitzky-Golay flags, shared by `pca analyze` and `pca regress`.
//
// Both commands take the same predictor-side preprocessing, and the refusals
// below are the same in both. Defining them once means the two commands cannot
// drift into accepting different things -- which is exactly what a user
// comparing an analysis with the regression built on it would trip over.

// SavGolOptions holds the command-line form of a Savitzky-Golay configuration.
type SavGolOptions struct {
	Window    int
	PolyOrder int
	Deriv     int
}

// addSavGolFlags registers the flags on a command.
func addSavGolFlags(cmd *cobra.Command, opts *SavGolOptions) {
	cmd.Flags().IntVar(&opts.Window, "savgol-window", 0,
		"Savitzky-Golay window length in variables (odd, >2); 0 disables the filter")
	cmd.Flags().IntVar(&opts.PolyOrder, "savgol-order", 2,
		"Polynomial order fitted within each Savitzky-Golay window")
	cmd.Flags().IntVar(&opts.Deriv, "savgol-deriv", 0,
		"Derivative order: 0 smooths, 1 and 2 take the first and second derivative")
}

// Enabled reports whether the filter was asked for.
func (s SavGolOptions) Enabled() bool { return s.Window > 0 }

// validate checks the flags against each other and against the rest of the run.
//
// The order/deriv check exists because those two flags have defaults, so
// supplying them without a window is silently a no-op -- the user writes
// `--savgol-deriv 1`, sees a result, and believes they took a derivative. The
// method and missing-value checks refuse combinations the engine cannot honour,
// rather than accepting the flag and ignoring it, which is how `--snv` came to
// be accepted and then dropped by temporal PCA.
func (s SavGolOptions) validate(cmd *cobra.Command, method, missingStrategy string) error {
	orderSet := cmd.Flags().Changed("savgol-order")
	derivSet := cmd.Flags().Changed("savgol-deriv")

	// A negative window is not "no filter": nobody types -5 meaning off. Cobra
	// accepts it, Enabled() reads it as disabled, and the run would proceed
	// unfiltered without a word -- the same silent no-op the order/deriv rule
	// below exists to prevent.
	if cmd.Flags().Changed("savgol-window") && s.Window < 0 {
		return fmt.Errorf("--savgol-window must be a positive odd length, or 0 to disable the filter, got %d", s.Window)
	}

	if !s.Enabled() {
		if orderSet || derivSet {
			var given []string
			if orderSet {
				given = append(given, "--savgol-order")
			}
			if derivSet {
				given = append(given, "--savgol-deriv")
			}
			return fmt.Errorf("%s has no effect without --savgol-window, which enables the filter; "+
				"add --savgol-window with an odd length, or drop %s",
				strings.Join(given, " and "), strings.Join(given, " and "))
		}
		return nil
	}

	if err := (core.SavGolConfig{
		WindowLength: s.Window,
		PolyOrder:    s.PolyOrder,
		Deriv:        s.Deriv,
	}).ValidateShape(); err != nil {
		return err
	}

	// Temporal PCA builds its own preprocessor and applies no row-wise stage at
	// all, so a filter set here would be accepted and quietly discarded.
	if strings.ToLower(method) == "temporal" {
		return fmt.Errorf("Savitzky-Golay filtering is not supported with --method temporal: " +
			"temporal PCA works along the time axis and applies no transform along the variable axis")
	}

	// Native missing-value handling leaves NaN in the matrix, and a window
	// spanning a hole would spread it across the whole window. A derivative
	// across a gap in a spectrum means nothing.
	if strings.ToLower(missingStrategy) == "native" {
		return fmt.Errorf("Savitzky-Golay filtering cannot be combined with --missing-strategy native: " +
			"the filter needs a complete spectrum, since a window spanning a missing value " +
			"would spread it across every variable in that window; " +
			"use --missing-strategy mean, median or drop instead")
	}

	return nil
}

// applyTo copies the settings onto a PCA configuration.
func (s SavGolOptions) applyTo(config *types.PCAConfig) {
	if !s.Enabled() {
		return
	}
	config.SavGolWindow = s.Window
	config.SavGolPolyOrder = s.PolyOrder
	config.SavGolDeriv = s.Deriv
}

// warnIfAxisNotContinuous reports, on the given writer, anything about the
// variable axis that undermines a Savitzky-Golay filter.
//
// These are warnings rather than refusals. A scientist may know something the
// statistic does not -- variables can be genuinely ordered without their names
// saying so, and an axis can be deliberately irregular. Refusing would override
// that judgement; saying nothing would let a meaningless derivative pass for a
// result. So the software states what it measured and leaves the decision where
// it belongs.
func warnIfAxisNotContinuous(w io.Writer, report core.AxisReport) {
	// Nothing usable to measure -- every row missing a value, or flat. Saying
	// "these variables do not form a continuum" here would state as a finding
	// something that was never established.
	if !report.Measurable {
		return
	}

	if !report.IsContinuous {
		_, _ = fmt.Fprintf(w, "Warning: the %d variables do not form a continuum "+
			"(adjacent values differ only %.1fx less than a random ordering would). "+
			"Savitzky-Golay fits a polynomial across neighbouring variables, so on data like this "+
			"a derivative mostly amplifies noise. Check that the columns are in a measured order.\n",
			report.Variables, report.SmoothnessFactor)
		return
	}

	// Continuous, but the axis itself may still be broken. The continuity ratio
	// cannot see this: removing a band from the middle of a spectrum leaves the
	// data every bit as smooth while making non-adjacent wavelengths neighbours.
	if report.NamesNumeric && !report.SpacingUniform {
		_, _ = fmt.Fprintf(w, "Warning: the variables are not evenly spaced (%d different step sizes "+
			"between neighbours). Savitzky-Golay treats them as equally spaced, so wherever a gap "+
			"falls the filter is combining variables that are not really adjacent. "+
			"Excluding columns from the middle of a spectrum does this.\n",
			report.DistinctSteps)
	}
}
