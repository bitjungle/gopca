import sys
import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler

# Load the wine dataset
wine_df = pd.read_csv('wine.csv', index_col=0)

# Use StandardScaler to normalize the data
scaler = StandardScaler()
wine_df.iloc[:, :13] = scaler.fit_transform(wine_df.iloc[:, :13])

# Perform PCA on the features
pca = PCA(n_components=2)
scores = pca.fit_transform(wine_df.iloc[:, :13])

# Plot PCA scores and loadings in subplots
fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 6))

# Left subplot: PCA scores
# Get unique species and assign colors
targets = wine_df['target']
unique_targets = targets.unique()
colors = ['red', 'blue', 'green']
color_map = dict(zip(unique_targets, colors))

# Create scatter plot with colors based on species
for i, spec in enumerate(unique_targets):
    mask = targets == spec
    ax1.scatter(scores[mask, 0], scores[mask, 1], 
               c=colors[i], label=spec, alpha=0.7)

# Add annotations for sample indices
for i, txt in enumerate(wine_df.index):
    ax1.annotate(txt, (scores[i, 0], scores[i, 1]), fontsize=6, alpha=0.7)

ax1.set_title('PCA Scores of Wine Data')
ax1.set_xlabel('PCA Component 1')
ax1.set_ylabel('PCA Component 2')
ax1.legend()
ax1.grid()

# Right subplot: PCA loadings
loadings = pca.components_.T * np.sqrt(pca.explained_variance_)
ax2.bar(range(len(loadings)), loadings[:, 0], label='PC1 Loadings', alpha=0.7)
ax2.bar(range(len(loadings)), loadings[:, 1], label='PC2 Loadings', alpha=0.7)
ax2.set_title('PCA Loadings')
ax2.set_xlabel('Feature Index')
ax2.set_ylabel('Loading Value')
ax2.legend()
ax2.grid()

plt.tight_layout()
plt.show()
