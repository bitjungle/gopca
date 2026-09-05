#!/usr/bin/env python3
"""
Generate reference Kernel PCA results using scikit-learn for validation.

This script generates reference results for Kernel PCA with various kernels
(RBF, polynomial, sigmoid, linear) to validate our implementation.

References:
- Mika et al. (1998): Kernel PCA and De-Noising in Feature Spaces
- Schölkopf et al. (1998): Nonlinear Component Analysis as a Kernel Eigenvalue Problem
"""

import sys
import os
import json
import numpy as np
import pandas as pd
from sklearn.decomposition import KernelPCA
from sklearn.preprocessing import StandardScaler
from sklearn.metrics.pairwise import rbf_kernel, polynomial_kernel, sigmoid_kernel, linear_kernel

# Add parent directory to path
sys.path.append(os.path.dirname(os.path.dirname(os.path.abspath(__file__))))


def compute_kernel_matrix(X, kernel='rbf', gamma=1.0, degree=3, coef0=1.0):
    """
    Compute kernel matrix for validation purposes.
    
    Parameters:
    -----------
    X : array-like
        Input data matrix (n_samples, n_features)
    kernel : str
        Kernel type: 'rbf', 'polynomial', 'sigmoid', 'linear'
    gamma : float
        Kernel coefficient for rbf, polynomial and sigmoid
    degree : int
        Degree for polynomial kernel
    coef0 : float
        Independent term in polynomial and sigmoid kernels
        
    Returns:
    --------
    K : array
        Kernel matrix
    """
    if kernel == 'rbf':
        K = rbf_kernel(X, gamma=gamma)
    elif kernel in ['poly', 'polynomial']:
        K = polynomial_kernel(X, degree=degree, gamma=gamma, coef0=coef0)
    elif kernel == 'sigmoid':
        K = sigmoid_kernel(X, gamma=gamma, coef0=coef0)
    elif kernel == 'linear':
        K = linear_kernel(X)
    else:
        raise ValueError(f"Unknown kernel: {kernel}")
    
    return K


def center_kernel_matrix(K):
    """
    Center kernel matrix in feature space.
    
    K_centered = K - 1_n K - K 1_n + 1_n K 1_n
    where 1_n is n x n matrix with all entries 1/n
    """
    n = K.shape[0]
    one_n = np.ones((n, n)) / n
    K_centered = K - one_n @ K - K @ one_n + one_n @ K @ one_n
    return K_centered


def generate_kernel_pca_reference(data, n_components=None, kernel='rbf', 
                                 preprocessing='mean_center', gamma=1.0, 
                                 degree=3, coef0=1.0, random_state=42):
    """
    Generate Kernel PCA reference results.
    
    Parameters:
    -----------
    data : array-like
        Input data matrix (n_samples, n_features)
    n_components : int, optional
        Number of components to compute
    kernel : str
        Kernel type: 'rbf', 'polynomial', 'sigmoid', 'linear'
    preprocessing : str
        Preprocessing method: 'none', 'mean_center', 'standardize'
    gamma : float
        Kernel coefficient for rbf, polynomial and sigmoid
    degree : int
        Degree for polynomial kernel
    coef0 : float
        Independent term in polynomial and sigmoid kernels
        
    Returns:
    --------
    dict : Reference results including scores, eigenvalues, kernel matrix, etc.
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
        # Divide by the sample standard deviation (ddof=1), not the population
        # one StandardScaler uses.
        #
        # For linear PCA the choice does not matter to what we compare: the
        # factor sqrt(n/(n-1)) cancels between theta and the inverse scaling, so
        # coefficients and predictions come out identical either way. For a
        # kernel it does not cancel. gamma multiplies squared distances, so
        # rescaling the inputs changes exp(-gamma*d^2) non-linearly, and the
        # eigenvalues move with it -- by about 0.4% on iris, which is far above
        # floating-point noise and far below anything a loose tolerance would
        # catch as a bug.
        #
        # GoPCA standardizes with ddof=1, so the reference does too. That keeps
        # the comparison a test of the kernel computation rather than of a
        # scaling convention neither implementation is wrong about.
        # Scale without centring, and divide by the sample standard deviation
        # (ddof=1). This is GoPCA's kernel pipeline, and the reference has to use
        # the same one or it is describing a different computation.
        #
        # Kernel PCA centres in kernel space, so GoPCA deliberately does not
        # centre the inputs -- see the "Practical Note on Preprocessing" in
        # docs/intro_to_pca.md. Whether that matters depends entirely on the
        # kernel:
        #
        #   RBF     depends only on ||x-y||, which no common translation changes,
        #           so centring the inputs makes no difference at all.
        #   linear  survives either way, because centring the kernel matrix of an
        #           uncentred linear kernel is the same as the linear kernel of
        #           centred data.
        #   poly    does not. (gamma<x,y> + coef0)^d is not translation
        #           invariant for d > 1, and centring the kernel matrix afterwards
        #           does not recover it. Centring here moved the leading iris
        #           eigenvalue from 116550 to 1255 -- a factor of 93, which looks
        #           like a catastrophic bug and is only a different pipeline.
        #
        # ddof=1 matters for a second reason. For linear PCA the factor
        # sqrt(n/(n-1)) cancels in the quantities worth comparing; for a kernel it
        # does not, since gamma multiplies squared distances. On iris that
        # difference alone moved the leading RBF eigenvalue by 0.4%.
        stds = X.std(axis=0, ddof=1)
        stds[stds == 0] = 1.0
        X_processed = X / stds
        preprocessing_params['stds'] = stds.tolist()
        preprocessing_params['ddof'] = 1
        preprocessing_params['centered'] = False
        
    else:  # 'none'
        X_processed = X.copy()
    
    # Compute kernel matrix manually for validation
    K = compute_kernel_matrix(X_processed, kernel=kernel, gamma=gamma, 
                             degree=degree, coef0=coef0)
    K_centered = center_kernel_matrix(K)
    
    # Perform Kernel PCA using sklearn
    kpca = KernelPCA(n_components=n_components, kernel=kernel, gamma=gamma,
                     degree=degree, coef0=coef0, eigen_solver='auto',
                     remove_zero_eig=True, random_state=random_state)
    
    scores = kpca.fit_transform(X_processed)
    
    # Get eigenvalues and eigenvectors
    eigenvalues = kpca.eigenvalues_
    eigenvectors = kpca.eigenvectors_
    
    # Calculate explained variance (approximate for kernel PCA)
    # In kernel PCA, we use eigenvalues of centered kernel matrix
    total_variance = np.sum(eigenvalues)
    explained_variance_ratio = eigenvalues[:n_components] / total_variance
    cumulative_variance = np.cumsum(explained_variance_ratio)
    
    # Verify kernel matrix properties
    kernel_symmetry_error = np.max(np.abs(K - K.T))
    
    # For linear kernel, compare with standard PCA
    linear_kernel_check = {}
    if kernel == 'linear':
        from sklearn.decomposition import PCA
        pca = PCA(n_components=n_components)
        pca_scores = pca.fit_transform(X_processed)
        
        # Check if kernel PCA with linear kernel approximates standard PCA
        # Note: There might be sign differences
        score_similarity = np.min([
            np.mean(np.abs(scores - pca_scores)),
            np.mean(np.abs(scores + pca_scores))
        ])
        linear_kernel_check['score_similarity'] = float(score_similarity)
        linear_kernel_check['matches_standard_pca'] = bool(score_similarity < 0.01)
    
    # Create result dictionary
    result = {
        'method': 'sklearn_kernel_pca',
        'kernel': kernel,
        'kernel_params': {
            'gamma': gamma,
            'degree': degree,
            'coef0': coef0
        },
        'preprocessing': preprocessing,
        'preprocessing_params': preprocessing_params,
        'n_samples': n_samples,
        'n_features': n_features,
        'n_components': n_components,
        'scores': scores.tolist(),
        'eigenvalues': eigenvalues[:n_components].tolist(),
        'eigenvectors': eigenvectors[:, :n_components].tolist(),
        'explained_variance_ratio': explained_variance_ratio.tolist(),
        'cumulative_variance': cumulative_variance.tolist(),
        'total_variance': float(total_variance),
        'kernel_matrix_shape': K.shape,
        'kernel_symmetry_error': float(kernel_symmetry_error),
    }
    
    if linear_kernel_check:
        result['linear_kernel_validation'] = linear_kernel_check
    
    # Add kernel matrix statistics
    result['kernel_matrix_stats'] = {
        'min': float(K.min()),
        'max': float(K.max()),
        'mean': float(K.mean()),
        'std': float(K.std()),
        'diagonal_mean': float(np.mean(np.diag(K)))
    }
    
    return result


def generate_nonlinear_test_data():
    """
    Generate nonlinear test datasets for Kernel PCA validation.
    """
    np.random.seed(42)
    datasets = []
    
    # 1. Swiss Roll (3D manifold)
    from sklearn.datasets import make_swiss_roll
    X_swiss, color = make_swiss_roll(n_samples=200, noise=0.1, random_state=42)
    datasets.append(('swiss_roll', X_swiss, color))
    
    # 2. Concentric circles
    from sklearn.datasets import make_circles
    X_circles, y_circles = make_circles(n_samples=200, noise=0.05, factor=0.3, random_state=42)
    datasets.append(('circles', X_circles, y_circles))
    
    # 3. Two moons
    from sklearn.datasets import make_moons
    X_moons, y_moons = make_moons(n_samples=200, noise=0.1, random_state=42)
    datasets.append(('moons', X_moons, y_moons))
    
    # 4. S-curve
    from sklearn.datasets import make_s_curve
    X_scurve, color_scurve = make_s_curve(n_samples=200, noise=0.1, random_state=42)
    datasets.append(('s_curve', X_scurve, color_scurve))
    
    return datasets


def validate_kernel_properties(result):
    """
    Validate mathematical properties specific to Kernel PCA.
    
    Returns:
    --------
    dict : Validation results
    """
    scores = np.array(result['scores'])
    eigenvalues = np.array(result['eigenvalues'])
    kernel = result['kernel']
    
    validations = {}
    
    # 1. Check eigenvalues are non-negative (kernel matrix is PSD)
    validations['eigenvalues_non_negative'] = bool(np.all(eigenvalues >= -1e-10))
    
    # 2. Check eigenvalues are in descending order
    validations['eigenvalues_descending'] = bool(np.all(np.diff(eigenvalues) <= 0))
    
    # 3. For RBF kernel, check bounded scores
    if kernel == 'rbf':
        # RBF kernel projects to bounded region
        score_norms = np.linalg.norm(scores, axis=1)
        validations['max_score_norm'] = float(np.max(score_norms))
        
    # 4. Check kernel matrix was symmetric
    validations['kernel_symmetry_verified'] = bool(result['kernel_symmetry_error'] < 1e-10)
    
    return validations


def main():
    """
    Generate Kernel PCA reference results for validation.
    """
    print("Generating Kernel PCA reference results...")
    
    # Ensure output directory exists
    output_dir = os.path.join(os.path.dirname(__file__), 'reference_results')
    os.makedirs(output_dir, exist_ok=True)
    
    # Test with standard datasets
    datasets = [
        ('iris', 'iris/iris.csv', 4),
        ('wine', 'wine/wine.csv', 10),
    ]
    
    # Test different kernels
    kernel_configs = [
        ('linear', {'gamma': 1.0}),
        ('rbf', {'gamma': 1.0}),
        ('rbf', {'gamma': 0.1}),
        ('rbf', {'gamma': 10.0}),
        ('poly', {'degree': 2, 'gamma': 1.0, 'coef0': 1.0}),
        ('poly', {'degree': 3, 'gamma': 1.0, 'coef0': 0.0}),
        # ('sigmoid', ...) omitted: GoPCA implements rbf, linear and poly only,
        # so a sigmoid reference is a file no test can ever consume (#845).
    ]
    
    for dataset_name, dataset_path, n_components in datasets:
        print(f"\nProcessing {dataset_name} dataset...")
        
        # Load data
        full_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), dataset_path)
        df = pd.read_csv(full_path, index_col=0)
        
        # Numeric columns, excluding #target columns -- the same convention
        # generate_reference_pca.py uses and the same one GoPCA applies.
        #
        # Without this the reference described a different problem than the one
        # GoPCA solves: on iris it fed species#target in as a fifth predictor,
        # which is both a mismatch and, since that column is the class label, a
        # decomposition of the answer alongside the question. A comparison
        # against it could never have passed, which is the likeliest reason
        # nobody ever wired these references up to a test (#845).
        numeric_cols = df.select_dtypes(include=[np.number]).columns
        feature_cols = [col for col in numeric_cols if not col.endswith('#target')]
        X = df[feature_cols].values
        
        for kernel, params in kernel_configs:
            kernel_name = f"{kernel}"
            if kernel == 'rbf':
                kernel_name += f"_gamma{params['gamma']}"
            elif kernel == 'poly':
                kernel_name += f"_deg{params['degree']}"
                
            print(f"  - Kernel: {kernel_name}")
            
            # Generate reference
            result = generate_kernel_pca_reference(
                X, n_components=n_components, 
                kernel=kernel,
                preprocessing='standardize',
                **params
            )
            
            # Validate properties
            validations = validate_kernel_properties(result)
            result['validations'] = validations
            
            # Save result
            output_file = f"reference_results/{dataset_name}_kpca_{kernel_name}.json"
            output_path = os.path.join(os.path.dirname(__file__), output_file)
            
            with open(output_path, 'w') as f:
                json.dump(result, f, indent=2)
            
            print(f"    Saved to {output_file}")
            if 'linear_kernel_validation' in result:
                print(f"    Linear kernel matches PCA: {result['linear_kernel_validation']['matches_standard_pca']}")
    
    # Generate results for nonlinear synthetic datasets
    print("\nGenerating results for nonlinear datasets...")
    nonlinear_datasets = generate_nonlinear_test_data()
    
    for dataset_name, X, labels in nonlinear_datasets:
        print(f"  - {dataset_name}")
        
        # Use RBF kernel for nonlinear data
        result = generate_kernel_pca_reference(
            X, n_components=2,
            kernel='rbf',
            preprocessing='standardize',
            gamma=1.0
        )
        
        validations = validate_kernel_properties(result)
        result['validations'] = validations
        result['labels'] = labels.tolist() if hasattr(labels, 'tolist') else labels
        
        output_file = f"reference_results/synthetic_{dataset_name}_kpca.json"
        output_path = os.path.join(os.path.dirname(__file__), output_file)
        
        with open(output_path, 'w') as f:
            json.dump(result, f, indent=2)
        
        print(f"    Saved to {output_file}")
    
    # Test Swiss Roll data specifically (it's in our testdata)
    swiss_roll_path = os.path.join(os.path.dirname(os.path.dirname(__file__)), 
                                   'swiss_roll', 'swiss_roll.csv')
    if os.path.exists(swiss_roll_path):
        print("\nProcessing Swiss Roll from testdata...")
        df = pd.read_csv(swiss_roll_path, index_col=0)
        X = df[['X', 'Y', 'Z']].values
        
        for gamma in [0.01, 0.1, 1.0, 10.0]:
            print(f"  - RBF kernel with gamma={gamma}")
            
            result = generate_kernel_pca_reference(
                X, n_components=2,
                kernel='rbf',
                preprocessing='standardize',
                gamma=gamma
            )
            
            validations = validate_kernel_properties(result)
            result['validations'] = validations
            
            output_file = f"reference_results/swiss_roll_kpca_rbf_gamma{gamma}.json"
            output_path = os.path.join(os.path.dirname(__file__), output_file)
            
            with open(output_path, 'w') as f:
                json.dump(result, f, indent=2)
            
            print(f"    Saved to {output_file}")
    
    print("\nKernel PCA reference generation complete!")


if __name__ == "__main__":
    main()