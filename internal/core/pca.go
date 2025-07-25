package core

import (
	"fmt"
	"math"

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
	X := matrixToDense(data)

	// Check if we're using NIPALS with native missing value handling
	usingNativeMissing := config.Method == "nipals" && config.MissingStrategy == types.MissingNative

	// Preprocessing using the Preprocessor class (skip if using native missing value handling)
	if !usingNativeMissing && (config.MeanCenter || config.StandardScale || config.RobustScale || config.ScaleOnly || config.SNV || config.VectorNorm) {
		// Create preprocessor with the appropriate settings
		p.preprocessor = NewPreprocessorWithScaleOnly(config.MeanCenter, config.StandardScale, config.RobustScale, config.ScaleOnly, config.SNV, config.VectorNorm)

		// Convert to types.Matrix for preprocessor
		typeMatrix := denseToMatrix(X)

		// Fit and transform
		processedData, err := p.preprocessor.FitTransform(typeMatrix)
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}

		// Convert back to mat.Dense
		X = matrixToDense(processedData)
	} else if usingNativeMissing && (config.MeanCenter || config.StandardScale || config.RobustScale || config.ScaleOnly || config.SNV || config.VectorNorm) {
		// Warn that preprocessing is skipped with native missing value handling
		// The NIPALS algorithm will handle centering internally if needed
	}

	// Select PCA method
	var scores, loadings *mat.Dense
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
		scores, loadings, err = p.svdAlgorithm(X, config.Components)
	case "nipals":
		if hasMissing && config.MissingStrategy == types.MissingNative {
			scores, loadings, err = p.nipalsAlgorithmWithMissing(X, config.Components)
		} else {
			scores, loadings, err = p.nipalsAlgorithm(X, config.Components)
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

	// Calculate explained variance
	explainedVar, cumulativeVar := p.calculateVariance(X, scores, loadings)

	// Generate component labels
	componentLabels := make([]string, actualComponents)
	for i := 0; i < actualComponents; i++ {
		componentLabels[i] = fmt.Sprintf("PC%d", i+1)
	}

	// Calculate explained variance ratio
	explainedVarRatio := make([]float64, len(explainedVar))
	if p.config.Method == "nipals" && p.config.MissingStrategy == types.MissingNative {
		// For NIPALS with missing values, we report eigenvalues, not percentages
		// Set ratio to eigenvalues for now (the UI should handle this differently)
		copy(explainedVarRatio, explainedVar)
	} else {
		// Traditional percentage calculation for complete data
		totalVar := 0.0
		for _, v := range explainedVar {
			totalVar += v
		}
		for i, v := range explainedVar {
			if totalVar > 0 {
				explainedVarRatio[i] = v / totalVar * 100
			}
		}
	}

	// Get preprocessing stats if applicable
	var means, stddevs []float64
	if p.preprocessor != nil {
		means = p.preprocessor.GetMeans()
		stddevs = p.preprocessor.GetStdDevs()
	}

	return &types.PCAResult{
		Scores:               denseToMatrix(scores),
		Loadings:             denseToMatrix(loadings),
		ExplainedVar:         explainedVar,
		ExplainedVarRatio:    explainedVarRatio,
		CumulativeVar:        cumulativeVar,
		ComponentLabels:      componentLabels,
		ComponentsComputed:   actualComponents,
		Method:               config.Method,
		PreprocessingApplied: config.MeanCenter || config.StandardScale || config.RobustScale,
		Means:                means,
		StdDevs:              stddevs,
	}, nil
}

// Transform applies the fitted PCA model to new data
func (p *PCAImpl) Transform(data types.Matrix) (types.Matrix, error) {
	if !p.fitted {
		return nil, fmt.Errorf("model not fitted: call Fit first")
	}

	// Convert to gonum matrix
	X := matrixToDense(data)

	// Apply same preprocessing as during fit
	if p.preprocessor != nil {
		// Convert to types.Matrix for preprocessor
		typeMatrix := denseToMatrix(X)

		// Transform using preprocessor
		processedData, err := p.preprocessor.Transform(typeMatrix)
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}

		// Convert back to mat.Dense
		X = matrixToDense(processedData)
	}

	// Project onto loadings
	n, _ := X.Dims()
	scores := mat.NewDense(n, p.nComponents, nil)
	scores.Mul(X, p.loadings)

	return denseToMatrix(scores), nil
}

// FitTransform fits the model and transforms the data in one step
func (p *PCAImpl) FitTransform(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	return p.Fit(data, config)
}

// nipalsAlgorithm implements the NIPALS algorithm for PCA
func (p *PCAImpl) nipalsAlgorithm(X *mat.Dense, nComponents int) (*mat.Dense, *mat.Dense, error) {
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
				return nil, nil, fmt.Errorf("score vector has zero variance at component %d", k+1)
			}
			p.ScaleVec(1.0/tNorm, p)

			// Normalize p
			pNorm := math.Sqrt(mat.Dot(p, p))
			if pNorm < tolerance {
				return nil, nil, fmt.Errorf("loading vector has zero variance at component %d", k+1)
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
			return nil, nil, fmt.Errorf("NIPALS did not converge for component %d", k+1)
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

	return T, P, nil
}

// nipalsAlgorithmWithMissing implements NIPALS with native missing value handling
func (p *PCAImpl) nipalsAlgorithmWithMissing(X *mat.Dense, nComponents int) (*mat.Dense, *mat.Dense, error) {
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
			nComponents = k
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
				return nil, nil, fmt.Errorf("loading vector has zero variance at component %d", k+1)
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
			return nil, nil, fmt.Errorf("NIPALS did not converge for component %d", k+1)
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

	return T, P, nil
}

// svdAlgorithm implements SVD-based PCA
func (p *PCAImpl) svdAlgorithm(X *mat.Dense, nComponents int) (*mat.Dense, *mat.Dense, error) {
	n, m := X.Dims()

	// Perform SVD: X = U * Σ * V^T
	var svd mat.SVD
	ok := svd.Factorize(X, mat.SVDThin)
	if !ok {
		return nil, nil, fmt.Errorf("SVD factorization failed")
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

	return scores, loadings, nil
}

// calculateVariance computes explained variance for each component
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
	if !(config.Method == "nipals" && config.MissingStrategy == types.MissingNative) {
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

// Helper functions for type conversion

func matrixToDense(m types.Matrix) *mat.Dense {
	if len(m) == 0 || len(m[0]) == 0 {
		return mat.NewDense(0, 0, nil)
	}

	rows, cols := len(m), len(m[0])
	data := make([]float64, rows*cols)
	for i := 0; i < rows; i++ {
		for j := 0; j < cols; j++ {
			data[i*cols+j] = m[i][j]
		}
	}
	return mat.NewDense(rows, cols, data)
}

func denseToMatrix(d *mat.Dense) types.Matrix {
	r, c := d.Dims()
	m := make(types.Matrix, r)
	for i := 0; i < r; i++ {
		m[i] = make([]float64, c)
		for j := 0; j < c; j++ {
			m[i][j] = d.At(i, j)
		}
	}
	return m
}
