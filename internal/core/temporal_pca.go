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

package core

import (
	"fmt"
	"math"
	"sync"

	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/mat"
	"gonum.org/v1/gonum/stat"
)

// TemporalPCAImpl implements the PCAEngine interface for Temporal PCA (SSA-style)
// Based on Singular Spectrum Analysis (SSA) methodology.
//
// References:
// - Broomhead & King (1986): "Extracting qualitative dynamics from experimental data"
// - Vautard & Ghil (1989): "Singular spectrum analysis in nonlinear dynamics"
// - Golyandina et al. (2001): "Analysis of Time Series Structure: SSA and related techniques"
// - Ghil et al. (2002): "Advanced spectral methods for climatic time series"
type TemporalPCAImpl struct {
	// Core model (following PCAImpl pattern)
	preprocessor *Preprocessor
	loadings     *mat.Dense // Flat [K × (p·L)] matrix for components
	nComponents  int
	fitted       bool

	// Temporal-specific parameters
	numLags      int       // Number of time lags (L)
	origVars     int       // Original number of variables (p)
	singularVals []float64 // Singular values from SVD
	explainedVar []float64 // Fraction of total variance per component; drives the VarianceExplained cutoff
	// eigenvalues holds the raw eigenvalue per component. It is what
	// PCAResult.ExplainedVar carries, so that field means the same thing here
	// as it does for linear and kernel PCA. Before V2 the temporal path put
	// the fraction there instead, so explained_variance meant one thing in two
	// of GoPCA's three methods and something else in the third.
	eigenvalues []float64

	// Configuration
	config types.PCAConfig

	// Store preprocessing config separately for Transform
	preprocessingConfig types.PreprocessingConfig
}

// NewTemporalPCAEngine creates a new Temporal PCA engine instance
func NewTemporalPCAEngine() types.PCAEngine {
	return &TemporalPCAImpl{}
}

// EstimateTemporalPCAMemory estimates memory usage for temporal PCA
// Returns estimated bytes and a warning message if memory usage is high
func EstimateTemporalPCAMemory(samples, variables, lags int) (bytes int64, warning string) {
	// Lag matrix size: (T-L+1) × (p·L) × 8 bytes per float64
	effectiveSamples := samples - lags + 1
	if effectiveSamples <= 0 {
		return 0, "Number of lags exceeds number of samples"
	}

	lagMatrixSize := int64(effectiveSamples) * int64(variables*lags) * 8
	// 25% overhead for double-buffering during computation
	overhead := lagMatrixSize / 4
	total := lagMatrixSize + overhead

	if total > 2*1024*1024*1024 { // >2GB
		warning = fmt.Sprintf("Large memory usage expected (%.2f GB). Consider reducing lags or using CLI.",
			float64(total)/(1024*1024*1024))
	}
	return total, warning
}

// buildLagMatrix constructs the lag/trajectory matrix for temporal PCA
// Creates a matrix where each row contains the current observation and L-1 previous observations
// Input: data [T × p], numLags L
// Output: lagMatrix [(T-L+1) × (p·L)]
//
// Algorithm complexity: O(T × p × L) where T is number of samples, p is variables, L is lags
func (t *TemporalPCAImpl) buildLagMatrix(data *mat.Dense) (*mat.Dense, error) {
	rows, cols := data.Dims()
	if t.numLags > rows {
		return nil, fmt.Errorf("number of lags (%d) exceeds number of samples (%d)", t.numLags, rows)
	}

	effectiveRows := rows - t.numLags + 1
	laggedCols := cols * t.numLags

	// Allocate lag matrix
	lagMatrix := mat.NewDense(effectiveRows, laggedCols, nil)

	// Build lag matrix using concurrent processing for large datasets
	if effectiveRows > 1000 && t.numLags > 10 {
		// Parallel construction for large matrices
		var wg sync.WaitGroup
		numWorkers := 4
		chunkSize := effectiveRows / numWorkers

		for w := 0; w < numWorkers; w++ {
			start := w * chunkSize
			end := start + chunkSize
			if w == numWorkers-1 {
				end = effectiveRows
			}

			wg.Add(1)
			go func(start, end int) {
				defer wg.Done()
				t.fillLagMatrixRange(data, lagMatrix, start, end, cols)
			}(start, end)
		}
		wg.Wait()
	} else {
		// Sequential construction for smaller matrices
		t.fillLagMatrixRange(data, lagMatrix, 0, effectiveRows, cols)
	}

	return lagMatrix, nil
}

// fillLagMatrixRange fills a range of rows in the lag matrix
func (t *TemporalPCAImpl) fillLagMatrixRange(data, lagMatrix *mat.Dense, startRow, endRow, origCols int) {
	for i := startRow; i < endRow; i++ {
		// For each row i in the lag matrix, we include observations from time i to i+L-1
		for lag := 0; lag < t.numLags; lag++ {
			for col := 0; col < origCols; col++ {
				// Place data[i+lag, col] into lagMatrix[i, lag*origCols + col]
				value := data.At(i+lag, col)
				lagMatrix.Set(i, lag*origCols+col, value)
			}
		}
	}
}

// GetLoadingForLag returns the loading vector for a specific variable and lag
// variable: 0-indexed variable number
// lag: 0-indexed lag number (0 = current time, 1 = t-1, etc.)
// component: 0-indexed component number
func (t *TemporalPCAImpl) GetLoadingForLag(variable, lag, component int) (float64, error) {
	if !t.fitted {
		return 0, fmt.Errorf("model not fitted")
	}
	if variable < 0 || variable >= t.origVars {
		return 0, fmt.Errorf("variable index %d out of range [0, %d)", variable, t.origVars)
	}
	if lag < 0 || lag >= t.numLags {
		return 0, fmt.Errorf("lag index %d out of range [0, %d)", lag, t.numLags)
	}
	if component < 0 || component >= t.nComponents {
		return 0, fmt.Errorf("component index %d out of range [0, %d)", component, t.nComponents)
	}

	// Loading matrix is [nComponents × (origVars * numLags)]
	// Column index for (variable, lag) pair
	colIdx := lag*t.origVars + variable
	return t.loadings.At(component, colIdx), nil
}

// validateTemporalPCAInput validates input for temporal PCA
func validateTemporalPCAInput(data types.Matrix, config types.PCAConfig) error {
	// Basic validation
	if len(data) == 0 {
		return fmt.Errorf("empty data matrix")
	}

	// Check temporal lags first
	if config.TemporalLags <= 0 {
		return fmt.Errorf("temporal PCA requires positive number of lags, got %d", config.TemporalLags)
	}

	n := len(data)
	m := len(data[0])

	// Check rectangular matrix
	for i, row := range data {
		if len(row) != m {
			return fmt.Errorf("inconsistent row length at index %d: expected %d, got %d", i, m, len(row))
		}
	}

	// For temporal PCA with L=1, we can work with a single sample
	if config.TemporalLags == 1 && n < 1 {
		return fmt.Errorf("insufficient samples: need at least 1, got %d", n)
	} else if config.TemporalLags > 1 && n < config.TemporalLags {
		return fmt.Errorf("insufficient samples (%d) for %d lags", n, config.TemporalLags)
	}

	if m < 1 {
		return fmt.Errorf("insufficient features: need at least 1, got %d", m)
	}

	// Check for NaN values (unless using native missing value handling)
	allowNaN := config.MissingStrategy == types.MissingNative
	if err := ValidateNaNValues(data, allowNaN); err != nil {
		return err
	}

	// Validate component count (skip if using variance explained criterion)
	if config.VarianceExplained <= 0 && config.Components > 0 {
		effectiveSamples := n - config.TemporalLags + 1
		maxComponents := utils.MinInt(effectiveSamples, m*config.TemporalLags)
		if config.Components > maxComponents {
			return fmt.Errorf("too many components requested: maximum %d, got %d", maxComponents, config.Components)
		}
	}

	return nil
}

// Fit trains the Temporal PCA model on the provided data
// Algorithm complexity: O((T-L+1) × (p×L)²) for SVD computation where T is samples, p is variables, L is lags
func (t *TemporalPCAImpl) Fit(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	// Validate input using temporal-specific validation
	if err := validateTemporalPCAInput(data, config); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	// Extract temporal-specific configuration
	t.numLags = config.TemporalLags

	t.config = config
	t.nComponents = config.Components

	// Convert input data to gonum matrix
	// types.Matrix is [][]float64
	dataMatrix := utils.SliceToMatrix([][]float64(data))
	rows, cols := dataMatrix.Dims()
	t.origVars = cols

	// Check if we have enough samples for the requested lags
	if t.numLags > rows {
		return nil, fmt.Errorf("insufficient samples (%d) for %d lags", rows, t.numLags)
	}

	// IMPORTANT: For proper SSA/temporal PCA, preprocess the ORIGINAL series
	// before building the lag matrix, not after
	preprocessedOriginal := dataMatrix
	if config.MeanCenter || config.StandardScale || config.RobustScale || config.ScaleOnly {
		// Apply preprocessing to the original series
		t.preprocessor = NewPreprocessorWithScaleOnly(
			config.MeanCenter,
			config.StandardScale,
			config.RobustScale,
			config.ScaleOnly,
			false, // SNV
			false, // VectorNorm
		)

		origData := utils.MatrixToSlice(dataMatrix)
		if err := t.preprocessor.Fit(types.Matrix(origData)); err != nil {
			return nil, fmt.Errorf("preprocessing fit failed: %w", err)
		}

		preprocessedData, err := t.preprocessor.Transform(types.Matrix(origData))
		if err != nil {
			return nil, fmt.Errorf("preprocessing transform failed: %w", err)
		}

		preprocessedOriginal = utils.SliceToMatrix([][]float64(preprocessedData))
	}

	// Build lag matrix from PREPROCESSED data
	lagMatrix, err := t.buildLagMatrix(preprocessedOriginal)
	if err != nil {
		return nil, fmt.Errorf("failed to build lag matrix: %w", err)
	}

	effectiveRows, laggedCols := lagMatrix.Dims()

	// Handle variance explained specification
	var targetComponents int
	if config.VarianceExplained > 0 {
		// Will determine components after SVD based on variance
		targetComponents = utils.MinInt(effectiveRows, laggedCols)
	} else {
		targetComponents = t.nComponents
		if targetComponents > utils.MinInt(effectiveRows, laggedCols) {
			targetComponents = utils.MinInt(effectiveRows, laggedCols)
		}
	}

	// Store preprocessing config for later use
	t.preprocessingConfig = types.PreprocessingConfig{
		Method:        types.PreprocessingTypeMeanCenter,
		MeanCenter:    config.MeanCenter,
		StandardScale: config.StandardScale,
		RobustScale:   config.RobustScale,
		ScaleOnly:     config.ScaleOnly,
		SNV:           config.SNV,
		VectorNorm:    config.VectorNorm,
	}

	// IMPORTANT: For proper SSA/temporal PCA, we do NOT apply column-wise preprocessing
	// to the lag matrix. The preprocessing was already applied to the original series
	// before building the lag matrix. This preserves the temporal structure.
	preprocessedData := lagMatrix

	// Note: We don't need laggedMeans and laggedScales for the lag matrix anymore
	// since we're not preprocessing it. The preprocessor is stored for the original series.

	// Handle edge case for single effective sample
	if effectiveRows == 1 {
		// For a single sample, we can't do meaningful PCA
		// Return a trivial result with one component
		t.nComponents = 1
		t.singularVals = []float64{1.0}
		t.explainedVar = []float64{1.0}
		t.eigenvalues = []float64{1.0}

		// Single loading vector (normalized row)
		t.loadings = mat.NewDense(1, laggedCols, nil)
		norm := 0.0
		for j := 0; j < laggedCols; j++ {
			val := preprocessedData.At(0, j)
			norm += val * val
		}
		norm = math.Sqrt(norm)
		if norm > 0 {
			for j := 0; j < laggedCols; j++ {
				t.loadings.Set(0, j, preprocessedData.At(0, j)/norm)
			}
		} else {
			// All zeros - set arbitrary loading
			t.loadings.Set(0, 0, 1.0)
		}

		// Single score
		scores := mat.NewDense(1, 1, []float64{norm})

		t.fitted = true

		// For single sample, create a trivial temporal eigenvector
		temporalEigenvectors := mat.NewDense(t.numLags, 1, nil)
		for i := 0; i < t.numLags; i++ {
			temporalEigenvectors.Set(i, 0, 1.0/math.Sqrt(float64(t.numLags))) // Normalized vector
		}

		// Generate component labels for single component
		componentLabels := []string{"PC1"}

		// For single sample, compute trivial variable importance
		variableImportance := make([][]float64, 1)
		variableImportance[0] = make([]float64, t.origVars)
		for var_ := 0; var_ < t.origVars; var_++ {
			sumSquared := 0.0
			for lag := 0; lag < t.numLags; lag++ {
				loading, _ := t.GetLoadingForLag(var_, lag, 0)
				sumSquared += loading * loading
			}
			variableImportance[0][var_] = math.Sqrt(sumSquared / float64(t.numLags))
		}

		return &types.PCAResult{
			Scores:                     utils.MatrixToSlice(scores),
			Loadings:                   utils.MatrixToSlice(t.loadings),
			ExplainedVar:               t.eigenvalues,
			ExplainedVarRatio:          []float64{1.0},   // Single component explains all of it
			CumulativeVar:              []float64{100.0}, // Cumulative should also be 100%
			ComponentLabels:            componentLabels,
			SingularValues:             t.singularVals,
			Method:                     "temporal",
			ComponentsComputed:         1,
			PreprocessingApplied:       t.preprocessor != nil,
			TemporalEigenvectors:       utils.MatrixToSlice(temporalEigenvectors), // Add trivial U matrix for single sample case
			TemporalVariableImportance: types.Matrix(variableImportance),          // Add variable importance for single sample case
		}, nil
	}

	// Perform SVD on the preprocessed lag matrix
	var svd mat.SVD
	ok := svd.Factorize(preprocessedData, mat.SVDThin)
	if !ok {
		return nil, fmt.Errorf("SVD factorization failed")
	}

	s := svd.Values(nil)
	minDim := len(s)

	// Get V matrix - VTo returns V, not V^T
	// V is [laggedCols × minDim] for thin SVD
	v := mat.NewDense(laggedCols, minDim, nil)
	svd.VTo(v)

	// Transpose V to get V^T for loadings
	vt := mat.NewDense(minDim, laggedCols, nil)
	vt.Copy(v.T())

	u := mat.NewDense(effectiveRows, minDim, nil)
	svd.UTo(u)

	// Calculate explained variance
	// Store ALL eigenvalues for proper variance calculation.
	//
	// The divisor matters for what explained_variance means. Linear PCA reports
	// lambda = sigma^2/(n-1), the variance along a component, and so does
	// scikit-learn. Temporal PCA reported sigma^2 instead, so the same field
	// carried a variance in two of GoPCA's methods and a sum of squares in the
	// third. Here n is the number of trajectory rows, since that is what the
	// decomposition actually operated on.
	//
	// The ratio and the VarianceExplained cutoff are unaffected: the divisor
	// cancels in eigenvalue/totalVar.
	varianceDivisor := float64(effectiveRows - 1)
	if varianceDivisor <= 0 {
		varianceDivisor = 1
	}
	allEigenvalues := make([]float64, len(s))
	for i, val := range s {
		allEigenvalues[i] = val * val / varianceDivisor
	}

	totalVar := 0.0
	for _, eigenval := range allEigenvalues {
		totalVar += eigenval
	}

	t.explainedVar = make([]float64, len(s))
	t.eigenvalues = make([]float64, len(s))
	cumVar := 0.0
	actualComponents := targetComponents

	for i := range s {
		variance := allEigenvalues[i] / totalVar
		t.explainedVar[i] = variance
		t.eigenvalues[i] = allEigenvalues[i]
		cumVar += variance

		// If using variance explained criterion, find cutoff
		if config.VarianceExplained > 0 && cumVar >= config.VarianceExplained && i < actualComponents {
			actualComponents = i + 1
			break
		}
	}

	// Store the determined number of components
	// Make sure we don't exceed the actual rank of the matrix
	maxComponents := len(s)
	if actualComponents > maxComponents {
		actualComponents = maxComponents
	}
	t.nComponents = actualComponents
	t.singularVals = s[:actualComponents]

	// Extract loadings (first nComponents rows of V^T)
	t.loadings = mat.NewDense(t.nComponents, laggedCols, nil)
	for i := 0; i < t.nComponents; i++ {
		for j := 0; j < laggedCols; j++ {
			t.loadings.Set(i, j, vt.At(i, j))
		}
	}

	// Calculate scores (U * S for first nComponents)
	scores := mat.NewDense(effectiveRows, t.nComponents, nil)
	for i := 0; i < effectiveRows; i++ {
		for j := 0; j < t.nComponents; j++ {
			scores.Set(i, j, u.At(i, j)*t.singularVals[j])
		}
	}

	t.fitted = true

	// Explained variance ratio, as a fraction of the total (see V2 note below)
	explainedVarRatio := make([]float64, t.nComponents)
	for i := 0; i < t.nComponents; i++ {
		explainedVarRatio[i] = t.explainedVar[i]
	}

	// Calculate cumulative variance (as cumulative sum of percentages)
	cumulativeVar := make([]float64, t.nComponents)
	cumSum := 0.0
	for i := 0; i < t.nComponents; i++ {
		cumSum += explainedVarRatio[i]
		cumulativeVar[i] = cumSum
	}

	// Build temporal eigenvectors matrix [numLags × nComponents] for visualization.
	//
	// For each component we show the signed loadings of its single most influential
	// channel across all lag positions. The dominant channel for component c is
	// defined as the variable v* with the largest RMS loading across lags:
	//
	//   v*(c) = argmax_v sqrt( mean_l loadings[c, l*p + v]^2 )
	//   temporalEigenvectors[lag, c] = loadings[c, lag*p + v*(c)]
	//
	// This approach preserves sign (so oscillatory components appear as sinusoids
	// and trend components as monotone curves) without the cancellation problem of
	// averaging channels with opposite-sign spatial loadings — which is the common
	// case in EEG where different brain regions load with opposite signs on the same
	// component (Broomhead & King, 1986; Vautard & Ghil, 1989).
	//
	// Layout of t.loadings: row c, column lag*origVars+v  →  loading of component c
	// for channel v at lag offset l.
	temporalEigenvectors := mat.NewDense(t.numLags, t.nComponents, nil)
	for comp := 0; comp < t.nComponents; comp++ {
		// Find the dominant channel for this component (highest RMS across lags)
		dominantVar := 0
		maxRMS := -1.0
		for v := 0; v < t.origVars; v++ {
			sumSq := 0.0
			for lag := 0; lag < t.numLags; lag++ {
				val := t.loadings.At(comp, lag*t.origVars+v)
				sumSq += val * val
			}
			rms := math.Sqrt(sumSq / float64(t.numLags))
			if rms > maxRMS {
				maxRMS = rms
				dominantVar = v
			}
		}
		// Store the signed loadings for the dominant channel across all lags
		for lag := 0; lag < t.numLags; lag++ {
			temporalEigenvectors.Set(lag, comp, t.loadings.At(comp, lag*t.origVars+dominantVar))
		}
	}

	// Generate component labels
	componentLabels := make([]string, t.nComponents)
	for i := 0; i < t.nComponents; i++ {
		componentLabels[i] = fmt.Sprintf("PC%d", i+1)
	}

	// Compute variable importance (aggregated loadings across lags)
	variableImportance, err := t.ComputeVariableImportance()
	if err != nil {
		// Log but don't fail - variable importance is optional
		fmt.Printf("Warning: failed to compute variable importance: %v\n", err)
		variableImportance = nil
	}

	// Prepare result
	result := &types.PCAResult{
		Scores:                     utils.MatrixToSlice(scores),
		Loadings:                   utils.MatrixToSlice(t.loadings),
		ExplainedVar:               t.eigenvalues[:t.nComponents],
		ExplainedVarRatio:          explainedVarRatio,
		CumulativeVar:              cumulativeVar,
		ComponentLabels:            componentLabels,
		SingularValues:             t.singularVals,
		Method:                     "temporal",
		ComponentsComputed:         t.nComponents,
		AllEigenvalues:             allEigenvalues, // Store all eigenvalues for diagnostic purposes
		PreprocessingApplied:       t.preprocessor != nil,
		TemporalEigenvectors:       utils.MatrixToSlice(temporalEigenvectors), // Add U matrix for temporal loadings visualization
		TemporalVariableImportance: types.Matrix(variableImportance),          // Add variable importance for temporal PCA
	}

	return result, nil
}

// Transform applies the fitted Temporal PCA model to new data
func (t *TemporalPCAImpl) Transform(data types.Matrix) (types.Matrix, error) {
	if !t.fitted {
		return nil, fmt.Errorf("model not fitted")
	}

	// Validate input data
	if len(data) == 0 {
		return nil, fmt.Errorf("empty data matrix")
	}
	if len(data[0]) == 0 {
		return nil, fmt.Errorf("empty data columns")
	}

	// Convert input data to gonum matrix
	// types.Matrix is [][]float64
	dataMatrix := utils.SliceToMatrix([][]float64(data))
	rows, cols := dataMatrix.Dims()

	if cols != t.origVars {
		return nil, fmt.Errorf("number of variables (%d) doesn't match training data (%d)", cols, t.origVars)
	}

	if rows < t.numLags {
		return nil, fmt.Errorf("insufficient samples (%d) for %d lags", rows, t.numLags)
	}

	// Apply same preprocessing to original series as during training
	preprocessedOriginal := dataMatrix
	if t.preprocessor != nil {
		origData := utils.MatrixToSlice(dataMatrix)
		preprocessedData, err := t.preprocessor.Transform(types.Matrix(origData))
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}
		preprocessedOriginal = utils.SliceToMatrix([][]float64(preprocessedData))
	}

	// Build lag matrix from preprocessed data
	lagMatrix, err := t.buildLagMatrix(preprocessedOriginal)
	if err != nil {
		return nil, fmt.Errorf("failed to build lag matrix: %w", err)
	}

	// Use lag matrix directly (no further preprocessing)
	preprocessedData := lagMatrix

	// Project onto principal components
	// scores = preprocessedData * loadings^T
	effectiveRows, _ := preprocessedData.Dims()
	scores := mat.NewDense(effectiveRows, t.nComponents, nil)
	scores.Mul(preprocessedData, t.loadings.T())

	return types.Matrix(utils.MatrixToSlice(scores)), nil
}

// FitTransform fits the model and transforms the data in one step
func (t *TemporalPCAImpl) FitTransform(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	return t.Fit(data, config)
}

// ReconstructionError computes the reconstruction error for each sample
// This is computed in the standardized lag space as per SSA methodology
// Algorithm complexity: O((T-L+1) × (p×L) × k) where k is the number of components
func (t *TemporalPCAImpl) ReconstructionError(data types.Matrix) ([]float64, error) {
	if !t.fitted {
		return nil, fmt.Errorf("model not fitted")
	}

	// Transform data to scores
	scores, err := t.Transform(data)
	if err != nil {
		return nil, fmt.Errorf("failed to transform data: %w", err)
	}

	// Convert to gonum matrices
	// types.Matrix is [][]float64
	scoresMatrix := utils.SliceToMatrix([][]float64(scores))
	dataMatrix := utils.SliceToMatrix([][]float64(data))

	// Apply same preprocessing to original series as during training
	preprocessedOriginal := dataMatrix
	if t.preprocessor != nil {
		origData := utils.MatrixToSlice(dataMatrix)
		preprocessedDataSlice, err := t.preprocessor.Transform(types.Matrix(origData))
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}
		preprocessedOriginal = utils.SliceToMatrix([][]float64(preprocessedDataSlice))
	}

	// Build lag matrix from preprocessed data
	lagMatrix, err := t.buildLagMatrix(preprocessedOriginal)
	if err != nil {
		return nil, fmt.Errorf("failed to build lag matrix: %w", err)
	}

	// Use lag matrix directly (no further preprocessing)
	preprocessedData := lagMatrix

	// Reconstruct: scores * loadings
	reconstructed := mat.NewDense(scoresMatrix.RawMatrix().Rows, t.loadings.RawMatrix().Cols, nil)
	reconstructed.Mul(scoresMatrix, t.loadings)

	// Calculate reconstruction error for each sample
	rows, cols := preprocessedData.Dims()
	errors := make([]float64, rows)

	for i := 0; i < rows; i++ {
		sumSqErr := 0.0
		for j := 0; j < cols; j++ {
			diff := preprocessedData.At(i, j) - reconstructed.At(i, j)
			sumSqErr += diff * diff
		}
		errors[i] = math.Sqrt(sumSqErr / float64(cols))
	}

	return errors, nil
}

// GetLagContributions returns the contribution of each lag to the total variance
// Returns a matrix where rows are components and columns are lags
func (t *TemporalPCAImpl) GetLagContributions() ([][]float64, error) {
	if !t.fitted {
		return nil, fmt.Errorf("model not fitted")
	}

	contributions := make([][]float64, t.nComponents)
	for comp := 0; comp < t.nComponents; comp++ {
		contributions[comp] = make([]float64, t.numLags)

		for lag := 0; lag < t.numLags; lag++ {
			// Sum squared loadings for all variables at this lag
			sumSq := 0.0
			for v := 0; v < t.origVars; v++ {
				loading, _ := t.GetLoadingForLag(v, lag, comp)
				sumSq += loading * loading
			}
			contributions[comp][lag] = sumSq
		}

		// Normalize to sum to 1 for each component
		total := 0.0
		for _, val := range contributions[comp] {
			total += val
		}
		if total > 0 {
			for lag := range contributions[comp] {
				contributions[comp][lag] /= total
			}
		}
	}

	return contributions, nil
}

// ComputeVariableImportance aggregates loadings across lags to show variable importance
// Returns a matrix where rows are components and columns are original variables
// Uses RMS (Root Mean Square) aggregation to capture overall contribution strength
func (t *TemporalPCAImpl) ComputeVariableImportance() ([][]float64, error) {
	if !t.fitted {
		return nil, fmt.Errorf("model not fitted")
	}

	// Create importance matrix [components × variables]
	importance := make([][]float64, t.nComponents)
	for comp := 0; comp < t.nComponents; comp++ {
		importance[comp] = make([]float64, t.origVars)

		for var_ := 0; var_ < t.origVars; var_++ {
			// Aggregate loadings across all lags using RMS
			sumSquared := 0.0
			for lag := 0; lag < t.numLags; lag++ {
				loading, _ := t.GetLoadingForLag(var_, lag, comp)
				sumSquared += loading * loading
			}
			// RMS aggregation
			importance[comp][var_] = math.Sqrt(sumSquared / float64(t.numLags))
		}
	}

	return importance, nil
}

// ComputeAutoCorrelation computes the autocorrelation function for lag selection guidance
// Returns autocorrelation values for each variable up to maxLag
func ComputeAutoCorrelation(data [][]float64, maxLag int) ([][]float64, error) {
	if len(data) == 0 || len(data[0]) == 0 {
		return nil, fmt.Errorf("empty data")
	}

	rows := len(data)
	cols := len(data[0])

	if maxLag >= rows {
		maxLag = rows - 1
	}

	acf := make([][]float64, cols)
	for col := 0; col < cols; col++ {
		acf[col] = make([]float64, maxLag+1)

		// Extract column
		series := make([]float64, rows)
		for i := 0; i < rows; i++ {
			series[i] = data[i][col]
		}

		// Compute mean
		mean := stat.Mean(series, nil)

		// Compute variance (lag 0)
		var variance float64
		for _, val := range series {
			diff := val - mean
			variance += diff * diff
		}
		variance /= float64(rows)

		// ACF at lag 0 is always 1
		acf[col][0] = 1.0

		// Compute ACF for each lag
		for lag := 1; lag <= maxLag; lag++ {
			covariance := 0.0
			count := 0
			for i := lag; i < rows; i++ {
				covariance += (series[i] - mean) * (series[i-lag] - mean)
				count++
			}
			if count > 0 && variance > 0 {
				acf[col][lag] = (covariance / float64(count)) / variance
			}
		}
	}

	return acf, nil
}
