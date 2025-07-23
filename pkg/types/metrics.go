package types

// PCAMetrics contains comprehensive diagnostic metrics for PCA model evaluation
type PCAMetrics struct {
	// MahalanobisDistances contains the Mahalanobis distance for each observation
	// in the transformed PC space, measuring multivariate distance from the mean
	MahalanobisDistances []float64

	// HotellingT2 contains Hotelling's T-squared statistic for each observation,
	// which is used for multivariate outlier detection
	HotellingT2 []float64

	// RSS (Residual Sum of Squares) contains the reconstruction error
	// for each observation when using the specified number of components
	RSS []float64

	// QResiduals contains the Q-statistic (SPE - Squared Prediction Error)
	// for each observation, measuring the lack of fit
	QResiduals []float64

	// OutlierMask indicates which observations are considered outliers
	// based on the specified significance level
	OutlierMask []bool

	// ContributionScores contains the contribution of each variable
	// to each principal component (variables x components)
	ContributionScores [][]float64

	// ConfidenceEllipse contains parameters for drawing confidence ellipses
	// in score plots for outlier visualization
	ConfidenceEllipse EllipseParams
}

// EllipseParams defines parameters for confidence ellipse visualization
type EllipseParams struct {
	// Center coordinates of the ellipse (typically the mean of scores)
	CenterX float64
	CenterY float64

	// Semi-major and semi-minor axes lengths
	MajorAxis float64
	MinorAxis float64

	// Rotation angle in radians
	Angle float64

	// Confidence level (e.g., 0.95 for 95% confidence)
	ConfidenceLevel float64
}

// MetricsConfig contains configuration options for metrics calculation
type MetricsConfig struct {
	// NumComponents specifies how many components to use for metrics calculation
	// If 0, uses all available components from the PCA result
	NumComponents int

	// SignificanceLevel for outlier detection (e.g., 0.01 for 1% significance)
	// Default is 0.01 if not specified
	SignificanceLevel float64

	// CalculateContributions determines whether to compute variable contributions
	// Default is true
	CalculateContributions bool

	// CalculateConfidenceEllipse determines whether to compute ellipse parameters
	// Default is true
	CalculateConfidenceEllipse bool
}

// MetricsCalculator defines the interface for PCA metrics computation
type MetricsCalculator interface {
	// CalculateMetrics computes all diagnostic metrics for the given PCA result
	// Parameters:
	//   - result: The PCA result containing scores, loadings, and eigenvalues
	//   - data: The original data matrix (observations x variables)
	//   - config: Configuration options for metrics calculation
	// Returns:
	//   - PCAMetrics containing all calculated metrics
	//   - error if calculation fails
	CalculateMetrics(result *PCAResult, data Matrix, config MetricsConfig) (*PCAMetrics, error)

	// DetectOutliers identifies outliers based on Hotelling's T² statistic
	// Parameters:
	//   - metrics: Previously calculated PCA metrics
	//   - significance: Significance level for outlier detection (e.g., 0.01)
	// Returns:
	//   - Boolean array indicating outliers (true = outlier)
	DetectOutliers(metrics *PCAMetrics, significance float64) []bool

	// CalculateContributions computes variable contributions to each PC
	// Parameters:
	//   - result: The PCA result containing loadings
	//   - data: The original data matrix
	// Returns:
	//   - Matrix of contributions (variables x components)
	CalculateContributions(result *PCAResult, data Matrix) [][]float64
}
