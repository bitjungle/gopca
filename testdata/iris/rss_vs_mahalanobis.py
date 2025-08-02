import sys
import os
import pandas as pd
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler
import matplotlib.pyplot as plt

# Add parent directory to Python path
sys.path.append(os.path.abspath(os.path.join(os.path.dirname(__file__), "..")))
# Import utility functions
from utils.utils import snv, calculate_rss, calculate_mahalanobis_distance

# Load the iris dataset
iris_df = pd.read_csv('iris.csv', index_col=0)

# Use StandardScaler to normalize the data
#scaler = StandardScaler()
#iris_df.iloc[:, :4] = scaler.fit_transform(iris_df.iloc[:, :4])

# Perform PCA on the first four columns (features)
pca = PCA(n_components=2)
scores = pca.fit_transform(iris_df.iloc[:, :4])

# Calculate RSS and Mahalanobis distance
md_values = calculate_mahalanobis_distance(scores)
rss_values = calculate_rss(iris_df.iloc[:, :4], pca, scores)

# Make a scatter plot of RSS vs. md 
plt.figure(figsize=(10, 6))
plt.scatter(rss_values, md_values, alpha=0.7)
plt.title('RSS vs. Mahalanobis Distance')
plt.xlabel('Residual Sum of Squares (RSS)')
plt.ylabel('Mahalanobis Distance')
plt.grid()
plt.show()
