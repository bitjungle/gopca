package core

import (
	"fmt"
	"math"

	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

// CalculateConfidenceEllipse computes the parameters for a confidence ellipse
// for a 2D set of points at the specified confidence level.
// Returns center coordinates, semi-major/minor axes, and rotation angle.
//
// Reference: Johnson & Wichern (2007) Applied Multivariate Statistical Analysis
func CalculateConfidenceEllipse(x, y []float64, confidenceLevel float64) (centerX, centerY, majorAxis, minorAxis, angle float64, err error) {
	if len(x) != len(y) {
		return 0, 0, 0, 0, 0, fmt.Errorf("x and y must have the same length")
	}

	n := len(x)
	if n < 3 {
		return 0, 0, 0, 0, 0, fmt.Errorf("need at least 3 points to calculate confidence ellipse")
	}

	// Calculate means (center of ellipse)
	centerX = stat.Mean(x, nil)
	centerY = stat.Mean(y, nil)

	// Build covariance matrix
	cov := mat.NewSymDense(2, nil)
	cov.SetSym(0, 0, stat.Variance(x, nil))
	cov.SetSym(1, 1, stat.Variance(y, nil))
	cov.SetSym(0, 1, stat.Covariance(x, y, nil))

	// Calculate eigenvalues and eigenvectors
	var eig mat.EigenSym
	ok := eig.Factorize(cov, true)
	if !ok {
		return 0, 0, 0, 0, 0, fmt.Errorf("failed to compute eigenvalues")
	}

	values := eig.Values(nil)
	vectors := mat.NewDense(2, 2, nil)
	eig.VectorsTo(vectors)

	// Ensure eigenvalues are positive
	if values[0] <= 0 || values[1] <= 0 {
		return 0, 0, 0, 0, 0, fmt.Errorf("covariance matrix is not positive definite")
	}

	// Sort eigenvalues (largest first)
	if values[0] < values[1] {
		values[0], values[1] = values[1], values[0]
		// Swap eigenvector columns
		v1 := mat.Col(nil, 0, vectors)
		v2 := mat.Col(nil, 1, vectors)
		vectors.SetCol(0, v2)
		vectors.SetCol(1, v1)
	}

	// Calculate chi-square value for the confidence level
	// For 2D data, degrees of freedom = 2
	chiSquare := chiSquareValue(confidenceLevel, 2)

	// Calculate semi-axes lengths
	majorAxis = math.Sqrt(chiSquare * values[0])
	minorAxis = math.Sqrt(chiSquare * values[1])

	// Calculate rotation angle from the first eigenvector
	angle = math.Atan2(vectors.At(1, 0), vectors.At(0, 0))

	return centerX, centerY, majorAxis, minorAxis, angle, nil
}

// chiSquareValue returns the chi-square value for a given confidence level and degrees of freedom.
// This is a simplified approximation for common confidence levels.
func chiSquareValue(confidenceLevel float64, df int) float64 {
	if df != 2 {
		// For now, we only support 2D ellipses
		return 5.991 // 95% confidence for df=2
	}

	// Chi-square values for df=2
	switch confidenceLevel {
	case 0.90:
		return 4.605
	case 0.95:
		return 5.991
	case 0.99:
		return 9.210
	default:
		// Default to 95% confidence
		return 5.991
	}
}

// CalculateGroupEllipses computes confidence ellipse parameters for each group in the data.
// scores is a 2D matrix where each row is an observation, columns are PC scores.
// groups is a slice indicating the group membership of each observation.
// pcX and pcY are the indices of the principal components to use (0-based).
func CalculateGroupEllipses(scores mat.Matrix, groups []string, pcX, pcY int, confidenceLevel float64) (map[string]EllipseParams, error) {
	rows, cols := scores.Dims()
	if pcX >= cols || pcY >= cols {
		return nil, fmt.Errorf("PC indices out of bounds")
	}

	// Group the data
	groupData := make(map[string]struct {
		x []float64
		y []float64
	})

	for i := 0; i < rows; i++ {
		group := groups[i]
		if _, exists := groupData[group]; !exists {
			groupData[group] = struct {
				x []float64
				y []float64
			}{
				x: make([]float64, 0),
				y: make([]float64, 0),
			}
		}
		data := groupData[group]
		data.x = append(data.x, scores.At(i, pcX))
		data.y = append(data.y, scores.At(i, pcY))
		groupData[group] = data
	}

	// Calculate ellipse for each group
	ellipses := make(map[string]EllipseParams)
	for group, data := range groupData {
		if len(data.x) < 3 {
			// Skip groups with too few points
			continue
		}

		centerX, centerY, majorAxis, minorAxis, angle, err := CalculateConfidenceEllipse(data.x, data.y, confidenceLevel)
		if err != nil {
			// Skip this group if ellipse calculation fails
			continue
		}

		ellipses[group] = EllipseParams{
			CenterX:         centerX,
			CenterY:         centerY,
			MajorAxis:       majorAxis,
			MinorAxis:       minorAxis,
			Angle:           angle,
			ConfidenceLevel: confidenceLevel,
		}
	}

	return ellipses, nil
}

// EllipseParams contains parameters for drawing a confidence ellipse
type EllipseParams struct {
	CenterX         float64
	CenterY         float64
	MajorAxis       float64
	MinorAxis       float64
	Angle           float64 // in radians
	ConfidenceLevel float64
}
