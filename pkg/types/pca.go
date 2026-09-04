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
	// MissingZero replaces missing values with zero
	MissingZero MissingValueStrategy = "zero"
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
	// Temporal PCA specific parameters
	TemporalLags      int     `json:"temporal_lags,omitempty"`      // Number of time lags for temporal PCA
	VarianceExplained float64 `json:"variance_explained,omitempty"` // Target explained variance (0.0-1.0)
	ImputeMethod      string  `json:"impute_method,omitempty"`      // "forward", "backward", "linear", "none"
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
	// VariableCorrelations holds the Pearson correlation between each
	// preprocessed variable and each retained component, as [variables][components].
	//
	// This is what a Circle of Correlations plots. It is NOT the loadings: the two
	// differ by a factor of sqrt(eigenvalue)/sd, so on standardised data the
	// correlations are systematically the larger of the two.
	//
	// Squared and summed over a variable's row, these give its communality in the
	// components retained — the share of that variable's variance those components
	// capture. The total reaches 1 only when the full basis is present; this matrix
	// holds the components actually computed, so for a truncated fit the row sums
	// to less than 1 and should not be read as an identity.
	//
	// That is exactly what makes the unit circle a benchmark rather than a bound
	// that is always met: an arrow of length 0.9 in a two-component plot says 81%
	// of that variable's variance lives in those two components.
	//
	// Empty when the engine has no preprocessed matrix to correlate against, which
	// is the case for kernel PCA and for NIPALS with native missing values.
	VariableCorrelations Matrix `json:"variable_correlations,omitempty"`

	// PreprocessedData is the exact matrix the engine operated on internally —
	// the data preprocessed with the same settings used for fitting, in the same
	// space as the reconstruction (scores·loadingsᵀ). Diagnostic metrics (Q/T²)
	// must be computed against this matrix. It is populated by Fit for linear PCA
	// (svd/nipals) and left nil for kernel/temporal PCA, where per-sample
	// reconstruction metrics do not apply. In-memory only; never serialized.
	PreprocessedData Matrix `json:"-"`
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
	// Additional fields for temporal PCA and other methods
	SingularValues []float64 `json:"singular_values,omitempty"` // Singular values from SVD
	// Temporal PCA specific fields
	TemporalEigenvectors       Matrix `json:"temporal_eigenvectors,omitempty"`        // Dominant-channel V-matrix loadings (lags × components): signed temporal shape of each component's most influential channel
	TemporalVariableImportance Matrix `json:"temporal_variable_importance,omitempty"` // Aggregated variable importance (components × variables)
	// Kernel PCA specific fields
	KernelType         string             `json:"kernel_type,omitempty"`         // Type of kernel used (rbf, linear, poly)
	KernelParams       map[string]float64 `json:"kernel_params,omitempty"`       // Kernel parameters (gamma, degree, coef0)
	KernelMatrix       Matrix             `json:"kernel_matrix,omitempty"`       // Kernel matrix (optional, for visualization)
	KernelEigenvectors Matrix             `json:"kernel_eigenvectors,omitempty"` // Eigenvectors for contribution analysis
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

	// Regression is present when the model also predicts a response, which makes
	// it a principal component regression model rather than a plain decomposition.
	// The field is optional and additive: a model without it is exactly what
	// earlier versions produced, and a model with it still satisfies the v1
	// schema, so the two remain interchangeable.
	Regression *RegressionModel `json:"regression,omitempty"`
}

// RegressionModel is the regression half of a principal component regression
// model: everything needed to turn component scores into a prediction.
//
// The decomposition it sits beside is the other half. Stored together, the pair
// is sufficient to predict from raw predictor values without the training data.
type RegressionModel struct {
	// Response names the column this model predicts.
	Response string `json:"response"`

	// Components is the number of leading components the regression uses. It may
	// be fewer than the decomposition carries, since the extra components are kept
	// for the variance profile.
	Components int `json:"components"`

	// ScoreCoefficients maps component scores to the response, one per retained
	// component. Their signs follow the component signs and so carry no meaning
	// on their own.
	ScoreCoefficients []float64 `json:"score_coefficients"`
	Intercept         float64   `json:"intercept"`

	// Coefficients and InterceptOriginal give the collapsed form
	// y = InterceptOriginal + x . Coefficients, which is what a downstream
	// consumer needs to predict without reimplementing the pipeline.
	//
	// OriginalScaleValid says whether that form exists. It is false under
	// row-wise preprocessing: SNV and vector normalization scale each sample by a
	// statistic of that same sample, so no fixed coefficient vector reproduces
	// their effect. Prediction is still possible through the full pipeline.
	//
	// Coefficients is omitted when there is no collapsed form, since a nil slice
	// is unambiguously absent. InterceptOriginal is always written, even when it
	// is zero: an intercept of exactly zero is a legitimate value, and omitting it
	// would be indistinguishable from a model that failed to record one.
	// OriginalScaleValid, not the presence of a field, is what says whether to use
	// the collapsed form.
	Coefficients       []float64 `json:"coefficients,omitempty"`
	InterceptOriginal  float64   `json:"intercept_original"`
	OriginalScaleValid bool      `json:"original_scale_valid"`

	ResponseMean float64 `json:"response_mean"`

	// RMSEC is the root mean square error of the training residuals. It describes
	// the fit and is not an estimate of future performance. Validation carries the
	// held-out figure.
	RMSEC float64 `json:"rmsec"`
	R2C   float64 `json:"r2c"`

	// Validation records how the component count was chosen, so a reported error
	// can be traced to the design that produced it.
	Validation *CVReport `json:"validation,omitempty"`
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
	Categorical map[string][]string `json:"categorical,omitempty"`

	// NumericTarget holds the #target columns, which routinely contain gaps: a
	// calibration set where only part of the samples were sent for reference
	// analysis is the normal case in chemometrics, not an edge case.
	//
	// It must be JSONFloat64, not float64. encoding/json refuses to marshal a
	// NaN at all, so with plain float64 a single unmeasured target made the whole
	// export fail — `pca analyze -f json` and `pca regress -o` both returned
	// "json: unsupported value: NaN" and wrote nothing, on exactly the datasets
	// this software exists to model. JSONFloat64 writes null and reads it back as
	// NaN, which is the round trip the rest of the output types already use.
	NumericTarget map[string][]JSONFloat64 `json:"numericTarget,omitempty"`
}
