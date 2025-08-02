#!/usr/bin/env python3
"""
Understand sklearn's sign convention for PCA.
"""

import numpy as np
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler

# Create a simple test dataset where we know what to expect
np.random.seed(42)
X = np.array([
    [1, 2],
    [2, 4],
    [3, 6],
    [4, 8],
    [5, 10]
])

print("Test data:")
print(X)
print("\nThis data lies on a line y = 2x, so PC1 should be along this line")

# Standardize
scaler = StandardScaler()
X_scaled = scaler.fit_transform(X)

# PCA
pca = PCA()
scores = pca.fit_transform(X_scaled)
loadings = pca.components_

print("\nPCA results:")
print(f"PC1 loadings: {loadings[0]}")
print(f"PC2 loadings: {loadings[1]}")

print("\nScores:")
for i, score in enumerate(scores):
    print(f"Point {i}: PC1={score[0]:.4f}, PC2={score[1]:.4f}")

print("\n\nSklearn's sign convention:")
print("According to sklearn documentation, the components_ matrix has shape (n_components, n_features)")
print("Each row is an eigenvector.")
print("\nSklearn uses SVD decomposition: X = U * S * V^T")
print("The components_ are the rows of V (right singular vectors)")
print("The signs are chosen so that the largest element in each component has a positive sign")
print("This is done for reproducibility across different runs and platforms")