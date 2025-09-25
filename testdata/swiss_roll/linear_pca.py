import numpy as np
import matplotlib.pyplot as plt
from sklearn.datasets import make_swiss_roll
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler

# Generate Swiss Roll data
X_raw, color = make_swiss_roll(n_samples=1000, noise=0.1, random_state=42)

# Apply mean centering and variance scaling
scaler = StandardScaler(with_mean=True, with_std=True)
X_scaled = scaler.fit_transform(X_raw)

# Use scaled data
X = X_scaled

# Linear PCA
pca = PCA(n_components=2)
X_pca = pca.fit_transform(X)

# Plotting
plt.figure(figsize=(6, 5))
plt.title("Linear PCA on Scaled Swiss Roll", fontsize=14)
plt.scatter(X_pca[:, 0], X_pca[:, 1], c=color, cmap='Spectral')
plt.xlabel("PC1")
plt.ylabel("PC2")
plt.grid(True, alpha=0.3)
plt.tight_layout()
plt.show()
