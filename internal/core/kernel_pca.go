package core

import (
	"fmt"
	"math"
	"sort"

	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/mat"
)

// KernelType represents the type of kernel function to use
type KernelType string

const (
	// KernelRBF is the Radial Basis Function (Gaussian) kernel
	KernelRBF KernelType = "rbf"
	// KernelLinear is the linear kernel (equivalent to standard PCA)
	KernelLinear KernelType = "linear"
	// KernelPoly is the polynomial kernel
	KernelPoly KernelType = "poly"
)

// KernelPCAImpl implements the PCAEngine interface for Kernel PCA
type KernelPCAImpl struct {
	config       types.PCAConfig
	kernelType   KernelType
	eigvals      []float64
	eigvecs      *mat.Dense
	trainingData types.Matrix
	fitted       bool
	// Precomputed values for centering
	trainKernelMeans []float64
	totalKernelMean  float64
	// Preprocessor for variance scaling
	preprocessor *Preprocessor
}

// NewKernelPCAEngine creates a new Kernel PCA engine
func NewKernelPCAEngine() types.PCAEngine {
	return &KernelPCAImpl{}
}

// validateKernelConfig validates kernel-specific configuration
func (kpca *KernelPCAImpl) validateKernelConfig(config types.PCAConfig) error {
	if config.KernelType == "" {
		return fmt.Errorf("kernel type must be specified for kernel PCA")
	}

	switch KernelType(config.KernelType) {
	case KernelRBF:
		if config.KernelGamma < 0 {
			return fmt.Errorf("gamma must be non-negative for RBF kernel")
		}
	case KernelPoly:
		if config.KernelGamma < 0 {
			return fmt.Errorf("gamma must be non-negative for polynomial kernel")
		}
		if config.KernelDegree < 1 {
			return fmt.Errorf("degree must be at least 1 for polynomial kernel")
		}
	case KernelLinear:
		// No specific validation needed
	default:
		return fmt.Errorf("unsupported kernel type: %s", config.KernelType)
	}

	return nil
}

// computeKernel computes the kernel function between two vectors
func (kpca *KernelPCAImpl) computeKernel(x, y []float64) (float64, error) {
	if len(x) != len(y) {
		return 0, fmt.Errorf("vectors must have the same length")
	}

	switch kpca.kernelType {
	case KernelRBF:
		sum := 0.0
		for i := range x {
			diff := x[i] - y[i]
			sum += diff * diff
		}
		return math.Exp(-kpca.config.KernelGamma * sum), nil

	case KernelLinear:
		sum := 0.0
		for i := range x {
			sum += x[i] * y[i]
		}
		return sum, nil

	case KernelPoly:
		sum := 0.0
		for i := range x {
			sum += x[i] * y[i]
		}
		return math.Pow(kpca.config.KernelGamma*sum+kpca.config.KernelCoef0, float64(kpca.config.KernelDegree)), nil

	default:
		return 0, fmt.Errorf("unsupported kernel type: %s", kpca.kernelType)
	}
}

// computeKernelMatrix computes the full kernel matrix
func (kpca *KernelPCAImpl) computeKernelMatrix(data types.Matrix) (*mat.Dense, error) {
	n := len(data)
	K := mat.NewDense(n, n, nil)

	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			val, err := kpca.computeKernel(data[i], data[j])
			if err != nil {
				return nil, fmt.Errorf("error computing kernel at (%d, %d): %w", i, j, err)
			}
			K.Set(i, j, val)
			if i != j {
				K.Set(j, i, val) // Kernel matrix is symmetric
			}
		}
	}

	return K, nil
}

// centerKernelMatrix centers the kernel matrix
func (kpca *KernelPCAImpl) centerKernelMatrix(K *mat.Dense) (*mat.Dense, error) {
	n, _ := K.Dims()

	// Compute row and column means
	rowMeans := make([]float64, n)
	colMeans := make([]float64, n)
	totalMean := 0.0

	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			val := K.At(i, j)
			rowMeans[i] += val
			colMeans[j] += val
			totalMean += val
		}
	}

	for i := 0; i < n; i++ {
		rowMeans[i] /= float64(n)
		colMeans[i] /= float64(n)
	}
	totalMean /= float64(n * n)

	// Store for transform method
	kpca.trainKernelMeans = colMeans
	kpca.totalKernelMean = totalMean

	// Center the kernel matrix
	Kc := mat.NewDense(n, n, nil)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			val := K.At(i, j) - rowMeans[i] - colMeans[j] + totalMean
			Kc.Set(i, j, val)
		}
	}

	return Kc, nil
}

// eigenDecomposition performs eigendecomposition and returns top k components
func (kpca *KernelPCAImpl) eigenDecomposition(K *mat.Dense, k int) ([]float64, *mat.Dense, error) {
	// Convert to symmetric matrix for eigendecomposition
	n, _ := K.Dims()
	symK := mat.NewSymDense(n, nil)
	for i := 0; i < n; i++ {
		for j := i; j < n; j++ {
			symK.SetSym(i, j, K.At(i, j))
		}
	}

	var eig mat.EigenSym
	if ok := eig.Factorize(symK, true); !ok {
		return nil, nil, fmt.Errorf("eigendecomposition failed")
	}

	vals := eig.Values(nil)
	var vecs mat.Dense
	eig.VectorsTo(&vecs)

	// Sort by descending eigenvalue
	nVals := len(vals)
	idx := make([]int, nVals)
	for i := range idx {
		idx[i] = i
	}
	sort.Slice(idx, func(i, j int) bool {
		return vals[idx[i]] > vals[idx[j]]
	})

	// Extract top k components
	if k > nVals {
		k = nVals
	}

	sortedVals := make([]float64, k)
	sortedVecs := mat.NewDense(nVals, k, nil)

	for i := 0; i < k; i++ {
		sortedVals[i] = vals[idx[i]]
		// Handle near-zero or negative eigenvalues
		if sortedVals[i] < 1e-10 {
			sortedVals[i] = 1e-10
		}
		for j := 0; j < nVals; j++ {
			sortedVecs.Set(j, i, vecs.At(j, idx[i]))
		}
	}

	return sortedVals, sortedVecs, nil
}

// Fit trains the Kernel PCA model on the provided data
func (kpca *KernelPCAImpl) Fit(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	// Validate configuration
	if err := kpca.validateKernelConfig(config); err != nil {
		return nil, fmt.Errorf("invalid kernel configuration: %w", err)
	}

	kpca.kernelType = KernelType(config.KernelType)

	// Validate data
	if len(data) == 0 || len(data[0]) == 0 {
		return nil, fmt.Errorf("empty data matrix")
	}

	nSamples := len(data)
	nFeatures := len(data[0])

	if config.Components > nSamples {
		return nil, fmt.Errorf("number of components (%d) cannot exceed number of samples (%d)",
			config.Components, nSamples)
	}

	// Set default gamma to 1/n_features if not specified (for RBF and polynomial kernels)
	if config.KernelGamma == 0 && (KernelType(config.KernelType) == KernelRBF || KernelType(config.KernelType) == KernelPoly) {
		config.KernelGamma = 1.0 / float64(nFeatures)
	}

	// Store the configuration after setting defaults
	kpca.config = config

	// Apply preprocessing if needed (only variance scaling, SNV, or vector norm for kernel PCA)
	processedData := data
	if config.ScaleOnly || config.SNV || config.VectorNorm {
		// Create preprocessor with only the allowed preprocessing options
		kpca.preprocessor = NewPreprocessorWithScaleOnly(
			false,             // no mean centering for kernel PCA
			false,             // no standard scale (includes centering)
			false,             // no robust scale (includes centering)
			config.ScaleOnly,  // variance scaling allowed
			config.SNV,        // SNV allowed
			config.VectorNorm, // vector norm allowed
		)

		// Fit and transform
		var err error
		processedData, err = kpca.preprocessor.FitTransform(data)
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}
	}

	// Store preprocessed training data for transform
	kpca.trainingData = make(types.Matrix, nSamples)
	for i := range processedData {
		kpca.trainingData[i] = make([]float64, nFeatures)
		copy(kpca.trainingData[i], processedData[i])
	}

	// Compute kernel matrix using preprocessed data
	K, err := kpca.computeKernelMatrix(processedData)
	if err != nil {
		return nil, fmt.Errorf("error computing kernel matrix: %w", err)
	}

	// Center kernel matrix
	Kc, err := kpca.centerKernelMatrix(K)
	if err != nil {
		return nil, fmt.Errorf("error centering kernel matrix: %w", err)
	}

	// Perform eigendecomposition
	eigvals, eigvecs, err := kpca.eigenDecomposition(Kc, config.Components)
	if err != nil {
		return nil, fmt.Errorf("error in eigendecomposition: %w", err)
	}

	kpca.eigvals = eigvals
	kpca.eigvecs = eigvecs
	kpca.fitted = true

	// Compute projections for training data
	scores := mat.NewDense(nSamples, config.Components, nil)
	for i := 0; i < config.Components; i++ {
		norm := math.Sqrt(eigvals[i])
		for j := 0; j < nSamples; j++ {
			scores.Set(j, i, eigvecs.At(j, i)/norm)
		}
	}

	// Convert scores to Matrix type
	scoresMatrix := make(types.Matrix, nSamples)
	for i := 0; i < nSamples; i++ {
		scoresMatrix[i] = make([]float64, config.Components)
		for j := 0; j < config.Components; j++ {
			scoresMatrix[i][j] = scores.At(i, j)
		}
	}

	// Calculate explained variance
	totalVar := 0.0
	for _, v := range eigvals {
		totalVar += v
	}

	explainedVar := make([]float64, config.Components)
	explainedVarRatio := make([]float64, config.Components)
	cumulativeVar := make([]float64, config.Components)
	cumSum := 0.0

	for i := 0; i < config.Components; i++ {
		explainedVar[i] = eigvals[i]
		explainedVarRatio[i] = eigvals[i] / totalVar * 100
		cumSum += explainedVarRatio[i]
		cumulativeVar[i] = cumSum
	}

	// Note: Loadings are not meaningful for Kernel PCA
	// We'll return an empty matrix for compatibility
	loadings := make(types.Matrix, 0)

	return &types.PCAResult{
		Scores:               scoresMatrix,
		Loadings:             loadings,
		ExplainedVar:         explainedVar,
		ExplainedVarRatio:    explainedVarRatio,
		CumulativeVar:        cumulativeVar,
		ComponentsComputed:   config.Components,
		Method:               "kernel",
		PreprocessingApplied: config.ScaleOnly || config.SNV || config.VectorNorm,
	}, nil
}

// Transform projects new data into the kernel PCA space
func (kpca *KernelPCAImpl) Transform(data types.Matrix) (types.Matrix, error) {
	if !kpca.fitted {
		return nil, fmt.Errorf("model must be fitted before transform")
	}

	// Apply the same preprocessing as during fit
	processedData := data
	if kpca.preprocessor != nil {
		var err error
		processedData, err = kpca.preprocessor.Transform(data)
		if err != nil {
			return nil, fmt.Errorf("preprocessing failed: %w", err)
		}
	}

	nTest := len(processedData)
	nTrain := len(kpca.trainingData)
	nComponents := kpca.config.Components

	// Compute kernel matrix between test and training data
	K := mat.NewDense(nTest, nTrain, nil)
	for i := 0; i < nTest; i++ {
		for j := 0; j < nTrain; j++ {
			val, err := kpca.computeKernel(processedData[i], kpca.trainingData[j])
			if err != nil {
				return nil, fmt.Errorf("error computing kernel at (%d, %d): %w", i, j, err)
			}
			K.Set(i, j, val)
		}
	}

	// Center the test kernel matrix using training statistics
	for i := 0; i < nTest; i++ {
		rowMean := 0.0
		for j := 0; j < nTrain; j++ {
			rowMean += K.At(i, j)
		}
		rowMean /= float64(nTrain)

		for j := 0; j < nTrain; j++ {
			val := K.At(i, j) - rowMean - kpca.trainKernelMeans[j] + kpca.totalKernelMean
			K.Set(i, j, val)
		}
	}

	// Project onto eigenvectors
	result := make(types.Matrix, nTest)
	for i := 0; i < nTest; i++ {
		result[i] = make([]float64, nComponents)
		for j := 0; j < nComponents; j++ {
			sum := 0.0
			norm := math.Sqrt(kpca.eigvals[j])
			for k := 0; k < nTrain; k++ {
				sum += K.At(i, k) * kpca.eigvecs.At(k, j)
			}
			result[i][j] = sum / norm
		}
	}

	return result, nil
}

// FitTransform fits the model and transforms the data in one step
func (kpca *KernelPCAImpl) FitTransform(data types.Matrix, config types.PCAConfig) (*types.PCAResult, error) {
	result, err := kpca.Fit(data, config)
	if err != nil {
		return nil, err
	}
	return result, nil
}
