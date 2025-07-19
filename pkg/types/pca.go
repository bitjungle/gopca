package types

// Matrix represents a 2D data matrix
type Matrix [][]float64

// PCAConfig holds configuration for PCA analysis
type PCAConfig struct {
	Components    int    `json:"components"`
	MeanCenter    bool   `json:"mean_center"`
	StandardScale bool   `json:"standard_scale"`
	Method        string `json:"method"` // "svd" or "eigen"
}

// PCAResult contains the results of PCA analysis
type PCAResult struct {
	Scores          Matrix    `json:"scores"`
	Loadings        Matrix    `json:"loadings"`
	ExplainedVar    []float64 `json:"explained_variance"`
	CumulativeVar   []float64 `json:"cumulative_variance"`
	ComponentLabels []string  `json:"component_labels"`
}

// PCAEngine defines the interface for PCA computation
type PCAEngine interface {
	Fit(data Matrix, config PCAConfig) (*PCAResult, error)
	Transform(data Matrix) (Matrix, error)
	FitTransform(data Matrix, config PCAConfig) (*PCAResult, error)
}
