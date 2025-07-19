package core

import (
	"fmt"
	"math"

	"github.com/bitjungle/complab/pkg/types"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
	"gonum.org/v1/gonum/stat/distuv"
)

// metricsCalculator implements the MetricsCalculator interface
type metricsCalculator struct{}

// NewMetricsCalculator creates a new instance of MetricsCalculator
func NewMetricsCalculator() types.MetricsCalculator {
	return &metricsCalculator{}
}

// CalculateMetrics computes all diagnostic metrics for the given PCA result
func (mc *metricsCalculator) CalculateMetrics(result *types.PCAResult, data types.Matrix, config types.MetricsConfig) (*types.PCAMetrics, error) {
	if result == nil {
		return nil, fmt.Errorf("PCA result cannot be nil")
	}
	if len(data) == 0 || len(data[0]) == 0 {
		return nil, fmt.Errorf("data matrix cannot be empty")
	}

	// Determine number of components to use
	nComponents := config.NumComponents
	if nComponents == 0 || nComponents > len(result.ExplainedVar) {
		nComponents = len(result.ExplainedVar)
	}

	// Extract scores for the specified number of components
	scores := extractScoresMatrix(result.Scores, nComponents)
	
	// Calculate Mahalanobis distances
	mahalanobis, err := calculateMahalanobisDistances(scores)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate Mahalanobis distances: %w", err)
	}

	// Calculate Hotelling's T² statistics
	hotelling, err := calculateHotellingT2(scores, len(data))
	if err != nil {
		return nil, fmt.Errorf("failed to calculate Hotelling's T²: %w", err)
	}

	// Calculate residuals
	rss, qResiduals, err := calculateResiduals(result, data, nComponents)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate residuals: %w", err)
	}

	// Detect outliers
	significance := config.SignificanceLevel
	if significance == 0 {
		significance = 0.01 // Default 1% significance level
	}
	outlierMask := detectOutliersFromT2(hotelling, len(data), nComponents, significance)

	metrics := &types.PCAMetrics{
		MahalanobisDistances: mahalanobis,
		HotellingT2:         hotelling,
		RSS:                 rss,
		QResiduals:          qResiduals,
		OutlierMask:         outlierMask,
	}

	// Calculate contributions if requested
	if config.CalculateContributions {
		metrics.ContributionScores = mc.CalculateContributions(result, data)
	}

	// Calculate confidence ellipse parameters if requested
	if config.CalculateConfidenceEllipse && nComponents >= 2 {
		metrics.ConfidenceEllipse = calculateConfidenceEllipse(scores, significance)
	}

	return metrics, nil
}

// extractScoresMatrix extracts the first n components from the scores
func extractScoresMatrix(scores types.Matrix, nComponents int) *mat.Dense {
	nSamples := len(scores)
	scoresMatrix := mat.NewDense(nSamples, nComponents, nil)
	
	for i := 0; i < nSamples; i++ {
		for j := 0; j < nComponents; j++ {
			scoresMatrix.Set(i, j, scores[i][j])
		}
	}
	
	return scoresMatrix
}

// calculateMahalanobisDistances computes Mahalanobis distances for each observation
func calculateMahalanobisDistances(scores *mat.Dense) ([]float64, error) {
	nSamples, nComponents := scores.Dims()
	
	// Calculate mean vector
	meanVec := make([]float64, nComponents)
	for j := 0; j < nComponents; j++ {
		col := mat.Col(nil, j, scores)
		meanVec[j] = stat.Mean(col, nil)
	}
	
	// Center the data
	centered := mat.NewDense(nSamples, nComponents, nil)
	for i := 0; i < nSamples; i++ {
		for j := 0; j < nComponents; j++ {
			centered.Set(i, j, scores.At(i, j)-meanVec[j])
		}
	}
	
	// Special case for single component
	if nComponents == 1 {
		distances := make([]float64, nSamples)
		variance := stat.Variance(mat.Col(nil, 0, scores), nil)
		if variance == 0 {
			return nil, fmt.Errorf("variance is zero, cannot calculate Mahalanobis distance")
		}
		
		for i := 0; i < nSamples; i++ {
			diff := scores.At(i, 0) - meanVec[0]
			distances[i] = math.Sqrt(diff * diff / variance)
		}
		return distances, nil
	}
	
	// Calculate covariance matrix
	var cov mat.SymDense
	stat.CovarianceMatrix(&cov, centered, nil)
	
	// Invert covariance matrix
	var invCov mat.Dense
	err := invCov.Inverse(&cov)
	if err != nil {
		return nil, fmt.Errorf("failed to invert covariance matrix: %w", err)
	}
	
	// Calculate Mahalanobis distances
	distances := make([]float64, nSamples)
	diff := mat.NewVecDense(nComponents, nil)
	temp := mat.NewVecDense(nComponents, nil)
	
	for i := 0; i < nSamples; i++ {
		// Get the difference from mean
		for j := 0; j < nComponents; j++ {
			diff.SetVec(j, centered.At(i, j))
		}
		
		// Calculate (x-μ)ᵀ * Σ⁻¹ * (x-μ)
		temp.MulVec(&invCov, diff)
		distances[i] = math.Sqrt(mat.Dot(diff, temp))
	}
	
	return distances, nil
}

// calculateHotellingT2 computes Hotelling's T-squared statistics
func calculateHotellingT2(scores *mat.Dense, nSamples int) ([]float64, error) {
	nRows, nComponents := scores.Dims()
	
	// Calculate mean vector
	meanVec := make([]float64, nComponents)
	for j := 0; j < nComponents; j++ {
		col := mat.Col(nil, j, scores)
		meanVec[j] = stat.Mean(col, nil)
	}
	
	// Center the data
	centered := mat.NewDense(nRows, nComponents, nil)
	for i := 0; i < nRows; i++ {
		for j := 0; j < nComponents; j++ {
			centered.Set(i, j, scores.At(i, j)-meanVec[j])
		}
	}
	
	// Special case for single component
	if nComponents == 1 {
		t2Stats := make([]float64, nRows)
		variance := stat.Variance(mat.Col(nil, 0, scores), nil)
		if variance == 0 {
			return nil, fmt.Errorf("variance is zero, cannot calculate Hotelling's T²")
		}
		
		for i := 0; i < nRows; i++ {
			diff := scores.At(i, 0) - meanVec[0]
			t2Stats[i] = diff * diff / variance
		}
		return t2Stats, nil
	}
	
	// Calculate covariance matrix
	var cov mat.SymDense
	stat.CovarianceMatrix(&cov, centered, nil)
	
	// Invert covariance matrix
	var invCov mat.Dense
	err := invCov.Inverse(&cov)
	if err != nil {
		return nil, fmt.Errorf("failed to invert covariance matrix: %w", err)
	}
	
	// Calculate Hotelling's T² statistics
	t2Stats := make([]float64, nRows)
	diff := mat.NewVecDense(nComponents, nil)
	temp := mat.NewVecDense(nComponents, nil)
	
	for i := 0; i < nRows; i++ {
		// Get the difference from mean
		for j := 0; j < nComponents; j++ {
			diff.SetVec(j, centered.At(i, j))
		}
		
		// Calculate (x-μ)ᵀ * Σ⁻¹ * (x-μ)
		temp.MulVec(&invCov, diff)
		t2Stats[i] = mat.Dot(diff, temp)
	}
	
	return t2Stats, nil
}

// calculateResiduals computes RSS and Q-residuals for each observation
func calculateResiduals(result *types.PCAResult, data types.Matrix, nComponents int) ([]float64, []float64, error) {
	nSamples := len(data)
	nVariables := len(data[0])
	
	// Convert data to matrix
	dataMatrix := mat.NewDense(nSamples, nVariables, nil)
	for i := 0; i < nSamples; i++ {
		for j := 0; j < nVariables; j++ {
			dataMatrix.Set(i, j, data[i][j])
		}
	}
	
	// Extract scores and loadings for the specified components
	scores := mat.NewDense(nSamples, nComponents, nil)
	loadings := mat.NewDense(nComponents, nVariables, nil)
	
	for i := 0; i < nSamples; i++ {
		for j := 0; j < nComponents; j++ {
			scores.Set(i, j, result.Scores[i][j])
		}
	}
	
	for i := 0; i < nComponents; i++ {
		for j := 0; j < nVariables; j++ {
			loadings.Set(i, j, result.Loadings[i][j])
		}
	}
	
	// Reconstruct data: X_reconstructed = scores * loadings + mean
	var reconstructed mat.Dense
	reconstructed.Mul(scores, loadings)
	
	// Note: In the current implementation, mean centering is handled internally
	// during PCA fit, so we don't add it back here. The residuals are calculated
	// from the centered data.
	
	// Calculate residuals
	rss := make([]float64, nSamples)
	qResiduals := make([]float64, nSamples)
	
	for i := 0; i < nSamples; i++ {
		sumSquares := 0.0
		for j := 0; j < nVariables; j++ {
			diff := dataMatrix.At(i, j) - reconstructed.At(i, j)
			sumSquares += diff * diff
		}
		rss[i] = sumSquares
		qResiduals[i] = sumSquares // Q-residuals are the same as RSS in this context
	}
	
	return rss, qResiduals, nil
}

// detectOutliersFromT2 identifies outliers based on Hotelling's T² statistics
func detectOutliersFromT2(t2Stats []float64, nSamples, nComponents int, significance float64) []bool {
	// Calculate F-distribution threshold
	// T² follows an F-distribution scaled by a factor
	// T² ~ (p(n-1)/(n-p)) * F(p, n-p)
	// where p = number of components, n = number of samples
	
	df1 := float64(nComponents)
	df2 := float64(nSamples - nComponents)
	
	if df2 <= 0 {
		// Not enough degrees of freedom for F-distribution
		// Return no outliers
		return make([]bool, len(t2Stats))
	}
	
	// Create F-distribution
	fDist := distuv.F{
		D1: df1,
		D2: df2,
	}
	
	// Calculate critical value
	fCritical := fDist.Quantile(1 - significance)
	
	// Scale factor for T² distribution
	scale := (df1 * float64(nSamples-1)) / (float64(nSamples) * df2)
	t2Threshold := scale * fCritical
	
	// Detect outliers
	outliers := make([]bool, len(t2Stats))
	for i, t2 := range t2Stats {
		outliers[i] = t2 > t2Threshold
	}
	
	return outliers
}

// DetectOutliers identifies outliers based on Hotelling's T² statistic
func (mc *metricsCalculator) DetectOutliers(metrics *types.PCAMetrics, significance float64) []bool {
	if metrics == nil || len(metrics.HotellingT2) == 0 {
		return []bool{}
	}
	
	// Estimate number of samples from the data
	// This is a simplified version - in practice, you'd pass this information
	nSamples := len(metrics.HotellingT2)
	
	// Estimate number of components (this is approximate)
	// In practice, this should be passed as a parameter
	nComponents := 2 // Default assumption
	
	return detectOutliersFromT2(metrics.HotellingT2, nSamples, nComponents, significance)
}

// CalculateContributions computes variable contributions to each PC
func (mc *metricsCalculator) CalculateContributions(result *types.PCAResult, data types.Matrix) [][]float64 {
	if result == nil || len(result.Loadings) == 0 {
		return nil
	}
	
	nComponents := len(result.Loadings)
	nVariables := len(result.Loadings[0])
	
	// Calculate contributions as squared loadings normalized by component
	contributions := make([][]float64, nVariables)
	for i := 0; i < nVariables; i++ {
		contributions[i] = make([]float64, nComponents)
	}
	
	for j := 0; j < nComponents; j++ {
		sumSquares := 0.0
		// Calculate sum of squared loadings for this component
		for i := 0; i < nVariables; i++ {
			sumSquares += result.Loadings[j][i] * result.Loadings[j][i]
		}
		
		// Normalize to get contributions
		if sumSquares > 0 {
			for i := 0; i < nVariables; i++ {
				contributions[i][j] = (result.Loadings[j][i] * result.Loadings[j][i]) / sumSquares
			}
		}
	}
	
	return contributions
}

// calculateConfidenceEllipse computes parameters for confidence ellipse visualization
func calculateConfidenceEllipse(scores *mat.Dense, significance float64) types.EllipseParams {
	nSamples, _ := scores.Dims()
	
	// Calculate means for first two components
	col1 := mat.Col(nil, 0, scores)
	col2 := mat.Col(nil, 1, scores)
	
	meanX := stat.Mean(col1, nil)
	meanY := stat.Mean(col2, nil)
	
	// Calculate covariance for first two components
	data2D := mat.NewDense(nSamples, 2, nil)
	for i := 0; i < nSamples; i++ {
		data2D.Set(i, 0, scores.At(i, 0))
		data2D.Set(i, 1, scores.At(i, 1))
	}
	
	var cov mat.SymDense
	stat.CovarianceMatrix(&cov, data2D, nil)
	
	// Eigendecomposition of 2x2 covariance matrix
	var eig mat.EigenSym
	ok := eig.Factorize(&cov, true)
	if !ok {
		// Return default ellipse if factorization fails
		return types.EllipseParams{
			CenterX:         meanX,
			CenterY:         meanY,
			MajorAxis:       1.0,
			MinorAxis:       1.0,
			Angle:           0.0,
			ConfidenceLevel: 1 - significance,
		}
	}
	
	values := eig.Values(nil)
	vectors := mat.NewDense(2, 2, nil)
	eig.VectorsTo(vectors)
	
	// Chi-squared value for confidence level
	chiSquared := distuv.ChiSquared{K: 2}
	chiValue := chiSquared.Quantile(1 - significance)
	
	// Calculate ellipse parameters
	majorAxis := 2 * math.Sqrt(chiValue*math.Max(values[0], values[1]))
	minorAxis := 2 * math.Sqrt(chiValue*math.Min(values[0], values[1]))
	
	// Angle is determined by the eigenvector corresponding to the largest eigenvalue
	var angle float64
	if values[0] > values[1] {
		angle = math.Atan2(vectors.At(1, 0), vectors.At(0, 0))
	} else {
		angle = math.Atan2(vectors.At(1, 1), vectors.At(0, 1))
	}
	
	return types.EllipseParams{
		CenterX:         meanX,
		CenterY:         meanY,
		MajorAxis:       majorAxis,
		MinorAxis:       minorAxis,
		Angle:           angle,
		ConfidenceLevel: 1 - significance,
	}
}