# Add Temporal PCA (Time-Delay PCA) Method for Time-Series Analysis

## Summary
Implement Temporal PCA as a new method in GoPCA's core engine to enable sophisticated time-series analysis. This method augments data with lagged values to capture temporal dynamics and autocorrelations, making GoPCA the definitive tool for both static and temporal data analysis.

## Motivation
While GoPCA excels at standard PCA (SVD, NIPALS) and Kernel PCA, it currently lacks specialized support for time-series data. Temporal PCA fills this gap by:
- Capturing temporal dynamics and autocorrelations in multivariate time series
- Enabling anomaly detection through reconstruction error over time
- Providing dimensionality reduction for time-dependent systems
- Supporting quality control in manufacturing processes with temporal patterns

This positions GoPCA as the complete PCA solution for data scientists, researchers, and engineers working with both static and temporal data.

## Technical Specification

### Core Algorithm
Temporal PCA extends standard PCA by creating a lag matrix where each time point is augmented with L previous time points:
- **Input**: Time series X ∈ ℝ^(T×p) (T timepoints, p variables)
- **Transform**: Build lag matrix Φ(X,L) ∈ ℝ^((T-L+1) × (p·L))
- **Apply**: Standard PCA via SVD on the lag matrix
- **Output**: Temporal PC scores, lag-specific loadings, reconstruction errors

Mathematical foundation:
```
For each time t ∈ [L-1, T-1]:
  row_t = [x_t, x_{t-1}, ..., x_{t-L+1}]  // Concatenated across all variables
```

### Integration Architecture

#### 1. Core Implementation (`internal/core/`)

```go
// temporal_pca.go - New file following GoPCA patterns
type TemporalPCAConfig struct {
    NumLags           int                  // Number of time lags (L >= 1)
    NumComponents     int                  // Number of components to retain
    VarianceExplained float64              // Alternative to NumComponents (0.0-1.0)
    TimeColumn        string               // Optional time column identifier
    Preprocessing     PreprocessingConfig  // Reuse existing preprocessing
    ImputeMethod      string               // "forward", "backward", "linear", "none"
}

type TemporalPCAModel struct {
    *basePCAModel     // Inherit common PCA functionality
    Config            TemporalPCAConfig
    NumLags           int
    OriginalVars      int                  // Number of original variables (p)
    TimeAlignment     []int64              // Maps output indices to original time
    LagLoadings       [][][]float64        // [component][variable][lag]
    LaggedMeans       []float64            // Mean of each lagged feature
    LaggedScales      []float64            // Scale of each lagged feature
}

// Implement PCAEngine interface
func (t *TemporalPCAModel) Fit(data [][]float64) error
func (t *TemporalPCAModel) Transform(data [][]float64) ([][]float64, error)
func (t *TemporalPCAModel) FitTransform(data [][]float64) ([][]float64, error)
func (t *TemporalPCAModel) InverseTransform(scores [][]float64) ([][]float64, error)
func (t *TemporalPCAModel) ReconstructionError(data [][]float64) ([]float64, error)
func (t *TemporalPCAModel) GetLoadings() [][]float64
func (t *TemporalPCAModel) GetExplainedVariance() []float64
```

#### 2. CLI Integration (`cmd/gopca-cli/`)

Extend existing CLI with temporal-specific options:
```bash
# Analyze time series with 24 lags (e.g., daily cycle in hourly data)
pca analyze --method temporal --temporal-lags 24 \
  --components 5 --time-column timestamp \
  timeseries.csv -o temporal_results.json

# Transform new data using fitted temporal model
pca transform temporal_model.json new_data.csv \
  -o transformed.csv --include-timestamps

# Compute reconstruction errors for anomaly detection
pca recon-error temporal_model.json test_data.csv \
  -o anomaly_scores.csv
```

#### 3. Desktop GUI Integration (`cmd/gopca-desktop/`)

##### Configuration Panel Extensions
```typescript
// New components in frontend/src/components/config/
interface TemporalPCAConfig {
  numLags: number;
  timeColumn?: string;
  imputeMethod: 'forward' | 'backward' | 'linear' | 'none';
  showAutoCorrelation: boolean;
}

// TemporalConfigPanel.tsx
- Lag count selector with presets (24=daily, 168=weekly, etc.)
- Time column dropdown (auto-detect from data)
- Imputation method selector
- Preview showing resulting matrix dimensions
- Auto-correlation plot for lag selection guidance
```

##### New Visualizations
1. **Temporal Loadings Heatmap** (`TemporalLoadingsPlot.tsx`):
   - 2D heatmap: rows=variables, columns=lags
   - Color intensity shows loading magnitude
   - Interactive tooltips with exact values
   - Export as high-resolution image

2. **Reconstruction Error Timeline** (`ReconErrorPlot.tsx`):
   - Time-series plot of reconstruction errors
   - Anomaly threshold lines (configurable percentiles)
   - Zoom/pan for detailed inspection
   - Synchronized with original data view

3. **Lag Contribution Plot** (`LagContributionPlot.tsx`):
   - Stacked bar chart showing variance by lag
   - Helps understand temporal importance
   - Per-component breakdown available

### Implementation Phases

#### Phase 1: Core Implementation (Estimated: 2-3 weeks)
- [ ] Implement efficient lag matrix construction
  - [ ] Concurrent processing for large datasets
  - [ ] Memory-efficient chunking strategy
- [ ] Add TemporalPCA to PCAEngine interface
- [ ] Integrate with existing preprocessing pipeline
- [ ] Implement core PCA computation via SVD
- [ ] Add reconstruction error calculation
- [ ] Create comprehensive unit tests (target >85% coverage)
- [ ] Validate against Python scikit-learn reference
- [ ] Add CLI support with basic commands

#### Phase 2: GUI Integration (Estimated: 2-3 weeks)
- [ ] Add temporal method to method selector
- [ ] Implement temporal configuration panel
- [ ] Create temporal loadings visualization
- [ ] Add reconstruction error timeline plot
- [ ] Implement lag contribution visualization
- [ ] Add contextual help and tooltips
- [ ] Update help documentation
- [ ] Add example datasets for tutorials

#### Phase 3: Advanced Features (Future Enhancement)
- [ ] Streaming mode for very large time series
- [ ] Multi-scale temporal analysis (multiple lag scales)
- [ ] Automated lag selection via information criteria
- [ ] GPU acceleration via CUDA bindings
- [ ] Real-time anomaly detection mode
- [ ] Export to time-series specific formats

### Key Design Decisions

#### Memory Management
```go
// Intelligent memory estimation before computation
func EstimateMemoryUsage(samples, variables, lags int) (bytes int64, warning string) {
    lagMatrixSize := int64(samples - lags + 1) * int64(variables * lags) * 8
    overhead := lagMatrixSize / 10 // ~10% overhead
    total := lagMatrixSize + overhead
    
    if total > 2*1024*1024*1024 { // >2GB
        warning = "Large memory usage expected. Consider reducing lags or using CLI."
    }
    return total, warning
}
```

#### Missing Data Handling
- **Forward fill**: Default for sensor data (last known value)
- **Linear interpolation**: For smooth processes
- **Error on gaps**: Option for critical applications
- Preserve temporal ordering during all operations

#### Time Alignment Rules
- Output index `i` corresponds to time `t = i + L - 1`
- Timestamps preserved in export formats
- Clear documentation of alignment in results

### Performance Requirements

1. **Scalability Targets**:
   - Handle up to 1M timepoints with L≤48
   - Process 10K variables with reasonable lags
   - Linear memory scaling O(T·p·L)
   - GUI responsive during computation (progress indicators)

2. **Optimization Strategies**:
   - Use gonum BLAS operations
   - Concurrent lag matrix construction
   - Optional chunked processing for huge datasets
   - Reuse memory buffers where possible

### Testing Requirements

#### Unit Tests (`internal/core/temporal_pca_test.go`)
```go
func TestLagMatrixConstruction(t *testing.T)
func TestTemporalPCAFit(t *testing.T)
func TestTimeAlignment(t *testing.T)
func TestReconstructionError(t *testing.T)
func TestEdgeCases(t *testing.T) // single lag, more lags than samples
func TestNumericalStability(t *testing.T)
func TestMissingDataHandling(t *testing.T)
```

#### Integration Tests
- End-to-end CLI workflow with real data
- GUI interaction testing
- Model serialization/deserialization
- Cross-validation with time series splits
- Performance benchmarks

#### Validation Against References
```python
# Reference implementation for testing
import numpy as np
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler

def temporal_pca_reference(X, n_lags, n_components):
    """Reference implementation for validation"""
    # Create lag matrix
    T, p = X.shape
    lag_matrix = np.hstack([X[n_lags-1-i:T-i] for i in range(n_lags)])
    
    # Standardize and apply PCA
    scaler = StandardScaler()
    X_scaled = scaler.fit_transform(lag_matrix)
    pca = PCA(n_components=n_components)
    scores = pca.fit_transform(X_scaled)
    
    return scores, pca.explained_variance_ratio_
```

### Documentation Requirements

#### User Guide (`docs/temporal-pca-guide.md`)
1. **Conceptual Introduction**
   - What is Temporal PCA?
   - When to use vs. standard PCA
   - Understanding lag matrices
   
2. **Practical Guidelines**
   - Choosing number of lags
   - Interpreting temporal loadings
   - Anomaly detection workflow
   
3. **Examples**
   - Manufacturing sensor monitoring
   - Financial time series analysis
   - Environmental data patterns

#### API Documentation
- Comprehensive godoc comments
- Mathematical references:
  - Broomhead & King (1986) - SSA foundations
  - Golyandina & Zhigljavsky (2013) - Modern SSA
  - Jolliffe & Cadima (2016) - PCA review
- Algorithm complexity: O(T·p·L) space, O(min(n²d, nd²)) time

#### GUI Help System
- Interactive tutorial for first-time users
- Contextual help for each parameter
- Visual explanation of lag matrix concept
- Best practices checklist

### Success Criteria

- [ ] Temporal PCA available in both CLI and GUI
- [ ] Results match reference implementation (tolerance: 1e-6)
- [ ] Memory usage scales linearly with theoretical expectations
- [ ] GUI remains responsive with progress indicators
- [ ] Comprehensive documentation with examples
- [ ] Test coverage >85% for core implementation
- [ ] All existing tests continue to pass
- [ ] Performance meets or exceeds targets
- [ ] Clean integration with existing visualizations

### Potential Challenges & Mitigations

1. **Large Memory Footprint**
   - *Challenge*: Lag matrix can be very large
   - *Mitigation*: Implement chunked processing, add memory warnings

2. **Numerical Stability**
   - *Challenge*: High lag counts can create ill-conditioned matrices
   - *Mitigation*: Use robust SVD, add condition number checks

3. **User Complexity**
   - *Challenge*: More parameters than standard PCA
   - *Mitigation*: Smart defaults, comprehensive help, example workflows

4. **Visualization Complexity**
   - *Challenge*: Temporal loadings are 3D (component × variable × lag)
   - *Mitigation*: Multiple coordinated views, interactive exploration

### Impact on Existing Code

- **Minimal breaking changes**: Existing PCA methods unchanged
- **Extended interfaces**: PCAEngine gains optional temporal methods
- **New dependencies**: None required (uses existing gonum)
- **Configuration changes**: Additional fields in PCAConfig
- **JSON schema**: Extended to support temporal models

### Future Enhancements

1. **Dynamic PCA**: Moving window analysis
2. **Multi-resolution**: Combined short and long-term lags  
3. **Causal PCA**: Directional temporal relationships
4. **Online learning**: Incremental model updates
5. **Forecast mode**: Predict future PC scores

## References

1. Broomhead, D.S., & King, G.P. (1986). Extracting qualitative dynamics from experimental data. *Physica D*, 20(2-3), 217-236.

2. Golyandina, N., & Zhigljavsky, A. (2013). *Singular Spectrum Analysis for Time Series*. Springer.

3. Jolliffe, I.T., & Cadima, J. (2016). Principal component analysis: a review and recent developments. *Philosophical Transactions A*, 374(2065).

4. Plaut, G., & Vautard, R. (1994). Spells of low-frequency oscillations and weather regimes in the Northern Hemisphere. *Journal of Atmospheric Sciences*, 51(2), 210-236.

## Conclusion

This enhancement would establish GoPCA as the most comprehensive PCA tool available, capable of handling both traditional cross-sectional data and sophisticated time-series analysis. The phased implementation approach ensures delivery of a high-quality, well-tested feature that maintains GoPCA's standards while significantly expanding its capabilities.

The addition of Temporal PCA aligns perfectly with GoPCA's mission to be "the definitive Principal Component Analysis application" by addressing a major use case in data analysis that currently requires specialized tools or custom implementations.