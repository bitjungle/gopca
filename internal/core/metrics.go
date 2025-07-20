package core

import (
	"fmt"
	"math"

	"github.com/bitjungle/complab/pkg/types"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat/distuv"
)

// PCAMetricsCalculator calculates advanced metrics for PCA results
type PCAMetricsCalculator struct {
	// PCA model parameters
	scores          *mat.Dense
	loadings        *mat.Dense
	mean            []float64
	stdDev          []float64
	nComponents     int
	nSamples        int
	nFeatures       int
	
	// Regularization for numerical stability
	regularization  float64
}

// NewPCAMetricsCalculator creates a new metrics calculator
func NewPCAMetricsCalculator(scores, loadings *mat.Dense, mean, stdDev []float64) *PCAMetricsCalculator {
	nSamples, nComponents := scores.Dims()
	nFeatures, _ := loadings.Dims()
	
	return &PCAMetricsCalculator{
		scores:         scores,
		loadings:       loadings,
		mean:           mean,
		stdDev:         stdDev,
		nComponents:    nComponents,
		nSamples:       nSamples,
		nFeatures:      nFeatures,
		regularization: 1e-8,
	}
}

// CalculateMetrics computes all metrics for each sample
func (m *PCAMetricsCalculator) CalculateMetrics(originalData types.Matrix) ([]types.SampleMetrics, error) {
	metrics := make([]types.SampleMetrics, m.nSamples)
	
	// Calculate covariance matrix of scores (regularized)
	scoresCov, err := m.calculateScoresCovariance()
	if err != nil {
		return nil, fmt.Errorf("failed to calculate scores covariance: %w", err)
	}
	
	// Calculate inverse of regularized covariance matrix
	var scoresCovInv mat.Dense
	err = scoresCovInv.Inverse(scoresCov)
	if err != nil {
		return nil, fmt.Errorf("failed to invert covariance matrix: %w", err)
	}
	
	// Calculate mean of scores (should be close to zero for centered data)
	scoreMeans := make([]float64, m.nComponents)
	for j := 0; j < m.nComponents; j++ {
		col := mat.Col(nil, j, m.scores)
		sum := 0.0
		for _, v := range col {
			sum += v
		}
		scoreMeans[j] = sum / float64(m.nSamples)
	}
	
	// Calculate metrics for each sample
	for i := 0; i < m.nSamples; i++ {
		// Get score vector for this sample
		scoreVec := mat.NewVecDense(m.nComponents, nil)
		for j := 0; j < m.nComponents; j++ {
			scoreVec.SetVec(j, m.scores.At(i, j))
		}
		
		// Calculate Hotelling's T²
		hotellingT2 := m.calculateHotellingT2(scoreVec, scoreMeans, &scoresCovInv)
		
		// Calculate Mahalanobis distance
		mahalanobis := m.calculateMahalanobisDistance(scoreVec, scoreMeans, &scoresCovInv)
		
		// Calculate Residual Sum of Squares (RSS)
		rss, err := m.calculateRSS(i, originalData)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate RSS for sample %d: %w", i, err)
		}
		
		// Determine if sample is an outlier based on Hotelling's T²
		isOutlier := m.isOutlier(hotellingT2)
		
		metrics[i] = types.SampleMetrics{
			HotellingT2: hotellingT2,
			Mahalanobis: mahalanobis,
			RSS:         rss,
			IsOutlier:   isOutlier,
		}
	}
	
	return metrics, nil
}

// calculateScoresCovariance computes the regularized covariance matrix of scores
func (m *PCAMetricsCalculator) calculateScoresCovariance() (*mat.Dense, error) {
	// Create covariance matrix
	cov := mat.NewDense(m.nComponents, m.nComponents, nil)
	
	// Calculate mean-centered scores
	centeredScores := mat.NewDense(m.nSamples, m.nComponents, nil)
	for j := 0; j < m.nComponents; j++ {
		col := mat.Col(nil, j, m.scores)
		mean := 0.0
		for _, v := range col {
			mean += v
		}
		mean /= float64(m.nSamples)
		
		for i := 0; i < m.nSamples; i++ {
			centeredScores.Set(i, j, m.scores.At(i, j)-mean)
		}
	}
	
	// Calculate covariance: (1/(n-1)) * X^T * X
	var temp mat.Dense
	temp.Mul(centeredScores.T(), centeredScores)
	cov.Scale(1.0/float64(m.nSamples-1), &temp)
	
	// Add regularization to diagonal
	for i := 0; i < m.nComponents; i++ {
		cov.Set(i, i, cov.At(i, i)+m.regularization)
	}
	
	return cov, nil
}

// calculateHotellingT2 computes Hotelling's T² statistic
func (m *PCAMetricsCalculator) calculateHotellingT2(scoreVec *mat.VecDense, means []float64, covInv *mat.Dense) float64 {
	// Calculate difference from mean
	diff := mat.NewVecDense(m.nComponents, nil)
	for i := 0; i < m.nComponents; i++ {
		diff.SetVec(i, scoreVec.AtVec(i)-means[i])
	}
	
	// T² = (x - μ)^T * Σ^(-1) * (x - μ)
	var temp mat.VecDense
	temp.MulVec(covInv, diff)
	t2 := mat.Dot(diff, &temp)
	
	return t2
}

// calculateMahalanobisDistance computes the Mahalanobis distance
func (m *PCAMetricsCalculator) calculateMahalanobisDistance(scoreVec *mat.VecDense, means []float64, covInv *mat.Dense) float64 {
	// Calculate difference from mean
	diff := mat.NewVecDense(m.nComponents, nil)
	for i := 0; i < m.nComponents; i++ {
		diff.SetVec(i, scoreVec.AtVec(i)-means[i])
	}
	
	// D² = (x - μ)^T * Σ^(-1) * (x - μ)
	var temp mat.VecDense
	temp.MulVec(covInv, diff)
	d2 := mat.Dot(diff, &temp)
	
	// Return the distance (not squared)
	return math.Sqrt(d2)
}

// calculateRSS computes the Residual Sum of Squares
func (m *PCAMetricsCalculator) calculateRSS(sampleIdx int, originalData types.Matrix) (float64, error) {
	// Reconstruct the sample from scores and loadings
	// X_reconstructed = scores * loadings^T
	
	// Get the score vector for this sample
	scoreVec := mat.NewVecDense(m.nComponents, nil)
	for j := 0; j < m.nComponents; j++ {
		scoreVec.SetVec(j, m.scores.At(sampleIdx, j))
	}
	
	// Calculate reconstructed values
	reconstructed := make([]float64, m.nFeatures)
	for j := 0; j < m.nFeatures; j++ {
		val := 0.0
		for k := 0; k < m.nComponents; k++ {
			val += scoreVec.AtVec(k) * m.loadings.At(j, k)
		}
		
		// Add back mean if data was centered
		if len(m.mean) > 0 {
			val += m.mean[j]
		}
		
		// Scale back if data was standardized
		if len(m.stdDev) > 0 && m.stdDev[j] > 0 {
			val *= m.stdDev[j]
		}
		
		reconstructed[j] = val
	}
	
	// Calculate sum of squared residuals
	rss := 0.0
	for j := 0; j < m.nFeatures; j++ {
		residual := originalData[sampleIdx][j] - reconstructed[j]
		rss += residual * residual
	}
	
	return rss, nil
}

// isOutlier determines if a sample is an outlier based on Hotelling's T²
func (m *PCAMetricsCalculator) isOutlier(hotellingT2 float64) bool {
	// Calculate critical value using F-distribution
	// T²_critical = p(n-1)/(n-p) * F_{p,n-p}(1-α)
	// where p = number of components, n = number of samples, α = significance level
	
	alpha := 0.001 // 99.9% confidence level - less sensitive to outliers
	p := float64(m.nComponents)
	n := float64(m.nSamples)
	
	if n <= p {
		// Cannot calculate threshold with insufficient samples
		return false
	}
	
	// Create F-distribution
	fDist := distuv.F{
		D1: p,
		D2: n - p,
	}
	
	// Calculate critical value
	fCritical := fDist.Quantile(1 - alpha)
	t2Critical := p * (n - 1) / (n - p) * fCritical
	
	return hotellingT2 > t2Critical
}

// CalculateMetricsFromPCAResult is a convenience function that calculates metrics directly from PCAResult
func CalculateMetricsFromPCAResult(result *types.PCAResult, originalData types.Matrix) ([]types.SampleMetrics, error) {
	// Convert result matrices to gonum matrices
	scores := matrixToDense(result.Scores)
	loadings := matrixToDense(result.Loadings)
	
	// Create metrics calculator
	calculator := NewPCAMetricsCalculator(scores, loadings, result.Means, result.StdDevs)
	
	// Calculate metrics
	return calculator.CalculateMetrics(originalData)
}