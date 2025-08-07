// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package main

import (
	"testing"
)

func TestCalculateEllipses(t *testing.T) {
	app := &App{}

	// Test with valid data - using data from core test that is known to work
	request := CalculateEllipsesRequest{
		Scores: [][]float64{
			// Group A - centered around (1, 1)
			{1.2, 1.1},
			{0.9, 1.3},
			{1.1, 0.8},
			{0.8, 1.2},
			{1.0, 1.0},
			// Group B - centered around (-1, -1)
			{-1.1, -0.9},
			{-0.9, -1.2},
			{-1.2, -1.1},
			{-0.8, -0.8},
			{-1.0, -1.0},
		},
		GroupLabels: []string{"A", "A", "A", "A", "A", "B", "B", "B", "B", "B"},
		XComponent:  0,
		YComponent:  1,
	}

	response := app.CalculateEllipses(request)

	if !response.Success {
		t.Errorf("Expected success but got error: %s", response.Error)
	}

	t.Logf("Response: Success=%v, 90=%d groups, 95=%d groups, 99=%d groups",
		response.Success, len(response.GroupEllipses90),
		len(response.GroupEllipses95), len(response.GroupEllipses99))

	if len(response.GroupEllipses90) == 0 {
		t.Error("Expected ellipses for 90% confidence level")
	}

	if len(response.GroupEllipses95) == 0 {
		t.Error("Expected ellipses for 95% confidence level")
	}

	if len(response.GroupEllipses99) == 0 {
		t.Error("Expected ellipses for 99% confidence level")
	}

	// Check that we have ellipses for both groups
	for _, ellipses := range []map[string]EllipseParams{response.GroupEllipses90, response.GroupEllipses95, response.GroupEllipses99} {
		if _, ok := ellipses["A"]; !ok {
			t.Error("Expected ellipse for group A")
		}
		if _, ok := ellipses["B"]; !ok {
			t.Error("Expected ellipse for group B")
		}
	}
}

func TestCalculateEllipsesWithInvalidData(t *testing.T) {
	app := &App{}

	// Test with empty data
	request := CalculateEllipsesRequest{
		Scores:      [][]float64{},
		GroupLabels: []string{},
		XComponent:  0,
		YComponent:  1,
	}

	response := app.CalculateEllipses(request)

	if response.Success {
		t.Error("Expected failure with empty data")
	}

	if response.Error == "" {
		t.Error("Expected error message")
	}

	// Test with mismatched lengths
	request = CalculateEllipsesRequest{
		Scores:      [][]float64{{1.0, 2.0}, {3.0, 4.0}},
		GroupLabels: []string{"A"},
		XComponent:  0,
		YComponent:  1,
	}

	response = app.CalculateEllipses(request)

	if response.Success {
		t.Error("Expected failure with mismatched lengths")
	}
}
