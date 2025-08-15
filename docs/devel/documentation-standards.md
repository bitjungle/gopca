# GoPCA Documentation Standards

This document defines the documentation standards for the GoPCA project, ensuring consistency and quality across all documentation.

## Documentation Hierarchy

```
README.md                    # Project overview, quick start
├── CLAUDE.md               # AI assistant and developer guide
├── SECURITY.md             # Security policy
├── docs/
│   ├── cli_reference.md   # CLI command reference
│   ├── data-format.md      # Data format specifications
│   ├── intro_to_pca.md     # PCA theory and concepts
│   └── devel/              # Developer documentation
│       ├── api-guidelines.md
│       ├── documentation-standards.md (this file)
│       └── ...
└── Package doc.go files    # Package-level documentation
```

## Documentation Types

### 1. User Documentation

**Purpose**: Help end users operate the software effectively.

**Requirements**:
- Clear, jargon-free language
- Step-by-step instructions
- Screenshots and examples
- Common use cases
- Troubleshooting section

**Files**:
- `README.md`
- `docs/cli_reference.md`
- `docs/intro_to_pca.md`
- `docs/intro_to_data_prep.md`
- `cmd/gopca-desktop/frontend/src/help/help-content.json`

### 2. Developer Documentation

**Purpose**: Enable developers to understand, maintain, and extend the codebase.

**Requirements**:
- Technical accuracy
- Architecture diagrams
- API references
- Code examples
- Development workflow

**Files**:
- `CLAUDE.md`
- `docs/devel/*.md`
- Package `doc.go` files
- Inline code comments

### 3. API Documentation

**Purpose**: Document public interfaces and their contracts.

**Requirements**:
- Every exported function, type, constant documented
- Parameter and return value descriptions
- Error conditions
- Usage examples
- Mathematical references

**Location**: In source code as Go doc comments

## Writing Standards

### Language and Tone

1. **Clarity**: Use simple, direct language
2. **Consistency**: Maintain consistent terminology
3. **Active Voice**: Prefer "The function calculates..." over "The calculation is performed by..."
4. **Present Tense**: Describe what code does, not what it will do

### Mathematical Documentation

When documenting mathematical algorithms:

1. **Use Standard Notation**
   ```
   X = U * Σ * V^T  (not X = U * S * V')
   ```

2. **Define Variables**
   ```
   Where:
   - X: Data matrix (n×m)
   - U: Left singular vectors (n×k)
   - Σ: Singular values (k×k diagonal)
   - V: Right singular vectors (m×k)
   ```

3. **Cite References**
   ```
   Reference: Golub & Van Loan (2013), Matrix Computations, Ch. 8
   ```

4. **Include Complexity**
   ```
   Time complexity: O(mn²) for m > n
   Space complexity: O(mn)
   ```

### Code Examples

1. **Complete and Runnable**
   ```go
   // Complete example that compiles and runs
   package main
   
   import (
       "fmt"
       "github.com/bitjungle/gopca/pkg/types"
   )
   
   func main() {
       data := types.Matrix{{1, 2}, {3, 4}}
       fmt.Println(data)
   }
   ```

2. **Highlight Key Concepts**
   ```go
   // Key insight: Deflation removes component contribution
   X = X - t * p^T  // Subtract rank-1 approximation
   ```

3. **Show Expected Output**
   ```go
   // Example:
   //   Input:  [[1, 2], [3, 4]]
   //   Output: [[0.5, 0.5], [0.5, 0.5]]
   ```

## Markdown Conventions

### Headers

```markdown
# Page Title (H1 - one per document)
## Major Section (H2)
### Subsection (H3)
#### Detail (H4 - rarely needed)
```

### Code Blocks

Always specify language:
```markdown
```go
func Example() {}
```

```bash
pca analyze data.csv
```
```

### Lists

Ordered for sequences:
```markdown
1. First step
2. Second step
3. Third step
```

Unordered for collections:
```markdown
- Feature one
- Feature two
- Feature three
```

### Tables

```markdown
| Column 1 | Column 2 | Column 3 |
|----------|----------|----------|
| Data     | Data     | Data     |
```

### Links

Internal:
```markdown
See [Configuration Guide](../configuration.md)
```

External:
```markdown
Based on [Scikit-learn's PCA](https://scikit-learn.org/stable/modules/generated/sklearn.decomposition.PCA.html)
```

## Documentation Maintenance

### Review Checklist

Before committing documentation:

- [ ] Spell check completed
- [ ] Links verified
- [ ] Code examples tested
- [ ] Mathematical notation consistent
- [ ] References complete
- [ ] Version numbers updated

### Update Triggers

Update documentation when:

1. **API Changes**: Any public interface modification
2. **Bug Fixes**: If they affect documented behavior
3. **New Features**: Complete documentation before merge
4. **User Feedback**: Address confusion or gaps
5. **Dependencies**: When external dependencies change

### Version Synchronization

Keep documentation synchronized with code:

```go
// Good: Documentation matches implementation
// SVD performs thin SVD when thin=true, full SVD when thin=false
func SVD(A *mat.Dense, thin bool) ...

// Bad: Documentation doesn't match parameters
// SVD performs singular value decomposition
func SVD(A *mat.Dense, method string, options SVDOptions) ...
```

## Quality Metrics

### Coverage

- 100% of exported functions documented
- All packages have `doc.go` files
- README covers all major features
- CLI help text complete and accurate

### Accuracy

- Code examples compile and run
- Mathematical formulas correct
- References valid and accessible
- No outdated information

### Usability

- New users can get started in < 5 minutes
- Common tasks have examples
- Error messages helpful
- Search terms lead to relevant docs

## Tools and Automation

### Documentation Generation

```bash
# Generate godoc locally
godoc -http=:6060

# Check documentation coverage
go doc -all ./...
```

### Linting

```bash
# Check Go documentation
golint ./...

# Check markdown
markdownlint docs/
```

### Link Checking

```bash
# Verify all links
markdown-link-check docs/**/*.md
```

## Examples of Good Documentation

### Function Documentation

```go
// CalculateEigenvalues computes eigenvalues from the covariance matrix.
//
// For a centered data matrix X, the covariance matrix is:
//   C = (1/(n-1)) * X^T * X
//
// Eigenvalues represent the variance along each principal component.
// They are returned in descending order.
//
// Parameters:
//   - cov: Symmetric positive semi-definite covariance matrix
//   - n: Number of eigenvalues to compute (0 for all)
//
// Returns eigenvalues in descending order and any numerical errors.
//
// Reference: Press et al. (2007), Numerical Recipes, Ch. 11.1
func CalculateEigenvalues(cov *mat.SymDense, n int) ([]float64, error) {
    // Implementation
}
```

### Package Documentation

```go
// Package core implements principal component analysis algorithms.
//
// This package provides the mathematical foundation for PCA, including:
//   - Data preprocessing (centering, scaling, normalization)
//   - Multiple PCA algorithms (SVD, NIPALS, Kernel)
//   - Statistical metrics (variance, correlations, outliers)
//   - Visualization support (biplots, scree plots)
//
// Algorithm Selection:
//
// Use SVD (default) for:
//   - Complete data without missing values
//   - Best numerical stability
//   - Fastest computation
//
// Use NIPALS for:
//   - Data with missing values
//   - Very wide matrices (features >> samples)
//   - Memory-constrained environments
//
// Use Kernel PCA for:
//   - Non-linear relationships
//   - Complex manifold structures
//   - Classification preprocessing
//
// Example:
//
//     config := types.PCAConfig{
//         Components: 3,
//         Method: "svd",
//         MeanCenter: true,
//     }
//     engine := core.NewPCAEngine(config)
//     result, err := engine.Fit(data)
//
// For theoretical background, see docs/intro_to_pca.md
package core
```

## Common Pitfalls

1. **Outdated Examples**: Test all code examples regularly
2. **Missing Error Cases**: Document all error conditions
3. **Assumed Knowledge**: Define technical terms
4. **Platform Differences**: Note OS-specific behavior
5. **Version Skew**: Keep docs synchronized with code

## References

- [Google Developer Documentation Style Guide](https://developers.google.com/style)
- [Microsoft Writing Style Guide](https://docs.microsoft.com/en-us/style-guide/welcome/)
- [Write the Docs](https://www.writethedocs.org/guide/)
- [The Documentation System](https://documentation.divio.com/)