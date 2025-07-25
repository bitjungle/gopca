# Kernel PCA in GoPCA

Kernel PCA is a non-linear dimensionality reduction technique that extends traditional PCA by using kernel functions to project data into higher-dimensional feature spaces where linear separation becomes possible.

## When to Use Kernel PCA

Use Kernel PCA when:
- Your data has non-linear relationships that linear PCA cannot capture
- You observe circular, spiral, or other complex patterns in your data
- Traditional PCA results show poor separation between known groups
- You work with data types that benefit from specific kernels (e.g., text, images, spectroscopy)

## Available Kernels

### RBF (Radial Basis Function) Kernel
The RBF or Gaussian kernel is the most commonly used kernel for non-linear data:
- **Formula**: `K(x, y) = exp(-γ ||x - y||²)`
- **Parameter**: `gamma` (γ) - controls the "reach" of the kernel
- **Use case**: General non-linear relationships, circular patterns

### Linear Kernel  
The linear kernel is equivalent to standard PCA:
- **Formula**: `K(x, y) = x · y`
- **Parameters**: None
- **Use case**: Testing, comparison with standard PCA

### Polynomial Kernel
The polynomial kernel captures polynomial relationships:
- **Formula**: `K(x, y) = (γ x · y + c₀)^d`
- **Parameters**: 
  - `gamma` (γ) - scaling factor
  - `degree` (d) - polynomial degree
  - `coef0` (c₀) - independent term
- **Use case**: Known polynomial relationships

## CLI Usage

### Basic RBF Kernel PCA
```bash
gopca-cli analyze --method kernel --kernel-type rbf --kernel-gamma 0.1 data.csv
```

### Polynomial Kernel with Custom Parameters
```bash
gopca-cli analyze --method kernel --kernel-type poly \
  --kernel-gamma 0.01 --kernel-degree 3 --kernel-coef0 1.0 \
  -c 5 data.csv
```

### Linear Kernel (equivalent to standard PCA)
```bash
gopca-cli analyze --method kernel --kernel-type linear data.csv
```

### Kernel PCA with Variance Scaling
```bash
# Scale features by standard deviation without centering
gopca-cli analyze --method kernel --kernel-type rbf --kernel-gamma 0.1 --scale-only data.csv

# Combine with row-wise preprocessing
gopca-cli analyze --method kernel --kernel-type rbf --snv --scale-only spectral_data.csv
```

## GUI Usage

1. Load your data file
2. In the "Configure PCA" section, select "Kernel PCA" from the Method dropdown
3. Configure kernel parameters:
   - Select kernel type (RBF, Linear, or Polynomial)
   - Adjust gamma parameter (for RBF and Polynomial)
   - Set degree and coef0 (for Polynomial only)
4. Click "Run PCA Analysis"

**Note**: Kernel PCA performs its own centering in kernel space, so standard preprocessing options that include mean centering (standard scale, robust scale) are not recommended. However, you can use:
- **Variance scaling**: Divides features by their standard deviation without centering - useful when features have different scales
- **Row-wise preprocessing**: SNV and vector normalization are fully compatible with Kernel PCA

## Parameter Guidelines

### Gamma Parameter
- **Small gamma** (e.g., 0.001-0.01): Smoother decision boundaries, each point has far-reaching influence
- **Large gamma** (e.g., 0.1-10): More complex boundaries, each point has local influence
- Start with `gamma = 1/n_features` as a rule of thumb

### Polynomial Degree
- **Degree 2**: Quadratic relationships
- **Degree 3**: Cubic relationships (default)
- Higher degrees risk overfitting

## Interpretation

### Scores
- Kernel PCA scores represent projections in the kernel-induced feature space
- Interpret similarly to regular PCA scores for visualization
- Distance relationships may differ from linear PCA

### Loadings
- Kernel PCA does not produce traditional loadings
- Variable contributions cannot be directly interpreted
- Focus on scores for analysis

### Explained Variance
- Represents variance captured in kernel space
- Not directly comparable to linear PCA explained variance
- Still useful for selecting number of components

## Performance Considerations

- Kernel PCA has O(n²) memory complexity due to kernel matrix
- Computation time scales with O(n³) for eigendecomposition
- For datasets >10,000 samples, consider:
  - Sampling approaches
  - Approximate kernel methods
  - Linear PCA as initial exploration

## Example: Circular Data

For data with circular patterns (e.g., two concentric circles):
```bash
# RBF kernel will separate the circles effectively
gopca-cli analyze --method kernel --kernel-type rbf --kernel-gamma 0.5 circles.csv

# Linear kernel (standard PCA) will fail to separate
gopca-cli analyze --method svd circles.csv
```

## References

- Schölkopf, B., Smola, A., & Müller, K. R. (1998). Nonlinear component analysis as a kernel eigenvalue problem.
- For mathematical details, see the implementation in `internal/core/kernel_pca.go`