import sys
import os
import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from sklearn.decomposition import PCA

# Add parent directory to Python path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))
# Import utility functions
from utils.utils import snv, vector_normalization

preprocess_type = 'snv'  # Choose 'snv' or 'vector_normalization'

# Load the corn spectra data
spectra_df = pd.read_csv('corn_m5spec.csv', sep=',')

# Apply SNV normalization to the spectral data
if preprocess_type == 'snv':
    spectra_processed_df = snv(spectra_df)
elif preprocess_type == 'vector_normalization':
    spectra_processed_df = vector_normalization(spectra_df)
else:
    spectra_processed_df = spectra_df.copy()

# Plot spectral data - spectras are in rows 
fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 6))

# Left subplot: Raw spectra
for i in range(spectra_df.shape[0]):
    ax1.plot(spectra_df.iloc[i], label=f'Sample {i+1}', alpha=0.5)
ax1.set_title('Raw Spectral Data')
ax1.set_xlabel('Wavelength Index')
ax1.set_ylabel('Reflectance')
ax1.grid()
# Set x-ticks every 100 variables
ax1.set_xticks(range(0, spectra_df.shape[1], 100))

# Right subplot: Processed spectra
for i in range(spectra_processed_df.shape[0]):
    ax2.plot(spectra_processed_df.iloc[i], label=f'Sample {i+1}', alpha=0.5)
ax2.set_title(f'Processed Spectral Data ({preprocess_type.upper()})')
ax2.set_xlabel('Wavelength Index')
ax2.set_ylabel('Reflectance')
ax2.grid()
# Set x-ticks every 100 variables
ax2.set_xticks(range(0, spectra_processed_df.shape[1], 100))

plt.tight_layout()
plt.show()

pca = PCA(n_components=2)
scores = pca.fit_transform(spectra_processed_df)

# Plot PCA scores and loadings in subplots
fig, (ax1, ax2) = plt.subplots(1, 2, figsize=(16, 6))

# Left subplot: PCA scores
ax1.scatter(scores[:, 0], scores[:, 1])
for i, txt in enumerate(spectra_processed_df.index):
    ax1.annotate(txt, (scores[i, 0], scores[i, 1]), fontsize=8)
ax1.set_title('PCA Scores of Corn Spectral Data')
ax1.set_xlabel('PCA Component 1')
ax1.set_ylabel('PCA Component 2')
ax1.grid()

# Right subplot: PCA loadings
loadings = pca.components_.T * np.sqrt(pca.explained_variance_)
ax2.plot(range(len(loadings)), loadings[:, 0], label='PC1 Loadings', linewidth=2)
ax2.plot(range(len(loadings)), loadings[:, 1], label='PC2 Loadings', linewidth=2)
ax2.set_title('PCA Loadings')
ax2.set_xlabel('Wavelength Index')
ax2.set_ylabel('Loading Value')
ax2.legend()
ax2.grid()
# Set x-ticks every 100 variables
ax2.set_xticks(range(0, len(loadings), 100))

plt.tight_layout()
plt.show()

# print(md_clean)