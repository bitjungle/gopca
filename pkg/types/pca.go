// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package types

// Matrix represents a 2D data matrix
type Matrix [][]float64

// MissingValueStrategy defines how to handle missing values
type MissingValueStrategy string

const (
	// MissingError returns an error when missing values are found
	MissingError MissingValueStrategy = "error"
	// MissingDrop removes rows containing missing values
	MissingDrop MissingValueStrategy = "drop"
	// MissingMean replaces missing values with column mean
	MissingMean MissingValueStrategy = "mean"
	// MissingMedian replaces missing values with column median
	MissingMedian MissingValueStrategy = "median"
	// MissingNative allows NIPALS to handle missing values natively (NIPALS only)
	MissingNative MissingValueStrategy = "native"
)

// PCAConfig holds configuration for PCA analysis
type PCAConfig struct {
	Components      int    `json:"components"`
	MeanCenter      bool   `json:"mean_center"`
	StandardScale   bool   `json:"standard_scale"`
	RobustScale     bool   `json:"robust_scale"`               // Robust scaling (median/MAD)
	ScaleOnly       bool   `json:"scale_only"`                 // Variance scaling: divide by std dev without mean centering
	SNV             bool   `json:"snv"`                        // Standard Normal Variate (row-wise normalization)
	VectorNorm      bool   `json:"vector_norm"`                // L2 normalization (row-wise)
	Method          string `json:"method"`                     // "svd", "eigen", "nipals", or "kernel"
	ExcludedRows    []int  `json:"excluded_rows,omitempty"`    // 0-based indices of rows to exclude
	ExcludedColumns []int  `json:"excluded_columns,omitempty"` // 0-based indices of columns to exclude
	// Missing value handling
	MissingStrategy MissingValueStrategy `json:"missing_strategy,omitempty"` // How to handle missing values
	// Kernel PCA specific parameters
	KernelType   string  `json:"kernel_type,omitempty"`   // "rbf", "linear", "poly"
	KernelGamma  float64 `json:"kernel_gamma,omitempty"`  // RBF/Poly parameter
	KernelDegree int     `json:"kernel_degree,omitempty"` // Poly parameter
	KernelCoef0  float64 `json:"kernel_coef0,omitempty"`  // Poly parameter
}

// PCAResult contains the results of PCA analysis
type PCAResult struct {
	Scores               Matrix    `json:"scores"`
	Loadings             Matrix    `json:"loadings"`
	ExplainedVar         []float64 `json:"explained_variance"`
	ExplainedVarRatio    []float64 `json:"explained_variance_ratio"` // Percentage of variance explained
	CumulativeVar        []float64 `json:"cumulative_variance"`
	ComponentLabels      []string  `json:"component_labels"`
	VariableLabels       []string  `json:"variable_labels,omitempty"` // Original variable names
	ComponentsComputed   int       `json:"components_computed"`       // Number of components actually computed
	Method               string    `json:"method"`                    // Method used (svd, nipals, kernel)
	PreprocessingApplied bool      `json:"preprocessing_applied"`     // Whether preprocessing was applied
	// Preprocessing statistics
	Means   []float64 `json:"means,omitempty"`   // Original feature means
	StdDevs []float64 `json:"stddevs,omitempty"` // Original feature std devs
	// Diagnostic metrics
	Metrics []SampleMetrics `json:"metrics,omitempty"` // Per-sample diagnostic metrics
	// Confidence limits for diagnostics
	T2Limit95 float64 `json:"t2_limit_95,omitempty"` // 95% confidence limit for T²
	T2Limit99 float64 `json:"t2_limit_99,omitempty"` // 99% confidence limit for T²
	QLimit95  float64 `json:"q_limit_95,omitempty"`  // 95% confidence limit for Q-residuals
	QLimit99  float64 `json:"q_limit_99,omitempty"`  // 99% confidence limit for Q-residuals
	// Eigencorrelations with metadata
	Eigencorrelations *EigencorrelationResult `json:"eigencorrelations,omitempty"`
	// All eigenvalues (including non-retained) for diagnostic calculations
	AllEigenvalues []float64 `json:"all_eigenvalues,omitempty"`
}

// EigencorrelationResult contains correlations between PC scores and metadata variables
type EigencorrelationResult struct {
	Correlations map[string][]float64 `json:"correlations"` // Variable name -> correlations with each PC
	PValues      map[string][]float64 `json:"pValues"`      // Variable name -> p-values
	Variables    []string             `json:"variables"`    // Order of variables
	Components   []string             `json:"components"`   // PC labels
	Method       string               `json:"method"`       // Correlation method used
}

// PCAEngine defines the interface for PCA computation
type PCAEngine interface {
	Fit(data Matrix, config PCAConfig) (*PCAResult, error)
	Transform(data Matrix) (Matrix, error)
	FitTransform(data Matrix, config PCAConfig) (*PCAResult, error)
}

// PCAOutputData represents complete PCA results for output
type PCAOutputData struct {
	Schema            string                  `json:"$schema,omitempty"`
	Metadata          ModelMetadata           `json:"metadata"`
	Preprocessing     PreprocessingInfo       `json:"preprocessing"`
	Model             ModelComponents         `json:"model"`
	Results           ResultsData             `json:"results"`
	Diagnostics       DiagnosticLimits        `json:"diagnostics,omitempty"`
	Eigencorrelations *EigencorrelationResult `json:"eigencorrelations,omitempty"`
	PreservedColumns  *PreservedColumns       `json:"preservedColumns,omitempty"`
}

// SampleData contains sample-space results
type SampleData struct {
	Names   []string        `json:"names"`   // Sample names from input
	Scores  Matrix          `json:"scores"`  // PC scores (n × c)
	Metrics []SampleMetrics `json:"metrics"` // Advanced metrics per sample
}

// FeatureData contains feature-space results
type FeatureData struct {
	Names    []string  `json:"names"`    // Feature names from input
	Loadings Matrix    `json:"loadings"` // Loadings (c × k)
	Means    []float64 `json:"means"`    // Original means (k)
	StdDevs  []float64 `json:"stddevs"`  // Original std devs (k)
}

// SampleMetrics contains advanced metrics for a sample
type SampleMetrics struct {
	HotellingT2 float64 `json:"hotelling_t2"`
	Mahalanobis float64 `json:"mahalanobis"`
	RSS         float64 `json:"rss"`
	IsOutlier   bool    `json:"is_outlier"`
}

// PCAMetadata contains analysis metadata
type PCAMetadata struct {
	NSamples           int       `json:"n_samples"`
	NFeatures          int       `json:"n_features"`
	NComponents        int       `json:"n_components"`
	Method             string    `json:"method"`
	Preprocessing      string    `json:"preprocessing"`
	ExplainedVariance  []float64 `json:"explained_variance"`
	CumulativeVariance []float64 `json:"cumulative_variance"`
}

// ModelMetadata contains metadata about the model and analysis
type ModelMetadata struct {
	AnalysisID      string      `json:"analysis_id"` // Unique identifier for this analysis
	SoftwareVersion string      `json:"software_version"`
	CreatedAt       string      `json:"created_at"`
	Software        string      `json:"software"`
	Config          ModelConfig `json:"config"`
	DataSource      *DataSource `json:"data_source,omitempty"` // Information about input data
	Description     string      `json:"description,omitempty"` // User-provided description
	Tags            []string    `json:"tags,omitempty"`        // User-defined tags for categorization
}

// DataSource contains information about the input data file
type DataSource struct {
	Filename      string `json:"filename"`                  // Original data file name
	Hash          string `json:"hash,omitempty"`            // SHA-256 hash of input data
	NRowsOriginal int    `json:"n_rows_original,omitempty"` // Number of rows before exclusions
	NColsOriginal int    `json:"n_cols_original,omitempty"` // Number of columns before exclusions
}

// ModelConfig contains the configuration used for PCA
type ModelConfig struct {
	Method          string               `json:"method"`
	NComponents     int                  `json:"n_components"`
	MissingStrategy MissingValueStrategy `json:"missing_strategy"`
	ExcludedRows    []int                `json:"excluded_rows,omitempty"`
	ExcludedColumns []int                `json:"excluded_columns,omitempty"`
	// Kernel PCA parameters
	KernelType   string  `json:"kernel_type,omitempty"`
	KernelGamma  float64 `json:"kernel_gamma,omitempty"`
	KernelDegree int     `json:"kernel_degree,omitempty"`
	KernelCoef0  float64 `json:"kernel_coef0,omitempty"`
}

// PreprocessingInfo contains all preprocessing configuration and parameters
type PreprocessingInfo struct {
	MeanCenter    bool                `json:"mean_center"`
	StandardScale bool                `json:"standard_scale"`
	RobustScale   bool                `json:"robust_scale"`
	ScaleOnly     bool                `json:"scale_only"`
	SNV           bool                `json:"snv"`
	VectorNorm    bool                `json:"vector_norm"`
	Parameters    PreprocessingParams `json:"parameters"`
}

// PreprocessingParams contains the fitted preprocessing parameters
type PreprocessingParams struct {
	FeatureMeans   []float64 `json:"feature_means,omitempty"`
	FeatureStdDevs []float64 `json:"feature_stddevs,omitempty"`
	FeatureMedians []float64 `json:"feature_medians,omitempty"`
	FeatureMADs    []float64 `json:"feature_mads,omitempty"`
	RowMeans       []float64 `json:"row_means,omitempty"`
	RowStdDevs     []float64 `json:"row_stddevs,omitempty"`
}

// ModelComponents contains the core PCA model components
type ModelComponents struct {
	Loadings               Matrix    `json:"loadings"`
	ExplainedVariance      []float64 `json:"explained_variance"`
	ExplainedVarianceRatio []float64 `json:"explained_variance_ratio"`
	CumulativeVariance     []float64 `json:"cumulative_variance"`
	ComponentLabels        []string  `json:"component_labels"`
	FeatureLabels          []string  `json:"feature_labels"`
}

// ResultsData contains the results of the PCA analysis
type ResultsData struct {
	Samples SamplesResults `json:"samples"`
}

// SamplesResults contains sample-specific results
type SamplesResults struct {
	Names   []string     `json:"names"`
	Scores  Matrix       `json:"scores"`
	Metrics *MetricsData `json:"metrics,omitempty"`
}

// MetricsData contains diagnostic metrics for samples
type MetricsData struct {
	HotellingT2 []float64 `json:"hotelling_t2"`
	Mahalanobis []float64 `json:"mahalanobis"`
	RSS         []float64 `json:"rss"`
	IsOutlier   []bool    `json:"is_outlier"`
}

// DiagnosticLimits contains statistical limits for diagnostics
type DiagnosticLimits struct {
	T2Limit95 float64 `json:"t2_limit_95,omitempty"`
	T2Limit99 float64 `json:"t2_limit_99,omitempty"`
	QLimit95  float64 `json:"q_limit_95,omitempty"`
	QLimit99  float64 `json:"q_limit_99,omitempty"`
}

// PreservedColumns contains columns that were excluded from PCA but preserved in output
type PreservedColumns struct {
	Categorical   map[string][]string  `json:"categorical,omitempty"`
	NumericTarget map[string][]float64 `json:"numericTarget,omitempty"`
}
