// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package integration

import "testing"

// ─── toSnakeCase ─────────────────────────────────────────────────────────────

func TestToSnakeCase_Simple(t *testing.T) {
	cases := []struct{ in, want string }{
		{"FooBar", "foo_bar"},
		{"Foo", "foo"},
		{"foo", "foo"},
		{"FooBarBaz", "foo_bar_baz"},
		{"", ""},
		{"A", "a"},
		{"ABC", "a_b_c"},
	}
	for _, c := range cases {
		got := toSnakeCase(c.in)
		if got != c.want {
			t.Errorf("toSnakeCase(%q) = %q, want %q", c.in, got, c.want)
		}
	}
}

// ─── StandardizeJSONTags ──────────────────────────────────────────────────────

func TestStandardizeJSONTags_NoOmitempty(t *testing.T) {
	cases := []struct{ field, want string }{
		{"FooBar", "fooBar"},
		{"Name", "name"},
		{"URL", "url"},
		{"PCAResult", "pcaResult"},
	}
	for _, c := range cases {
		got := StandardizeJSONTags(c.field, false)
		if got != c.want {
			t.Errorf("StandardizeJSONTags(%q, false) = %q, want %q", c.field, got, c.want)
		}
	}
}

func TestStandardizeJSONTags_WithOmitempty(t *testing.T) {
	got := StandardizeJSONTags("FooBar", true)
	want := "fooBar,omitempty"
	if got != want {
		t.Errorf("StandardizeJSONTags(%q, true) = %q, want %q", "FooBar", got, want)
	}
}

// ─── CheckJSONConsistency ─────────────────────────────────────────────────────

func TestCheckJSONConsistency_NonStruct(t *testing.T) {
	_, err := CheckJSONConsistency("not-a-struct")
	if err == nil {
		t.Error("expected error for non-struct input, got nil")
	}
}

func TestCheckJSONConsistency_SimpleStruct(t *testing.T) {
	type Sample struct {
		Name string `json:"name"`
		Age  int    `json:"age"`
	}
	results, err := CheckJSONConsistency(Sample{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(results) != 2 {
		t.Errorf("expected 2 results, got %d", len(results))
	}
}

func TestCheckJSONConsistency_SkipsDashTag(t *testing.T) {
	type Sample struct {
		Hidden string `json:"-"`
		Shown  string `json:"shown"`
	}
	results, err := CheckJSONConsistency(Sample{})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Only "Shown" should be in results; "Hidden" has json:"-"
	if len(results) != 1 {
		t.Errorf("expected 1 result (skipping json:\"-\"), got %d", len(results))
	}
}

func TestCheckJSONConsistency_PointerInput(t *testing.T) {
	type Sample struct {
		Value int `json:"value"`
	}
	results, err := CheckJSONConsistency(&Sample{})
	if err != nil {
		t.Fatalf("unexpected error for pointer: %v", err)
	}
	if len(results) != 1 {
		t.Errorf("expected 1 result, got %d", len(results))
	}
}

// ─── ValidateJSONMarshaling ───────────────────────────────────────────────────

func TestValidateJSONMarshaling_ValidStruct(t *testing.T) {
	// ValidateJSONMarshaling always reconstructs a pointer, so the caller must pass
	// a pointer to avoid a "data loss" false positive from the fmt.Sprintf comparison.
	type Simple struct {
		X int    `json:"x"`
		Y string `json:"y"`
	}
	if err := ValidateJSONMarshaling(&Simple{X: 1, Y: "hello"}); err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestValidateJSONMarshaling_EmptyStruct(t *testing.T) {
	type Empty struct {
		Val float64 `json:"val"`
	}
	if err := ValidateJSONMarshaling(&Empty{}); err != nil {
		t.Errorf("unexpected error for empty struct: %v", err)
	}
}
