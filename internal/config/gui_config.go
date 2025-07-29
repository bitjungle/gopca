package config

// GUIConfig holds configuration for the GUI application
type GUIConfig struct {
	// Visualization configuration
	Visualization VisualizationConfig `json:"visualization"`

	// UI configuration
	UI UIConfig `json:"ui"`
}

// VisualizationConfig holds visualization-related configuration
type VisualizationConfig struct {
	// Threshold for showing variables in loadings plot
	LoadingsVariableThreshold int `json:"loadings_variable_threshold"`

	// Threshold for correlation circle
	CorrelationThreshold float64 `json:"correlation_threshold"`

	// Elbow threshold for scree plot
	ElbowThreshold float64 `json:"elbow_threshold"`

	// Diagnostic plot thresholds
	MahalanobisThreshold float64 `json:"mahalanobis_threshold"`
	RSSThreshold         float64 `json:"rss_threshold"`

	// Default confidence level for ellipses
	DefaultConfidenceLevel float64 `json:"default_confidence_level"`
}

// UIConfig holds UI-related configuration
type UIConfig struct {
	// Maximum rows to show in data preview
	DataPreviewMaxRows int `json:"data_preview_max_rows"`

	// Maximum columns to show in data preview
	DataPreviewMaxCols int `json:"data_preview_max_cols"`

	// Default zoom factor for plots
	DefaultZoomFactor float64 `json:"default_zoom_factor"`
}

// DefaultGUIConfig returns the default GUI configuration
func DefaultGUIConfig() *GUIConfig {
	return &GUIConfig{
		Visualization: VisualizationConfig{
			LoadingsVariableThreshold: 50,
			CorrelationThreshold:      0.3,
			ElbowThreshold:            80.0,
			MahalanobisThreshold:      3.0,
			RSSThreshold:              0.03,
			DefaultConfidenceLevel:    0.95,
		},
		UI: UIConfig{
			DataPreviewMaxRows: 10,
			DataPreviewMaxCols: 10,
			DefaultZoomFactor:  0.8,
		},
	}
}
