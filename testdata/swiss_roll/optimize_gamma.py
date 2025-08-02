import numpy as np
import matplotlib.pyplot as plt
from sklearn.datasets import make_swiss_roll
from sklearn.decomposition import KernelPCA

# Generate Swiss Roll data
X, color = make_swiss_roll(n_samples=1000, noise=0.1, random_state=42)

# Kernel PCA parameters
gamma_values = [0.33, 0.1, 0.05, 0.01]
kernel = "rbf"

# Plotting
fig, axes = plt.subplots(1, 4, figsize=(16, 4))

# Loop over gamma values and create subplots
for i, gamma in enumerate(gamma_values):
    # Kernel PCA with RBF kernel
    kpca = KernelPCA(n_components=2, kernel=kernel, gamma=gamma)
    X_kpca = kpca.fit_transform(X)
    
    # Plot Kernel PCA result
    axes[i].scatter(X_kpca[:, 0], X_kpca[:, 1], c=color, cmap='Spectral')
    axes[i].set_title(f"Kernel PCA (γ={gamma})")
    axes[i].set_xlabel("KPC1")
    axes[i].set_ylabel("KPC2")
    axes[i].grid(True, alpha=0.3)

plt.tight_layout()
plt.show()
