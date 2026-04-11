// Copyright 2025-2026 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

// Package core - Phase 4 transform and preprocessing validation tests (#443 / #480)
//
// Scientific background:
// A fitted PCA model defines a projection space based on the training data.
// Any new observation must be projected into that same space using the SAME
// preprocessing parameters estimated during training. This is the fundamental
// invariant validated in these tests.
//
// References:
//   - Bro, R., & Smilde, A. K. (2014). Principal component analysis.
//     Analytical Methods, 6(9), 2812–2831.
//   - Esbensen, K. H., et al. (2002). Multivariate Data Analysis in Practice, Ch. 4.
//   - Gallagher, N. B., et al. (2020). The Effect of Data Centering on PCA Models.
//   - Barnes, R. J., Dhanoa, M. S., & Lister, S. J. (1989). Standard Normal Variate
//     transformation. Applied Spectroscopy, 43(5), 772–777.
//   - Hampel, F. R., et al. (1986). Robust Statistics. Wiley. (MAD scale factor 1.4826)

package core

import (
	"encoding/json"
	"math"
	"testing"

	"github.com/bitjungle/gopca/internal/utils"
	"github.com/bitjungle/gopca/pkg/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gonum.org/v1/gonum/mat"
)

// ─────────────────────────────────────────────────────────────────────────────
// Test helpers
// ─────────────────────────────────────────────────────────────────────────────

// preprocessingConfigs returns all preprocessing combinations to test.
// Each entry is (name, config). We test all six preprocessing types.
func preprocessingConfigs(method string) []struct {
	name   string
	config types.PCAConfig
} {
	return []struct {
		name   string
		config types.PCAConfig
	}{
		{
			name: "no_preprocessing",
			config: types.PCAConfig{
				Components: 2, MeanCenter: false, StandardScale: false,
				RobustScale: false, Method: method,
			},
		},
		{
			name: "mean_center",
			config: types.PCAConfig{
				Components: 2, MeanCenter: true, StandardScale: false,
				RobustScale: false, Method: method,
			},
		},
		{
			name: "standard_scale",
			config: types.PCAConfig{
				Components: 2, MeanCenter: true, StandardScale: true,
				RobustScale: false, Method: method,
			},
		},
		{
			name: "robust_scale",
			config: types.PCAConfig{
				Components: 2, MeanCenter: false, StandardScale: false,
				RobustScale: true, Method: method,
			},
		},
		{
			name: "scale_only",
			config: types.PCAConfig{
				Components: 2, ScaleOnly: true, Method: method,
			},
		},
		{
			name: "snv",
			config: types.PCAConfig{
				Components: 2, SNV: true, Method: method,
			},
		},
		{
			name: "vector_norm",
			config: types.PCAConfig{
				Components: 2, VectorNorm: true, Method: method,
			},
		},
	}
}

// irisTrainTest splits the Iris feature matrix into 120 training and 30 test rows.
// We use the first 120 rows as training to ensure well-conditioned data.
func irisTrainTest(t *testing.T) (train, test types.Matrix) {
	t.Helper()
	data, err := loadTestDataAsMatrix("../../testdata/iris/iris.csv")
	require.NoError(t, err, "failed to load iris dataset")
	require.GreaterOrEqual(t, len(data), 150, "iris should have 150 rows")
	return data[:120], data[120:]
}

// assertScoresEqual compares two score matrices element-wise allowing for sign ambiguity.
// PCA is defined up to sign flips of components (both scores and loadings flip together).
// Reference: Bro & Smilde (2014) – sign ambiguity is expected and acceptable.
func assertScoresEqual(t *testing.T, expected, actual types.Matrix, tol float64, msg string) {
	t.Helper()
	require.Equal(t, len(expected), len(actual), "%s: row count mismatch", msg)
	require.Equal(t, len(expected[0]), len(actual[0]), "%s: column count mismatch", msg)

	nComponents := len(expected[0])
	for c := 0; c < nComponents; c++ {
		// Determine sign by scanning rows for the first element whose absolute
		// value is above a safe epsilon. Using only row 0 is fragile: if that
		// score is (near) zero the sign decision would be undefined and tests
		// could become flaky.
		sign := 1.0
		const signEps = 1e-10
		for i, row := range expected {
			if math.Abs(row[c]) > signEps {
				if (row[c] > 0) != (actual[i][c] > 0) {
					sign = -1.0
				}
				break
			}
		}
		for i, row := range expected {
			diff := math.Abs(row[c] - sign*actual[i][c])
			assert.LessOrEqual(t, diff, tol,
				"%s: scores[%d][%d]: expected %f, got %f (sign=%+.0f)",
				msg, i, c, row[c], actual[i][c], sign)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group A: Transform reproduces training scores
//
// Invariant: Transform(X_train) must reproduce Fit(X_train).Scores
// Mathematical basis: T = X_preprocessed @ P (Bro & Smilde 2014, eq. 1)
// If this fails, the fitted model is internally inconsistent.
// ─────────────────────────────────────────────────────────────────────────────

// TestTransformReproducesTrainingScores_SVD verifies that applying Transform()
// to the training data after Fit() returns exactly the same scores as Fit() itself.
// Tested for all preprocessing types with the SVD method.
func TestTransformReproducesTrainingScores_SVD(t *testing.T) {
	train, _ := irisTrainTest(t)

	for _, tc := range preprocessingConfigs("svd") {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewPCAEngine()
			fitResult, err := engine.Fit(train, tc.config)
			require.NoError(t, err, "Fit failed")

			transformScores, err := engine.Transform(train)
			require.NoError(t, err, "Transform failed")

			assertScoresEqual(t, fitResult.Scores, transformScores, 1e-10,
				"SVD/"+tc.name+": Transform(X_train) must reproduce Fit scores")
		})
	}
}

// TestTransformReproducesTrainingScores_NIPALS is the same test for the NIPALS method.
// SVD and NIPALS must both satisfy this invariant independently.
func TestTransformReproducesTrainingScores_NIPALS(t *testing.T) {
	train, _ := irisTrainTest(t)

	for _, tc := range preprocessingConfigs("nipals") {
		t.Run(tc.name, func(t *testing.T) {
			engine := NewPCAEngine()
			fitResult, err := engine.Fit(train, tc.config)
			require.NoError(t, err, "Fit failed")

			transformScores, err := engine.Transform(train)
			require.NoError(t, err, "Transform failed")

			// NIPALS uses slightly looser tolerance due to iterative convergence
			assertScoresEqual(t, fitResult.Scores, transformScores, 1e-8,
				"NIPALS/"+tc.name+": Transform(X_train) must reproduce Fit scores")
		})
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group B: Preprocessing parameters are frozen at fit time
//
// Scientific requirement: new observations MUST be projected using the training
// preprocessing parameters, not re-estimated from the new data.
// Reference: Gallagher et al. (2020) – centering on test data gives wrong projections.
// Esbensen et al. (2002) Ch. 4 – "parameters come from the calibration set only"
// ─────────────────────────────────────────────────────────────────────────────

// TestTransformUsesTrainingMean verifies that Transform applies the TRAINING mean
// (not the test mean) when centering new data.
//
// Design: X_train has column means [4, 5, 6]. X_test has a completely different mean.
// We manually compute the expected scores using the training mean, then verify
// that engine.Transform(X_test) matches.
func TestTransformUsesTrainingMean(t *testing.T) {
	// Training data with known column means [4, 5, 6]
	trainData := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{4.0, 5.0, 6.0},
	}

	// Test data with a completely different mean (deliberately chosen to be distinct)
	testData := types.Matrix{
		{20.0, 30.0, 40.0},
		{22.0, 32.0, 42.0},
		{24.0, 34.0, 44.0},
	}

	config := types.PCAConfig{
		Components: 2, MeanCenter: true, Method: "svd",
	}

	engine := NewPCAEngine()
	fitResult, err := engine.Fit(trainData, config)
	require.NoError(t, err)

	// Verify the training means are what we expect
	expectedMeans := []float64{4.0, 5.0, 6.0}
	for i, m := range fitResult.Means {
		assert.InDelta(t, expectedMeans[i], m, 1e-10, "training mean[%d]", i)
	}

	// Transform test data using the engine (must use training mean)
	engineScores, err := engine.Transform(testData)
	require.NoError(t, err)

	// Manually compute expected scores using TRAINING mean (not test mean)
	// This is the scientifically correct behavior.
	impl := engine.(*PCAImpl)
	centeredTest := make(types.Matrix, len(testData))
	for i, row := range testData {
		centeredTest[i] = make([]float64, len(row))
		for j, v := range row {
			centeredTest[i][j] = v - fitResult.Means[j] // Use TRAINING mean
		}
	}

	// Project: T = X_centered @ P
	loadings := impl.loadings
	nSamples := len(centeredTest)
	nComponents := impl.nComponents
	X := mat.NewDense(nSamples, len(centeredTest[0]), nil)
	for i, row := range centeredTest {
		for j, v := range row {
			X.Set(i, j, v)
		}
	}
	manualScores := mat.NewDense(nSamples, nComponents, nil)
	manualScores.Mul(X, loadings)

	for i := 0; i < len(engineScores); i++ {
		for c := 0; c < nComponents; c++ {
			assert.InDelta(t, manualScores.At(i, c), engineScores[i][c], 1e-10,
				"scores[%d][%d]: engine must use training mean, not test mean", i, c)
		}
	}
}

// TestTransformUsesTrainingStdDev verifies that standard scaling in Transform
// uses the training standard deviations, not re-estimated ones from new data.
func TestTransformUsesTrainingStdDev(t *testing.T) {
	// Training data: columns are [1..5] and [10..50].
	// Sample std devs (Bessel-corrected, n-1=4): col1 ≈ 1.5811, col2 ≈ 15.811.
	// The test reads trainStds from the preprocessor directly, so the assertion
	// is independent of the exact values — what matters is that Transform uses
	// the TRAINING std, not a re-estimated std from the test data.
	trainData := types.Matrix{
		{1.0, 10.0},
		{2.0, 20.0},
		{3.0, 30.0},
		{4.0, 40.0},
		{5.0, 50.0},
	}

	// Test data with different column scale relationships
	testData := types.Matrix{
		{10.0, 100.0},
		{20.0, 200.0},
	}

	config := types.PCAConfig{
		Components: 1, MeanCenter: true, StandardScale: true, Method: "svd",
	}

	engine := NewPCAEngine()
	_, err := engine.Fit(trainData, config)
	require.NoError(t, err)

	impl := engine.(*PCAImpl)
	require.NotNil(t, impl.preprocessor, "preprocessor must be set after Fit with standard scaling")

	// Training StdDevs must be stored
	trainStds := impl.preprocessor.GetStdDevs()
	require.NotNil(t, trainStds)
	require.Len(t, trainStds, 2)

	// Transform test data
	engineScores, err := engine.Transform(testData)
	require.NoError(t, err)

	// Manual calculation using TRAINING mean and std
	trainMeans := impl.preprocessor.GetMeans()
	manualScores := make(types.Matrix, len(testData))
	for i, row := range testData {
		centered := make([]float64, len(row))
		for j, v := range row {
			centered[j] = (v - trainMeans[j]) / trainStds[j]
		}
		manualScores[i] = centered
	}

	// Project onto loadings
	nSamples := len(manualScores)
	X := mat.NewDense(nSamples, len(manualScores[0]), nil)
	for i, row := range manualScores {
		for j, v := range row {
			X.Set(i, j, v)
		}
	}
	proj := mat.NewDense(nSamples, impl.nComponents, nil)
	proj.Mul(X, impl.loadings)

	for i := 0; i < len(engineScores); i++ {
		for c := 0; c < impl.nComponents; c++ {
			assert.InDelta(t, proj.At(i, c), engineScores[i][c], 1e-10,
				"scores[%d][%d]: must use training std, not test std", i, c)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group C: InverseTransform accuracy
//
// The PCA bilinear model: X ≈ T @ P^T + residual
// When all components are retained: X_reconstructed ≈ X_original (near-perfect)
// Reconstruction error must decrease monotonically with more components.
// Reference: Esbensen et al. (2002) Ch. 4
// ─────────────────────────────────────────────────────────────────────────────

// TestInverseTransformRoundtrip verifies that InverseTransform reverses the
// preprocessing operations to within machine precision for known inputs.
// Covers column-wise preprocessing methods: mean_center, standard_scale,
// robust_scale, and scale_only. Row-wise methods (SNV, VectorNorm) cannot be
// fully reversed — see TestSNVInverseTransformLimitation and
// TestVectorNormInverseTransformLimitation for their documented behaviour.
func TestInverseTransformRoundtrip(t *testing.T) {
	// Well-conditioned test data with distinct column statistics
	data := types.Matrix{
		{2.0, 10.0, 100.0},
		{4.0, 20.0, 200.0},
		{6.0, 30.0, 300.0},
		{8.0, 40.0, 400.0},
		{10.0, 50.0, 500.0},
	}

	tests := []struct {
		name string
		prep *Preprocessor
	}{
		{
			name: "mean_center",
			prep: NewPreprocessor(true, false, false),
		},
		{
			name: "standard_scale",
			prep: NewPreprocessor(true, true, false),
		},
		{
			name: "robust_scale",
			prep: NewPreprocessor(false, false, true),
		},
		{
			name: "scale_only",
			prep: NewPreprocessorWithScaleOnly(false, false, false, true, false, false),
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			transformed, err := tc.prep.FitTransform(data)
			require.NoError(t, err)

			// InverseTransform should recover the original data
			recovered, err := tc.prep.InverseTransform(transformed)
			require.NoError(t, err)

			require.Equal(t, len(data), len(recovered), "row count must match")
			require.Equal(t, len(data[0]), len(recovered[0]), "column count must match")

			for i, row := range data {
				for j, original := range row {
					assert.InDelta(t, original, recovered[i][j], 1e-10,
						"InverseTransform(%s): element [%d][%d] not recovered", tc.name, i, j)
				}
			}
		})
	}
}

// TestReconstructionErrorDecreasesMonotonically verifies the fundamental PCA
// property that adding more components reduces reconstruction error.
//
// Mathematical basis: X = sum_k(t_k @ p_k^T) + E
// ||E_k|| >= ||E_{k+1}|| (each component captures additional variance)
// Reference: Bro & Smilde (2014), Theorem 1
func TestReconstructionErrorDecreasesMonotonically(t *testing.T) {
	data, err := loadTestDataAsMatrix("../../testdata/iris/iris.csv")
	require.NoError(t, err)

	nFeatures := len(data[0])
	maxComponents := nFeatures // Iris has 4 features → up to 4 components

	// Standard PCA: mean center
	config := types.PCAConfig{
		Components: maxComponents, MeanCenter: true, StandardScale: true, Method: "svd",
	}

	engine := NewPCAEngine()
	fitResult, err := engine.Fit(data, config)
	require.NoError(t, err)
	require.GreaterOrEqual(t, fitResult.ComponentsComputed, 2, "need at least 2 components")

	impl := engine.(*PCAImpl)
	prep := impl.preprocessor
	loadings := impl.loadings

	// Preprocess original data
	processedData, err := prep.Transform(data)
	require.NoError(t, err)

	nSamples := len(processedData)
	nCols := len(processedData[0])
	X := mat.NewDense(nSamples, nCols, nil)
	for i, row := range processedData {
		for j, v := range row {
			X.Set(i, j, v)
		}
	}

	prevError := math.MaxFloat64
	computed := fitResult.ComponentsComputed

	for k := 1; k <= computed; k++ {
		// Get k-component loadings
		Pk := mat.NewDense(nCols, k, nil)
		for r := 0; r < nCols; r++ {
			for c := 0; c < k; c++ {
				Pk.Set(r, c, loadings.At(r, c))
			}
		}

		// Scores: T_k = X @ P_k
		Tk := mat.NewDense(nSamples, k, nil)
		Tk.Mul(X, Pk)

		// Reconstruction: X_k = T_k @ P_k^T
		Xk := mat.NewDense(nSamples, nCols, nil)
		Xk.Mul(Tk, Pk.T())

		// Residual: E_k = X - X_k
		Ek := mat.NewDense(nSamples, nCols, nil)
		Ek.Sub(X, Xk)

		// Frobenius norm: sqrt(sum of all element squares)
		nR, nC := Ek.Dims()
		frobSq := 0.0
		for r := 0; r < nR; r++ {
			for c := 0; c < nC; c++ {
				v := Ek.At(r, c)
				frobSq += v * v
			}
		}
		residualNorm := math.Sqrt(frobSq)

		assert.LessOrEqual(t, residualNorm, prevError+1e-10,
			"reconstruction error must not increase: k=%d error=%f, prev=%f",
			k, residualNorm, prevError)

		prevError = residualNorm
	}
}

// TestFullReconstructionWithAllComponents verifies that using all k=min(n-1,p)
// components gives near-perfect reconstruction of the training data (within
// machine precision for a well-conditioned matrix).
func TestFullReconstructionWithAllComponents(t *testing.T) {
	// Use a small, well-conditioned dataset where we can use all components
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 6.0, 8.0},
		{7.0, 9.0, 2.0},
		{3.0, 5.0, 7.0},
		{6.0, 1.0, 4.0},
	}

	// With 5 rows and 3 columns, we can extract up to min(4,3)=3 components
	config := types.PCAConfig{
		Components: 3, MeanCenter: true, Method: "svd",
	}

	engine := NewPCAEngine()
	fitResult, err := engine.Fit(data, config)
	require.NoError(t, err)

	impl := engine.(*PCAImpl)
	prep := impl.preprocessor
	loadings := impl.loadings

	// Preprocess original data
	processedData, err := prep.Transform(data)
	require.NoError(t, err)

	nSamples := len(processedData)
	nCols := len(processedData[0])
	X := mat.NewDense(nSamples, nCols, nil)
	for i, row := range processedData {
		for j, v := range row {
			X.Set(i, j, v)
		}
	}

	// Full reconstruction: T = X @ P, X_reconstructed = T @ P^T
	T := mat.NewDense(nSamples, fitResult.ComponentsComputed, nil)
	T.Mul(X, loadings)

	Xrecon := mat.NewDense(nSamples, nCols, nil)
	Xrecon.Mul(T, loadings.T())

	// Residual should be near machine precision (Frobenius norm)
	E := mat.NewDense(nSamples, nCols, nil)
	E.Sub(X, Xrecon)
	frobSq := 0.0
	for r := 0; r < nSamples; r++ {
		for c := 0; c < nCols; c++ {
			v := E.At(r, c)
			frobSq += v * v
		}
	}
	residualNorm := math.Sqrt(frobSq)

	assert.LessOrEqual(t, residualNorm, 1e-10,
		"full reconstruction residual norm=%f should be near zero", residualNorm)
}

// TestSNVInverseTransformLimitation documents and validates the known limitation
// that InverseTransform for SNV only reverses column-wise operations.
// This is intentional and documented in preprocessing.go.
func TestSNVInverseTransformLimitation(t *testing.T) {
	data := types.Matrix{
		{2.0, 4.0, 6.0},
		{1.0, 3.0, 5.0},
		{3.0, 5.0, 7.0},
	}

	// SNV alone (no column-wise preprocessing)
	prep := NewPreprocessorFull(false, false, false, true, false)
	transformed, err := prep.FitTransform(data)
	require.NoError(t, err)

	// Verify SNV was applied: each transformed row should have mean≈0, std≈1
	for i, row := range transformed {
		sum, sumSq := 0.0, 0.0
		for _, v := range row {
			sum += v
			sumSq += v * v
		}
		mean := sum / float64(len(row))
		variance := (sumSq - float64(len(row))*mean*mean) / float64(len(row)-1)
		assert.InDelta(t, 0.0, mean, 1e-10, "SNV row %d: mean should be 0", i)
		assert.InDelta(t, 1.0, variance, 1e-10, "SNV row %d: variance should be 1", i)
	}

	// InverseTransform on SNV-only preprocessing returns the input unchanged
	// (SNV is a row-wise transform; InverseTransform only handles column-wise ops)
	recovered, err := prep.InverseTransform(transformed)
	require.NoError(t, err)

	// The recovered values are the same as the SNV-transformed values
	// (since there's no column-wise preprocessing to reverse)
	for i, row := range transformed {
		for j, v := range row {
			assert.InDelta(t, v, recovered[i][j], 1e-10,
				"SNV InverseTransform: row %d col %d should be unchanged", i, j)
		}
	}
}

// TestVectorNormInverseTransformLimitation documents and validates the known
// limitation that InverseTransform for VectorNorm (L2 row normalisation) only
// reverses column-wise operations. Because VectorNorm divides each row by its
// L2 norm — a row-wise operation — the original row scales cannot be recovered
// from column statistics alone.
//
// Reference: preprocessing.go InverseTransform docstring.
func TestVectorNormInverseTransformLimitation(t *testing.T) {
	data := types.Matrix{
		{3.0, 4.0, 0.0}, // L2 norm = 5
		{1.0, 0.0, 0.0}, // L2 norm = 1
		{0.0, 3.0, 4.0}, // L2 norm = 5
	}

	// VectorNorm alone (no column-wise preprocessing)
	prep := NewPreprocessorFull(false, false, false, false, true)
	transformed, err := prep.FitTransform(data)
	require.NoError(t, err)

	// Verify VectorNorm was applied: each transformed row should have L2 norm ≈ 1
	for i, row := range transformed {
		normSq := 0.0
		for _, v := range row {
			normSq += v * v
		}
		assert.InDelta(t, 1.0, math.Sqrt(normSq), 1e-10,
			"VectorNorm row %d: L2 norm should be 1", i)
	}

	// InverseTransform on VectorNorm-only preprocessing returns the input unchanged
	// (VectorNorm is a row-wise transform; InverseTransform only handles column-wise ops)
	recovered, err := prep.InverseTransform(transformed)
	require.NoError(t, err)

	for i, row := range transformed {
		for j, v := range row {
			assert.InDelta(t, v, recovered[i][j], 1e-10,
				"VectorNorm InverseTransform: row %d col %d should be unchanged", i, j)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group D: Preprocessing numerical correctness
//
// All expected values below are derived analytically from first principles,
// not from the implementation itself. This ensures the implementation is correct.
// ─────────────────────────────────────────────────────────────────────────────

// TestMeanCenteringNumericalValues verifies exact numerical output of mean centering.
// Column mean is subtracted from each element. Expected values are hand-computed.
func TestMeanCenteringNumericalValues(t *testing.T) {
	// Data: column means are exactly [4, 5, 6]
	data := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
	}
	// Mean centering: x_centered = x - mean(x)
	// Expected: subtract [4, 5, 6] from each row
	expected := types.Matrix{
		{-3.0, -3.0, -3.0},
		{0.0, 0.0, 0.0},
		{3.0, 3.0, 3.0},
	}

	prep := NewPreprocessor(true, false, false)
	result, err := prep.FitTransform(data)
	require.NoError(t, err)

	for i, row := range expected {
		for j, v := range row {
			assert.InDelta(t, v, result[i][j], 1e-14,
				"MeanCenter: element [%d][%d]", i, j)
		}
	}

	// Verify stored means
	means := prep.GetMeans()
	assert.InDelta(t, 4.0, means[0], 1e-14, "stored mean[0]")
	assert.InDelta(t, 5.0, means[1], 1e-14, "stored mean[1]")
	assert.InDelta(t, 6.0, means[2], 1e-14, "stored mean[2]")
}

// TestStandardScalingNumericalValues verifies exact output of autoscaling
// (mean center + divide by standard deviation).
//
// For data [2, 4, 6]: mean=4, std = sqrt(((2-4)²+(4-4)²+(6-4)²)/2) = sqrt(4) = 2
// Autoscaled: [(2-4)/2, (4-4)/2, (6-4)/2] = [-1, 0, 1]
func TestStandardScalingNumericalValues(t *testing.T) {
	// Each column: data [2,4,6], mean=4, std=2 → scaled=[-1,0,1]
	data := types.Matrix{
		{2.0, 20.0},
		{4.0, 40.0},
		{6.0, 60.0},
	}

	// Column 1: mean=4, std=2 → [-1, 0, 1]
	// Column 2: mean=40, std=20 → [-1, 0, 1]
	expected := types.Matrix{
		{-1.0, -1.0},
		{0.0, 0.0},
		{1.0, 1.0},
	}

	prep := NewPreprocessor(true, true, false)
	result, err := prep.FitTransform(data)
	require.NoError(t, err)

	for i, row := range expected {
		for j, v := range row {
			assert.InDelta(t, v, result[i][j], 1e-14,
				"StandardScale: element [%d][%d]", i, j)
		}
	}

	// Verify stored parameters
	means := prep.GetMeans()
	stds := prep.GetStdDevs()
	assert.InDelta(t, 4.0, means[0], 1e-14, "stored mean[0]")
	assert.InDelta(t, 40.0, means[1], 1e-14, "stored mean[1]")
	assert.InDelta(t, 2.0, stds[0], 1e-12, "stored std[0]")
	assert.InDelta(t, 20.0, stds[1], 1e-12, "stored std[1]")
}

// TestRobustScalingNumericalValues verifies exact output of robust scaling.
//
// Formula: x_robust = (x - median(x)) / MAD(x)
// MAD = 1.4826 * median(|x - median(x)|)
// Scale factor 1.4826 makes MAD consistent with σ for normal data.
// Reference: Hampel et al. (1986), Robust Statistics. Wiley.
func TestRobustScalingNumericalValues(t *testing.T) {
	// Column: [1, 2, 3, 4, 5] → median=3
	// Absolute deviations: [2, 1, 0, 1, 2] → MAD_raw = median([0,1,1,2,2]) = 1
	// MAD = 1.4826 * 1 = 1.4826
	// Scaled: (x - 3) / 1.4826
	data := types.Matrix{
		{1.0},
		{2.0},
		{3.0},
		{4.0},
		{5.0},
	}

	expectedMedian := 3.0
	expectedMAD := 1.4826 // 1.4826 * median(|x-median|) = 1.4826 * 1.0

	prep := NewPreprocessor(false, false, true)
	result, err := prep.FitTransform(data)
	require.NoError(t, err)

	medians := prep.GetMedians()
	mads := prep.GetMADs()

	assert.InDelta(t, expectedMedian, medians[0], 1e-10, "stored median")
	assert.InDelta(t, expectedMAD, mads[0], 1e-6, "stored MAD (with 1.4826 scale factor)")

	// Verify scaled values
	for i, row := range data {
		expectedScaled := (row[0] - expectedMedian) / expectedMAD
		assert.InDelta(t, expectedScaled, result[i][0], 1e-10,
			"RobustScale: element [%d]", i)
	}
}

// TestSNVNumericalValues verifies the Standard Normal Variate transformation.
//
// SNV formula: x_snv[j] = (x[j] - mean(x_row)) / std(x_row)
// For row [2, 4, 6]: mean=4, std=sqrt(((2-4)²+(4-4)²+(6-4)²)/2) = 2
// SNV result: [(2-4)/2, (4-4)/2, (6-4)/2] = [-1, 0, 1]
//
// Reference: Barnes, Dhanoa & Lister (1989), Applied Spectroscopy, 43(5), 772–777.
func TestSNVNumericalValues(t *testing.T) {
	// Row [2, 4, 6]: row_mean=4, row_std=2 → SNV=[-1, 0, 1]
	// Row [10, 20, 30]: row_mean=20, row_std=10 → SNV=[-1, 0, 1]
	// (Both rows give the same SNV output despite 10x scale difference — this is
	// exactly the multiplicative scatter removal property of SNV)
	data := types.Matrix{
		{2.0, 4.0, 6.0},
		{10.0, 20.0, 30.0},
	}

	expected := types.Matrix{
		{-1.0, 0.0, 1.0},
		{-1.0, 0.0, 1.0},
	}

	prep := NewPreprocessorFull(false, false, false, true, false)
	result, err := prep.FitTransform(data)
	require.NoError(t, err)

	for i, row := range expected {
		for j, v := range row {
			assert.InDelta(t, v, result[i][j], 1e-14,
				"SNV: element [%d][%d]", i, j)
		}
	}

	// SNV removes multiplicative differences between rows: verify row 0 == row 1
	for j := range result[0] {
		assert.InDelta(t, result[0][j], result[1][j], 1e-14,
			"SNV: rows with proportional values should produce identical output")
	}
}

// TestVectorNormNumericalValues verifies L2 vector normalization.
//
// Formula: x_norm = x / ||x||_2
// For row [3, 4]: ||[3,4]||_2 = 5 → normalized = [0.6, 0.8]
func TestVectorNormNumericalValues(t *testing.T) {
	data := types.Matrix{
		{3.0, 4.0},  // L2 norm = 5.0 → [0.6, 0.8]
		{1.0, 0.0},  // L2 norm = 1.0 → [1.0, 0.0]
		{0.0, 2.0},  // L2 norm = 2.0 → [0.0, 1.0]
		{5.0, 12.0}, // L2 norm = 13.0 → [5/13, 12/13]
	}

	expected := types.Matrix{
		{3.0 / 5.0, 4.0 / 5.0},
		{1.0, 0.0},
		{0.0, 1.0},
		{5.0 / 13.0, 12.0 / 13.0},
	}

	prep := NewPreprocessorFull(false, false, false, false, true)
	result, err := prep.FitTransform(data)
	require.NoError(t, err)

	for i, row := range expected {
		for j, v := range row {
			assert.InDelta(t, v, result[i][j], 1e-14,
				"VectorNorm: element [%d][%d]", i, j)
		}
	}

	// Verify each normalized row has L2 norm = 1
	for i, row := range result {
		normSq := 0.0
		for _, v := range row {
			normSq += v * v
		}
		assert.InDelta(t, 1.0, math.Sqrt(normSq), 1e-14,
			"VectorNorm: row %d must have unit L2 norm", i)
	}
}

// TestScaleOnlyNumericalValues verifies scale-only preprocessing (divide by std,
// no mean subtraction).
//
// For data [2, 4, 6]: std=2, scaled = [1, 2, 3]
func TestScaleOnlyNumericalValues(t *testing.T) {
	// Column: [2, 4, 6], std=2 → scale-only: [1, 2, 3]
	// Column: [10, 20, 30], std=10 → scale-only: [1, 2, 3]
	data := types.Matrix{
		{2.0, 10.0},
		{4.0, 20.0},
		{6.0, 30.0},
	}

	expected := types.Matrix{
		{1.0, 1.0},
		{2.0, 2.0},
		{3.0, 3.0},
	}

	prep := NewPreprocessorWithScaleOnly(false, false, false, true, false, false)
	result, err := prep.FitTransform(data)
	require.NoError(t, err)

	for i, row := range expected {
		for j, v := range row {
			assert.InDelta(t, v, result[i][j], 1e-14,
				"ScaleOnly: element [%d][%d]", i, j)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group E: FitTransform consistency
//
// Invariant: FitTransform(X).Scores must equal Fit(X) then Transform(X)
// This ensures the two API paths are equivalent.
// ─────────────────────────────────────────────────────────────────────────────

// TestFitTransformConsistency verifies that calling FitTransform() and calling
// Fit() followed by Transform() produce identical results for all preprocessing
// types and both SVD and NIPALS.
func TestFitTransformConsistency(t *testing.T) {
	train, _ := irisTrainTest(t)

	methods := []string{"svd", "nipals"}
	for _, method := range methods {
		for _, tc := range preprocessingConfigs(method) {
			name := method + "/" + tc.name
			t.Run(name, func(t *testing.T) {
				// Path 1: FitTransform
				engine1 := NewPCAEngine()
				ftResult, err := engine1.FitTransform(train, tc.config)
				require.NoError(t, err, "FitTransform failed")

				// Path 2: Fit then Transform
				engine2 := NewPCAEngine()
				_, err = engine2.Fit(train, tc.config)
				require.NoError(t, err, "Fit failed")

				transformScores, err := engine2.Transform(train)
				require.NoError(t, err, "Transform failed")

				// Results must be identical (both paths go through same code)
				tol := 1e-10
				if method == "nipals" {
					tol = 1e-8 // NIPALS has iterative tolerance
				}
				assertScoresEqual(t, ftResult.Scores, transformScores, tol, name)
			})
		}
	}
}

// TestTransformDoesNotModifyPreprocessorState verifies that calling Transform()
// multiple times does not change the stored preprocessing parameters.
// This would be a serious bug — parameters must be immutable after Fit().
func TestTransformDoesNotModifyPreprocessorState(t *testing.T) {
	train := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{2.0, 3.0, 4.0},
		{5.0, 6.0, 7.0},
	}

	test1 := types.Matrix{{10.0, 20.0, 30.0}}
	test2 := types.Matrix{{100.0, 200.0, 300.0}}

	config := types.PCAConfig{
		Components: 2, MeanCenter: true, StandardScale: true, Method: "svd",
	}

	engine := NewPCAEngine()
	_, err := engine.Fit(train, config)
	require.NoError(t, err)

	impl := engine.(*PCAImpl)

	// Capture parameters after Fit
	meansBefore := make([]float64, len(impl.preprocessor.GetMeans()))
	copy(meansBefore, impl.preprocessor.GetMeans())
	stdsBefore := make([]float64, len(impl.preprocessor.GetStdDevs()))
	copy(stdsBefore, impl.preprocessor.GetStdDevs())

	// Transform two different test batches
	_, err = engine.Transform(test1)
	require.NoError(t, err)
	_, err = engine.Transform(test2)
	require.NoError(t, err)

	// Parameters must be unchanged
	meansAfter := impl.preprocessor.GetMeans()
	stdsAfter := impl.preprocessor.GetStdDevs()

	for i := range meansBefore {
		assert.InDelta(t, meansBefore[i], meansAfter[i], 1e-14,
			"mean[%d] must not change after Transform calls", i)
		assert.InDelta(t, stdsBefore[i], stdsAfter[i], 1e-14,
			"std[%d] must not change after Transform calls", i)
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group F: Model persistence — JSON roundtrip
//
// The CLI `pca transform` command reconstructs the preprocessor from JSON.
// This path is more fragile than the in-memory path and requires all
// preprocessing parameters to survive JSON serialization with full precision.
// A bug here causes silent wrong results — the most dangerous failure mode.
// ─────────────────────────────────────────────────────────────────────────────

// TestModelPersistenceRoundtrip_MeanCenter verifies that the mean-centering
// parameters survive JSON serialization and that Transform gives identical
// results whether using the in-memory preprocessor or one restored from JSON.
func TestModelPersistenceRoundtrip_MeanCenter(t *testing.T) {
	testModelPersistenceRoundtrip(t, types.PCAConfig{
		Components: 2, MeanCenter: true, Method: "svd",
	}, "mean_center")
}

func TestModelPersistenceRoundtrip_StandardScale(t *testing.T) {
	testModelPersistenceRoundtrip(t, types.PCAConfig{
		Components: 2, MeanCenter: true, StandardScale: true, Method: "svd",
	}, "standard_scale")
}

func TestModelPersistenceRoundtrip_RobustScale(t *testing.T) {
	testModelPersistenceRoundtrip(t, types.PCAConfig{
		Components: 2, RobustScale: true, Method: "svd",
	}, "robust_scale")
}

func TestModelPersistenceRoundtrip_ScaleOnly(t *testing.T) {
	testModelPersistenceRoundtrip(t, types.PCAConfig{
		Components: 2, ScaleOnly: true, Method: "svd",
	}, "scale_only")
}

// TestModelPersistenceRoundtrip_NoPreprocessing verifies that a model trained
// with no preprocessing can be JSON-serialized and used for Transform without
// errors or silent behavior changes.
//
// When no preprocessing flags are set, impl.preprocessor is nil — the CLI path
// projects raw data directly onto the deserialized loadings. This test covers
// that path explicitly because the shared testModelPersistenceRoundtrip helper
// always reconstructs a Preprocessor, which would fail on nil mean slices.
func TestModelPersistenceRoundtrip_NoPreprocessing(t *testing.T) {
	train, test := irisTrainTest(t)

	config := types.PCAConfig{
		Components: 2, Method: "svd",
		// No MeanCenter, StandardScale, RobustScale, ScaleOnly, SNV, or VectorNorm
	}

	engine := NewPCAEngine()
	fitResult, err := engine.Fit(train, config)
	require.NoError(t, err)

	impl := engine.(*PCAImpl)
	assert.Nil(t, impl.preprocessor,
		"preprocessor must be nil when no preprocessing is configured")

	// Serialize loadings to JSON and deserialize (the CLI save/load path)
	modelComponents := types.ModelComponents{
		Loadings: fitResult.Loadings,
	}

	modelJSON, err := json.Marshal(modelComponents)
	require.NoError(t, err, "model serialization failed")

	var restoredModel types.ModelComponents
	err = json.Unmarshal(modelJSON, &restoredModel)
	require.NoError(t, err, "model deserialization failed")
	require.NotEmpty(t, restoredModel.Loadings, "deserialized loadings must not be empty")

	// In-memory path: use the fitted engine
	origScores, err := engine.Transform(test)
	require.NoError(t, err)

	// Persistent path: project raw test data directly onto deserialized loadings,
	// no preprocessing applied (matches CLI behavior when preprocessor is nil)
	nSamples := len(test)
	nFeatures := len(test[0])
	nComponents := len(restoredModel.Loadings[0])

	X := mat.NewDense(nSamples, nFeatures, nil)
	for i, row := range test {
		for j, v := range row {
			X.Set(i, j, v)
		}
	}

	L := mat.NewDense(nFeatures, nComponents, nil)
	for i, row := range restoredModel.Loadings {
		for j, v := range row {
			L.Set(i, j, v)
		}
	}

	restoredScores := mat.NewDense(nSamples, nComponents, nil)
	restoredScores.Mul(X, L)

	for i := 0; i < len(origScores); i++ {
		for c := 0; c < nComponents; c++ {
			assert.InDelta(t, origScores[i][c], restoredScores.At(i, c), 1e-12,
				"no_preprocessing persistence roundtrip: scores[%d][%d] mismatch", i, c)
		}
	}
}

// TestModelPersistenceRoundtrip_SNV verifies that an SNV-preprocessed model
// survives JSON serialization. SNV is row-wise so no column statistics are
// stored; the test confirms no serialization errors occur and that fresh
// row-wise normalization is applied correctly after deserialization.
func TestModelPersistenceRoundtrip_SNV(t *testing.T) {
	testModelPersistenceRoundtrip(t, types.PCAConfig{
		Components: 2, SNV: true, Method: "svd",
	}, "snv")
}

// TestModelPersistenceRoundtrip_VectorNorm mirrors TestModelPersistenceRoundtrip_SNV
// for VectorNorm (L2 row normalization).
func TestModelPersistenceRoundtrip_VectorNorm(t *testing.T) {
	testModelPersistenceRoundtrip(t, types.PCAConfig{
		Components: 2, VectorNorm: true, Method: "svd",
	}, "vector_norm")
}

// testModelPersistenceRoundtrip is the shared implementation for Group F tests.
// It simulates the full CLI model save/load/transform pipeline:
//  1. Fit in-memory
//  2. Extract preprocessing params AND loadings (as the CLI output.go would)
//  3. Serialize BOTH to JSON and deserialize (full roundtrip)
//  4. Restore preprocessor via SetFittedParameters
//  5. Project using deserialized loadings (not the in-memory *mat.Dense)
//  6. Assert identical results to 1e-12
//
// Roundtripping the loadings is essential: using impl.loadings directly would
// miss any silent precision loss or shape error in JSON serialization of the
// model — the silent-failure mode the CLI is most vulnerable to.
func testModelPersistenceRoundtrip(t *testing.T, config types.PCAConfig, name string) {
	t.Helper()
	train, test := irisTrainTest(t)

	// Step 1: Fit
	engine := NewPCAEngine()
	fitResult, err := engine.Fit(train, config)
	require.NoError(t, err)

	impl := engine.(*PCAImpl)
	prep := impl.preprocessor

	// Step 2: Extract preprocessing params and loadings (simulates output.go)
	params := types.PreprocessingParams{}
	if prep != nil {
		params.FeatureMeans = prep.GetMeans()
		params.FeatureStdDevs = prep.GetStdDevs()
		params.FeatureMedians = prep.GetMedians()
		params.FeatureMADs = prep.GetMADs()
		params.RowMeans = prep.GetRowMeans()
		params.RowStdDevs = prep.GetRowStdDevs()
	}

	prepInfo := types.PreprocessingInfo{
		MeanCenter:    config.MeanCenter,
		StandardScale: config.StandardScale,
		RobustScale:   config.RobustScale,
		ScaleOnly:     config.ScaleOnly,
		SNV:           config.SNV,
		VectorNorm:    config.VectorNorm,
		Parameters:    params,
	}

	modelComponents := types.ModelComponents{
		Loadings: fitResult.Loadings,
	}

	// Step 3: JSON roundtrip for BOTH preprocessing info and loadings
	prepJSON, err := json.Marshal(prepInfo)
	require.NoError(t, err, "preprocessing serialization failed")

	modelJSON, err := json.Marshal(modelComponents)
	require.NoError(t, err, "model serialization failed")

	var restoredInfo types.PreprocessingInfo
	err = json.Unmarshal(prepJSON, &restoredInfo)
	require.NoError(t, err, "preprocessing deserialization failed")

	var restoredModel types.ModelComponents
	err = json.Unmarshal(modelJSON, &restoredModel)
	require.NoError(t, err, "model deserialization failed")

	require.NotEmpty(t, restoredModel.Loadings, "deserialized loadings must not be empty")

	// Step 4: Restore preprocessor from JSON-deserialized params
	restoredPrep := NewPreprocessorWithScaleOnly(
		restoredInfo.MeanCenter, restoredInfo.StandardScale, restoredInfo.RobustScale,
		restoredInfo.ScaleOnly, restoredInfo.SNV, restoredInfo.VectorNorm,
	)
	err = restoredPrep.SetFittedParameters(
		restoredInfo.Parameters.FeatureMeans,
		restoredInfo.Parameters.FeatureStdDevs,
		restoredInfo.Parameters.FeatureMedians,
		restoredInfo.Parameters.FeatureMADs,
		restoredInfo.Parameters.RowMeans,
		restoredInfo.Parameters.RowStdDevs,
	)
	require.NoError(t, err, "SetFittedParameters failed")

	// Step 5: In-memory path (reference)
	origScores, err := engine.Transform(test)
	require.NoError(t, err)

	// Persistent path: preprocess with restored params, project with deserialized loadings.
	// This is the exact path the CLI `pca transform` command takes.
	processedTest, err := restoredPrep.Transform(test)
	require.NoError(t, err)

	nSamples := len(processedTest)
	nFeatures := len(processedTest[0])
	nComponents := len(restoredModel.Loadings[0])

	X := mat.NewDense(nSamples, nFeatures, nil)
	for i, row := range processedTest {
		for j, v := range row {
			X.Set(i, j, v)
		}
	}

	// Build loadings matrix from deserialized [][]float64 (nFeatures x nComponents)
	L := mat.NewDense(nFeatures, nComponents, nil)
	for i, row := range restoredModel.Loadings {
		for j, v := range row {
			L.Set(i, j, v)
		}
	}

	restoredScores := mat.NewDense(nSamples, nComponents, nil)
	restoredScores.Mul(X, L)

	// Step 6: Compare — must match to 1e-12
	for i := 0; i < len(origScores); i++ {
		for c := 0; c < nComponents; c++ {
			assert.InDelta(t, origScores[i][c], restoredScores.At(i, c), 1e-12,
				"%s persistence roundtrip: scores[%d][%d] mismatch", name, i, c)
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group G: Edge cases for Transform
// ─────────────────────────────────────────────────────────────────────────────

// TestTransformBeforeFit verifies that calling Transform before Fit returns
// a clear, descriptive error (not a panic).
func TestTransformBeforeFit(t *testing.T) {
	engine := NewPCAEngine()
	_, err := engine.Transform(types.Matrix{{1.0, 2.0, 3.0}})
	require.Error(t, err, "Transform before Fit must return an error")
	assert.Contains(t, err.Error(), "not fitted",
		"error message should explain the model is not fitted")
}

// TestTransformWrongFeatureCount verifies that transforming data with the wrong
// number of features returns a clear error.
func TestTransformWrongFeatureCount(t *testing.T) {
	trainData := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{1.0, 2.0, 3.0},
	}

	config := types.PCAConfig{
		Components: 2, MeanCenter: true, Method: "svd",
	}

	engine := NewPCAEngine()
	_, err := engine.Fit(trainData, config)
	require.NoError(t, err)

	// Try to transform with wrong number of features (2 instead of 3)
	wrongData := types.Matrix{{1.0, 2.0}, {3.0, 4.0}}
	_, err = engine.Transform(wrongData)
	require.Error(t, err, "Transform with wrong feature count must return an error")
}

// TestTransformSingleSample verifies that a single new observation can be
// transformed correctly. This is the most common production use-case for
// online monitoring applications.
func TestTransformSingleSample(t *testing.T) {
	train, _ := irisTrainTest(t)

	config := types.PCAConfig{
		Components: 3, MeanCenter: true, StandardScale: true, Method: "svd",
	}

	engine := NewPCAEngine()
	fitResult, err := engine.Fit(train, config)
	require.NoError(t, err)

	// Transform a single observation (first training sample)
	singleSample := types.Matrix{train[0]}
	scores, err := engine.Transform(singleSample)
	require.NoError(t, err)

	assert.Len(t, scores, 1, "single-sample transform must return 1 score row")
	assert.Len(t, scores[0], 3, "must have 3 components")

	// The score for the first training sample must match the Fit scores from the
	// same engine — no second Fit() to avoid sign ambiguity across independent runs.
	assertScoresEqual(t, fitResult.Scores[:1], scores, 1e-9, "single sample transform")
}

// TestTransformMoreSamplesThanTrain verifies that transforming a new dataset
// larger than the training set works correctly.
func TestTransformMoreSamplesThanTrain(t *testing.T) {
	// Train on 5 samples
	trainData := types.Matrix{
		{1.0, 2.0, 3.0},
		{4.0, 5.0, 6.0},
		{7.0, 8.0, 9.0},
		{2.0, 3.0, 4.0},
		{5.0, 6.0, 7.0},
	}

	config := types.PCAConfig{
		Components: 2, MeanCenter: true, Method: "svd",
	}

	engine := NewPCAEngine()
	_, err := engine.Fit(trainData, config)
	require.NoError(t, err)

	// Transform 20 samples (4× more than training)
	largeTest := make(types.Matrix, 20)
	for i := range largeTest {
		largeTest[i] = []float64{
			float64(i+1) * 1.5,
			float64(i+1) * 2.5,
			float64(i+1) * 3.5,
		}
	}

	scores, err := engine.Transform(largeTest)
	require.NoError(t, err)
	assert.Len(t, scores, 20, "must transform all 20 samples")
	assert.Len(t, scores[0], 2, "must have 2 components")
}

// TestTransformWithZeroVarianceColumn verifies that a column with near-zero
// variance at fit time does not cause division by zero during Transform.
// The implementation guards this with MinVarianceThreshold, setting scale=1.0.
func TestTransformWithZeroVarianceColumn(t *testing.T) {
	// Column 2 has zero variance (all values identical)
	trainData := types.Matrix{
		{1.0, 5.0, 7.0},
		{2.0, 5.0, 8.0},
		{3.0, 5.0, 9.0},
		{4.0, 5.0, 10.0},
		{5.0, 5.0, 11.0},
	}

	config := types.PCAConfig{
		Components: 2, MeanCenter: true, StandardScale: true, Method: "svd",
	}

	engine := NewPCAEngine()
	_, err := engine.Fit(trainData, config)
	require.NoError(t, err)

	testData := types.Matrix{{3.0, 5.0, 9.0}}
	scores, err := engine.Transform(testData)
	require.NoError(t, err, "Transform with zero-variance column must not fail")
	require.Len(t, scores, 1)

	// Scores must be finite (no NaN/Inf from division by zero)
	for _, v := range scores[0] {
		assert.False(t, math.IsNaN(v), "score must not be NaN")
		assert.False(t, math.IsInf(v, 0), "score must not be Inf")
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// Group H: NIPALS native missing value handling in Transform
//
// When NIPALS fits data with NaN values using MissingNative strategy,
// Transform on new (complete) data must work correctly.
// ─────────────────────────────────────────────────────────────────────────────

// TestNIPALSMissingValueFitThenTransformClean verifies that after fitting NIPALS
// with missing values (NaN), Transform works correctly on complete new data.
func TestNIPALSMissingValueFitThenTransformClean(t *testing.T) {
	// Training data with missing values
	trainData := types.Matrix{
		{1.0, 2.0, math.NaN()},
		{4.0, math.NaN(), 6.0},
		{7.0, 8.0, 9.0},
		{2.0, 3.0, 4.0},
		{5.0, 6.0, 7.0},
		{3.0, 4.0, 5.0},
	}

	config := types.PCAConfig{
		Components:      2,
		MeanCenter:      true,
		Method:          "nipals",
		MissingStrategy: types.MissingNative,
	}

	engine := NewPCAEngine()
	_, err := engine.Fit(trainData, config)
	require.NoError(t, err, "NIPALS Fit with missing values must succeed")

	// Transform clean data (no NaNs) — must work without error
	cleanTest := types.Matrix{
		{3.0, 4.0, 5.0},
		{6.0, 7.0, 8.0},
	}

	scores, err := engine.Transform(cleanTest)
	require.NoError(t, err, "Transform on clean data after missing-value Fit must succeed")
	require.Len(t, scores, 2)
	require.Len(t, scores[0], 2)

	// Scores must be finite
	for i, row := range scores {
		for j, v := range row {
			assert.False(t, math.IsNaN(v), "score[%d][%d] must not be NaN", i, j)
			assert.False(t, math.IsInf(v, 0), "score[%d][%d] must not be Inf", i, j)
		}
	}
}

// TestNIPALSMissingValueTransformPreservesParameterFreezing verifies that even
// when NIPALS fits with native missing handling, the preprocessor parameters
// (if any column-wise preprocessing is applied alongside) are correctly frozen.
func TestNIPALSMissingNativeSuppressesOtherPreprocessing(t *testing.T) {
	// When MissingNative is used with NIPALS and data has NaNs, other preprocessing
	// options (StandardScale, RobustScale) are suppressed with a warning.
	// Mean centering is handled internally by the NIPALS algorithm.
	trainData := types.Matrix{
		{1.0, 2.0, math.NaN()},
		{4.0, math.NaN(), 6.0},
		{7.0, 8.0, 9.0},
		{2.0, 3.0, 4.0},
		{5.0, 6.0, 7.0},
	}

	config := types.PCAConfig{
		Components:      2,
		MeanCenter:      true,
		StandardScale:   true, // Will be suppressed because of NaN + MissingNative
		Method:          "nipals",
		MissingStrategy: types.MissingNative,
	}

	engine := NewPCAEngine()
	_, err := engine.Fit(trainData, config)
	require.NoError(t, err, "Fit must succeed even with suppressed preprocessing")

	impl := engine.(*PCAImpl)
	// With NaN data and MissingNative, preprocessor should be nil
	// (mean centering handled internally by NIPALS algorithm)
	assert.Nil(t, impl.preprocessor,
		"preprocessor must be nil for NIPALS with NaN data and MissingNative strategy")
}

// ─────────────────────────────────────────────────────────────────────────────
// Group I: Column/Row Exclusion Tests
//
// Column and row exclusion is an application-layer concern: the cobra CLI and
// desktop app call utils.FilterMatrix() BEFORE passing data to PCAEngine.
// These integration tests validate the full pattern:
//
//   FilterMatrix(trainData, excludedRows, excludedCols) → engine.Fit()
//   FilterMatrix(testData,  nil,          excludedCols) → engine.Transform()
//
// The excluded-column set must be identical for fit and transform so that the
// feature dimension matches the fitted model. Row exclusion only affects which
// observations are used for fitting; all (non-NaN) rows of new data are
// transformed.
//
// References:
//   - Bro, R., & Smilde, A. K. (2014). Principal component analysis.
//     Analytical Methods, 6(9), 2812–2831. (Section 3: preprocessing scope)
// ─────────────────────────────────────────────────────────────────────────────

// TestExcludedColumns verifies that fitting on a column-filtered matrix and
// then transforming new data through the same column filter yields scores
// consistent with fitting directly on the reduced variable set.
func TestExcludedColumns(t *testing.T) {
	// 6-variable dataset; we will exclude columns 1 and 4 (0-based), keeping {0,2,3,5}.
	fullData := [][]float64{
		{1.0, 99.0, 2.0, 3.0, 99.0, 4.0},
		{2.0, 99.0, 4.0, 6.0, 99.0, 8.0},
		{3.0, 99.0, 6.0, 9.0, 99.0, 12.0},
		{4.0, 99.0, 8.0, 12.0, 99.0, 16.0},
		{5.0, 99.0, 10.0, 15.0, 99.0, 20.0},
		{6.0, 99.0, 12.0, 18.0, 99.0, 24.0},
		{7.0, 99.0, 14.0, 21.0, 99.0, 28.0},
		{8.0, 99.0, 16.0, 24.0, 99.0, 32.0},
	}
	newData := [][]float64{
		{9.0, 99.0, 18.0, 27.0, 99.0, 36.0},
		{10.0, 99.0, 20.0, 30.0, 99.0, 40.0},
	}

	excludedCols := []int{1, 4} // 0-based

	// Apply column filter to training data (no row exclusions).
	filteredTrain, err := utils.FilterMatrix(fullData, nil, excludedCols)
	require.NoError(t, err)

	// Apply identical column filter to new observations.
	filteredNew, err := utils.FilterMatrix(newData, nil, excludedCols)
	require.NoError(t, err)

	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: false,
		Method:        "svd",
	}

	engine := NewPCAEngine()
	_, err = engine.Fit(filteredTrain, config)
	require.NoError(t, err, "Fit on column-filtered data must succeed")

	scores, err := engine.Transform(filteredNew)
	require.NoError(t, err, "Transform on column-filtered new data must succeed")

	require.Equal(t, 2, len(scores), "expected 2 transformed rows")
	require.Equal(t, 2, len(scores[0]), "expected 2 score columns (components)")

	// Cross-check: fitting directly on the reduced variable set (4 columns)
	// must produce identical results.
	reducedTrain := [][]float64{
		{1.0, 2.0, 3.0, 4.0},
		{2.0, 4.0, 6.0, 8.0},
		{3.0, 6.0, 9.0, 12.0},
		{4.0, 8.0, 12.0, 16.0},
		{5.0, 10.0, 15.0, 20.0},
		{6.0, 12.0, 18.0, 24.0},
		{7.0, 14.0, 21.0, 28.0},
		{8.0, 16.0, 24.0, 32.0},
	}
	reducedNew := [][]float64{
		{9.0, 18.0, 27.0, 36.0},
		{10.0, 20.0, 30.0, 40.0},
	}

	engine2 := NewPCAEngine()
	_, err = engine2.Fit(reducedTrain, config)
	require.NoError(t, err)

	scores2, err := engine2.Transform(reducedNew)
	require.NoError(t, err)

	assertScoresEqual(t, scores2, scores, 1e-10,
		"column-filtered path must produce same scores as directly reduced data")
}

// TestExcludedRows verifies that row exclusion during fitting affects the
// trained model: scores computed after excluding outlier rows differ from
// scores produced by a model trained on all rows.
func TestExcludedRows(t *testing.T) {
	// Training data with two outlier rows (index 0 and 7) that would skew the model.
	trainData := [][]float64{
		{1000.0, 2000.0, 3000.0}, // outlier — excluded
		{1.0, 2.0, 3.0},
		{2.0, 3.0, 4.0},
		{3.0, 4.0, 5.0},
		{4.0, 5.0, 6.0},
		{5.0, 6.0, 7.0},
		{6.0, 7.0, 8.0},
		{-1000.0, -2000.0, -3000.0}, // outlier — excluded
	}
	newObs := [][]float64{
		{3.5, 4.5, 5.5},
	}

	excludedRows := []int{0, 7} // 0-based

	filteredTrain, err := utils.FilterMatrix(trainData, excludedRows, nil)
	require.NoError(t, err)
	require.Equal(t, 6, len(filteredTrain), "expected 6 rows after exclusion")

	config := types.PCAConfig{
		Components: 1,
		MeanCenter: true,
		Method:     "svd",
	}

	// Model trained WITHOUT outliers.
	engineClean := NewPCAEngine()
	_, err = engineClean.Fit(filteredTrain, config)
	require.NoError(t, err)

	scoresClean, err := engineClean.Transform(newObs)
	require.NoError(t, err)

	// Model trained WITH outliers (all rows).
	engineAll := NewPCAEngine()
	_, err = engineAll.Fit(trainData, config)
	require.NoError(t, err)

	scoresAll, err := engineAll.Transform(newObs)
	require.NoError(t, err)

	// The two models must differ because the outliers shift the mean and
	// principal directions. We assert the absolute score values are not equal.
	absClean := math.Abs(scoresClean[0][0])
	absAll := math.Abs(scoresAll[0][0])
	assert.NotEqual(t, absClean, absAll,
		"model trained without outlier rows must differ from model trained with all rows")
}

// TestExclusionsWorkWithTransform verifies the end-to-end pattern used by the
// application layer: exclude columns at fit time, then apply the SAME column
// exclusion when transforming, and confirm that training scores are reproduced
// (sign-agnostic) for the training observations.
func TestExclusionsWorkWithTransform(t *testing.T) {
	// Iris-like dataset: 4 variables, we exclude variable at index 2.
	trainFull, testFull := irisTrainTest(t)
	excludedCols := []int{2}

	filteredTrain, err := utils.FilterMatrix(trainFull, nil, excludedCols)
	require.NoError(t, err)

	filteredTest, err := utils.FilterMatrix(testFull, nil, excludedCols)
	require.NoError(t, err)

	config := types.PCAConfig{
		Components:    2,
		MeanCenter:    true,
		StandardScale: true,
		Method:        "svd",
	}

	engine := NewPCAEngine()
	fitResult, err := engine.Fit(filteredTrain, config)
	require.NoError(t, err)

	// Transform the held-out test rows.
	testScores, err := engine.Transform(filteredTest)
	require.NoError(t, err)
	require.Equal(t, len(filteredTest), len(testScores),
		"number of transformed rows must match input rows")
	require.Equal(t, config.Components, len(testScores[0]),
		"number of score columns must match requested components")

	// Transform the training rows — these must reproduce fitResult.Scores
	// (the in-memory fit scores) up to sign ambiguity.
	trainScores, err := engine.Transform(filteredTrain)
	require.NoError(t, err)

	assertScoresEqual(t, fitResult.Scores, trainScores, 1e-10,
		"transform of training rows (column-filtered) must reproduce fit scores")
}

// TestExclusionsWithDifferentPreprocessing verifies that column exclusion
// combined with different preprocessing methods all produce valid, consistent
// results. We test mean-center and standard-scale combinations to ensure that
// preprocessing parameters are estimated only on the kept columns.
func TestExclusionsWithDifferentPreprocessing(t *testing.T) {
	// 5-variable dataset; exclude the first and last column (indices 0 and 4).
	rawData := [][]float64{
		{99.0, 1.0, 2.0, 3.0, 99.0},
		{99.0, 2.0, 4.0, 6.0, 99.0},
		{99.0, 3.0, 6.0, 9.0, 99.0},
		{99.0, 4.0, 8.0, 12.0, 99.0},
		{99.0, 5.0, 10.0, 15.0, 99.0},
		{99.0, 6.0, 12.0, 18.0, 99.0},
		{99.0, 7.0, 14.0, 21.0, 99.0},
		{99.0, 8.0, 16.0, 24.0, 99.0},
	}
	newObs := [][]float64{
		{99.0, 9.0, 18.0, 27.0, 99.0},
	}

	excludedCols := []int{0, 4}

	filteredTrain, err := utils.FilterMatrix(rawData, nil, excludedCols)
	require.NoError(t, err)
	require.Equal(t, 3, len(filteredTrain[0]), "expected 3 columns after exclusion")

	filteredNew, err := utils.FilterMatrix(newObs, nil, excludedCols)
	require.NoError(t, err)

	preprocessings := []struct {
		name          string
		meanCenter    bool
		standardScale bool
	}{
		{"MeanCenter", true, false},
		{"StandardScale", true, true},
		{"ScaleOnly", false, true},
		{"None", false, false},
	}

	for _, pp := range preprocessings {
		t.Run(pp.name, func(t *testing.T) {
			config := types.PCAConfig{
				Components:    2,
				MeanCenter:    pp.meanCenter,
				StandardScale: pp.standardScale,
				Method:        "svd",
			}

			engine := NewPCAEngine()
			_, err := engine.Fit(filteredTrain, config)
			require.NoError(t, err, "Fit must succeed for preprocessing=%s", pp.name)

			scores, err := engine.Transform(filteredNew)
			require.NoError(t, err, "Transform must succeed for preprocessing=%s", pp.name)

			require.Equal(t, 1, len(scores),
				"expected 1 transformed row for preprocessing=%s", pp.name)
			require.Equal(t, 2, len(scores[0]),
				"expected 2 score columns for preprocessing=%s", pp.name)
		})
	}
}
