// Copyright 2025 bitjungle - Rune Mathisen. All rights reserved.
// Use of this source code is governed by the MIT license
// that can be found in the LICENSE file.
// The author respectfully requests that it not be used for
// military, warfare, or surveillance applications.

package csv

import (
	"time"

	"github.com/bitjungle/gopca/internal/core"
	"github.com/bitjungle/gopca/pkg/types"
)

// ConvertToPCAOutputData converts PCAResult and Data to PCAOutputData for export
// This function is shared between CLI and Desktop applications
func ConvertToPCAOutputData(result *types.PCAResult, data *Data, includeMetrics bool,
	config types.PCAConfig, preprocessor *core.Preprocessor,
	categoricalData map[string][]string, targetData map[string][]float64) *types.PCAOutputData {

	// Create timestamp
	createdAt := time.Now().Format(time.RFC3339)

	// Create model metadata
	// Use the actual method from the result, not the config
	metadata := types.ModelMetadata{
		Version:   "1.0",
		CreatedAt: createdAt,
		Software:  "gopca",
		Config: types.ModelConfig{
			Method:          result.Method,             // Use actual method from result
			NComponents:     result.ComponentsComputed, // Use actual components computed
			MissingStrategy: config.MissingStrategy,
			ExcludedRows:    config.ExcludedRows,
			ExcludedColumns: config.ExcludedColumns,
		},
	}

	// Only include kernel parameters for kernel PCA
	if result.Method == "kernel" {
		metadata.Config.KernelType = config.KernelType
		// Only include relevant parameters based on kernel type
		switch config.KernelType {
		case "rbf":
			metadata.Config.KernelGamma = config.KernelGamma
		case "poly", "polynomial":
			metadata.Config.KernelGamma = config.KernelGamma
			metadata.Config.KernelDegree = config.KernelDegree
			metadata.Config.KernelCoef0 = config.KernelCoef0
			// For linear kernel, only kernel_type is needed
		}
	}

	// Create preprocessing info
	preprocessingInfo := types.PreprocessingInfo{
		MeanCenter:    config.MeanCenter,
		StandardScale: config.StandardScale,
		RobustScale:   config.RobustScale,
		ScaleOnly:     config.ScaleOnly,
		SNV:           config.SNV,
		VectorNorm:    config.VectorNorm,
		Parameters:    types.PreprocessingParams{},
	}

	// Add preprocessing parameters if preprocessor was used
	if preprocessor != nil {
		preprocessingInfo.Parameters.FeatureMeans = preprocessor.GetMeans()
		preprocessingInfo.Parameters.FeatureStdDevs = preprocessor.GetStdDevs()
		preprocessingInfo.Parameters.FeatureMedians = preprocessor.GetMedians()
		preprocessingInfo.Parameters.FeatureMADs = preprocessor.GetMADs()
		preprocessingInfo.Parameters.RowMeans = preprocessor.GetRowMeans()
		preprocessingInfo.Parameters.RowStdDevs = preprocessor.GetRowStdDevs()
	}

	// Create model components
	modelComponents := types.ModelComponents{
		Loadings:               result.Loadings,
		ExplainedVariance:      result.ExplainedVar,
		ExplainedVarianceRatio: result.ExplainedVarRatio,
		CumulativeVariance:     result.CumulativeVar,
		ComponentLabels:        result.ComponentLabels,
		FeatureLabels:          data.Headers,
	}

	// Create results data
	resultsData := types.ResultsData{
		Samples: types.SamplesResults{
			Names:  data.RowNames,
			Scores: result.Scores,
		},
	}

	// Add metrics if requested (skip for kernel PCA as it doesn't have loadings)
	if includeMetrics && result.Method != "kernel" && data.Matrix != nil {
		metrics, err := core.CalculateMetricsFromPCAResult(result, data.Matrix)
		if err == nil && metrics != nil {
			metricsData := &types.MetricsData{
				HotellingT2: make([]float64, len(metrics)),
				Mahalanobis: make([]float64, len(metrics)),
				RSS:         make([]float64, len(metrics)),
				IsOutlier:   make([]bool, len(metrics)),
			}
			for i, m := range metrics {
				metricsData.HotellingT2[i] = m.HotellingT2
				metricsData.Mahalanobis[i] = m.Mahalanobis
				metricsData.RSS[i] = m.RSS
				metricsData.IsOutlier[i] = m.IsOutlier
			}
			resultsData.Samples.Metrics = metricsData
		}
	} else if includeMetrics && result.Method == "kernel" {
		// For kernel PCA, we can't calculate RSS but we can still calculate some metrics if we have them in the result
		if len(result.Metrics) > 0 {
			metricsData := &types.MetricsData{
				HotellingT2: make([]float64, len(result.Metrics)),
				Mahalanobis: make([]float64, len(result.Metrics)),
				RSS:         make([]float64, len(result.Metrics)),
				IsOutlier:   make([]bool, len(result.Metrics)),
			}
			for i, m := range result.Metrics {
				metricsData.HotellingT2[i] = m.HotellingT2
				metricsData.Mahalanobis[i] = m.Mahalanobis
				metricsData.RSS[i] = m.RSS
				metricsData.IsOutlier[i] = m.IsOutlier
			}
			resultsData.Samples.Metrics = metricsData
		}
	}

	// Create diagnostic limits
	diagnostics := types.DiagnosticLimits{
		T2Limit95: result.T2Limit95,
		T2Limit99: result.T2Limit99,
		QLimit95:  result.QLimit95,
		QLimit99:  result.QLimit99,
	}

	// Add preserved columns if provided
	var preservedColumns *types.PreservedColumns
	if len(categoricalData) > 0 || len(targetData) > 0 {
		preservedColumns = &types.PreservedColumns{
			Categorical:   categoricalData,
			NumericTarget: targetData,
		}
	}

	return &types.PCAOutputData{
		Schema:            "https://github.com/bitjungle/gopca/schemas/v1/pca-output.schema.json",
		Metadata:          metadata,
		Preprocessing:     preprocessingInfo,
		Model:             modelComponents,
		Results:           resultsData,
		Diagnostics:       diagnostics,
		Eigencorrelations: result.Eigencorrelations,
		PreservedColumns:  preservedColumns,
	}
}
