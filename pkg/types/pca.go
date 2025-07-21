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
)

// PCAConfig holds configuration for PCA analysis
type PCAConfig struct {
	Components      int     `json:"components"`
	MeanCenter      bool    `json:"mean_center"`
	StandardScale   bool    `json:"standard_scale"`
	Method          string  `json:"method"` // "svd", "eigen", "nipals", or "kernel"
	ExcludedRows    []int   `json:"excluded_rows,omitempty"`    // 0-based indices of rows to exclude
	ExcludedColumns []int   `json:"excluded_columns,omitempty"` // 0-based indices of columns to exclude
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
}

// PCAEngine defines the interface for PCA computation
type PCAEngine interface {
	Fit(data Matrix, config PCAConfig) (*PCAResult, error)
	Transform(data Matrix) (Matrix, error)
	FitTransform(data Matrix, config PCAConfig) (*PCAResult, error)
}

// PCAOutputData represents complete PCA results for output
type PCAOutputData struct {
	Samples  SampleData  `json:"samples"`
	Features FeatureData `json:"features"`
	Metadata PCAMetadata `json:"metadata"`
}

// SampleData contains sample-space results
type SampleData struct {
	Names   []string       `json:"names"`   // Sample names from input
	Scores  Matrix         `json:"scores"`  // PC scores (n × c)
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
	NSamples            int       `json:"n_samples"`
	NFeatures           int       `json:"n_features"`
	NComponents         int       `json:"n_components"`
	Method              string    `json:"method"`
	Preprocessing       string    `json:"preprocessing"`
	ExplainedVariance   []float64 `json:"explained_variance"`
	CumulativeVariance  []float64 `json:"cumulative_variance"`
}
