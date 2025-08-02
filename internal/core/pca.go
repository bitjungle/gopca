package core

import (
	"fmt"
	"math"

	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/mat"
)

// PCAImpl implements the PCAEngine interface
type PCAImpl struct {
	// Fitted model parameters
	preprocessor *Preprocessor
	loadings     *mat.Dense
	nComponents  int
	fitted       bool

	// Configuration
	config types.PCAConfig
}

// NewPCAEngine creates a new PCA engine instance
func NewPCAEngine() types.PCAEngine {
	return &PCAImpl{}
}

// NewPCAEngineForMethod creates a PCA engine for the specified method
func NewPCAEngineForMethod(method string) types.PCAEngine {
	switch method {
	case "kernel":
		return NewKernelPCAEngine()
	default:
		return NewPCAEngine()
	}
}

// Fit trains the PCA model on the provided data
func (p *PCAImpl) Fit(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	if err := p.validateInput(data, config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	p.config = config

	// Convert to gonum matrix
	X := utils.MatrixToDense(data)

	// Check if we're using NIPALS with native missing value handling
	usingNativeMissing := config.Method == "nipals" && config.MissingStrategy == types.MissingNative

	// Preprocessing using the Preprocessor class (skip if using native missing value handling)
	// Note: For NIPALS with missing values, mean centering is handled within the algorithm
	if !usingNativeMissing && (config.MeanCenter || config.StandardScale || config.RobustScale || config.ScaleOnly || config.SNV || config.VectorNorm) {
		// Create preprocessor with the appropriate settings
		p.preprocessor = NewPreprocessorWithScaleOnly(config.MeanCenter, config.StandardScale, config.RobustScale, config.ScaleOnly, config.SNV, config.VectorNorm)

		// Convert to types.Matrix for preprocessor
		typeMatrix := utils.DenseToMatrix(X)

		// Fit and transform
		processedData, err := p.preprocessor.FitTransform(typeMatrix)
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}

		// Convert back to mat.Dense
		X = utils.MatrixToDense(processedData)
	} else if usingNativeMissing && (config.StandardScale || config.RobustScale || config.ScaleOnly || config.SNV || config.VectorNorm) {
		// Log warning: preprocessing (except mean centering) is not supported with native missing value handling
		// Mean centering is handled internally by the NIPALS algorithm for missing data
		fmt.Printf("Warning: Preprocessing options (except mean centering) are not supported with NIPALS native missing value handling. These options were ignored.\n")
	}

	// Select PCA method
	var scores, loadings *mat.Dense
	var allEigenvalues []float64
	var err error

	// Check if we should use NIPALS with native missing value handling
	hasMissing := false
	if config.Method == "nipals" && config.MissingStrategy == types.MissingNative {
		// Check for NaN values in the data
		r, c := X.Dims()
		for i := 0; i < r && !hasMissing; i++ {
			for j := 0; j < c; j++ {
				if math.IsNaN(X.At(i, j)) {
					hasMissing = true
					break
				}
			}
		}
	}

	switch config.Method {
	case "svd", "":
		scores, loadings, allEigenvalues, err = p.svdAlgorithm(X, config.Components)
	case "nipals":
		if hasMissing && config.MissingStrategy == types.MissingNative {
			scores, loadings, allEigenvalues, err = p.nipalsAlgorithmWithMissing(X, config.Components)
		} else {
			scores, loadings, allEigenvalues, err = p.nipalsAlgorithm(X, config.Components)
		}
	default:
		return nil, fmt.Errorf("unknown PCA method: %s", config.Method)
	}

	if err != nil {
		return nil, fmt.Errorf("PCA computation failed: %w", err)
	}

	// Store loadings for transform
	p.loadings = loadings
	_, actualComponents := scores.Dims()
	p.nComponents = actualComponents
	p.fitted = true

	// Get eigenvalues for variance calculations
	var eigenvalues []float64
	if allEigenvalues != nil {
		// Use eigenvalues from the algorithm (SVD or NIPALS)
		eigenvalues = allEigenvalues[:actualComponents]
	} else {
		// Fallback: calculate from scores (shouldn't happen with current algorithms)
		eigenvalues = p.calculateEigenvaluesFromScores(scores)
	}

	// Generate component labels
	componentLabels := make([]string, actualComponents)
	for i := 0; i < actualComponents; i++ {
		componentLabels[i] = fmt.Sprintf("PC%d", i+1)
	}

	// Calculate explained variance ratio
	explainedVarRatio := make([]float64, len(eigenvalues))
	cumulativeVar := make([]float64, len(eigenvalues))

	if p.config.Method == "nipals" && p.config.MissingStrategy == types.MissingNative {
		// For NIPALS with missing values, we cannot calculate true percentages
		// because total variance is undefined with missing data
		// Instead, show relative proportions of extracted components
		totalExtractedVar := 0.0
		for _, v := range eigenvalues {
			totalExtractedVar += v
		}
		cumSum := 0.0
		for i, v := range eigenvalues {
			if totalExtractedVar > 0 {
				explainedVarRatio[i] = v / totalExtractedVar * 100
			}
			cumSum += explainedVarRatio[i]
			cumulativeVar[i] = cumSum
		}
	} else {
		// Standard calculation using all eigenvalues for total variance
		totalVar := 0.0
		if allEigenvalues != nil {
			// Sum ALL eigenvalues for total variance
			for _, v := range allEigenvalues {
				totalVar += v
			}
		} else {
			// Fallback to selected eigenvalues
			for _, v := range eigenvalues {
				totalVar += v
			}
		}

		cumSum := 0.0
		for i, v := range eigenvalues {
			if totalVar > 0 {
				explainedVarRatio[i] = v / totalVar * 100
			}
			cumSum += explainedVarRatio[i]
			cumulativeVar[i] = cumSum
		}
	}

	// Get preprocessing stats if applicable
	var means, stddevs []float64
	if p.preprocessor != nil {
		means = p.preprocessor.GetMeans()
		stddevs = p.preprocessor.GetStdDevs()
	}

	return &types.PCAResult{
		Scores:               utils.DenseToMatrix(scores),
		Loadings:             utils.DenseToMatrix(loadings),
		ExplainedVar:         eigenvalues,
		ExplainedVarRatio:    explainedVarRatio,
		CumulativeVar:        cumulativeVar,
		ComponentLabels:      componentLabels,
		ComponentsComputed:   actualComponents,
		Method:               config.Method,
		PreprocessingApplied: config.MeanCenter || config.StandardScale || config.RobustScale,
		Means:                means,
		StdDevs:              stddevs,
		AllEigenvalues:       allEigenvalues,
	}, nil
}

// Transform applies the fitted PCA model to new data
func (p *PCAImpl) Transform(data types.Matrix) (types.Matrix, error) {
	if !p.fitted {
		return nil, fmt.Errorf("model not fitted: call Fit first")
	}

	// Convert to gonum matrix
	X := utils.MatrixToDense(data)

	// Apply same preprocessing as during fit
	if p.preprocessor != nil {
		// Convert to types.Matrix for preprocessor
		typeMatrix := utils.DenseToMatrix(X)

		// Transform using preprocessor
		processedData, err := p.preprocessor.Transform(typeMatrix)
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}

		// Convert back to mat.Dense
		X = utils.MatrixToDense(processedData)
	}

	// Project onto loadings
	n, _ := X.Dims()
	scores := mat.NewDense(n, p.nComponents, nil)
	scores.Mul(X, p.loadings)

	return utils.DenseToMatrix(scores), nil
}

// FitTransform fits the model and transforms the data in one step
func (p *PCAImpl) FitTransform(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	return p.Fit(data, config)
}

// nipalsAlgorithm implements the NIPALS (Nonlinear Iterative Partial Least Squares) algorithm for PCA
// Reference: Wold, H. (1966). Estimation of principal components and related models by iterative least squares.
// In P.R. Krishnaiah (Ed.), Multivariate Analysis (pp. 391-420). Academic Press.
func (p *PCAImpl) nipalsAlgorithm(X *mat.Dense, nComponents int) (*mat.Dense, *mat.Dense, []float64, error) {
	n, m := X.Dims()

	// Initialize matrices
	T := mat.NewDense(n, nComponents, nil) // Scores
	P := mat.NewDense(m, nComponents, nil) // Loadings

	// Working copy of X for deflation
	Xwork := mat.NewDense(n, m, nil)
	Xwork.Copy(X)

	// Tolerance for convergence
	const tolerance = 1e-8
	const maxIter = 1000

	for k := 0; k < nComponents; k++ {
		// Initialize score vector t with column having maximum variance
		t := mat.NewVecDense(n, nil)
		maxVar := 0.0
		maxVarCol := 0

		// Find column with maximum variance
		for j := 0; j < m; j++ {
			col := mat.Col(nil, j, Xwork)
			var sum, sumSq float64
			for _, v := range col {
				sum += v
				sumSq += v * v
			}
			variance := sumSq/float64(n) - (sum/float64(n))*(sum/float64(n))
			if variance > maxVar {
				maxVar = variance
				maxVarCol = j
			}
		}

		// Check if remaining variance is too small
		if maxVar < tolerance {
			// No more meaningful components, reduce number of components
			nComponents = k //nolint:ineffassign // Used in the return statement
			T = T.Slice(0, n, 0, k).(*mat.Dense)
			P = P.Slice(0, m, 0, k).(*mat.Dense)
			break
		}

		// Initialize t with the column having maximum variance
		col := mat.Col(nil, maxVarCol, Xwork)
		for i := 0; i < n; i++ {
			t.SetVec(i, col[i])
		}

		// Power iteration
		converged := false
		var tOld *mat.VecDense
		var p *mat.VecDense

		for iter := 0; iter < maxIter; iter++ {
			// Save old t for convergence check
			tOld = mat.NewVecDense(n, nil)
			tOld.CopyVec(t)

			// p = X^T * t / (t^T * t)
			p = mat.NewVecDense(m, nil)
			p.MulVec(Xwork.T(), t)
			tNorm := mat.Dot(t, t)
			if tNorm < tolerance {
				return nil, nil, nil, fmt.Errorf("score vector has zero variance at component %d", k+1)
			}
			p.ScaleVec(1.0/tNorm, p)

			// Normalize p
			pNorm := math.Sqrt(mat.Dot(p, p))
			if pNorm < tolerance {
				return nil, nil, nil, fmt.Errorf("loading vector has zero variance at component %d", k+1)
			}
			p.ScaleVec(1.0/pNorm, p)

			// t = X * p / (p^T * p)
			t.MulVec(Xwork, p)
			pNormSq := mat.Dot(p, p)
			t.ScaleVec(1.0/pNormSq, t)

			// Check convergence
			diff := mat.NewVecDense(n, nil)
			diff.SubVec(t, tOld)
			if mat.Norm(diff, 2) < tolerance {
				converged = true
				break
			}
		}

		if !converged {
			return nil, nil, nil, fmt.Errorf("NIPALS did not converge for component %d", k+1)
		}

		// Store component
		tData := make([]float64, n)
		pData := make([]float64, m)
		for i := 0; i < n; i++ {
			tData[i] = t.AtVec(i)
		}
		for i := 0; i < m; i++ {
			pData[i] = p.AtVec(i)
		}
		T.SetCol(k, tData)
		P.SetCol(k, pData)

		// Deflate X: X = X - t * p^T
		tMat := mat.NewDense(n, 1, tData)
		pMat := mat.NewDense(1, m, pData)
		deflation := mat.NewDense(n, m, nil)
		deflation.Mul(tMat, pMat)
		Xwork.Sub(Xwork, deflation)
	}

	// Calculate eigenvalues from scores for retained components
	// For NIPALS, eigenvalue = variance of the score vector
	allEigenvalues := make([]float64, nComponents)
	for i := 0; i < nComponents; i++ {
		scoreCol := mat.Col(nil, i, T)
		var eigenvalue float64
		for _, v := range scoreCol {
			eigenvalue += v * v
		}
		allEigenvalues[i] = eigenvalue / float64(n-1)
	}

	// Calculate residual variance to estimate eigenvalues for non-retained components
	// The residual matrix Xwork contains what's left after extracting nComponents
	// We estimate remaining eigenvalues based on the residual variance
	if m > nComponents {
		// Calculate total residual variance
		var residualVar float64
		for i := 0; i < n; i++ {
			for j := 0; j < m; j++ {
				val := Xwork.At(i, j)
				residualVar += val * val
			}
		}
		residualVar /= float64(n - 1)

		// Estimate eigenvalues for non-retained components
		// Distribute residual variance equally among remaining components
		remainingComponents := m - nComponents
		if remainingComponents > 0 {
			// Extend eigenvalues array to include all possible components
			extendedEigenvalues := make([]float64, m)
			copy(extendedEigenvalues, allEigenvalues)

			// Distribute residual variance among remaining components
			avgResidualEigenvalue := residualVar / float64(remainingComponents)
			for i := nComponents; i < m; i++ {
				extendedEigenvalues[i] = avgResidualEigenvalue
			}
			allEigenvalues = extendedEigenvalues
		}
	}

	return T, P, allEigenvalues, nil
}

// nipalsAlgorithmWithMissing implements NIPALS with native missing value handling
func (p *PCAImpl) nipalsAlgorithmWithMissing(X *mat.Dense, nComponents int) (*mat.Dense, *mat.Dense, []float64, error) {
	n, m := X.Dims()

	// Initialize matrices
	T := mat.NewDense(n, nComponents, nil) // Scores
	P := mat.NewDense(m, nComponents, nil) // Loadings

	// Working copy of X for deflation
	Xwork := mat.NewDense(n, m, nil)
	Xwork.Copy(X)

	// Calculate column means using only non-missing values if mean centering is requested
	columnMeans := make([]float64, m)
	if p.config.MeanCenter {
		for j := 0; j < m; j++ {
			sum := 0.0
			count := 0
			for i := 0; i < n; i++ {
				val := Xwork.At(i, j)
				if !math.IsNaN(val) {
					sum += val
					count++
				}
			}
			if count > 0 {
				columnMeans[j] = sum / float64(count)
			}
		}

		// Center the data by subtracting column means from non-missing values
		for i := 0; i < n; i++ {
			for j := 0; j < m; j++ {
				val := Xwork.At(i, j)
				if !math.IsNaN(val) {
					Xwork.Set(i, j, val-columnMeans[j])
				}
			}
		}
	}

	// Tolerance for convergence
	const tolerance = 1e-8
	const maxIter = 1000

	for k := 0; k < nComponents; k++ {
		// Initialize score vector t with column having maximum non-missing variance
		t := mat.NewVecDense(n, nil)
		maxVar := 0.0
		maxVarCol := 0

		// Find column with maximum variance (considering only non-missing values)
		for j := 0; j < m; j++ {
			var sum, sumSq float64
			count := 0
			for i := 0; i < n; i++ {
				v := Xwork.At(i, j)
				if !math.IsNaN(v) {
					sum += v
					sumSq += v * v
					count++
				}
			}
			if count > 0 {
				mean := sum / float64(count)
				variance := sumSq/float64(count) - mean*mean
				if variance > maxVar {
					maxVar = variance
					maxVarCol = j
				}
			}
		}

		// Check if remaining variance is too small
		if maxVar < tolerance {
			// No more meaningful components, reduce number of components
			T = T.Slice(0, n, 0, k).(*mat.Dense)
			P = P.Slice(0, m, 0, k).(*mat.Dense)
			break
		}

		// Initialize t with the column having maximum variance (only non-missing values)
		for i := 0; i < n; i++ {
			v := Xwork.At(i, maxVarCol)
			if !math.IsNaN(v) {
				t.SetVec(i, v)
			} else {
				// Initialize missing positions with column mean
				var colSum float64
				colCount := 0
				for ii := 0; ii < n; ii++ {
					vv := Xwork.At(ii, maxVarCol)
					if !math.IsNaN(vv) {
						colSum += vv
						colCount++
					}
				}
				if colCount > 0 {
					t.SetVec(i, colSum/float64(colCount))
				} else {
					t.SetVec(i, 0)
				}
			}
		}

		// Power iteration with missing value handling
		converged := false
		var tOld *mat.VecDense
		var p *mat.VecDense

		for iter := 0; iter < maxIter; iter++ {
			// Save old t for convergence check
			tOld = mat.NewVecDense(n, nil)
			tOld.CopyVec(t)

			// p = X^T * t / (t^T * t), handling missing values
			p = mat.NewVecDense(m, nil)
			for j := 0; j < m; j++ {
				numerator := 0.0
				denominator := 0.0
				count := 0
				for i := 0; i < n; i++ {
					xVal := Xwork.At(i, j)
					tVal := t.AtVec(i)
					if !math.IsNaN(xVal) && !math.IsNaN(tVal) {
						numerator += xVal * tVal
						denominator += tVal * tVal
						count++
					}
				}
				if count > 0 && denominator > tolerance {
					p.SetVec(j, numerator/denominator)
				} else {
					p.SetVec(j, 0)
				}
			}

			// Normalize p
			pNorm := 0.0
			for j := 0; j < m; j++ {
				pVal := p.AtVec(j)
				if !math.IsNaN(pVal) {
					pNorm += pVal * pVal
				}
			}
			pNorm = math.Sqrt(pNorm)
			if pNorm < tolerance {
				return nil, nil, nil, fmt.Errorf("loading vector has zero variance at component %d", k+1)
			}
			p.ScaleVec(1.0/pNorm, p)

			// t = X * p / (p^T * p), handling missing values
			for i := 0; i < n; i++ {
				numerator := 0.0
				denominator := 0.0
				count := 0
				for j := 0; j < m; j++ {
					xVal := Xwork.At(i, j)
					pVal := p.AtVec(j)
					if !math.IsNaN(xVal) && !math.IsNaN(pVal) {
						numerator += xVal * pVal
						denominator += pVal * pVal
						count++
					}
				}
				if count > 0 && denominator > tolerance {
					t.SetVec(i, numerator/denominator)
				} else {
					// If no valid data for this sample, keep previous value
					t.SetVec(i, tOld.AtVec(i))
				}
			}

			// Check convergence
			diff := mat.NewVecDense(n, nil)
			diff.SubVec(t, tOld)
			if mat.Norm(diff, 2) < tolerance {
				converged = true
				break
			}
		}

		if !converged {
			return nil, nil, nil, fmt.Errorf("NIPALS did not converge for component %d", k+1)
		}

		// Store component
		tData := make([]float64, n)
		pData := make([]float64, m)
		for i := 0; i < n; i++ {
			tData[i] = t.AtVec(i)
		}
		for i := 0; i < m; i++ {
			pData[i] = p.AtVec(i)
		}
		T.SetCol(k, tData)
		P.SetCol(k, pData)

		// Deflate X: X = X - t * p^T, only for non-missing values
		for i := 0; i < n; i++ {
			for j := 0; j < m; j++ {
				if !math.IsNaN(Xwork.At(i, j)) {
					Xwork.Set(i, j, Xwork.At(i, j)-tData[i]*pData[j])
				}
			}
		}
	}

	// Calculate eigenvalues from scores for retained components
	// T might have been sliced to fewer columns if convergence stopped early
	_, actualComponents := T.Dims()
	allEigenvalues := make([]float64, actualComponents)
	for i := 0; i < actualComponents; i++ {
		scoreCol := mat.Col(nil, i, T)
		var eigenvalue float64
		count := 0
		for _, v := range scoreCol {
			if !math.IsNaN(v) {
				eigenvalue += v * v
				count++
			}
		}
		if count > 0 {
			allEigenvalues[i] = eigenvalue / float64(n-1)
		}
	}

	// Calculate residual variance to estimate eigenvalues for non-retained components
	// The residual matrix Xwork contains what's left after extracting actualComponents
	if m > actualComponents {
		// Calculate total residual variance (only for non-missing values)
		var residualVar float64
		count := 0
		for i := 0; i < n; i++ {
			for j := 0; j < m; j++ {
				val := Xwork.At(i, j)
				if !math.IsNaN(val) {
					residualVar += val * val
					count++
				}
			}
		}
		if count > 0 {
			residualVar /= float64(n - 1)

			// Estimate eigenvalues for non-retained components
			// Distribute residual variance equally among remaining components
			remainingComponents := m - actualComponents
			if remainingComponents > 0 {
				// Extend eigenvalues array to include all possible components
				extendedEigenvalues := make([]float64, m)
				copy(extendedEigenvalues, allEigenvalues)

				// Distribute residual variance among remaining components
				avgResidualEigenvalue := residualVar / float64(remainingComponents)
				for i := actualComponents; i < m; i++ {
					extendedEigenvalues[i] = avgResidualEigenvalue
				}
				allEigenvalues = extendedEigenvalues
			}
		}
	}

	return T, P, allEigenvalues, nil
}

// svdAlgorithm implements SVD-based PCA using Singular Value Decomposition
// The scores are computed as T = U * Σ and loadings as P = V
// Reference: Jolliffe, I.T. (2002). Principal Component Analysis (2nd ed.). Springer.
func (p *PCAImpl) svdAlgorithm(X *mat.Dense, nComponents int) (*mat.Dense, *mat.Dense, []float64, error) {
	n, m := X.Dims()

	// Perform SVD: X = U * Σ * V^T
	var svd mat.SVD
	ok := svd.Factorize(X, mat.SVDThin)
	if !ok {
		return nil, nil, nil, fmt.Errorf("SVD factorization failed")
	}

	// Get U and V matrices
	var u, v mat.Dense
	svd.UTo(&u)
	svd.VTo(&v)

	// Get singular values
	s := svd.Values(nil)

	// Check if we have enough components
	actualComponents := nComponents
	if len(s) < nComponents {
		actualComponents = len(s)
	}

	// Truncate to requested number of components
	uTrunc := u.Slice(0, n, 0, actualComponents).(*mat.Dense)
	vTrunc := v.Slice(0, m, 0, actualComponents).(*mat.Dense)

	// Create diagonal matrix with singular values
	sigma := mat.NewDiagDense(actualComponents, s[:actualComponents])

	// Scores = U * Σ
	scores := mat.NewDense(n, actualComponents, nil)
	scores.Mul(uTrunc, sigma)

	// Loadings = V (columns are the principal components)
	loadings := mat.NewDense(m, actualComponents, nil)
	loadings.Copy(vTrunc)

	// Convert singular values to eigenvalues
	// eigenvalue = (singular value)^2 / (n-1)
	allEigenvalues := make([]float64, len(s))
	for i, sv := range s {
		allEigenvalues[i] = (sv * sv) / float64(n-1)
	}

	return scores, loadings, allEigenvalues, nil
}

// calculateEigenvaluesFromScores computes eigenvalues from score matrix
// This is a fallback method when eigenvalues are not provided by the algorithm
func (p *PCAImpl) calculateEigenvaluesFromScores(scores *mat.Dense) []float64 {
	n, k := scores.Dims()
	eigenvalues := make([]float64, k)

	for i := 0; i < k; i++ {
		scoreCol := mat.Col(nil, i, scores)
		var sum float64
		for _, v := range scoreCol {
			sum += v * v
		}
		eigenvalues[i] = sum / float64(n-1)
	}

	return eigenvalues
}

// calculateVariance computes explained variance for each component
// DEPRECATED: This function is kept for reference but should not be used
func (p *PCAImpl) calculateVariance(X, scores, loadings *mat.Dense) ([]float64, []float64) {
	n, m := X.Dims()
	_, k := scores.Dims()

	// For NIPALS with missing values, we cannot calculate traditional explained variance
	// because the total variance of the data with missing values is not well-defined.
	// Instead, we return the eigenvalues (variance of each component) directly.
	if p.config.Method == "nipals" && p.config.MissingStrategy == types.MissingNative {
		explainedVar := make([]float64, k)

		// Calculate eigenvalues from scores
		for i := 0; i < k; i++ {
			scoreCol := mat.Col(nil, i, scores)
			var eigenvalue float64
			for _, v := range scoreCol {
				eigenvalue += v * v
			}
			eigenvalue /= float64(n - 1)
			explainedVar[i] = eigenvalue
		}

		// For cumulative variance with missing data, we just sum the eigenvalues
		// Note: These are not percentages but absolute values
		cumulativeVar := make([]float64, k)
		cumSum := 0.0
		for i := 0; i < k; i++ {
			cumSum += explainedVar[i]
			cumulativeVar[i] = cumSum
		}

		return explainedVar, cumulativeVar
	}

	// Original variance calculation for complete data
	totalVar := 0.0
	for j := 0; j < m; j++ {
		col := mat.Col(nil, j, X)
		for _, v := range col {
			totalVar += v * v
		}
	}
	totalVar /= float64(n - 1)

	// Variance explained by each component
	explainedVar := make([]float64, k)
	for i := 0; i < k; i++ {
		scoreCol := mat.Col(nil, i, scores)
		var componentVar float64
		for _, v := range scoreCol {
			componentVar += v * v
		}
		explainedVar[i] = componentVar / float64(n-1) / totalVar * 100
	}

	// Cumulative variance
	cumulativeVar := make([]float64, k)
	cumSum := 0.0
	for i := 0; i < k; i++ {
		cumSum += explainedVar[i]
		cumulativeVar[i] = cumSum
	}

	return explainedVar, cumulativeVar
}

// validateInput checks input data and configuration
func (p *PCAImpl) validateInput(data types.Matrix, config types.PCAConfig) error {
	if len(data) == 0 {
		return fmt.Errorf("empty data matrix")
	}

	n := len(data)
	m := len(data[0])

	// Check rectangular matrix
	for i, row := range data {
		if len(row) != m {
			return fmt.Errorf("inconsistent row length at index %d: expected %d, got %d", i, m, len(row))
		}
	}

	// Check dimensions
	if n < 2 {
		return fmt.Errorf("insufficient samples: need at least 2, got %d", n)
	}

	if m < 1 {
		return fmt.Errorf("insufficient features: need at least 1, got %d", m)
	}

	// Check for NaN values (unless using NIPALS with native missing value handling)
	if config.Method != "nipals" || config.MissingStrategy != types.MissingNative {
		for i := 0; i < n; i++ {
			for j := 0; j < m; j++ {
				if math.IsNaN(data[i][j]) {
					return fmt.Errorf("NaN value found at row %d, column %d - use missing value handling before PCA", i+1, j+1)
				}
			}
		}
	}

	// Check components
	maxComponents := n
	if m < n {
		maxComponents = m
	}

	if config.Components <= 0 {
		return fmt.Errorf("number of components must be positive, got %d", config.Components)
	}

	if config.Components > maxComponents {
		return fmt.Errorf("too many components requested: maximum %d, got %d", maxComponents, config.Components)
	}

	return nil
}

// SetPreprocessor sets the preprocessor for the PCA engine
func (p *PCAImpl) SetPreprocessor(preprocessor *Preprocessor) {
	p.preprocessor = preprocessor
}

// SetLoadings sets the loadings matrix and marks the engine as fitted
func (p *PCAImpl) SetLoadings(loadings types.Matrix, nComponents int) error {
	if len(loadings) == 0 || len(loadings[0]) == 0 {
		return fmt.Errorf("loadings matrix cannot be empty")
	}

	// The saved loadings matrix has shape (n_features, n_components)
	// This is already the correct shape for X @ V multiplication
	// Just convert to Dense matrix
	p.loadings = utils.MatrixToDense(loadings)
	p.nComponents = nComponents
	p.fitted = true
	return nil
}
