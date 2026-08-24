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
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// defaultAnalyzeOptions mirrors the flag defaults of `pca analyze`, so the
// agreement test compares validate against analyze as a user would run it.
func defaultAnalyzeOptions(outputDir string) *AnalyzeOptions {
	return &AnalyzeOptions{
		Components:      2,
		Method:          "svd",
		MeanCenter:      true,
		Scale:           "none",
		Delimiter:       ",",
		NAValues:        ",NA,N/A,nan,NaN,null,NULL,m",
		MissingStrategy: "error",
		MissingPercent:  50,
		OutputFormat:    "table",
		OutputDir:       outputDir,
		OutputScores:    true,
		OutputLoadings:  true,
		OutputVariance:  true,
	}
}

// TestIssue810_ValidateAgreesWithAnalyze is the guard that matters. `pca
// validate` exists to predict whether `pca analyze` will work, so the two must
// not disagree about the same file. Before the fix, validate reported "Data is
// ready for PCA analysis" for a file with no numeric columns and analyze then
// refused it.
//
// Only configuration-independent outcomes are compared. validate takes no
// --method or --components, so it cannot and should not predict failures that
// depend on them; every fixture here therefore has either no numeric features
// or at least two, so the component count is never what decides the result.
func TestIssue810_ValidateAgreesWithAnalyze(t *testing.T) {
	for _, tt := range []struct {
		name       string
		csv        string
		wantAccept bool
	}{
		{
			name:       "ordinary numeric table",
			csv:        "id,a,b,c\ns1,1,2,3\ns2,4,5,6\ns3,7,8,10\ns4,2,9,4\n",
			wantAccept: true,
		},
		{
			name:       "no numeric columns at all",
			csv:        "id,a,b\ns1,x,y\ns2,p,q\ns3,m,n\n",
			wantAccept: false,
		},
		{
			name:       "only one sample",
			csv:        "id,a,b,c\ns1,1,2,3\n",
			wantAccept: false,
		},
		{
			name:       "a constant column is a warning, not a blocker",
			csv:        "id,a,b,c\ns1,1,2,5\ns2,4,5,5\ns3,7,8,5\ns4,2,9,5\n",
			wantAccept: true,
		},
	} {
		t.Run(tt.name, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "data.csv")
			if err := os.WriteFile(path, []byte(tt.csv), 0o600); err != nil {
				t.Fatalf("write fixture: %v", err)
			}

			validateErr := runValidate(&ValidateOptions{Delimiter: ","}, path)
			analyzeErr := runAnalyze(defaultAnalyzeOptions(filepath.Join(dir, "out")), path)

			validateAccepted := validateErr == nil
			analyzeAccepted := analyzeErr == nil

			if validateAccepted != analyzeAccepted {
				t.Errorf("validate and analyze disagree about the same file:\n"+
					"  validate accepted = %v (%v)\n"+
					"  analyze  accepted = %v (%v)",
					validateAccepted, validateErr, analyzeAccepted, analyzeErr)
			}
			if validateAccepted != tt.wantAccept {
				t.Errorf("validate accepted = %v, want %v (err: %v)",
					validateAccepted, tt.wantAccept, validateErr)
			}
		})
	}
}

// TestIssue810_ValidateNamesTheProblem checks the message, not just the exit
// status: a user who is told their file is unusable should learn why.
func TestIssue810_ValidateNamesTheProblem(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "alltext.csv")
	if err := os.WriteFile(path, []byte("id,a,b\ns1,x,y\ns2,p,q\ns3,m,n\n"), 0o600); err != nil {
		t.Fatalf("write fixture: %v", err)
	}

	err := runValidate(&ValidateOptions{Delimiter: ","}, path)
	if err == nil {
		t.Fatal("a file with no numeric columns was reported ready for PCA")
	}
	if !strings.Contains(err.Error(), "insufficient features") {
		t.Errorf("error does not say what is wrong with the file: %v", err)
	}
}
