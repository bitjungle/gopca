package core

import (
	"fmt"
	"math"
	"sort"

	"github.com/bitjungle/gopca/pkg/types"
	"gonum.org/v1/gonum/stat"
)

// Preprocessor handles data preprocessing for PCA
type Preprocessor struct {
	// Preprocessing parameters
	MeanCenter    bool
	StandardScale bool
	RobustScale   bool
	SNV           bool
	VectorNorm    bool

	// Fitted parameters
	mean        []float64
	scale       []float64
	originalStd []float64 // Original standard deviations before scaling
	median      []float64
	mad         []float64
	fitted      bool

	// SNV parameters (stored for potential inverse transform)
	rowMeans   []float64
	rowStdDevs []float64
}

// NewPreprocessor creates a new preprocessor instance
func NewPreprocessor(meanCenter, standardScale, robustScale bool) *Preprocessor {
	return &Preprocessor{
		MeanCenter:    meanCenter,
		StandardScale: standardScale,
		RobustScale:   robustScale,
	}
}

// NewPreprocessorWithSNV creates a new preprocessor instance with SNV support
func NewPreprocessorWithSNV(meanCenter, standardScale, robustScale, snv bool) *Preprocessor {
	return &Preprocessor{
		MeanCenter:    meanCenter,
		StandardScale: standardScale,
		RobustScale:   robustScale,
		SNV:           snv,
	}
}

// NewPreprocessorFull creates a new preprocessor instance with all options
func NewPreprocessorFull(meanCenter, standardScale, robustScale, snv, vectorNorm bool) *Preprocessor {
	return &Preprocessor{
		MeanCenter:    meanCenter,
		StandardScale: standardScale,
		RobustScale:   robustScale,
		SNV:           snv,
		VectorNorm:    vectorNorm,
	}
}

// FitTransform fits the preprocessor and transforms the data
func (p *Preprocessor) FitTransform(data types.Matrix) (types.Matrix, error) {
	// If row-wise preprocessing is enabled, we need to fit column statistics on row-normalized data
	if p.SNV || p.VectorNorm {
		// First apply row-wise preprocessing
		dataForFit := make(types.Matrix, len(data))
		for i := range data {
			dataForFit[i] = make([]float64, len(data[i]))
			copy(dataForFit[i], data[i])

			if p.SNV {
				// Apply SNV to this row
				rowMean := stat.Mean(dataForFit[i], nil)
				rowStdDev := stat.StdDev(dataForFit[i], nil)

				if rowStdDev < 1e-8 {
					// Just center if std dev is too small
					for j := range dataForFit[i] {
						dataForFit[i][j] -= rowMean
					}
				} else {
					for j := range dataForFit[i] {
						dataForFit[i][j] = (dataForFit[i][j] - rowMean) / rowStdDev
					}
				}
			} else if p.VectorNorm {
				// Apply L2 normalization
				rowNorm := 0.0
				for j := range dataForFit[i] {
					rowNorm += dataForFit[i][j] * dataForFit[i][j]
				}
				rowNorm = math.Sqrt(rowNorm)

				if rowNorm > 1e-8 {
					for j := range dataForFit[i] {
						dataForFit[i][j] /= rowNorm
					}
				}
			}
		}

		// Fit column statistics on row-normalized data
		if err := p.Fit(dataForFit); err != nil {
			return nil, err
		}
	} else {
		// Standard case: fit on original data
		if err := p.Fit(data); err != nil {
			return nil, err
		}
	}

	return p.Transform(data)
}

// Fit calculates preprocessing parameters from the data
func (p *Preprocessor) Fit(data types.Matrix) error {
	if len(data) == 0 || len(data[0]) == 0 {
		return fmt.Errorf("empty data matrix")
	}

	n, m := len(data), len(data[0])

	// Initialize parameter arrays
	p.mean = make([]float64, m)
	p.scale = make([]float64, m)
	p.originalStd = make([]float64, m)
	p.median = make([]float64, m)
	p.mad = make([]float64, m)

	// Calculate parameters for each feature
	for j := 0; j < m; j++ {
		col := make([]float64, n)
		for i := 0; i < n; i++ {
			col[i] = data[i][j]
		}

		// Mean
		p.mean[j] = stat.Mean(col, nil)

		// Always calculate original standard deviation
		p.originalStd[j] = stat.StdDev(col, nil)

		// Standard deviation for scaling
		if p.StandardScale {
			p.scale[j] = p.originalStd[j]
			if p.scale[j] < 1e-8 {
				p.scale[j] = 1.0 // Avoid division by zero
			}
		} else {
			p.scale[j] = 1.0
		}

		// Robust scaling parameters
		if p.RobustScale {
			// Sort the column data for quantile calculation
			sortedCol := make([]float64, len(col))
			copy(sortedCol, col)
			sort.Float64s(sortedCol)

			p.median[j] = stat.Quantile(0.5, stat.Empirical, sortedCol, nil)
			p.mad[j] = medianAbsoluteDeviation(col, p.median[j])
			if p.mad[j] < 1e-8 {
				p.mad[j] = 1.0 // Avoid division by zero
			}
		}
	}

	p.fitted = true
	return nil
}

// Transform applies the preprocessing to data
func (p *Preprocessor) Transform(data types.Matrix) (types.Matrix, error) {
	if !p.fitted {
		return nil, fmt.Errorf("preprocessor not fitted: call Fit first")
	}

	if len(data) == 0 || len(data[0]) == 0 {
		return nil, fmt.Errorf("empty data matrix")
	}

	n, m := len(data), len(data[0])
	if m != len(p.mean) {
		return nil, fmt.Errorf("data has %d features, expected %d", m, len(p.mean))
	}

	// Create output matrix - start with copy of input data
	result := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		result[i] = make([]float64, m)
		copy(result[i], data[i])
	}

	// Apply row-wise preprocessing first (SNV or Vector Normalization)
	if p.SNV || p.VectorNorm {
		// Initialize storage for row statistics if needed
		if p.rowMeans == nil {
			p.rowMeans = make([]float64, n)
			p.rowStdDevs = make([]float64, n)
		}

		for i := 0; i < n; i++ {
			if p.SNV {
				// Calculate row mean and std dev
				rowMean := stat.Mean(result[i], nil)
				rowStdDev := stat.StdDev(result[i], nil)

				// Store for potential inverse transform
				p.rowMeans[i] = rowMean
				p.rowStdDevs[i] = rowStdDev

				// Apply SNV: (x - row_mean) / row_std
				if rowStdDev < 1e-8 {
					// Handle case where row has near-zero variance
					// Just center the row without scaling
					for j := 0; j < m; j++ {
						result[i][j] -= rowMean
					}
				} else {
					for j := 0; j < m; j++ {
						result[i][j] = (result[i][j] - rowMean) / rowStdDev
					}
				}
			} else if p.VectorNorm {
				// Calculate L2 norm of the row
				rowNorm := 0.0
				for j := 0; j < m; j++ {
					rowNorm += result[i][j] * result[i][j]
				}
				rowNorm = math.Sqrt(rowNorm)

				// Store norm for potential inverse transform (in rowStdDevs for simplicity)
				p.rowStdDevs[i] = rowNorm

				// Apply L2 normalization
				if rowNorm > 1e-8 {
					for j := 0; j < m; j++ {
						result[i][j] /= rowNorm
					}
				}
			}
		}
	}

	// Then apply column-wise preprocessing
	for i := 0; i < n; i++ {
		for j := 0; j < m; j++ {
			val := result[i][j]

			// Apply centering and scaling
			if p.RobustScale {
				// Robust scaling: (x - median) / MAD
				val = (val - p.median[j]) / p.mad[j]
			} else {
				// Standard scaling
				if p.MeanCenter {
					val -= p.mean[j]
				}
				if p.StandardScale {
					val /= p.scale[j]
				}
			}

			result[i][j] = val
		}
	}

	return result, nil
}

// InverseTransform reverses the preprocessing
// Note: When SNV is combined with column-wise preprocessing, the inverse transform
// only reverses the column-wise operations. Full reversal of SNV after column
// preprocessing would require storing the full transformed matrix.
func (p *Preprocessor) InverseTransform(data types.Matrix) (types.Matrix, error) {
	if !p.fitted {
		return nil, fmt.Errorf("preprocessor not fitted")
	}

	n, m := len(data), len(data[0])
	if m != len(p.mean) {
		return nil, fmt.Errorf("data has %d features, expected %d", m, len(p.mean))
	}

	// Create output matrix
	result := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		result[i] = make([]float64, m)
		for j := 0; j < m; j++ {
			val := data[i][j]

			// Reverse scaling and centering (column-wise operations only)
			if p.RobustScale {
				// Reverse robust scaling
				val = val*p.mad[j] + p.median[j]
			} else {
				// Reverse standard scaling
				if p.StandardScale {
					val *= p.scale[j]
				}
				if p.MeanCenter {
					val += p.mean[j]
				}
			}

			result[i][j] = val
		}
	}

	// Note: SNV reversal is not performed when combined with column preprocessing
	// as it would require the intermediate state after SNV but before column operations

	return result, nil
}

// medianAbsoluteDeviation calculates MAD for robust scaling
func medianAbsoluteDeviation(data []float64, median float64) float64 {
	deviations := make([]float64, len(data))
	for i, v := range data {
		deviations[i] = math.Abs(v - median)
	}

	// Sort deviations for quantile calculation
	sort.Float64s(deviations)

	// MAD = median(|x - median(x)|)
	return stat.Quantile(0.5, stat.Empirical, deviations, nil) * 1.4826 // Scale factor for consistency with std dev
}

// HandleMissingValues provides strategies for dealing with missing data
type MissingValueStrategy string

const (
	MissingMean   MissingValueStrategy = "mean"
	MissingMedian MissingValueStrategy = "median"
	MissingZero   MissingValueStrategy = "zero"
	MissingDrop   MissingValueStrategy = "drop"
)

// ImputeMissing handles missing values in the data
func ImputeMissing(data types.Matrix, strategy MissingValueStrategy) (types.Matrix, error) {
	if len(data) == 0 || len(data[0]) == 0 {
		return data, nil
	}

	n, m := len(data), len(data[0])

	switch strategy {
	case MissingDrop:
		// Remove rows with any NaN values
		cleanRows := [][]float64{}
		for i := 0; i < n; i++ {
			hasNaN := false
			for j := 0; j < m; j++ {
				if math.IsNaN(data[i][j]) {
					hasNaN = true
					break
				}
			}
			if !hasNaN {
				cleanRows = append(cleanRows, data[i])
			}
		}
		return cleanRows, nil

	case MissingMean, MissingMedian, MissingZero:
		// Calculate imputation values for each column
		imputeValues := make([]float64, m)

		for j := 0; j < m; j++ {
			validValues := []float64{}
			for i := 0; i < n; i++ {
				if !math.IsNaN(data[i][j]) {
					validValues = append(validValues, data[i][j])
				}
			}

			if len(validValues) == 0 {
				return nil, fmt.Errorf("column %d has all missing values", j)
			}

			switch strategy {
			case MissingMean:
				imputeValues[j] = stat.Mean(validValues, nil)
			case MissingMedian:
				sortedValues := make([]float64, len(validValues))
				copy(sortedValues, validValues)
				sort.Float64s(sortedValues)
				imputeValues[j] = stat.Quantile(0.5, stat.Empirical, sortedValues, nil)
			case MissingZero:
				imputeValues[j] = 0.0
			}
		}

		// Create output with imputed values
		result := make(types.Matrix, n)
		for i := 0; i < n; i++ {
			result[i] = make([]float64, m)
			for j := 0; j < m; j++ {
				if math.IsNaN(data[i][j]) {
					result[i][j] = imputeValues[j]
				} else {
					result[i][j] = data[i][j]
				}
			}
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unknown missing value strategy: %s", strategy)
	}
}

// SelectRowsColumns provides utilities for data subsetting
func SelectRowsColumns(data types.Matrix, rows, cols []int) (types.Matrix, error) {
	if len(data) == 0 || len(data[0]) == 0 {
		return nil, fmt.Errorf("empty data matrix")
	}

	// Validate indices
	for _, r := range rows {
		if r < 0 || r >= len(data) {
			return nil, fmt.Errorf("row index %d out of bounds [0, %d)", r, len(data))
		}
	}

	for _, c := range cols {
		if c < 0 || c >= len(data[0]) {
			return nil, fmt.Errorf("column index %d out of bounds [0, %d)", c, len(data[0]))
		}
	}

	// Create subset
	result := make(types.Matrix, len(rows))
	for i, r := range rows {
		result[i] = make([]float64, len(cols))
		for j, c := range cols {
			result[i][j] = data[r][c]
		}
	}

	return result, nil
}

// RemoveOutliers removes outliers based on z-score
func RemoveOutliers(data types.Matrix, threshold float64) (types.Matrix, []int, error) {
	if len(data) == 0 || len(data[0]) == 0 {
		return data, []int{}, nil
	}

	n, m := len(data), len(data[0])

	// First, calculate mean and std dev for each column
	means := make([]float64, m)
	stdDevs := make([]float64, m)

	for j := 0; j < m; j++ {
		col := make([]float64, n)
		for i := 0; i < n; i++ {
			col[i] = data[i][j]
		}
		means[j] = stat.Mean(col, nil)
		stdDevs[j] = stat.StdDev(col, nil)
	}

	// Now check each row for outliers
	keepRows := []int{}
	for i := 0; i < n; i++ {
		isOutlier := false
		for j := 0; j < m; j++ {
			if stdDevs[j] > 0 {
				zScore := math.Abs((data[i][j] - means[j]) / stdDevs[j])
				if zScore > threshold {
					isOutlier = true
					break
				}
			}
		}

		if !isOutlier {
			keepRows = append(keepRows, i)
		}
	}

	// Create cleaned data
	cleanData := make(types.Matrix, len(keepRows))
	for idx, row := range keepRows {
		cleanData[idx] = data[row]
	}

	return cleanData, keepRows, nil
}

// VariableTransform applies mathematical transformations to variables
type TransformType string

const (
	TransformLog        TransformType = "log"
	TransformSqrt       TransformType = "sqrt"
	TransformSquare     TransformType = "square"
	TransformReciprocal TransformType = "reciprocal"
)

// ApplyTransform applies a mathematical transformation to specified columns
func ApplyTransform(data types.Matrix, columns []int, transform TransformType) (types.Matrix, error) {
	if len(data) == 0 || len(data[0]) == 0 {
		return data, nil
	}

	n, m := len(data), len(data[0])

	// Validate columns
	for _, c := range columns {
		if c < 0 || c >= m {
			return nil, fmt.Errorf("column index %d out of bounds", c)
		}
	}

	// Create transformed data
	result := make(types.Matrix, n)
	for i := 0; i < n; i++ {
		result[i] = make([]float64, m)
		copy(result[i], data[i])
	}

	// Apply transformation to specified columns
	for _, col := range columns {
		for i := 0; i < n; i++ {
			val := result[i][col]

			switch transform {
			case TransformLog:
				if val <= 0 {
					return nil, fmt.Errorf("cannot take log of non-positive value at [%d,%d]", i, col)
				}
				result[i][col] = math.Log(val)

			case TransformSqrt:
				if val < 0 {
					return nil, fmt.Errorf("cannot take sqrt of negative value at [%d,%d]", i, col)
				}
				result[i][col] = math.Sqrt(val)

			case TransformSquare:
				result[i][col] = val * val

			case TransformReciprocal:
				if math.Abs(val) < 1e-10 {
					return nil, fmt.Errorf("cannot take reciprocal of zero at [%d,%d]", i, col)
				}
				result[i][col] = 1.0 / val

			default:
				return nil, fmt.Errorf("unknown transform type: %s", transform)
			}
		}
	}

	return result, nil
}

// GetMeans returns the fitted mean values
func (p *Preprocessor) GetMeans() []float64 {
	if !p.fitted {
		return nil
	}
	return p.mean
}

// GetStdDevs returns the fitted standard deviation values (original, before scaling)
func (p *Preprocessor) GetStdDevs() []float64 {
	if !p.fitted {
		return nil
	}
	return p.originalStd
}

// IsSNVEnabled returns whether SNV preprocessing is enabled
func (p *Preprocessor) IsSNVEnabled() bool {
	return p.SNV
}

// GetVarianceByColumn calculates variance for each column
func GetVarianceByColumn(data types.Matrix) ([]float64, error) {
	if len(data) == 0 || len(data[0]) == 0 {
		return nil, fmt.Errorf("empty data matrix")
	}

	m := len(data[0])
	variances := make([]float64, m)

	for j := 0; j < m; j++ {
		col := make([]float64, len(data))
		for i := 0; i < len(data); i++ {
			col[i] = data[i][j]
		}
		variances[j] = stat.Variance(col, nil)
	}

	return variances, nil
}

// GetColumnRanks returns column indices sorted by variance (descending)
func GetColumnRanks(data types.Matrix) ([]int, error) {
	variances, err := GetVarianceByColumn(data)
	if err != nil {
		return nil, err
	}

	// Create index-variance pairs
	type varPair struct {
		index    int
		variance float64
	}

	pairs := make([]varPair, len(variances))
	for i, v := range variances {
		pairs[i] = varPair{index: i, variance: v}
	}

	// Sort by variance (descending)
	sort.Slice(pairs, func(i, j int) bool {
		return pairs[i].variance > pairs[j].variance
	})

	// Extract sorted indices
	ranks := make([]int, len(pairs))
	for i, p := range pairs {
		ranks[i] = p.index
	}

	return ranks, nil
}
