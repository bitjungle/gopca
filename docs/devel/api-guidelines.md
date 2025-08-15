# GoPCA API Documentation Guidelines

This guide establishes standards for documenting APIs in the GoPCA codebase.

## Overview

Clear, comprehensive API documentation is essential for maintainability and usability. This guide covers standards for documenting Go code, with emphasis on scientific computing and mathematical algorithms.

## Documentation Principles

1. **Completeness**: Every exported entity must be documented
2. **Clarity**: Use clear, concise language
3. **Accuracy**: Keep documentation synchronized with code
4. **Examples**: Provide usage examples for complex APIs
5. **References**: Cite mathematical sources

## Function Documentation

### Basic Structure

```go
// FunctionName performs a specific action with inputs.
// 
// Additional details about the algorithm or approach.
// Special cases or edge conditions to be aware of.
//
// Parameters:
//   - param1: Description of first parameter
//   - param2: Description of second parameter
//
// Returns the computed result and any error encountered.
// Error conditions include invalid input, numerical instability, etc.
func FunctionName(param1 Type1, param2 Type2) (Result, error)
```

### Mathematical Functions

```go
// SVD performs Singular Value Decomposition on matrix A.
//
// Decomposes A into U * Σ * V^T where:
//   - U: Left singular vectors (m×m orthogonal matrix)
//   - Σ: Diagonal matrix of singular values
//   - V: Right singular vectors (n×n orthogonal matrix)
//
// Reference: Golub & Van Loan (2013), Matrix Computations, 4th ed., Ch. 8
//
// Algorithm complexity: O(min(mn², m²n))
// Memory usage: O(mn) for thin SVD, O(m² + n²) for full SVD
func SVD(A *mat.Dense, thin bool) (U, S, V *mat.Dense, error)
```

## Type Documentation

### Structs

```go
// PCAConfig configures the PCA analysis parameters.
// It controls preprocessing, algorithm selection, and output options.
type PCAConfig struct {
    // Components specifies the number of principal components to compute.
    // Must be positive and not exceed min(samples, features).
    Components int
    
    // MeanCenter indicates whether to center data to zero mean.
    // This is typically required for PCA unless data is pre-centered.
    MeanCenter bool
    
    // Method selects the PCA algorithm: "svd", "nipals", or "kernel".
    // Default is "svd" for complete data.
    Method string
}
```

### Interfaces

```go
// PCAEngine defines the contract for PCA algorithm implementations.
// Different algorithms (SVD, NIPALS, Kernel) implement this interface
// to provide a uniform API for PCA analysis.
//
// Thread safety: Implementations are not required to be thread-safe.
// Concurrent access should be synchronized by the caller.
type PCAEngine interface {
    // Fit computes the PCA model from training data.
    // Returns error if data is invalid or algorithm fails to converge.
    Fit(data Matrix) error
    
    // Transform projects new data onto principal components.
    // Requires Fit to be called first.
    Transform(data Matrix) (Matrix, error)
}
```

## Constants and Enums

```go
// PCA method identifiers
const (
    // MethodSVD uses Singular Value Decomposition (fastest for complete data)
    MethodSVD = "svd"
    
    // MethodNIPALS uses Nonlinear Iterative Partial Least Squares (handles missing data)
    MethodNIPALS = "nipals"
    
    // MethodKernel uses Kernel PCA for non-linear relationships
    MethodKernel = "kernel"
)
```

## Package Documentation

Create a `doc.go` file for each package:

```go
// Package core implements the mathematical algorithms for PCA analysis.
// It provides multiple PCA implementations optimized for different scenarios.
//
// Algorithms:
//   - SVD: Fast, accurate for complete data
//   - NIPALS: Iterative, handles missing data
//   - Kernel PCA: Non-linear dimensionality reduction
//
// The package uses gonum for numerical operations and follows
// standard mathematical notation from multivariate statistics.
//
// Example:
//
//     engine := core.NewPCAEngine(config)
//     result, err := engine.Fit(data)
//     if err != nil {
//         log.Fatal(err)
//     }
//     scores := result.Scores
//
// For mathematical background, see:
//   - Jolliffe (2002), Principal Component Analysis
//   - Wold (1966), Estimation of principal components
package core
```

## Error Documentation

```go
var (
    // ErrInsufficientData indicates the data matrix has too few samples.
    // PCA requires at least 2 samples for meaningful analysis.
    ErrInsufficientData = errors.New("insufficient data: need at least 2 samples")
    
    // ErrSingularMatrix indicates the covariance matrix is singular.
    // This typically occurs with perfectly correlated variables.
    ErrSingularMatrix = errors.New("singular covariance matrix")
)
```

## Inline Comments

### Algorithm Steps

```go
// NIPALS algorithm main loop
for k := 0; k < nComponents; k++ {
    // Step 1: Initialize scores with column of maximum variance
    t := X.Col(maxVarIdx)
    
    // Step 2: Iterate until convergence
    for iter := 0; iter < maxIter; iter++ {
        // Project X onto t to get loadings: p = X^T * t / (t^T * t)
        p = X.T().Mul(t).Scale(1.0 / t.Dot(t))
        
        // Normalize loadings to unit length
        p = p.Scale(1.0 / p.Norm())
        
        // Update scores: t = X * p / (p^T * p)
        tNew = X.Mul(p).Scale(1.0 / p.Dot(p))
        
        // Check convergence: ||t_new - t|| < tolerance
        if tNew.Sub(t).Norm() < tolerance {
            break
        }
        t = tNew
    }
    
    // Step 3: Deflate X by removing component contribution
    // X = X - t * p^T (removes variance explained by this component)
    X = X.Sub(t.Outer(p))
}
```

### Complex Logic

```go
// Determine optimal gamma for RBF kernel using median heuristic
// This provides a reasonable default when gamma is not specified
if gamma == 0 {
    // Calculate pairwise distances
    distances := calculatePairwiseDistances(X)
    
    // Use median distance as characteristic length scale
    // This ensures the kernel captures typical data separation
    gamma = 1.0 / (2.0 * median(distances) * median(distances))
}
```

## Testing Documentation

```go
// TestSVD_Reproducibility verifies that SVD produces consistent results
// across multiple runs with the same input data.
// This is critical for scientific reproducibility.
func TestSVD_Reproducibility(t *testing.T) {
    // Test implementation...
}

// TestNIPALS_MissingData validates NIPALS handling of missing values.
// Uses synthetic data with known ground truth from R implementation.
//
// Reference output generated with:
//   R> library(nipals)
//   R> result <- nipals(data, ncomp=3, center=TRUE, scale=FALSE)
func TestNIPALS_MissingData(t *testing.T) {
    // Test implementation...
}
```

## Best Practices

1. **Update Documentation with Code**: When modifying functionality, update docs in the same commit
2. **Use godoc Format**: Follow standard Go documentation conventions
3. **Include Examples**: Add runnable examples in `_test.go` files
4. **Document Assumptions**: State any assumptions about input data
5. **Explain Deviations**: Document why code deviates from standard approaches
6. **Version Compatibility**: Note any version-specific behavior

## Tools

- `go doc`: View documentation from command line
- `godoc -http=:6060`: Browse documentation locally
- `golint`: Check documentation completeness
- VS Code Go extension: Inline documentation hints

## References

- [Effective Go - Commentary](https://golang.org/doc/effective_go.html#commentary)
- [Godoc: documenting Go code](https://blog.golang.org/godoc-documenting-go-code)
- [Go Code Review Comments](https://github.com/golang/go/wiki/CodeReviewComments)