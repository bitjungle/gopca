import sys
import os
import pandas as pd
from sklearn.decomposition import PCA
import matplotlib.pyplot as plt

# Add parent directory to Python path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))
# Import utility functions
from utils.utils import snv, calculate_rss, calculate_mahalanobis_distance

# Load and preprocess the corn spectra data
spectra_df = pd.read_csv('corn_m5spec.csv', sep=',')
spectra_processed_df = snv(spectra_df)

# Fit PCA model
pca = PCA(n_components=2)
scores = pca.fit_transform(spectra_processed_df)

# Calculate RSS and Mahalanobis distance
md_values = calculate_mahalanobis_distance(scores)
rss_values = calculate_rss(spectra_processed_df, pca, scores)

# Make a scatter plot of RSS vs. md 
plt.figure(figsize=(10, 6))
plt.scatter(rss_values, md_values, alpha=0.7)
plt.title('RSS vs. Mahalanobis Distance')
plt.xlabel('Residual Sum of Squares (RSS)')
plt.ylabel('Mahalanobis Distance')
plt.grid()
plt.show()
