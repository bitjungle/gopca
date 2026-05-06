# David Thompson - 2014 - Nonlinear Dimensionality Reduction: KPCA

**Source:** JPL-Caltech Virtual Summer School on Big Data Analytics  
**Video:** https://www.youtube.com/watch?v=HbDHohXPLnU  
**Speaker:** David Thompson, NASA Jet Propulsion Laboratory / Caltech  

---

## Summary

This lecture introduces nonlinear dimensionality reduction, with Kernel PCA (KPCA) as the central method.

### Why linear methods fail

Linear PCA finds directions of maximum variance in the original feature space. When data lies on a curved manifold (e.g., a Swiss Roll, concentric rings, or a spiral), the principal components cut across the manifold rather than following its intrinsic geometry. The result is poor separation and lost structure.

### The kernel trick

KPCA extends PCA by implicitly mapping data into a high-dimensional feature space φ(x) where the manifold becomes approximately linear. Instead of computing φ(x) explicitly, KPCA uses a **kernel function** k(xᵢ, xⱼ) = φ(xᵢ)·φ(xⱼ) that evaluates dot products in feature space without ever constructing φ. The PCA eigenvalue problem is then solved on the N×N kernel matrix K rather than the p×p covariance matrix.

**Common kernels:**
- Gaussian (RBF): k(x,z) = exp(−γ‖x−z‖²) — most widely used; γ controls neighbourhood width
- Polynomial: k(x,z) = (x·z + c)^d
- Linear: k(x,z) = x·z — reduces to standard PCA

### Practical guidance

- KPCA is most useful when linear PCA produces poor scores-plot separation or when the data structure is known to be nonlinear (e.g., spectroscopic calibration curves, image manifolds)
- The γ parameter (RBF bandwidth) must be tuned — too small gives no generalisation, too large collapses to a linear method
- KPCA solves an N×N eigenproblem; for large N (>10,000 samples) this becomes the main computational bottleneck
- Nonlinear methods require more data to be reliable: use KPCA only when the dataset is large enough to sample the manifold densely

### Key takeaway

Use linear PCA first. If the scores plot shows curved or folded structure that cannot be resolved by rotating components, switch to Kernel PCA with an RBF kernel. Start with γ = 1/p (number of features) and adjust from there.
