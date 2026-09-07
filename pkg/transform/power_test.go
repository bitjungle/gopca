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

package transform

import (
	"math"
	"strconv"
	"strings"
	"testing"
)

func powerFixture(values ...string) Input {
	data := make([][]string, len(values))
	for i, v := range values {
		data[i] = []string{v}
	}
	return Input{
		Data: data, Headers: []string{"X"},
		ColumnTypes: map[string]string{"X": "numeric"},
		Rows:        len(values), Columns: 1,
	}
}

// TestBoxCoxRefusesNonPositive covers the boundary between the two families.
//
// Box-Cox is undefined at and below zero. The column is refused whole rather
// than transformed in part, for the reason #861 established: a column holding
// some transformed and some raw values carries two scales in one variable and
// nothing downstream can detect it. The message names Yeo-Johnson, because
// that is the answer rather than a consolation.
func TestBoxCoxRefusesNonPositive(t *testing.T) {
	for _, values := range [][]string{
		{"1", "0", "3"},
		{"1", "-2", "3"},
	} {
		res, err := Apply(powerFixture(values...), Options{Type: BoxCox, Columns: []string{"X"}})
		if err != nil {
			t.Fatalf("Apply: %v", err)
		}
		for i, want := range values {
			if res.Data[i][0] != want {
				t.Errorf("row %d = %q, want %q untouched", i, res.Data[i][0], want)
			}
		}
		if len(res.TransformedColumns) != 0 {
			t.Errorf("a refused column must not be reported as transformed: %v",
				res.TransformedColumns)
		}
		joined := strings.Join(res.Messages, " ")
		for _, want := range []string{"left unchanged", "Yeo-Johnson"} {
			if !strings.Contains(joined, want) {
				t.Errorf("message should mention %q, got %v", want, res.Messages)
			}
		}
	}
}

// TestYeoJohnsonHandlesWhatBoxCoxCannot is the reason it is offered.
func TestYeoJohnsonHandlesWhatBoxCoxCannot(t *testing.T) {
	for name, values := range map[string][]string{
		"zeros":     {"0", "1", "2", "0", "5"},
		"negatives": {"-3", "-1", "0", "2", "4"},
	} {
		t.Run(name, func(t *testing.T) {
			res, err := Apply(powerFixture(values...), Options{Type: YeoJohnson, Columns: []string{"X"}})
			if err != nil {
				t.Fatalf("Apply: %v", err)
			}
			if len(res.TransformedColumns) != 1 {
				t.Fatalf("expected the column to transform, got %v", res.TransformedColumns)
			}
			changed := false
			for i, original := range values {
				if res.Data[i][0] != original {
					changed = true
				}
			}
			if !changed {
				t.Error("the column was reported as transformed but nothing changed")
			}
		})
	}
}

// TestPowerTransformReportsLambda covers the requirement that the parameter be
// visible.
//
// A transform whose parameter the user cannot see is one they cannot report in
// a paper or reproduce anywhere else.
func TestPowerTransformReportsLambda(t *testing.T) {
	res, err := Apply(powerFixture("1", "2", "3", "8", "20"),
		Options{Type: BoxCox, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	joined := strings.Join(res.Messages, " ")
	for _, want := range []string{"λ =", "maximum likelihood", "Box-Cox"} {
		if !strings.Contains(joined, want) {
			t.Errorf("message should mention %q, got %v", want, res.Messages)
		}
	}

	// A supplied lambda is reported as supplied, not as fitted.
	fixed := 0.5
	res, err = Apply(powerFixture("1", "2", "3", "8", "20"),
		Options{Type: BoxCox, Columns: []string{"X"}, Lambda: &fixed})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	joined = strings.Join(res.Messages, " ")
	if !strings.Contains(joined, "as supplied") {
		t.Errorf("a supplied λ should say so, got %v", res.Messages)
	}
	if strings.Contains(joined, "maximum likelihood") {
		t.Errorf("a supplied λ must not claim to be fitted: %v", res.Messages)
	}
}

// TestPowerTransformLambdaZeroIsMeaningful is why Lambda is a pointer.
//
// Zero is the logarithm, a perfectly ordinary member of both families, so
// there is no number free to mean "not set". A plain float64 would make an
// explicit λ = 0 indistinguishable from an absent one and silently estimate
// instead.
func TestPowerTransformLambdaZeroIsMeaningful(t *testing.T) {
	zero := 0.0
	res, err := Apply(powerFixture("1", "2", "4", "8"),
		Options{Type: BoxCox, Columns: []string{"X"}, Lambda: &zero})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if !strings.Contains(strings.Join(res.Messages, " "), "as supplied") {
		t.Errorf("λ = 0 must be honoured as a value, not read as unset: %v", res.Messages)
	}
	// At λ = 0 Box-Cox is the natural logarithm.
	for i, input := range []float64{1, 2, 4, 8} {
		got, err := strconv.ParseFloat(res.Data[i][0], 64)
		if err != nil {
			t.Fatalf("%v", err)
		}
		if diff := math.Abs(got - math.Log(input)); diff > 1e-5 {
			t.Errorf("row %d: got %g, want ln(%g) = %g", i, got, input, math.Log(input))
		}
	}
}

// TestPowerTransformSkipsBlanks matches the other numeric transforms: a cell
// that was never a number cannot be put into the wrong units by being left.
func TestPowerTransformSkipsBlanks(t *testing.T) {
	res, err := Apply(powerFixture("1", "", "4", "N/A", "9"),
		Options{Type: YeoJohnson, Columns: []string{"X"}})
	if err != nil {
		t.Fatalf("Apply: %v", err)
	}
	if res.Data[1][0] != "" {
		t.Errorf("a blank became %q", res.Data[1][0])
	}
	if res.Data[3][0] != "N/A" {
		t.Errorf("an unparseable cell became %q", res.Data[3][0])
	}
}

// TestEstimateLambdaOnSymmetricData checks the estimator does not over-correct.
//
// Data that is already roughly symmetric needs almost no transform, and the
// maximum likelihood lambda should come out near 1 -- the identity, up to a
// shift and scale. An estimator that always found a strong transform would
// still pass a test that only checked it produced *a* number.
func TestEstimateLambdaOnSymmetricData(t *testing.T) {
	// A symmetric spread about 50.
	values := []float64{}
	for i := -20; i <= 20; i++ {
		values = append(values, 50+float64(i)*0.5)
	}
	lambda := estimateLambda(values, BoxCox)
	if math.Abs(lambda-1) > 0.6 {
		t.Errorf("λ = %.3f for symmetric data, expected near 1 (little correction)", lambda)
	}

	// A strongly right-skewed column should call for a much stronger one.
	skewed := []float64{}
	for i := 1; i <= 40; i++ {
		skewed = append(skewed, math.Exp(float64(i)/6))
	}
	skewedLambda := estimateLambda(skewed, BoxCox)
	if skewedLambda >= lambda {
		t.Errorf("skewed data fitted λ = %.3f, not below the symmetric %.3f",
			skewedLambda, lambda)
	}
}
