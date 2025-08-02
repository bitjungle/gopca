import numpy as np
import matplotlib.pyplot as plt
from sklearn.datasets import make_swiss_roll
from sklearn.decomposition import PCA, KernelPCA

# Generate Swiss Roll data
X, color = make_swiss_roll(n_samples=1000, noise=0.1, random_state=42)

# Kernel PCA parameters
gamma = 0.01
kernel = "rbf"

# Standard PCA (linear)
pca = PCA(n_components=2)
X_pca = pca.fit_transform(X)

# Kernel PCA with RBF kernel
kpca = KernelPCA(n_components=2, kernel=kernel, gamma=gamma)
X_kpca = kpca.fit_transform(X)

# Plotting
fig = plt.figure(figsize=(15, 4))

# 1. Original Swiss Roll in 3D
ax = fig.add_subplot(131, projection='3d')
ax.scatter(X[:, 0], X[:, 2], X[:, 1], c=color, cmap=plt.cm.Spectral)
ax.set_title("Original Swiss Roll (3D)")
ax.set_xlabel("X")
ax.set_ylabel("Z")
ax.set_zlabel("Y")

# 2. Linear PCA result
ax = fig.add_subplot(132)
ax.scatter(X_pca[:, 0], X_pca[:, 1], c=color, cmap=plt.cm.Spectral)
ax.set_title("Linear PCA (2D)")
ax.set_xlabel("PC1")
ax.set_ylabel("PC2")

# 3. Kernel PCA result
ax = fig.add_subplot(133)
ax.scatter(X_kpca[:, 0], X_kpca[:, 1], c=color, cmap=plt.cm.Spectral)
ax.set_title(f"Kernel PCA ({kernel}), gamma={gamma} (2D)")
ax.set_xlabel("KPC1")
ax.set_ylabel("KPC2")

plt.tight_layout()
plt.show()
