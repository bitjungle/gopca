import numpy as np
import matplotlib.pyplot as plt
from sklearn.datasets import make_swiss_roll
from sklearn.decomposition import KernelPCA
from sklearn.preprocessing import StandardScaler

# Generate Swiss Roll data
X_raw, color = make_swiss_roll(n_samples=1000, noise=0.1, random_state=42)
X_scaled = StandardScaler(with_mean=False, with_std=True).fit_transform(X_raw)

use_scaled = True
X = X_scaled if use_scaled else X_raw

# Polynomial kernel settings
degrees = [2, 3, 4]
gammas = [0.01, 0.1, 1.0]
kernel = "poly"
coef0 = 1  # Typically set to 1 to avoid homogeneous polynomial

# Plotting grid: rows = degrees, cols = gamma
fig, axes = plt.subplots(len(degrees), len(gammas), figsize=(15, 10))
fig.suptitle(f"scikit-learn: Kernel PCA (polynomial kernel), scaled={use_scaled}", fontsize=16)

# Iterate over degree × gamma
for i, degree in enumerate(degrees):
    for j, gamma in enumerate(gammas):
        kpca = KernelPCA(n_components=2, kernel=kernel, degree=degree, gamma=gamma, coef0=coef0)
        X_kpca = kpca.fit_transform(X)

        ax = axes[i, j]
        sc = ax.scatter(X_kpca[:, 0], X_kpca[:, 1], c=color, cmap='Spectral', s=10)
        ax.set_title(f"degree={degree}, γ={gamma}")
        ax.set_xlabel("KPC1")
        ax.set_ylabel("KPC2")
        ax.grid(True, alpha=0.3)

plt.tight_layout(rect=[0, 0, 1, 0.96])
plt.show()
