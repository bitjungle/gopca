// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

package types

import (
	"os"
	"regexp"
	"testing"
)

// declaredStrategyPattern matches a MissingValueStrategy constant declaration
// and captures the string it is bound to.
var declaredStrategyPattern = regexp.MustCompile(
	`\bMissing[A-Za-z]+\s+MissingValueStrategy\s*=\s*"([a-z]+)"`)

// TestAllMissingValueStrategiesIsComplete keeps the summary list honest.
//
// AllMissingValueStrategies is hand-maintained, and a hand-maintained list of
// constants falls behind the constants sooner or later — which is the very
// failure it was added to prevent one level up, where the JSON schema's enum had
// fallen behind the code. A list that can silently go stale is no better than
// the schema that could.
//
// Reading the source is unusual in a unit test and is the point: Go has no
// reflection over constants, so the only way to compare the list against the
// declarations is to read the declarations. The alternative — asserting the list
// has six entries, or listing them again in the test — restates the list rather
// than checking it, and would pass unchanged when a seventh constant appeared.
func TestAllMissingValueStrategiesIsComplete(t *testing.T) {
	source, err := os.ReadFile("pca.go")
	if err != nil {
		t.Fatalf("reading pca.go: %v", err)
	}

	matches := declaredStrategyPattern.FindAllStringSubmatch(string(source), -1)
	if len(matches) == 0 {
		t.Fatal("no MissingValueStrategy constants were found in pca.go, so this " +
			"test is passing by not looking; check the pattern against the source")
	}

	declared := make([]string, 0, len(matches))
	for _, m := range matches {
		declared = append(declared, m[1])
	}

	listed := make(map[string]bool, len(AllMissingValueStrategies))
	for _, s := range AllMissingValueStrategies {
		listed[string(s)] = true
	}

	for _, name := range declared {
		if !listed[name] {
			t.Errorf("MissingValueStrategy %q is declared in pca.go but missing from "+
				"AllMissingValueStrategies; the JSON schemas are validated against "+
				"that list, so leaving it out makes the schema reject a value the "+
				"code accepts", name)
		}
	}

	if len(AllMissingValueStrategies) != len(declared) {
		t.Errorf("AllMissingValueStrategies has %d entries but %d constants are "+
			"declared: %v against %v",
			len(AllMissingValueStrategies), len(declared),
			AllMissingValueStrategies, declared)
	}
}
