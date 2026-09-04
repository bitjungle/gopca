// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

package cobra

import (
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

// TestDocumentedFlagsExist checks that every long flag named in the CLI
// reference is one the CLI actually accepts.
//
// A documented flag is a promise. Three of them were broken when this test was
// written, and each one wasted the time of any reader who trusted the page:
//
//	--exclude-cols   the flag is --exclude-columns
//	--quiet          no such flag, and no equivalent
//	--version        the version is a subcommand, `pca version`
//
// All three answered with "Error: unknown flag". None of them could have been
// caught by a Go test, a build, or a linter, because nothing connected the prose
// to the command tree — the documentation and the code were simply two
// independent assertions about the same thing, and only one of them was checked.
//
// The direction matters. This asserts documentation ⊆ implementation, so a flag
// the docs invent fails. It deliberately does not assert the converse: an
// undocumented flag is a gap worth closing but not a lie, and requiring every
// flag to appear here would turn each new one into a documentation chore
// enforced in the wrong place.
func TestDocumentedFlagsExist(t *testing.T) {
	reference := filepath.Join("..", "..", "docs", "cli_reference.md")
	content, err := os.ReadFile(reference)
	if err != nil {
		t.Fatalf("reading %s: %v", reference, err)
	}

	real := acceptedFlags(NewRootCommand())
	documented := flagsMentionedIn(string(content))

	if len(documented) == 0 {
		t.Fatal("no flags were extracted from the CLI reference, so this test is " +
			"passing by not looking; check the extractor against the document")
	}

	var missing []string
	for _, flag := range documented {
		if _, ok := real[flag]; !ok {
			missing = append(missing, flag)
		}
	}
	sort.Strings(missing)

	if len(missing) > 0 {
		t.Errorf("%s documents %d flag(s) the CLI does not accept: %s\n"+
			"Each one gives the reader \"Error: unknown flag\". Correct the "+
			"documentation, or implement the flag if it should exist.",
			reference, len(missing), strings.Join(missing, ", "))
	}
}

// TestDocumentedFlagsExtractorFindsFlags guards the extractor.
//
// The check above is worth nothing if flagsMentionedIn quietly returns an empty
// list, and a regex is exactly the sort of thing that starts matching nothing
// after an unrelated edit. TestDocumentedFlagsExist fails loudly on an empty
// result for the same reason; this pins the shapes it must handle.
func TestDocumentedFlagsExtractorFindsFlags(t *testing.T) {
	doc := "Use `--scale standard` or `--no-mean-centering`.\n" +
		"- `--help, -h` - show help\n" +
		"    pca analyze --exclude-rows 1,2 file.csv\n" +
		"---\n" +
		"An em-dash — and a bare -- should not appear.\n"

	got := flagsMentionedIn(doc)
	want := []string{"--exclude-rows", "--help", "--no-mean-centering", "--scale"}

	if strings.Join(got, " ") != strings.Join(want, " ") {
		t.Errorf("extracted %v, want %v", got, want)
	}
}

// acceptedFlags collects every long flag on the command tree, from the root and
// from every subcommand, local and persistent alike.
func acceptedFlags(root *cobra.Command) map[string]struct{} {
	flags := make(map[string]struct{})
	record := func(flag *pflag.Flag) { flags["--"+flag.Name] = struct{}{} }

	var walk func(*cobra.Command)
	walk = func(cmd *cobra.Command) {
		// Cobra registers --help lazily, when it first needs to show help, so a
		// tree that has never been executed does not have it yet. Without this the
		// test reports --help as undocumented-but-used, which is both wrong and
		// exactly the sort of false alarm that gets a check disabled.
		cmd.InitDefaultHelpFlag()

		cmd.Flags().VisitAll(record)
		cmd.PersistentFlags().VisitAll(record)
		cmd.InheritedFlags().VisitAll(record)
		for _, child := range cmd.Commands() {
			walk(child)
		}
	}
	walk(root)

	return flags
}

// documentedFlagPattern matches a long flag: two dashes, then a letter, then
// letters, digits and internal dashes.
//
// Requiring a letter immediately after the dashes is what keeps markdown's
// horizontal rules and its runs of dashes out of the result. Matching them and
// filtering later would work too, but this way the pattern says what a flag is.
var documentedFlagPattern = regexp.MustCompile(`--[a-z][a-z0-9]*(?:-[a-z0-9]+)*`)

// flagsMentionedIn returns the distinct long flags named in a document, sorted.
func flagsMentionedIn(document string) []string {
	seen := make(map[string]struct{})
	for _, match := range documentedFlagPattern.FindAllString(document, -1) {
		seen[match] = struct{}{}
	}
	out := make([]string, 0, len(seen))
	for flag := range seen {
		out = append(out, flag)
	}
	sort.Strings(out)
	return out
}
