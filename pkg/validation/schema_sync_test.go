// GoPCA Suite
//
// Copyright © 2025-2026 Rune Mathisen <devel@bitjungle.com>
//
// This file is part of GoPCA Suite.
//
// See LICENSE for the full license terms.

package validation

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
)

// The v1 schemas exist in two copies and cannot be reduced to one.
//
//	schemas/v1/                 the published copy; what the $schema URLs point
//	                            at, and what a human edits
//	pkg/validation/schemas/v1/  the embedded copy; //go:embed cannot reach
//	                            outside its own package directory, so the bytes
//	                            the binary carries must live here
//
// Duplication that a language rule makes unavoidable is duplication that has to
// be watched instead of removed. It had already gone unwatched: the embedded
// copy omitted "zero" from the MissingValueStrategy enum for long enough that nobody
// remembers when it happened, while `pca analyze --missing-strategy zero` worked
// the whole time. Nothing failed, because nothing compared them.

// The versions to check. Derived from SupportedSchemaVersions rather than
// listed here, so a v3 is covered the moment the package can validate it --
// enumerating them by hand is how the CI package list fell thirteen packages
// behind (#836), and a schema copy that nothing compares is exactly the state
// #835 was filed about.
func schemaDirsFor(version string) (published, embedded string) {
	return filepath.Join("..", "..", "schemas", version), filepath.Join("schemas", version)
}

// TestSchemaCopiesAreIdentical fails on any difference between the two copies.
//
// The published copy is the source: it is the one the $schema URLs name and the
// one a person edits. `make sync-schemas` copies it over the embedded one.
func TestSchemaCopiesAreIdentical(t *testing.T) {
	if len(SupportedSchemaVersions) == 0 {
		t.Fatal("no supported schema versions, so this test is checking nothing")
	}
	for _, version := range SupportedSchemaVersions {
		t.Run(version, func(t *testing.T) {
			assertSchemaCopiesMatch(t, version)
		})
	}
}

func assertSchemaCopiesMatch(t *testing.T, version string) {
	t.Helper()
	publishedSchemaDir, embeddedSchemaDir := schemaDirsFor(version)
	published := schemaFilesIn(t, publishedSchemaDir)
	embedded := schemaFilesIn(t, embeddedSchemaDir)

	if len(published) == 0 {
		t.Fatalf("no schemas found in %s, so this test is passing by not looking",
			publishedSchemaDir)
	}

	names := map[string]bool{}
	for name := range published {
		names[name] = true
	}
	for name := range embedded {
		names[name] = true
	}
	sorted := make([]string, 0, len(names))
	for name := range names {
		sorted = append(sorted, name)
	}
	sort.Strings(sorted)

	for _, name := range sorted {
		a, inPublished := published[name]
		b, inEmbedded := embedded[name]

		switch {
		case !inEmbedded:
			t.Errorf("%s exists in %s but not in %s: the binary would not carry it",
				name, publishedSchemaDir, embeddedSchemaDir)
		case !inPublished:
			t.Errorf("%s exists in %s but not in %s: the binary carries a schema "+
				"that is not published", name, embeddedSchemaDir, publishedSchemaDir)
		case !bytes.Equal(a, b):
			t.Errorf("%s differs between the two copies. %s is the source; run "+
				"`make sync-schemas` to copy it over %s.\n%s",
				name, publishedSchemaDir, embeddedSchemaDir, firstDifference(a, b))
		}
	}
}

// TestSchemaMissingStrategyMatchesTheCode is the check the drift above should
// have run into long before a human noticed it.
//
// Comparing the two copies catches them disagreeing with each other. It cannot
// catch them agreeing with each other and both being wrong, which is the state
// they would have been in had the enum been written short in the first place.
// The enum is a claim about what the software accepts, so it is checked against
// the software.
func TestSchemaMissingStrategyMatchesTheCode(t *testing.T) {
	var dirs []string
	for _, version := range SupportedSchemaVersions {
		published, embedded := schemaDirsFor(version)
		dirs = append(dirs, published, embedded)
	}
	for _, dir := range dirs {
		t.Run(dir, func(t *testing.T) {
			raw, err := os.ReadFile(filepath.Join(dir, "common.schema.json"))
			if err != nil {
				t.Fatalf("reading common.schema.json: %v", err)
			}

			var doc struct {
				Definitions map[string]struct {
					Enum []string `json:"enum"`
				} `json:"definitions"`
			}
			if err := json.Unmarshal(raw, &doc); err != nil {
				t.Fatalf("parsing common.schema.json: %v", err)
			}

			definition, ok := doc.Definitions["MissingValueStrategy"]
			if !ok {
				t.Fatalf("common.schema.json has no MissingValueStrategy definition; the "+
					"available ones are %v", definitionNames(doc.Definitions))
			}
			if len(definition.Enum) == 0 {
				t.Fatal("MissingValueStrategy has no enum, so this test cannot check anything")
			}

			inSchema := map[string]bool{}
			for _, value := range definition.Enum {
				inSchema[value] = true
			}

			for _, strategy := range types.AllMissingValueStrategies {
				if !inSchema[string(strategy)] {
					t.Errorf("the code accepts missing strategy %q and the schema "+
						"rejects it; a model written with that setting fails "+
						"validation against a schema describing the tool that "+
						"wrote it", strategy)
				}
			}

			inCode := map[string]bool{}
			for _, strategy := range types.AllMissingValueStrategies {
				inCode[string(strategy)] = true
			}
			for _, value := range definition.Enum {
				if !inCode[value] {
					t.Errorf("the schema permits missing strategy %q, which the code "+
						"does not implement", value)
				}
			}
		})
	}
}

// schemaFilesIn reads every .json file in a directory, keyed by base name.
func schemaFilesIn(t *testing.T, dir string) map[string][]byte {
	t.Helper()

	entries, err := os.ReadDir(dir)
	if err != nil {
		t.Fatalf("reading %s: %v", dir, err)
	}

	files := map[string][]byte{}
	for _, entry := range entries {
		if entry.IsDir() || filepath.Ext(entry.Name()) != ".json" {
			continue
		}
		content, err := os.ReadFile(filepath.Join(dir, entry.Name()))
		if err != nil {
			t.Fatalf("reading %s: %v", entry.Name(), err)
		}
		files[entry.Name()] = content
	}
	return files
}

// firstDifference reports the first line that differs, so a failure names the
// change rather than only asserting that one exists.
func firstDifference(a, b []byte) string {
	left := bytes.Split(a, []byte("\n"))
	right := bytes.Split(b, []byte("\n"))

	for i := 0; i < len(left) && i < len(right); i++ {
		if !bytes.Equal(left[i], right[i]) {
			return fmt.Sprintf("  first difference at line %d:\n    published: %s\n    embedded:  %s",
				i+1, left[i], right[i])
		}
	}
	return fmt.Sprintf("  the files share a common prefix but differ in length: "+
		"%d lines against %d", len(left), len(right))
}

func definitionNames(definitions map[string]struct {
	Enum []string `json:"enum"`
}) []string {
	names := make([]string, 0, len(definitions))
	for name := range definitions {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
