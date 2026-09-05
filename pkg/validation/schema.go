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

// Package validation provides JSON schema validation for PCA models.
//
// The schemas under schemas/v1 are the definition of a valid model file, and
// this package enforces them. That was not always so: for a long time the
// schemas were embedded, two of them were read into strings, neither was ever
// parsed, and a set of hand-written structural checks stood in their place
// under a comment reading "Full schema validation would require resolving all
// $ref references". The schema files therefore constrained nothing at runtime,
// and editing one had no effect on what pca transform would accept (#834).
//
// Resolving the $ref chain turned out to be straightforward. Every schema
// carries an absolute $id and refers to its neighbours by relative filename, so
// each reference resolves against the referrer's $id to exactly the $id of the
// schema being referred to. Pre-loading all seven under their own $id therefore
// resolves the whole graph without a single network fetch -- which matters,
// because a validator that reaches the network is one that fails in an air-gapped
// build and stalls in a slow one.
package validation

import (
	"embed"
	"encoding/json"
	"fmt"
	"io/fs"
	"path"
	"sort"
	"strings"
	"sync"

	"github.com/xeipuuv/gojsonschema"
)

//go:embed schemas/v1/*.json schemas/v2/*.json
var schemaFS embed.FS

// mainSchemaFile is the entry point of the schema graph; every other schema in
// the directory is reachable from it by $ref.
const mainSchemaFile = "pca-output.schema.json"

// ModelValidator validates PCA model JSON data against the v1 schemas.
type ModelValidator struct {
	version string

	// compiled caches one graph per schema version. Compiling parses seven
	// documents, and both callers -- pca transform and the desktop's export --
	// validate repeatedly within one process. A validator may see both versions
	// in one run, since a v1 model file stays readable.
	mu       sync.Mutex
	compiled map[string]*gojsonschema.Schema
}

// NewModelValidator creates a validator for the given schema version.
//
// Compilation is deferred to the first validation so that constructing a
// validator cannot fail for reasons a caller can do nothing about.
func NewModelValidator(version string) (*ModelValidator, error) {
	if version == "" {
		version = "v1"
	}
	dir := path.Join("schemas", version)
	if _, err := fs.Stat(schemaFS, dir); err != nil {
		return nil, fmt.Errorf("unknown schema version %q: no embedded schemas at %s", version, dir)
	}
	return &ModelValidator{version: version}, nil
}

// schema compiles the graph for a version, once per version.
func (v *ModelValidator) schema(version string) (*gojsonschema.Schema, error) {
	v.mu.Lock()
	defer v.mu.Unlock()
	if v.compiled == nil {
		v.compiled = make(map[string]*gojsonschema.Schema, len(SupportedSchemaVersions))
	}
	if compiled, ok := v.compiled[version]; ok {
		return compiled, nil
	}
	compiled, err := compileSchemaGraph(version)
	if err != nil {
		return nil, err
	}
	v.compiled[version] = compiled
	return compiled, nil
}

// compileSchemaGraph loads every embedded schema for a version and compiles the
// main one against them.
//
// The referenced schemas are registered before compiling so that gojsonschema
// resolves each $ref from the pre-loaded set rather than dereferencing the
// absolute URL in its $id. Nothing here touches the network.
func compileSchemaGraph(version string) (*gojsonschema.Schema, error) {
	dir := path.Join("schemas", version)
	entries, err := fs.ReadDir(schemaFS, dir)
	if err != nil {
		return nil, fmt.Errorf("reading embedded schemas: %w", err)
	}

	loader := gojsonschema.NewSchemaLoader()
	var main gojsonschema.JSONLoader
	var registered []string

	// Sorted, so a compilation failure reports the same schema every time
	// regardless of directory order.
	names := make([]string, 0, len(entries))
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".json") {
			names = append(names, e.Name())
		}
	}
	sort.Strings(names)

	for _, name := range names {
		raw, err := fs.ReadFile(schemaFS, path.Join(dir, name))
		if err != nil {
			return nil, fmt.Errorf("reading %s: %w", name, err)
		}
		if name == mainSchemaFile {
			main = gojsonschema.NewBytesLoader(raw)
			continue
		}
		if err := loader.AddSchemas(gojsonschema.NewBytesLoader(raw)); err != nil {
			return nil, fmt.Errorf("registering %s: %w", name, err)
		}
		registered = append(registered, name)
	}

	if main == nil {
		return nil, fmt.Errorf("%s is missing from the embedded schemas for %s", mainSchemaFile, version)
	}

	compiled, err := loader.Compile(main)
	if err != nil {
		return nil, fmt.Errorf("compiling %s against %v: %w", mainSchemaFile, registered, err)
	}
	return compiled, nil
}

// ValidateModel checks a model file against the v1 schema graph.
//
// Errors name the failing path, because "invalid model" tells a user nothing
// they can act on and a path tells them exactly where to look.
func (v *ModelValidator) ValidateModel(data []byte) error {
	var probe interface{}
	if err := json.Unmarshal(data, &probe); err != nil {
		return fmt.Errorf("invalid JSON: %w", err)
	}

	// The $schema field, when present, is a claim about which version wrote the
	// file. It is checked separately from the shape because a file written
	// against a future schema should say so rather than producing a list of
	// shape errors that all stem from the version mismatch.
	// A model is judged against the version it declares, not against whatever
	// this build happens to write. explained_variance_ratio is a percentage in
	// v1 and a fraction in v2, so the same numbers are valid under one schema
	// and out of range under the other -- validating a v1 file against v2 would
	// reject a perfectly good model.
	version := v.version
	if model, ok := probe.(map[string]interface{}); ok {
		version = versionOf(model, v.version)
		if declared, ok := model["$schema"].(string); ok {
			if err := checkSchemaVersion(declared, version); err != nil {
				return err
			}
		}
	}

	schema, err := v.schema(version)
	if err != nil {
		return err
	}

	result, err := schema.Validate(gojsonschema.NewBytesLoader(data))
	if err != nil {
		return fmt.Errorf("validating against the %s schema: %w", version, err)
	}
	if !result.Valid() {
		return describeFailures(result.Errors())
	}

	// The schema has approved the shape. What it cannot judge is whether the
	// fields agree with each other; see validateSemantics.
	if model, ok := probe.(map[string]interface{}); ok {
		return validateSemantics(model)
	}
	return nil
}

// checkSchemaVersion accepts the canonical URL and the relative spellings a file
// may reasonably carry.
func checkSchemaVersion(declared, version string) error {
	suffix := fmt.Sprintf("schemas/%s/%s", version, mainSchemaFile)
	if strings.HasSuffix(declared, suffix) {
		return nil
	}
	return fmt.Errorf("unknown schema version: %s (this build validates against %s)",
		declared, suffix)
}

// versionOf reads the schema version a model file declares.
//
// A v1 model reports explained_variance_ratio as a percentage and a v2 model as
// a fraction, so the same numbers are valid under one schema and out of range
// under the other. The file has to be judged against the version it was written
// for, which is exactly what $schema is there to say.
//
// Files with no $schema are treated as the current version. That is the only
// available guess, and it is the right one for anything this build wrote.
func versionOf(model map[string]interface{}, fallback string) string {
	declared, ok := model["$schema"].(string)
	if !ok {
		return fallback
	}
	for _, candidate := range SupportedSchemaVersions {
		if strings.Contains(declared, "schemas/"+candidate+"/") {
			return candidate
		}
	}
	return fallback
}

// SupportedSchemaVersions are the schema versions this build can validate
// against, newest first. CurrentSchemaVersion is what it writes.
var SupportedSchemaVersions = []string{"v2", "v1"}

// CurrentSchemaVersion is the version new model files declare.
const CurrentSchemaVersion = "v2"

// maxReportedFailures caps how many problems one error mentions.
//
// A model whose shape is badly wrong produces one failure per array element,
// which on a spectral dataset runs to hundreds of near-identical lines. The
// first few name the problem; the rest only bury it.
const maxReportedFailures = 8

func describeFailures(failures []gojsonschema.ResultError) error {
	var b strings.Builder
	b.WriteString("model does not match the schema:")
	for i, f := range failures {
		if i == maxReportedFailures {
			fmt.Fprintf(&b, "\n  ... and %d more", len(failures)-maxReportedFailures)
			break
		}
		field := f.Field()
		if field == "(root)" {
			fmt.Fprintf(&b, "\n  %s", f.Description())
			continue
		}
		fmt.Fprintf(&b, "\n  %s: %s", field, f.Description())
	}
	return fmt.Errorf("%s", b.String())
}

// ValidateWithSchema validates data against one named schema in the graph,
// rather than against a whole model file.
func ValidateWithSchema(data []byte, schemaName string, version string) error {
	if version == "" {
		version = "v1"
	}
	if !strings.HasSuffix(schemaName, ".json") {
		schemaName += ".schema.json"
	}

	raw, err := fs.ReadFile(schemaFS, path.Join("schemas", version, schemaName))
	if err != nil {
		return fmt.Errorf("unknown schema %q for version %s", schemaName, version)
	}

	// The neighbours are registered here too: a fragment schema may still refer
	// to common.schema.json for a shared definition.
	loader := gojsonschema.NewSchemaLoader()
	entries, err := fs.ReadDir(schemaFS, path.Join("schemas", version))
	if err != nil {
		return fmt.Errorf("reading embedded schemas: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || e.Name() == schemaName || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		other, err := fs.ReadFile(schemaFS, path.Join("schemas", version, e.Name()))
		if err != nil {
			return fmt.Errorf("reading %s: %w", e.Name(), err)
		}
		if err := loader.AddSchemas(gojsonschema.NewBytesLoader(other)); err != nil {
			return fmt.Errorf("registering %s: %w", e.Name(), err)
		}
	}

	compiled, err := loader.Compile(gojsonschema.NewBytesLoader(raw))
	if err != nil {
		return fmt.Errorf("compiling %s: %w", schemaName, err)
	}

	result, err := compiled.Validate(gojsonschema.NewBytesLoader(data))
	if err != nil {
		return fmt.Errorf("validating against %s: %w", schemaName, err)
	}
	if result.Valid() {
		return nil
	}
	return describeFailures(result.Errors())
}

// --- semantic checks -------------------------------------------------------

// validateSemantics enforces the invariants a JSON Schema cannot state.
//
// Draft-07 constrains each value on its own: types, ranges, enums, required
// keys. It has no way to say that one field's value must agree with another's
// length, so a model claiming twenty components while carrying two coefficients
// satisfies the schema completely. Those cross-field agreements are exactly the
// ones that make a model file self-contradictory rather than merely malformed,
// and they are the reason this function survives while the rest of the old
// hand-written validation was deleted in favour of the schema.
//
// The boundary is deliberate and worth keeping: shape belongs to the schema,
// agreement between fields belongs here. Anything expressible in the schema
// should be moved there, so there is one place to look for each kind of rule.
func validateSemantics(model map[string]interface{}) error {
	regression, ok := model["regression"].(map[string]interface{})
	if !ok {
		return nil // absent or not an object; the schema has already judged it
	}

	components, hasComponents := numberField(regression, "components")
	gamma, hasGamma := arrayField(regression, "score_coefficients")

	if hasComponents && hasGamma && int(components) != len(gamma) {
		return fmt.Errorf("model does not match the schema:\n  regression: claims %d components "+
			"but carries %d score coefficients; one per retained component is required",
			int(components), len(gamma))
	}

	// A model may legitimately have no collapsed original-scale form, under SNV
	// or vector normalisation. What it may not do is claim one and omit it: a
	// consumer trusting original_scale_valid would then read coefficients that
	// are not there and predict from nothing.
	if valid, ok := regression["original_scale_valid"].(bool); ok && valid {
		coefficients, hasCoefficients := arrayField(regression, "coefficients")
		if !hasCoefficients || len(coefficients) == 0 {
			return fmt.Errorf("model does not match the schema:\n  regression: " +
				"original_scale_valid is true but the model does not carry the " +
				"coefficients that claim promises")
		}
	}

	return nil
}

func numberField(m map[string]interface{}, key string) (float64, bool) {
	v, ok := m[key].(float64)
	return v, ok
}

func arrayField(m map[string]interface{}, key string) ([]interface{}, bool) {
	v, ok := m[key].([]interface{})
	return v, ok
}
