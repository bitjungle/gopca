import pandas as pd
import numpy as np
from scipy.spatial.distance import mahalanobis

def snv(X:pd.DataFrame) -> pd.DataFrame:
    '''Standard Normal Variate

    Barnes, R. J., Dhanoa, M. S., & Lister, S. J. (1989). 
      Standard Normal Variate Transformation and De-Trending of Near-Infrared 
      Diffuse Reflectance Spectra. Applied Spectroscopy, 43(5), 772–777. 
      https://doi.org/10.1366/0003702894202201
    '''
    sample_mean = X.mean(axis=1)
    sample_std = X.std(axis=1)
    X_new = ((X.T - sample_mean)/sample_std).T
    return pd.DataFrame(data=X_new, index=X.index, columns=X.columns)

# Define a function for Vector Normalization (L2 Norm)
def vector_normalization(X: pd.DataFrame) -> pd.DataFrame:
    """Vector Normalization (L2 Norm)"""
    norm = np.linalg.norm(X, axis=1, keepdims=True)
    return X.div(norm, axis=0)

def calculate_rss(original_data, pca_model, pca_scores):
    '''Calculate Residual Sum of Squares (RSS) for PCA reconstruction
    
    RSS measures the difference between original data and PCA reconstruction.
    Lower RSS values indicate better reconstruction quality.
    
    Args:
        original_data: Original preprocessed data
        pca_model: Fitted PCA model
        pca_scores: PCA scores (transformed data)
    
    Returns:
        List of RSS values for each sample
    '''
    # Reconstruct the data using PCA inverse transform
    reconstructed_data = pca_model.inverse_transform(pca_scores)
    
    # Calculate RSS for each sample (sum of squared residuals)
    rss_list = []
    for i in range(len(original_data)):
        residuals = original_data.iloc[i] - reconstructed_data[i]
        rss = np.sum(residuals ** 2)
        rss_list.append(rss)
    
    return [float(x) for x in rss_list]

def calculate_mahalanobis_distance(scores):
    '''Calculate Mahalanobis distance for PCA scores
    
    Args:
        scores: PCA scores (transformed data)
    
    Returns:
        List of Mahalanobis distances for each sample
    '''
    mean_vec = scores.mean(axis=0)
    cov = np.cov(scores, rowvar=False)
    inv_covmat = np.linalg.inv(cov)
    md = [mahalanobis(row, mean_vec, inv_covmat) for row in scores]
    return [float(x) for x in md]
