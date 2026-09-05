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
	"os"
	"path/filepath"
	"strings"
	"testing"

	pkgcsv "github.com/bitjungle/gopca/pkg/csv"
)

// TestRowNameHeaderSurvivesRoundTrip is the prerequisite for #859 and a bug in
// its own right.
//
// A file read and written back had its row-name column renamed to nothing:
// "Prove,By,Kvalitet" became ",By,Kvalitet". The name was dropped by the parser
// and no field existed to carry it, so promoting a column to row names could
// never have been undone -- there would be no name to restore.
//
// It went unnoticed because every dataset in this repo's testdata leaves that
// header blank, which is a common CSV convention. Files from instruments and
// LIMS systems routinely name it, and those are the files this loses.
func TestRowNameHeaderSurvivesRoundTrip(t *testing.T) {
	tests := []struct {
		name    string
		content string
		want    string
	}{
		{
			// Mixed text and numeric, which routes through CombineColumns --
			// a second place the header has to be carried across.
			name:    "named row-name column",
			content: "Prove,By,Score\nP1,Oslo,10\nP2,Bergen,12\n",
			want:    "Prove,By,Score",
		},
		{
			name:    "blank row-name header stays blank",
			content: ",By,Score\nP1,Oslo,10\nP2,Bergen,12\n",
			want:    ",By,Score",
		},
		{
			name:    "numeric data with a named id column",
			content: "SampleID,A,B\nS1,1,2\nS2,3,4\n",
			want:    "SampleID,A,B",
		},
	}

	app := &App{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data, err := app.parseCSVContent(tt.content, ".csv")
			if err != nil {
				t.Fatalf("parseCSVContent: %v", err)
			}

			out := filepath.Join(t.TempDir(), "out.csv")
			opts := pkgcsv.DefaultOptions()
			opts.HasHeaders = true
			opts.HasRowNames = len(data.RowNames) > 0
			written := &pkgcsv.Data{
				Headers:        data.Headers,
				RowNames:       data.RowNames,
				RowNamesHeader: data.RowNamesHeader,
				StringData:     data.Data,
				Rows:           data.Rows,
				Columns:        data.Columns,
			}
			if err := pkgcsv.SaveFile(out, written, opts); err != nil {
				t.Fatalf("SaveFile: %v", err)
			}

			raw, err := os.ReadFile(out)
			if err != nil {
				t.Fatalf("reading back: %v", err)
			}
			got := strings.SplitN(strings.TrimSpace(string(raw)), "\n", 2)[0]
			got = strings.TrimRight(got, "\r")
			if got != tt.want {
				t.Errorf("header line round-tripped as %q, want %q", got, tt.want)
			}
		})
	}
}
