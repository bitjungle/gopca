#!/usr/bin/env python3
"""
Generate reference Temporal PCA (SSA) results for validation.

This script implements Singular Spectrum Analysis (SSA) for time series,
which is the basis of our Temporal PCA implementation.

References:
- Golyandina et al. (2015): Multivariate and 2D Extensions of SSA with Rssa
- Ghil et al. (2002): Advanced Spectral Methods for Climatic Time Series
- Broomhead & King (1986): Extracting qualitative dynamics from experimental data
"""

import sys
import os
import json
import numpy as np
import pandas as pd
from scipy import linalg

# Add parent directory to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


def build_trajectory_matrix(X, window_length):
    """
    Build trajectory (Hankel) matrix for SSA.
    
    For multivariate time series, creates a stacked Hankel matrix.
    
    Parameters:
    -----------
    X : array-like
        Time series data (n_timepoints, n_variables)
    window_length : int
        Window length (number of lags)
        
    Returns:
    --------
    trajectory_matrix : array
        Trajectory matrix with lagged vectors
    """
    if len(X.shape) == 1:
        X = X.reshape(-1, 1)
    
    n_timepoints, n_variables = X.shape
    K = n_timepoints - window_length + 1  # Number of lagged vectors
    
    if K < 1:
        raise ValueError(f"Window length {window_length} too large for series length {n_timepoints}")
    
    # Build trajectory matrix
    trajectory_matrix = np.zeros((K, window_length * n_variables))
    
    for i in range(K):
        # Extract window and flatten (variables are concatenated)
        window = X[i:i+window_length, :].flatten()
        trajectory_matrix[i, :] = window
    
    return trajectory_matrix


def diagonal_averaging(matrix, original_shape, window_length):
    """
    Perform diagonal averaging (hankelization) to reconstruct time series.
    
    Parameters:
    -----------
    matrix : array
        Matrix to be diagonally averaged
    original_shape : tuple
        Original time series shape (n_timepoints, n_variables)
    window_length : int
        Window length used in trajectory matrix
        
    Returns:
    --------
    reconstructed : array
        Reconstructed time series
    """
    n_timepoints, n_variables = original_shape
    K = n_timepoints - window_length + 1
    
    reconstructed = np.zeros(original_shape)
    counts = np.zeros(n_timepoints)
    
    # Reshape matrix to separate variables
    matrix_reshaped = matrix.reshape(K, window_length, n_variables)
    
    for i in range(K):
        for j in range(window_length):
            reconstructed[i+j, :] += matrix_reshaped[i, j, :]
            counts[i+j] += 1
    
    # Average
    for i in range(n_timepoints):
        if counts[i] > 0:
            reconstructed[i, :] /= counts[i]
    
    return reconstructed


def generate_temporal_pca_reference(data, window_length, n_components=None,
                                   preprocessing='mean_center'):
    """
    Generate Temporal PCA (SSA) reference results.
    
    Parameters:
    -----------
    data : array-like
        Time series data (n_timepoints, n_variables)
    window_length : int
        Window length for SSA (number of lags)
    n_components : int, optional
        Number of components to compute
    preprocessing : str
        Preprocessing method: 'none', 'mean_center', 'standardize'
        
    Returns:
    --------
    dict : Reference results including scores, singular values, reconstructed components
    """
    X = np.array(data, dtype=np.float64)
    if len(X.shape) == 1:
        X = X.reshape(-1, 1)
    
    n_timepoints, n_variables = X.shape
    
    # Store preprocessing parameters
    preprocessing_params = {}
    
    # Apply preprocessing
    if preprocessing == 'mean_center':
        means = X.mean(axis=0)
        X_processed = X - means
        preprocessing_params['means'] = means.tolist()
        
    elif preprocessing == 'standardize':
        means = X.mean(axis=0)
        stds = X.std(axis=0)
        stds[stds == 0] = 1.0  # Avoid division by zero
        X_processed = (X - means) / stds
        preprocessing_params['means'] = means.tolist()
        preprocessing_params['stds'] = stds.tolist()
        
    else:  # 'none'
        X_processed = X.copy()
    
    # Build trajectory matrix
    trajectory_matrix = build_trajectory_matrix(X_processed, window_length)
    K, M = trajectory_matrix.shape  # K = n_timepoints - window_length + 1, M = window_length * n_variables
    
    if n_components is None:
        n_components = min(K, M)
    
    # Perform SVD
    U, s, Vt = linalg.svd(trajectory_matrix, full_matrices=False)
    
    # Limit to requested components
    n_components = min(n_components, len(s))
    U = U[:, :n_components]
    s = s[:n_components]
    Vt = Vt[:n_components, :]
    
    # Calculate scores (temporal coefficients)
    scores = U * s
    
    # Calculate loadings (eigenvectors showing temporal patterns)
    loadings = Vt.T
    
    # Calculate explained variance
    eigenvalues = s**2 / (K - 1)
    total_variance = np.sum(eigenvalues)
    explained_variance_ratio = eigenvalues / total_variance
    cumulative_variance = np.cumsum(explained_variance_ratio)
    
    # Reconstruct components
    reconstructed_components = []
    for i in range(n_components):
        # Reconstruct from single component
        component_matrix = s[i] * np.outer(U[:, i], Vt[i, :])
        reconstructed = diagonal_averaging(component_matrix, X_processed.shape, window_length)
        reconstructed_components.append(reconstructed.tolist())
    
    # Group components for trend and oscillations (simple grouping)
    # Component 0 is typically trend, pairs of components often form oscillations
    trend_reconstruction = None
    if n_components > 0:
        trend_matrix = s[0] * np.outer(U[:, 0], Vt[0, :])
        trend_reconstruction = diagonal_averaging(trend_matrix, X_processed.shape, window_length)
    
    # Calculate reconstruction error
    full_reconstruction = np.zeros_like(X_processed)
    for i in range(n_components):
        component_matrix = s[i] * np.outer(U[:, i], Vt[i, :])
        full_reconstruction += diagonal_averaging(component_matrix, X_processed.shape, window_length)
    
    reconstruction_error = np.mean((X_processed - full_reconstruction)**2)
    
    # Verify Hankel structure preservation
    # Check if the trajectory matrix has the expected Hankel structure
    hankel_structure_check = verify_hankel_structure(trajectory_matrix, X_processed, window_length)
    
    # Create result dictionary
    result = {
        'method': 'temporal_pca_ssa',
        'window_length': window_length,
        'preprocessing': preprocessing,
        'preprocessing_params': preprocessing_params,
        'n_timepoints': n_timepoints,
        'n_variables': n_variables,
        'n_components': n_components,
        'trajectory_matrix_shape': [K, M],
        'scores': scores.tolist(),
        'loadings': loadings.tolist(),
        'singular_values': s.tolist(),
        'eigenvalues': eigenvalues.tolist(),
        'explained_variance_ratio': explained_variance_ratio.tolist(),
        'cumulative_variance': cumulative_variance.tolist(),
        'total_variance': float(total_variance),
        'reconstructed_components': reconstructed_components,
        'reconstruction_error': float(reconstruction_error),
        'hankel_structure_valid': hankel_structure_check
    }
    
    if trend_reconstruction is not None:
        result['trend_component'] = trend_reconstruction.tolist()
    
    return result


def verify_hankel_structure(trajectory_matrix, original_data, window_length):
    """
    Verify that the trajectory matrix has proper Hankel structure.
    
    Returns:
    --------
    bool : True if Hankel structure is valid
    """
    if len(original_data.shape) == 1:
        original_data = original_data.reshape(-1, 1)
    
    n_timepoints, n_variables = original_data.shape
    K = n_timepoints - window_length + 1
    
    # Check a few elements to verify Hankel structure
    for i in range(min(5, K-1)):
        for j in range(min(5, window_length-1)):
            for v in range(n_variables):
                idx1 = j * n_variables + v
                idx2 = (j + 1) * n_variables + v
                
                # Elements on the anti-diagonal should be equal
                val1 = trajectory_matrix[i, idx2]
                val2 = trajectory_matrix[i + 1, idx1]
                
                if abs(val1 - val2) > 1e-10:
                    return False
    
    return True


def generate_synthetic_time_series():
    """
    Generate synthetic time series for SSA validation.
    """
    np.random.seed(42)
    t = np.linspace(0, 4*np.pi, 200)
    
    series_list = []
    
    # 1. Pure sine wave
    sine = np.sin(t)
    series_list.append(('sine_wave', sine))
    
    # 2. Trend + seasonal
    trend = 0.1 * t
    seasonal = np.sin(2*t) + 0.5 * np.cos(4*t)
    trend_seasonal = trend + seasonal + 0.1 * np.random.randn(len(t))
    series_list.append(('trend_seasonal', trend_seasonal))
    
    # 3. Multiple frequencies
    multi_freq = np.sin(t) + 0.5 * np.sin(3*t) + 0.3 * np.sin(7*t) + 0.1 * np.random.randn(len(t))
    series_list.append(('multi_frequency', multi_freq))
    
    # 4. Exponentially damped oscillation
    damped = np.exp(-0.01 * t) * np.sin(2*t)
    series_list.append(('damped_oscillation', damped))
    
    # 5. Multivariate series
    multivariate = np.column_stack([
        np.sin(t) + 0.1 * np.random.randn(len(t)),
        np.cos(t) + 0.1 * np.random.randn(len(t)),
        0.5 * np.sin(2*t) + 0.1 * np.random.randn(len(t))
    ])
    series_list.append(('multivariate', multivariate))
    
    return series_list


def validate_ssa_properties(result):
    """
    Validate mathematical properties specific to SSA/Temporal PCA.
    
    Returns:
    --------
    dict : Validation results
    """
    scores = np.array(result['scores'])
    singular_values = np.array(result['singular_values'])
    eigenvalues = np.array(result['eigenvalues'])
    
    validations = {}
    
    # 1. Check singular values are non-negative and descending
    validations['singular_values_non_negative'] = bool(np.all(singular_values >= 0))
    validations['singular_values_descending'] = bool(np.all(np.diff(singular_values) <= 0))
    
    # 2. Check eigenvalues are non-negative
    validations['eigenvalues_non_negative'] = bool(np.all(eigenvalues >= -1e-10))
    
    # 3. Check Hankel structure was preserved
    validations['hankel_structure_preserved'] = result['hankel_structure_valid']
    
    # 4. Check explained variance sums to <= 1
    total_explained = np.sum(result['explained_variance_ratio'])
    validations['total_explained_variance'] = float(total_explained)
    validations['variance_sum_valid'] = bool(total_explained <= 1.01)
    
    # 5. Check trajectory matrix dimensions
    K, M = result['trajectory_matrix_shape']
    n_timepoints = result['n_timepoints']
    n_variables = result['n_variables']
    window_length = result['window_length']
    
    expected_K = n_timepoints - window_length + 1
    expected_M = window_length * n_variables
    
    validations['trajectory_matrix_dims_correct'] = bool(K == expected_K and M == expected_M)
    
    return validations


def main():
    """
    Generate Temporal PCA (SSA) reference results for validation.
    """
    print("Generating Temporal PCA (SSA) reference results...")
    
    # Ensure output directory exists
    output_dir = os.path.join(os.path.dirname(__file__), 'reference_results')
    os.makedirs(output_dir, exist_ok=True)
    
    # Test with stock data (time series)
    stocks_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), 
                               'stocks', 'stock-data-AAPL.csv')
    
    if os.path.exists(stocks_path):
        print("\nProcessing stock price data...")
        df = pd.read_csv(stocks_path)
        
        # Use closing prices
        if 'Close' in df.columns:
            prices = df['Close'].values
            
            # Test different window lengths
            window_lengths = [10, 20, 50]
            
            for window_length in window_lengths:
                if window_length < len(prices):
                    print(f"  - Window length: {window_length}")
                    
                    result = generate_temporal_pca_reference(
                        prices, 
                        window_length=window_length,
                        n_components=min(10, window_length),
                        preprocessing='mean_center'
                    )
                    
                    validations = validate_ssa_properties(result)
                    result['validations'] = validations
                    
                    output_file = f"reference_results/stocks_temporal_pca_w{window_length}.json"
                    output_path = os.path.join(os.path.dirname(__file__), output_file)
                    
                    with open(output_path, 'w') as f:
                        json.dump(result, f, indent=2)
                    
                    print(f"    Saved to {output_file}")
                    print(f"    First 3 singular values: {result['singular_values'][:3]}")
    
    # Generate synthetic time series tests
    print("\nGenerating synthetic time series tests...")
    synthetic_series = generate_synthetic_time_series()
    
    for series_name, series_data in synthetic_series:
        print(f"  - {series_name}")
        
        # Use window length = 1/4 of series length
        if len(series_data.shape) == 1:
            series_length = len(series_data)
        else:
            series_length = series_data.shape[0]
        
        window_length = max(10, series_length // 4)
        
        result = generate_temporal_pca_reference(
            series_data,
            window_length=window_length,
            n_components=min(10, window_length),
            preprocessing='mean_center'
        )
        
        validations = validate_ssa_properties(result)
        result['validations'] = validations
        
        output_file = f"reference_results/synthetic_{series_name}_temporal_pca.json"
        output_path = os.path.join(os.path.dirname(__file__), output_file)
        
        with open(output_path, 'w') as f:
            json.dump(result, f, indent=2)
        
        print(f"    Saved to {output_file}")
        print(f"    Explained variance (first 3): {result['explained_variance_ratio'][:3]}")
    
    # Test with multivariate data (using iris as time series)
    print("\nTesting multivariate temporal PCA...")
    iris_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), 
                             'iris', 'iris.csv')
    
    if os.path.exists(iris_path):
        df = pd.read_csv(iris_path, index_col=0)
        numeric_cols = df.select_dtypes(include=[np.number]).columns
        X = df[numeric_cols].values
        
        # Treat as multivariate time series
        window_length = 20
        print(f"  - Iris data as multivariate series (window={window_length})")
        
        result = generate_temporal_pca_reference(
            X,
            window_length=window_length,
            n_components=8,
            preprocessing='standardize'
        )
        
        validations = validate_ssa_properties(result)
        result['validations'] = validations
        
        output_file = "reference_results/iris_temporal_pca_multivariate.json"
        output_path = os.path.join(os.path.dirname(__file__), output_file)
        
        with open(output_path, 'w') as f:
            json.dump(result, f, indent=2)
        
        print(f"    Saved to {output_file}")
        print(f"    Trajectory matrix shape: {result['trajectory_matrix_shape']}")
    
    print("\nTemporal PCA (SSA) reference generation complete!")


if __name__ == "__main__":
    main()