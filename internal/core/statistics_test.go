package core

import (
	"math"
	"testing"

	"gonum.org/v1/gonum/mat"
)

func TestCalculateConfidenceEllipse(t *testing.T) {
	tests := []struct {
		name            string
		x               []float64
		y               []float64
		confidenceLevel float64
		wantErr         bool
	}{
		{
			name:            "valid circular data",
			x:               []float64{0, 1, 0, -1, 0.5, -0.5, 0.5, -0.5},
			y:               []float64{1, 0, -1, 0, 0.5, 0.5, -0.5, -0.5},
			confidenceLevel: 0.95,
			wantErr:         false,
		},
		{
			name:            "valid elliptical data",
			x:               []float64{0, 2, 0, -2, 1, -1, 1, -1},
			y:               []float64{0.5, 0, -0.5, 0, 0.25, 0.25, -0.25, -0.25},
			confidenceLevel: 0.95,
			wantErr:         false,
		},
		{
			name:            "too few points",
			x:               []float64{0, 1},
			y:               []float64{0, 1},
			confidenceLevel: 0.95,
			wantErr:         true,
		},
		{
			name:            "mismatched lengths",
			x:               []float64{0, 1, 2},
			y:               []float64{0, 1},
			confidenceLevel: 0.95,
			wantErr:         true,
		},
		{
			name:            "90% confidence",
			x:               []float64{0, 1, 0, -1, 0.5, -0.5, 0.5, -0.5},
			y:               []float64{1, 0, -1, 0, 0.5, 0.5, -0.5, -0.5},
			confidenceLevel: 0.90,
			wantErr:         false,
		},
		{
			name:            "99% confidence",
			x:               []float64{0, 1, 0, -1, 0.5, -0.5, 0.5, -0.5},
			y:               []float64{1, 0, -1, 0, 0.5, 0.5, -0.5, -0.5},
			confidenceLevel: 0.99,
			wantErr:         false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			centerX, centerY, majorAxis, minorAxis, angle, err := CalculateConfidenceEllipse(tt.x, tt.y, tt.confidenceLevel)

			if (err != nil) != tt.wantErr {
				t.Errorf("CalculateConfidenceEllipse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr {
				// Check that center is approximately at the mean
				meanX := mean(tt.x)
				meanY := mean(tt.y)
				if math.Abs(centerX-meanX) > 1e-10 || math.Abs(centerY-meanY) > 1e-10 {
					t.Errorf("Center mismatch: got (%f, %f), want (%f, %f)", centerX, centerY, meanX, meanY)
				}

				// Check that axes are positive
				if majorAxis <= 0 || minorAxis <= 0 {
					t.Errorf("Invalid axes: majorAxis=%f, minorAxis=%f", majorAxis, minorAxis)
				}

				// Check that major axis is larger than minor axis
				if majorAxis < minorAxis {
					t.Errorf("Major axis should be larger than minor axis: majorAxis=%f, minorAxis=%f", majorAxis, minorAxis)
				}

				// Check that angle is in valid range
				if angle < -math.Pi || angle > math.Pi {
					t.Errorf("Invalid angle: %f", angle)
				}
			}
		})
	}
}

func TestCalculateGroupEllipses(t *testing.T) {
	// Create test data with two groups
	scores := mat.NewDense(10, 2, []float64{
		// Group A - centered around (1, 1)
		1.2, 1.1,
		0.9, 1.3,
		1.1, 0.8,
		0.8, 1.2,
		1.0, 1.0,
		// Group B - centered around (-1, -1)
		-1.1, -0.9,
		-0.8, -1.2,
		-1.2, -1.1,
		-0.9, -0.8,
		-1.0, -1.0,
	})

	groups := []string{"A", "A", "A", "A", "A", "B", "B", "B", "B", "B"}

	ellipses, err := CalculateGroupEllipses(scores, groups, 0, 1, 0.95)
	if err != nil {
		t.Fatalf("CalculateGroupEllipses() error = %v", err)
	}

	// Check that we got ellipses for both groups
	if len(ellipses) != 2 {
		t.Errorf("Expected 2 ellipses, got %d", len(ellipses))
	}

	// Check group A ellipse
	if ellipseA, ok := ellipses["A"]; ok {
		if math.Abs(ellipseA.CenterX-1.0) > 0.2 || math.Abs(ellipseA.CenterY-1.0) > 0.2 {
			t.Errorf("Group A center mismatch: got (%f, %f), expected near (1, 1)", ellipseA.CenterX, ellipseA.CenterY)
		}
		if ellipseA.MajorAxis <= 0 || ellipseA.MinorAxis <= 0 {
			t.Errorf("Group A invalid axes: majorAxis=%f, minorAxis=%f", ellipseA.MajorAxis, ellipseA.MinorAxis)
		}
	} else {
		t.Error("Missing ellipse for group A")
	}

	// Check group B ellipse
	if ellipseB, ok := ellipses["B"]; ok {
		if math.Abs(ellipseB.CenterX+1.0) > 0.2 || math.Abs(ellipseB.CenterY+1.0) > 0.2 {
			t.Errorf("Group B center mismatch: got (%f, %f), expected near (-1, -1)", ellipseB.CenterX, ellipseB.CenterY)
		}
		if ellipseB.MajorAxis <= 0 || ellipseB.MinorAxis <= 0 {
			t.Errorf("Group B invalid axes: majorAxis=%f, minorAxis=%f", ellipseB.MajorAxis, ellipseB.MinorAxis)
		}
	} else {
		t.Error("Missing ellipse for group B")
	}
}

func TestCalculateGroupEllipsesEdgeCases(t *testing.T) {
	tests := []struct {
		name   string
		scores mat.Matrix
		groups []string
		pcX    int
		pcY    int
	}{
		{
			name: "group with too few points",
			scores: mat.NewDense(5, 2, []float64{
				1.0, 1.1,
				1.2, 0.9,
				3.0, 3.0, // Group B has only 1 point
				0.8, 1.3,
				1.1, 1.0,
			}),
			groups: []string{"A", "A", "B", "A", "A"},
			pcX:    0,
			pcY:    1,
		},
		{
			name: "PC indices out of bounds",
			scores: mat.NewDense(3, 2, []float64{
				1, 1,
				2, 2,
				3, 3,
			}),
			groups: []string{"A", "A", "A"},
			pcX:    2, // Out of bounds
			pcY:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ellipses, err := CalculateGroupEllipses(tt.scores, tt.groups, tt.pcX, tt.pcY, 0.95)

			if tt.name == "PC indices out of bounds" {
				if err == nil {
					t.Error("Expected error for out of bounds PC indices")
				}
			} else if tt.name == "group with too few points" {
				if err != nil {
					t.Errorf("Unexpected error: %v", err)
				}
				// Should only have ellipse for group A
				if len(ellipses) != 1 {
					t.Errorf("Expected 1 ellipse, got %d", len(ellipses))
				}
				if _, ok := ellipses["A"]; !ok {
					t.Error("Missing ellipse for group A")
				}
				if _, ok := ellipses["B"]; ok {
					t.Error("Should not have ellipse for group B (too few points)")
				}
			}
		})
	}
}

// Helper function for testing
func mean(x []float64) float64 {
	sum := 0.0
	for _, v := range x {
		sum += v
	}
	return sum / float64(len(x))
}
