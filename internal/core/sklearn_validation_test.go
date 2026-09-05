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
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"

	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// SklearnReference represents the reference PCA results from sklearn
// Used to validate GoPCA implementations against the industry standard
type SklearnReference struct {
	Method                     string                 `json:"method"`
	Preprocessing              string                 `json:"preprocessing"`
	NSamples                   int                    `json:"n_samples"`
	NFeatures                  int                    `json:"n_features"`
	NComponents                int                    `json:"n_components"`
	ExplainedVarianceRatio     []float64              `json:"explained_variance_ratio"`
	CumulativeVariance         []float64              `json:"cumulative_variance"`
	SingularValues             []float64              `json:"singular_values"`
	Eigenvalues                []float64              `json:"eigenvalues"`
	Scores                     [][]float64            `json:"scores"`
	Loadings                   [][]float64            `json:"loadings"`
	TotalVariance              float64                `json:"total_variance"`
	ConditionNumber            float64                `json:"condition_number"`
	LoadingsOrthogonalityError float64                `json:"loadings_orthogonality_error"`
	ReconstructionErrors       []float64              `json:"reconstruction_errors"`
	Validations                map[string]interface{} `json:"validations"`
	PreprocessingParams        map[string][]float64   `json:"preprocessing_params"`
}

// ValidationTolerance defines acceptable differences for different data conditions
type ValidationTolerance struct {
	Base           float64 // Base tolerance for well-conditioned data
	IllConditioned float64 // Tolerance for ill-conditioned data (condition > 1e6)
	NearSingular   float64 // Tolerance for near-singular data (condition > 1e12)
}

var (
	// Standard tolerances based on numerical analysis best practices
	varianceTolerance = ValidationTolerance{
		Base:           1e-6,
		IllConditioned: 1e-4,
		NearSingular:   1e-2,
	}

	singularValueTolerance = ValidationTolerance{
		Base:           1e-8,
		IllConditioned: 1e-6,
		NearSingular:   1e-4,
	}
)

// loadSklearnReference loads a reference result from JSON file
func loadSklearnReference(path string) (*SklearnReference, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read reference file %s: %w", path, err)
	}

	var ref SklearnReference
	if err := json.Unmarshal(data, &ref); err != nil {
		return nil, fmt.Errorf("failed to parse reference JSON from %s: %w", path, err)
	}

	// Validate loaded data
	if ref.NSamples == 0 || ref.NFeatures == 0 {
		return nil, fmt.Errorf("invalid reference data: n_samples=%d, n_features=%d",
			ref.NSamples, ref.NFeatures)
	}

	return &ref, nil
}

// checkReferenceFiles checks if sklearn reference files exist and skips the test if not
func checkReferenceFiles(t *testing.T) {
	t.Helper()

	// Check if the reference results directory exists
	refDir := filepath.Join("..", "..", "testdata", "validation", "reference_results")
	if _, err := os.Stat(refDir); os.IsNotExist(err) {
		t.Skip("Sklearn reference files not found. Generate them with: make generate-sklearn-reference")
	}

	// Check if at least one reference file exists
	testFile := filepath.Join(refDir, "iris_pca_mean_center.json")
	if _, err := os.Stat(testFile); os.IsNotExist(err) {
		t.Skip("Sklearn reference files not found. Generate them with: make generate-sklearn-reference")
	}
}

// getTolerance returns the appropriate tolerance based on condition number
func getTolerance(conditionNumber float64, tol ValidationTolerance) float64 {
	if conditionNumber > 1e12 {
		return tol.NearSingular
	}
	if conditionNumber > 1e6 {
		return tol.IllConditioned
	}
	return tol.Base
}

// compareVectors compares two vectors element-wise with tolerance
func compareVectors(a, b []float64, tolerance float64, name string) error {
	if len(a) != len(b) {
		return fmt.Errorf("%s: length mismatch: got %d, expected %d", name, len(a), len(b))
	}

	for i := range a {
		if math.IsNaN(a[i]) || math.IsNaN(b[i]) {
			return fmt.Errorf("%s[%d]: NaN detected", name, i)
		}

		// Use relative tolerance for large values
		diff := math.Abs(a[i] - b[i])
		maxVal := math.Max(math.Abs(a[i]), math.Abs(b[i]))

		if maxVal > 1.0 {
			// Relative comparison for large values
			if diff/maxVal > tolerance {
				return fmt.Errorf("%s[%d]: relative difference %.2e exceeds tolerance %.2e (got %f, expected %f)",
					name, i, diff/maxVal, tolerance, a[i], b[i])
			}
		} else {
			// Absolute comparison for small values
			if diff > tolerance {
				return fmt.Errorf("%s[%d]: absolute difference %.2e exceeds tolerance %.2e (got %f, expected %f)",
					name, i, diff, tolerance, a[i], b[i])
			}
		}
	}

	return nil
}

// resolveSignAmbiguity handles the sign indeterminacy of eigenvectors
// Returns the sign-corrected matrix that minimizes difference with reference
func resolveSignAmbiguity(gopca, sklearn [][]float64) [][]float64 {
	if len(gopca) == 0 || len(sklearn) == 0 {
		return gopca
	}

	nRows := len(gopca)
	nCols := len(gopca[0])

	// Create a copy to avoid modifying original
	aligned := make([][]float64, nRows)
	for i := range aligned {
		aligned[i] = make([]float64, nCols)
		copy(aligned[i], gopca[i])
	}

	// For each component (column), determine best sign
	for j := 0; j < nCols; j++ {
		sumDiffPositive := 0.0
		sumDiffNegative := 0.0

		for i := 0; i < nRows; i++ {
			sumDiffPositive += math.Abs(aligned[i][j] - sklearn[i][j])
			sumDiffNegative += math.Abs(-aligned[i][j] - sklearn[i][j])
		}

		// Flip sign if it reduces total difference
		if sumDiffNegative < sumDiffPositive {
			for i := 0; i < nRows; i++ {
				aligned[i][j] = -aligned[i][j]
			}
		}
	}

	return aligned
}

// compareMatrices compares two matrices with sign ambiguity handling
func compareMatrices(gopca, sklearn [][]float64, tolerance float64, name string) error {
	if len(gopca) != len(sklearn) {
		return fmt.Errorf("%s: row count mismatch: got %d, expected %d",
			name, len(gopca), len(sklearn))
	}

	if len(gopca) == 0 {
		return nil
	}

	if len(gopca[0]) != len(sklearn[0]) {
		return fmt.Errorf("%s: column count mismatch: got %d, expected %d",
			name, len(gopca[0]), len(sklearn[0]))
	}

	// Resolve sign ambiguity
	aligned := resolveSignAmbiguity(gopca, sklearn)

	// Compare element-wise
	maxDiff := 0.0

	for i := range aligned {
		for j := range aligned[i] {
			diff := math.Abs(aligned[i][j] - sklearn[i][j])
			maxVal := math.Max(math.Abs(aligned[i][j]), math.Abs(sklearn[i][j]))

			relativeDiff := diff
			if maxVal > 1.0 {
				relativeDiff = diff / maxVal
			}

			if relativeDiff > maxDiff {
				maxDiff = relativeDiff
			}

			if relativeDiff > tolerance {
				return fmt.Errorf("%s[%d,%d]: difference %.2e exceeds tolerance %.2e (got %f, expected %f)",
					name, i, j, relativeDiff, tolerance, aligned[i][j], sklearn[i][j])
			}
		}
	}

	return nil
}

// TestValidateAgainstSklearn validates GoPCA against sklearn reference implementations
// This is the main validation test that ensures mathematical correctness
// Reference: sklearn.decomposition.PCA - the industry standard implementation
func TestValidateAgainstSklearn(t *testing.T) {
	// Check if reference files exist, skip if not
	checkReferenceFiles(t)

	// Test configurations
	datasets := []string{"iris", "wine", "corn"}
	preprocessingMethods := []string{"mean_center", "standardize"}

	for _, dataset := range datasets {
		for _, preprocessing := range preprocessingMethods {
			testName := fmt.Sprintf("%s_%s", dataset, preprocessing)
			t.Run(testName, func(t *testing.T) {
				// Load reference data
				refPath := filepath.Join("..", "..", "testdata", "validation", "reference_results",
					fmt.Sprintf("%s_pca_%s.json", dataset, preprocessing))

				ref, err := loadSklearnReference(refPath)
				require.NoError(t, err, "Failed to load reference data")

				// Load the same dataset
				dataPath := filepath.Join("..", "..", "testdata", dataset, fmt.Sprintf("%s.csv", dataset))
				data, err := loadTestDataAsMatrix(dataPath)
				require.NoError(t, err, "Failed to load test data")

				// Configure PCA to match sklearn settings
				config := types.PCAConfig{
					Components:    ref.NComponents,
					MeanCenter:    preprocessing == "mean_center" || preprocessing == "standardize",
					StandardScale: preprocessing == "standardize",
					Method:        "svd",
				}

				// Run GoPCA
				engine := NewPCAEngine()
				result, err := engine.Fit(data, config)
				require.NoError(t, err, "PCA fit failed")

				// Determine appropriate tolerance based on condition number
				tol := getTolerance(ref.ConditionNumber, varianceTolerance)

				// Validate explained variance ratios
				// GoPCA returns percentages (0-100), sklearn returns fractions (0-1)
				// Convert GoPCA percentages to fractions for comparison
				// Compared directly. Both are fractions of 1 as of V2; GoPCA used
				// to report percentages here and this comparison divided by 100
				// to bridge the difference (#848).
				err = compareVectors(result.ExplainedVarRatio, ref.ExplainedVarianceRatio,
					tol, "explained_variance_ratio")
				assert.NoError(t, err)

				// Validate singular values (if provided by the implementation)
				// Note: Not all PCA implementations provide singular values directly
				if len(result.SingularValues) > 0 {
					svTol := getTolerance(ref.ConditionNumber, singularValueTolerance)
					err = compareVectors(result.SingularValues, ref.SingularValues,
						svTol, "singular_values")
					assert.NoError(t, err)
				}

				// Validate cumulative variance, likewise directly.
				err = compareVectors(result.CumulativeVar, ref.CumulativeVariance,
					tol, "cumulative_variance")
				assert.NoError(t, err)

				// Validate total variance preservation
				// GoPCA percentages should sum to 100, not 1.0
				totalVar := 0.0
				for _, v := range result.ExplainedVarRatio {
					totalVar += v
				}
				assert.InDelta(t, 1.0, totalVar, tol,
					"Total explained variance should sum to 1.0")

				// Log validation success
				t.Logf("✓ Validated %s with condition number %.2e", testName, ref.ConditionNumber)
			})
		}
	}
}

// TestValidateScoresAndLoadings validates the scores and loadings matrices
// Handles sign ambiguity and validates reconstruction
func TestValidateScoresAndLoadings(t *testing.T) {
	// Check if reference files exist, skip if not
	checkReferenceFiles(t)

	// Focus on well-conditioned iris dataset for detailed validation
	refPath := filepath.Join("..", "..", "testdata", "validation", "reference_results",
		"iris_pca_mean_center.json")

	ref, err := loadSklearnReference(refPath)
	require.NoError(t, err)

	// Load iris data
	dataPath := filepath.Join("..", "..", "testdata", "iris", "iris.csv")
	data, err := loadTestDataAsMatrix(dataPath)
	require.NoError(t, err)

	config := types.PCAConfig{
		Components: ref.NComponents,
		MeanCenter: true,
		Method:     "svd",
	}

	engine := NewPCAEngine()
	result, err := engine.Fit(data, config)
	require.NoError(t, err)

	// Test loadings with sign ambiguity handling
	t.Run("Loadings", func(t *testing.T) {
		err := compareMatrices(result.Loadings, ref.Loadings,
			1e-6, "loadings")
		assert.NoError(t, err)

		// Verify orthogonality: V^T * V = I
		// Reference: Jolliffe & Cadima (2016), Phil. Trans. R. Soc. A
		orthError := calculateOrthogonalityError(result.Loadings)
		assert.Less(t, orthError, 1e-10,
			"Loadings should be orthonormal (V^T * V = I)")
	})

	// Test scores with sign ambiguity handling
	t.Run("Scores", func(t *testing.T) {
		err := compareMatrices(result.Scores, ref.Scores,
			1e-6, "scores")
		assert.NoError(t, err)
	})

	// Test reconstruction: X ≈ Scores * Loadings^T + mean
	t.Run("Reconstruction", func(t *testing.T) {
		// Skip reconstruction test if means are not available
		// (not all PCA implementations store the mean)
		if len(result.Means) == 0 {
			t.Skip("Means not available in PCA result, skipping reconstruction test")
		}

		// For centered data, reconstruction is just Scores * Loadings^T (without adding mean)
		reconstructed := reconstructData(result.Scores, result.Loadings, nil)

		// Apply mean-centering to original data for comparison
		centeredData := make([][]float64, len(data))
		for i := range data {
			centeredData[i] = make([]float64, len(data[i]))
			for j := range data[i] {
				centeredData[i][j] = data[i][j] - result.Means[j]
			}
		}

		// Calculate reconstruction error against centered data
		mse := 0.0
		count := 0
		for i := range centeredData {
			for j := range centeredData[i] {
				diff := reconstructed[i][j] - centeredData[i][j]
				mse += diff * diff
				count++
			}
		}
		mse /= float64(count)
		rmse := math.Sqrt(mse)

		// With 4 components for 4-feature iris data, reconstruction should be very good
		// but not perfect due to numerical precision
		assert.Less(t, rmse, 1e-3,
			"Reconstruction error should be small when using all principal components")
	})
}

// TestNIPALSValidation tests that NIPALS gives same results as SVD for complete data
func TestNIPALSValidation(t *testing.T) {
	// Check if reference files exist, skip if not
	checkReferenceFiles(t)

	// NIPALS should produce identical results to SVD for complete data (no missing values)
	// Reference: Wold, H. (1966). Estimation of principal components by iterative least squares.

	datasets := []struct {
		name          string
		dataPath      string
		refPath       string
		preprocessing string
	}{
		{
			name:          "iris_mean_center",
			dataPath:      filepath.Join("..", "..", "testdata", "iris", "iris.csv"),
			refPath:       filepath.Join("..", "..", "testdata", "validation", "reference_results", "iris_pca_mean_center.json"),
			preprocessing: "mean_center",
		},
		{
			name:          "wine_standardize",
			dataPath:      filepath.Join("..", "..", "testdata", "wine", "wine.csv"),
			refPath:       filepath.Join("..", "..", "testdata", "validation", "reference_results", "wine_pca_standardize.json"),
			preprocessing: "standardize",
		},
	}

	for _, ds := range datasets {
		t.Run(ds.name, func(t *testing.T) {
			// Load reference
			ref, err := loadSklearnReference(ds.refPath)
			require.NoError(t, err, "Failed to load reference")

			// Load data
			data, err := loadTestDataAsMatrix(ds.dataPath)
			require.NoError(t, err, "Failed to load data")

			// Configure PCA with NIPALS method
			config := types.PCAConfig{
				Components:    ref.NComponents,
				MeanCenter:    ds.preprocessing == "mean_center" || ds.preprocessing == "standardize",
				StandardScale: ds.preprocessing == "standardize",
				Method:        "nipals", // Use NIPALS instead of SVD
			}

			// Run NIPALS
			engine := NewPCAEngine()
			result, err := engine.Fit(data, config)
			require.NoError(t, err, "NIPALS fit failed")

			// Compare with sklearn (SVD) results
			// Tolerance may be slightly higher due to iterative nature of NIPALS
			tol := 1e-5

			// Compared directly: both sides are fractions of 1 as of V2 (#848).
			err = compareVectors(result.ExplainedVarRatio, ref.ExplainedVarianceRatio,
				tol, "explained_variance_ratio")
			assert.NoError(t, err, "NIPALS should match SVD for complete data")

			// Note: NIPALS may have sign ambiguity in loadings/scores
			// but explained variance should be identical
		})
	}
}

// TestMathematicalProperties validates fundamental mathematical properties
// Based on academic references from docs/references/_summaries.txt
func TestMathematicalProperties(t *testing.T) {
	// Check if reference files exist, skip if not
	checkReferenceFiles(t)

	refPath := filepath.Join("..", "..", "testdata", "validation", "reference_results",
		"iris_pca_standardize.json")

	ref, err := loadSklearnReference(refPath)
	require.NoError(t, err)

	dataPath := filepath.Join("..", "..", "testdata", "iris", "iris.csv")
	data, err := loadTestDataAsMatrix(dataPath)
	require.NoError(t, err)

	config := types.PCAConfig{
		Components:    ref.NComponents,
		MeanCenter:    true,
		StandardScale: true,
		Method:        "svd",
	}

	engine := NewPCAEngine()
	result, err := engine.Fit(data, config)
	require.NoError(t, err)

	// Test 1: Eigenvalue ordering
	// Reference: Golub & Van Loan (2013), Matrix Computations
	t.Run("EigenvalueOrdering", func(t *testing.T) {
		for i := 1; i < len(result.SingularValues); i++ {
			assert.GreaterOrEqual(t, result.SingularValues[i-1], result.SingularValues[i],
				"Eigenvalues should be in descending order: λ_%d >= λ_%d", i-1, i)
		}
	})

	// Test 2: Orthogonality of loadings
	// Reference: Jolliffe & Cadima (2016)
	t.Run("LoadingsOrthogonality", func(t *testing.T) {
		orthError := calculateOrthogonalityError(result.Loadings)
		assert.Less(t, orthError, ref.LoadingsOrthogonalityError*10,
			"Orthogonality error should be comparable to sklearn")
	})

	// Test 3: Variance preservation
	t.Run("VariancePreservation", func(t *testing.T) {
		totalVar := 0.0
		for _, v := range result.ExplainedVarRatio {
			totalVar += v
		}
		// GoPCA returns percentages, so total should be 100
		assert.InDelta(t, 1.0, totalVar, 1e-8,
			"Total explained variance must sum to 1.0")
	})

	// Test 4: Mahalanobis distance relationship
	// Reference: Brereton (2015), J. Chemometrics
	t.Run("MahalanobisDistance", func(t *testing.T) {
		// For each sample, calculate Mahalanobis distance
		// D² = Σ(score²/eigenvalue)
		for i := 0; i < min(10, len(result.Scores)); i++ { // Test first 10 samples
			mahalanobis := 0.0
			for j := range result.Scores[i] {
				// Use eigenvalues directly (already computed from singular values)
				eigenvalue := result.ExplainedVar[j]
				if eigenvalue > 1e-10 { // Avoid division by near-zero
					mahalanobis += result.Scores[i][j] * result.Scores[i][j] / eigenvalue
				}
			}

			// Mahalanobis distance should be finite and reasonable
			assert.False(t, math.IsNaN(mahalanobis), "Mahalanobis distance is NaN")
			assert.False(t, math.IsInf(mahalanobis, 0), "Mahalanobis distance is infinite")
		}
	})
}

// Helper functions

// loadTestDataAsMatrix loads CSV data as a types.Matrix
func loadTestDataAsMatrix(path string) (types.Matrix, error) {
	// Open the CSV file
	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("failed to open CSV file %s: %w", path, err)
	}
	defer func() {
		_ = file.Close()
	}()

	// Use mixed parser to handle both numeric and categorical columns
	format := types.DefaultCSVFormat()

	// Parse the CSV file with mixed types and targets
	// This will automatically separate numeric columns from categorical and target columns
	csvData, _, _, err := types.ParseCSVMixedWithTargets(file, format, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to parse CSV file %s: %w", path, err)
	}

	// Return the numeric matrix (excludes categorical and target columns)
	if len(csvData.Matrix) == 0 {
		return nil, fmt.Errorf("no numeric data found in %s", path)
	}

	return csvData.Matrix, nil
}

// calculateOrthogonalityError calculates ||V^T * V - I||_F (Frobenius norm)
func calculateOrthogonalityError(loadings [][]float64) float64 {
	if len(loadings) == 0 || len(loadings[0]) == 0 {
		return 0
	}

	n := len(loadings)
	m := len(loadings[0])

	error := 0.0

	// Calculate V^T * V
	for i := 0; i < m; i++ {
		for j := 0; j < m; j++ {
			dot := 0.0
			for k := 0; k < n; k++ {
				dot += loadings[k][i] * loadings[k][j]
			}

			// Compare with identity matrix
			expected := 0.0
			if i == j {
				expected = 1.0
			}

			diff := dot - expected
			error += diff * diff
		}
	}

	return math.Sqrt(error)
}

// reconstructData reconstructs the original data from PCA components
func reconstructData(scores, loadings [][]float64, mean []float64) [][]float64 {
	if len(scores) == 0 || len(loadings) == 0 {
		return nil
	}

	nSamples := len(scores)
	nFeatures := len(loadings)

	reconstructed := make([][]float64, nSamples)
	for i := range reconstructed {
		reconstructed[i] = make([]float64, nFeatures)

		// X_reconstructed = Scores * Loadings^T + mean
		for j := 0; j < nFeatures; j++ {
			sum := 0.0
			for k := 0; k < len(scores[i]); k++ {
				sum += scores[i][k] * loadings[j][k]
			}

			if mean != nil && j < len(mean) {
				sum += mean[j]
			}

			reconstructed[i][j] = sum
		}
	}

	return reconstructed
}
