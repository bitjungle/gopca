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
	"strings"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/spf13/cobra"
)

// newSavGolTestCmd builds a command carrying the flags, so the tests exercise
// the real registration and the real "was this flag set" logic rather than a
// hand-built struct. The distinction matters: the order/deriv rule depends
// entirely on cmd.Flags().Changed, which a struct literal cannot express.
func newSavGolTestCmd(args ...string) (*cobra.Command, *SavGolOptions, error) {
	opts := &SavGolOptions{}
	cmd := &cobra.Command{Use: "test", RunE: func(*cobra.Command, []string) error { return nil }}
	addSavGolFlags(cmd, opts)
	cmd.SetArgs(args)
	cmd.SilenceUsage = true
	cmd.SilenceErrors = true
	err := cmd.Execute()
	return cmd, opts, err
}

func TestSavGolFlagValidation(t *testing.T) {
	tests := []struct {
		name            string
		args            []string
		method          string
		missingStrategy string
		wantErr         bool
		contains        string
	}{
		{"nothing set", nil, "svd", "error", false, ""},
		{"complete and valid", []string{"--savgol-window", "11", "--savgol-order", "2", "--savgol-deriv", "1"}, "svd", "error", false, ""},
		{"window alone uses the defaults", []string{"--savgol-window", "11"}, "svd", "error", false, ""},

		// The trap this rule exists for: the user believes they took a
		// derivative, and the flag was silently inert.
		{"deriv without window", []string{"--savgol-deriv", "1"}, "svd", "error", true, "no effect without --savgol-window"},
		{"order without window", []string{"--savgol-order", "3"}, "svd", "error", true, "no effect without --savgol-window"},
		{"both without window", []string{"--savgol-order", "3", "--savgol-deriv", "1"}, "svd", "error", true, "--savgol-order and --savgol-deriv"},

		// Shape, rejected before the file is even opened.
		{"even window", []string{"--savgol-window", "10"}, "svd", "error", true, "must be odd"},

		// A negative window is not a request to disable the filter. Cobra takes
		// it, Enabled() reads it as off, and the run would otherwise proceed
		// unfiltered without a word.
		{"negative window", []string{"--savgol-window", "-5"}, "svd", "error", true, "must be a positive odd length"},
		{"explicit zero disables", []string{"--savgol-window", "0"}, "svd", "error", false, ""},
		{"order not below window", []string{"--savgol-window", "5", "--savgol-order", "5"}, "svd", "error", true, "less than the window length"},
		{"deriv above order", []string{"--savgol-window", "7", "--savgol-deriv", "3"}, "svd", "error", true, "zero everywhere"},

		// Combinations the engine cannot honour. Accepting these and ignoring
		// them is the failure mode --snv already has with temporal PCA.
		{"temporal", []string{"--savgol-window", "11"}, "temporal", "error", true, "not supported with --method temporal"},
		{"temporal, mixed case", []string{"--savgol-window", "11"}, "Temporal", "error", true, "not supported with --method temporal"},
		{"native missing values", []string{"--savgol-window", "11"}, "nipals", "native", true, "missing-strategy native"},
		{"native, mixed case", []string{"--savgol-window", "11"}, "nipals", "Native", true, "missing-strategy native"},

		// Without a window, an unsupported method is not this flag's business.
		{"temporal with no filter", nil, "temporal", "error", false, ""},
		{"native with no filter", nil, "nipals", "native", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cmd, opts, err := newSavGolTestCmd(tt.args...)
			if err != nil {
				t.Fatalf("parsing flags %v: %v", tt.args, err)
			}
			err = opts.validate(cmd, tt.method, tt.missingStrategy)
			if !tt.wantErr {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error, got none")
			}
			if !strings.Contains(err.Error(), tt.contains) {
				t.Errorf("error %q does not mention %q", err.Error(), tt.contains)
			}
		})
	}
}

// TestSavGolFlagsApplyToConfig checks the last hop on the command-line side.
// The flags could parse and validate perfectly and still never reach the engine.
func TestSavGolFlagsApplyToConfig(t *testing.T) {
	_, opts, err := newSavGolTestCmd("--savgol-window", "15", "--savgol-order", "3", "--savgol-deriv", "2")
	if err != nil {
		t.Fatalf("parsing flags: %v", err)
	}
	var config types.PCAConfig
	opts.applyTo(&config)
	if config.SavGolWindow != 15 || config.SavGolPolyOrder != 3 || config.SavGolDeriv != 2 {
		t.Errorf("config got window=%d order=%d deriv=%d, want 15/3/2",
			config.SavGolWindow, config.SavGolPolyOrder, config.SavGolDeriv)
	}

	// And with no window, nothing is written -- including the order and deriv
	// defaults, which would otherwise land in the model file as settings that
	// were never used.
	_, empty, err := newSavGolTestCmd()
	if err != nil {
		t.Fatalf("parsing flags: %v", err)
	}
	var untouched types.PCAConfig
	empty.applyTo(&untouched)
	if untouched.SavGolWindow != 0 || untouched.SavGolPolyOrder != 0 || untouched.SavGolDeriv != 0 {
		t.Errorf("an unset filter wrote window=%d order=%d deriv=%d into the config",
			untouched.SavGolWindow, untouched.SavGolPolyOrder, untouched.SavGolDeriv)
	}
}

// TestSavGolFlagsRegisteredOnBothCommands guards the symmetry. `pca regress`
// takes the same predictor-side preprocessing as `pca analyze`, and a user who
// explores with one and then models with the other would otherwise find the
// flag missing exactly when they came to rely on it.
func TestSavGolFlagsRegisteredOnBothCommands(t *testing.T) {
	for _, tc := range []struct {
		name string
		cmd  *cobra.Command
	}{
		{"analyze", NewAnalyzeCommand()},
		{"regress", NewRegressCommand()},
	} {
		t.Run(tc.name, func(t *testing.T) {
			for _, flag := range []string{"savgol-window", "savgol-order", "savgol-deriv"} {
				if tc.cmd.Flags().Lookup(flag) == nil {
					t.Errorf("pca %s does not accept --%s", tc.name, flag)
				}
			}
		})
	}
}
