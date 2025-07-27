# Kernel PCA: Unrolling the Hidden Patterns in Your Data

## Understanding Kernel PCA Through the Swiss Roll

Imagine trying to flatten a cinnamon roll onto a plate—without tearing it—so you could study its delicious spiral pattern. In the world of data analysis, this is exactly the kind of challenge we face with complex datasets like the Swiss Roll. The Swiss Roll is a set of data points curled up in three-dimensional space, resembling a rolled-up sheet of paper. But beneath its twisty appearance, it actually has a simple, two-dimensional structure waiting to be discovered.

Traditional Principal Component Analysis (PCA) is excellent at finding patterns, but it can only see the world in straight lines. It flattens data along straight axes that capture the most variance, which means it misses the deeper, curved relationships hidden within structures like the Swiss Roll.

**Enter Kernel PCA**—a more sophisticated version of PCA that doesn't just look for straight lines, but can also "see" and unravel complex, nonlinear shapes. It accomplishes this using a mathematical technique called the "kernel trick," which allows it to compare data points as if they were mapped into a much higher-dimensional space. In this new perspective, those complicated curves and spirals often become straight and easy to separate.

With Kernel PCA, the Swiss Roll can be "unrolled" onto the plate, revealing its simple two-dimensional form. This makes patterns and relationships much more visible and understandable—something that standard PCA simply cannot achieve.

## When Should You Use Kernel PCA?

Consider Kernel PCA when you encounter:
- **Circular or spiral patterns** in your data visualization
- **Poor separation** between known groups using standard PCA
- **Non-linear relationships** that seem obvious but aren't captured by linear methods
- **Complex data types** like spectroscopy, where relationships aren't simply linear

Common applications include:
- Image recognition and computer vision
- Spectroscopic data analysis
- Biological data with complex interactions
- Any dataset where you suspect "hidden" curved structures

## Available Kernels: Different Lenses for Your Data

Kernel PCA works with different "flavors" of kernels—each giving the algorithm its own way of measuring similarity and unfolding the data:

### RBF (Radial Basis Function) Kernel
The **RBF kernel** is the Swiss Army knife of kernels—versatile and powerful for most non-linear problems:
- **Best for**: General non-linear patterns, circular data, unknown relationships
- **Key parameter**: `gamma` controls how "flexible" the unrolling is
  - Small gamma (0.001-0.01): Gentle, smooth unrolling
  - Large gamma (0.1-10): Tight, local unrolling
  - **Default**: 1/number_of_features (automatically calculated in GoPCA)

### Linear Kernel
The **linear kernel** gives you standard PCA:
- **Best for**: When you want to compare Kernel PCA with regular PCA
- **No parameters needed**

### Polynomial Kernel
The **polynomial kernel** finds polynomial relationships:
- **Best for**: Data with known polynomial patterns
- **Parameters**: 
  - `degree`: How complex the polynomial (2=quadratic, 3=cubic)
  - `gamma`: Scaling factor
  - `coef0`: Independent term

## Using Kernel PCA in GoPCA

### GUI Quick Start

1. **Load your data** using the file upload or sample datasets (try Swiss Roll!)
2. **Select "Kernel PCA"** from the Method dropdown
3. **Choose your kernel**:
   - For most cases, start with RBF
   - The gamma parameter is automatically set to 1/n_features when you load data
4. **Click "Run PCA Analysis"** and explore your unrolled data!

### CLI Examples

Basic RBF Kernel PCA (with automatic gamma):
```bash
gopca-cli analyze --method kernel --kernel-type rbf data.csv
```

Specify custom gamma for tighter patterns:
```bash
gopca-cli analyze --method kernel --kernel-type rbf --kernel-gamma 0.5 data.csv
```

Try polynomial kernel for known polynomial relationships:
```bash
gopca-cli analyze --method kernel --kernel-type poly \
  --kernel-degree 3 --kernel-gamma 0.1 data.csv
```

## Understanding Your Results

### The Scores Plot
- Shows your data points in the new "unrolled" space
- Points that were tangled in curves now appear separated
- Use for clustering, outlier detection, and pattern recognition

### Explained Variance
- Tells you how much pattern each component captures
- Different from regular PCA (measured in kernel space)
- Still useful for choosing how many components to keep

### What About Loadings?
Unlike regular PCA, Kernel PCA doesn't produce traditional loadings. Think of it this way: when you unroll the Swiss Roll, you're not just rotating your viewpoint (like regular PCA does)—you're fundamentally transforming the space. This transformation is too complex to express as simple variable contributions.

## Practical Tips

### Starting Parameters
1. **Always start with RBF kernel** unless you have specific knowledge about your data
2. **Use the default gamma** (1/n_features) as your starting point
3. **Try 2-3 components first** to visualize the main patterns

### Preprocessing Options
Kernel PCA handles its own centering in kernel space, so avoid:
- ❌ Mean centering
- ❌ Standard scaling (includes centering)
- ❌ Robust scaling (includes centering)

Compatible preprocessing:
- ✅ **Variance scaling only**: Useful when features have very different scales
- ✅ **SNV**: For spectroscopic data
- ✅ **Vector normalization**: For normalized comparisons

### Performance Considerations
- Works great for datasets up to ~5,000 samples
- For larger datasets (>10,000 samples), consider:
  - Sampling your data first
  - Starting with regular PCA for initial exploration
  - Using variance scaling to improve numerical stability

## Example: Exploring the Swiss Roll

Let's unroll the Swiss Roll dataset included with GoPCA:

```bash
# Using the GUI:
# 1. Click "Swiss Roll" in the sample datasets
# 2. Observe how gamma was automatically set to 0.333 (1/3 features)
# 3. Select "Kernel PCA" with RBF kernel
# 4. Click "Run PCA Analysis"
# 5. See the beautifully unrolled 2D structure!

# Using the CLI:
gopca-cli analyze --method kernel --kernel-type rbf \
  --kernel-gamma 0.333 swiss_roll.csv -o swiss_roll_unrolled.json
```

Compare this with regular PCA to see the dramatic difference:
```bash
gopca-cli analyze --method svd swiss_roll.csv -o swiss_roll_linear.json
```

## Summary

Kernel PCA is your tool for discovering hidden simplicity in complex data. Like unrolling a Swiss Roll to reveal its true two-dimensional nature, Kernel PCA helps you see through the twists and turns of non-linear data to find the meaningful patterns underneath.

Remember: 
- Start simple (RBF kernel, default gamma)
- Visualize your results
- Compare with regular PCA to appreciate the difference
- Let your data's structure guide your parameter choices

The Swiss Roll is just one delicious example of how Kernel PCA can turn a twisted problem into a straightforward insight. Your data might be hiding its own rolled-up patterns, waiting to be discovered!

## Technical References

For those interested in the mathematical details:
- Schölkopf, B., Smola, A., & Müller, K. R. (1998). Nonlinear component analysis as a kernel eigenvalue problem
- Implementation details in `internal/core/kernel_pca.go`