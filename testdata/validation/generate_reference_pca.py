#!/usr/bin/env python3
"""
Generate reference PCA results using scikit-learn for validation.

This script generates reference results that match the expected behavior
of standard PCA implementations (SVD and NIPALS methods).

References:
- Bro & Smilde (2014): Principal component analysis
- Jolliffe & Cadima (2016): Principal component analysis: A review
- Shlens (2014): A Tutorial on Principal Component Analysis
"""

import sys
import os
import json
import numpy as np
import pandas as pd
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler
from scipy import stats

# Add parent directory to path for utils
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


def generate_pca_reference(data, n_components=None, preprocessing='mean_center', 
                          method='svd', random_state=42):
    """
    Generate PCA reference results with various preprocessing options.
    
    Parameters:
    -----------
    data : array-like
        Input data matrix (n_samples, n_features)
    n_components : int, optional
        Number of components to compute
    preprocessing : str
        Preprocessing method: 'none', 'mean_center', 'standardize', 'robust'
    method : str
        PCA method: 'svd' (uses sklearn's default) or 'nipals' simulation
    random_state : int
        Random seed for reproducibility
        
    Returns:
    --------
    dict : Reference results including scores, loadings, explained variance, etc.
    """
    X = np.array(data, dtype=np.float64)
    n_samples, n_features = X.shape
    
    if n_components is None:
        n_components = min(n_samples, n_features)
    
    # Store preprocessing parameters
    preprocessing_params = {}
    
    # Apply preprocessing
    if preprocessing == 'mean_center':
        means = X.mean(axis=0)
        X_processed = X - means
        preprocessing_params['means'] = means.tolist()
        
    elif preprocessing == 'standardize':
        scaler = StandardScaler()
        X_processed = scaler.fit_transform(X)
        preprocessing_params['means'] = scaler.mean_.tolist()
        preprocessing_params['stds'] = scaler.scale_.tolist()
        
    elif preprocessing == 'robust':
        # Robust scaling using median and MAD
        medians = np.median(X, axis=0)
        mads = stats.median_abs_deviation(X, axis=0)
        mads[mads == 0] = 1.0  # Avoid division by zero
        X_processed = (X - medians) / mads
        preprocessing_params['medians'] = medians.tolist()
        preprocessing_params['mads'] = mads.tolist()
        
    else:  # 'none'
        X_processed = X.copy()
    
    # Perform PCA
    pca = PCA(n_components=n_components, svd_solver='full', random_state=random_state)
    scores = pca.fit_transform(X_processed)
    
    # Get loadings (components transposed)
    loadings = pca.components_.T
    
    # Calculate additional statistics
    eigenvalues = pca.explained_variance_
    explained_variance_ratio = pca.explained_variance_ratio_
    cumulative_variance = np.cumsum(explained_variance_ratio)
    
    # Calculate singular values (for comparison with Go implementation)
    # singular_value = sqrt(eigenvalue * (n-1))
    singular_values = np.sqrt(eigenvalues * (n_samples - 1))
    
    # Verify orthogonality of loadings
    loadings_orthogonality = np.max(np.abs(loadings.T @ loadings - np.eye(n_components)))
    
    # Calculate reconstruction error for each number of components
    reconstruction_errors = []
    for k in range(1, n_components + 1):
        X_reconstructed = scores[:, :k] @ loadings[:, :k].T
        if preprocessing in ['mean_center', 'standardize', 'robust']:
            error = np.mean((X_processed - X_reconstructed) ** 2)
        else:
            error = np.mean((X - X_reconstructed) ** 2)
        reconstruction_errors.append(error)
    
    # Create result dictionary
    result = {
        'method': 'sklearn_pca',
        'preprocessing': preprocessing,
        'preprocessing_params': preprocessing_params,
        'n_samples': n_samples,
        'n_features': n_features,
        'n_components': n_components,
        'scores': scores.tolist(),
        'loadings': loadings.tolist(),
        'eigenvalues': eigenvalues.tolist(),
        'singular_values': singular_values.tolist(),
        'explained_variance_ratio': explained_variance_ratio.tolist(),
        'cumulative_variance': cumulative_variance.tolist(),
        'total_variance': float(np.sum(eigenvalues)),
        'loadings_orthogonality_error': float(loadings_orthogonality),
        'reconstruction_errors': reconstruction_errors,
    }
    
    # Add condition number for numerical stability assessment
    if preprocessing != 'none' and n_features > 1:
        cov_matrix = np.cov(X_processed.T)
        if cov_matrix.ndim == 2:  # Only if we have a matrix
            eigenvals_cov = np.linalg.eigvalsh(cov_matrix)
            eigenvals_cov = eigenvals_cov[eigenvals_cov > 1e-10]  # Filter near-zero
            if len(eigenvals_cov) > 0:
                condition_number = eigenvals_cov[-1] / eigenvals_cov[0]
                result['condition_number'] = float(condition_number)
    
    return result


def validate_mathematical_properties(result):
    """
    Validate mathematical properties of PCA results.
    
    Returns:
    --------
    dict : Validation results
    """
    scores = np.array(result['scores'])
    loadings = np.array(result['loadings'])
    eigenvalues = np.array(result['eigenvalues'])
    explained_var = np.array(result['explained_variance_ratio'])
    
    n_samples, n_components = scores.shape
    
    validations = {}
    
    # 1. Check orthogonality of scores (normalized)
    scores_normalized = scores / np.sqrt(n_samples - 1)
    scores_orthogonality = np.max(np.abs(scores_normalized.T @ scores_normalized - np.diag(eigenvalues)))
    validations['scores_orthogonality_error'] = float(scores_orthogonality)
    
    # 2. Check eigenvalues are in descending order
    validations['eigenvalues_descending'] = bool(np.all(np.diff(eigenvalues) <= 0))
    
    # 3. Check explained variance sums to <= 1
    total_explained = np.sum(explained_var)
    validations['total_explained_variance'] = float(total_explained)
    validations['variance_sum_valid'] = bool(total_explained <= 1.01)  # Allow small numerical error
    
    # 4. Check cumulative variance is monotonic
    cumulative = result['cumulative_variance']
    validations['cumulative_monotonic'] = bool(np.all(np.diff(cumulative) >= 0))
    
    # 5. Verify Mahalanobis distance relationship (Brereton, 2015)
    # Mahalanobis distance squared = sum of squared standardized PC scores
    mahalanobis_distances = np.sum((scores / np.sqrt(eigenvalues)) ** 2, axis=1)
    validations['mean_mahalanobis_distance'] = float(np.mean(mahalanobis_distances))
    
    return validations


def generate_test_cases():
    """
    Generate various test cases for validation.
    """
    test_cases = []
    
    # 1. Well-conditioned random data
    np.random.seed(42)
    X_good = np.random.randn(50, 10)
    test_cases.append(('well_conditioned', X_good))
    
    # 2. Ill-conditioned data
    X_ill = np.random.randn(50, 10)
    X_ill[:, -1] = X_ill[:, 0] * 1e-8  # Make last column nearly collinear
    test_cases.append(('ill_conditioned', X_ill))
    
    # 3. Rank-deficient data
    X_rank = np.random.randn(20, 10)
    X_rank[:, 5:] = X_rank[:, :5] @ np.random.randn(5, 5)  # Make rank 5
    test_cases.append(('rank_deficient', X_rank))
    
    # 4. More variables than samples
    X_wide = np.random.randn(10, 30)
    test_cases.append(('wide_data', X_wide))
    
    # 5. Single variable
    X_single = np.random.randn(50, 1)
    test_cases.append(('single_variable', X_single))
    
    return test_cases


def main():
    """
    Generate reference results for all test datasets.
    """
    print("Generating PCA reference results...")
    
    # Test standard datasets
    datasets = [
        ('iris', 'iris/iris.csv', 4),
        ('wine', 'wine/wine.csv', 13),
        ('corn', 'corn/corn.csv', 20),  # High-dimensional, limit components
    ]
    
    preprocessing_methods = ['mean_center', 'standardize']
    
    for dataset_name, dataset_path, n_components in datasets:
        print(f"\nProcessing {dataset_name} dataset...")
        
        # Load data
        full_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), dataset_path)
        df = pd.read_csv(full_path, index_col=0)
        
        # Get numeric columns only, excluding target columns (those with #target suffix)
        numeric_cols = df.select_dtypes(include=[np.number]).columns
        # Exclude columns that end with #target
        feature_cols = [col for col in numeric_cols if not col.endswith('#target')]
        X = df[feature_cols].values
        
        for preprocessing in preprocessing_methods:
            print(f"  - Preprocessing: {preprocessing}")
            
            # Generate reference
            result = generate_pca_reference(X, n_components=n_components, 
                                           preprocessing=preprocessing)
            
            # Validate mathematical properties
            validations = validate_mathematical_properties(result)
            result['validations'] = validations
            
            # Save result
            output_file = f"reference_results/{dataset_name}_pca_{preprocessing}.json"
            output_path = os.path.join(os.path.dirname(__file__), output_file)
            
            with open(output_path, 'w') as f:
                json.dump(result, f, indent=2)
            
            print(f"    Saved to {output_file}")
            print(f"    Condition number: {result.get('condition_number', 'N/A'):.2e}")
            print(f"    Orthogonality error: {result['loadings_orthogonality_error']:.2e}")
    
    # Generate synthetic test cases
    print("\nGenerating synthetic test cases...")
    test_cases = generate_test_cases()
    
    for case_name, X in test_cases:
        print(f"  - {case_name}")
        result = generate_pca_reference(X, preprocessing='mean_center')
        validations = validate_mathematical_properties(result)
        result['validations'] = validations
        
        output_file = f"reference_results/synthetic_{case_name}.json"
        output_path = os.path.join(os.path.dirname(__file__), output_file)
        
        with open(output_path, 'w') as f:
            json.dump(result, f, indent=2)
        
        print(f"    Saved to {output_file}")
    
    print("\nReference generation complete!")


if __name__ == "__main__":
    main()