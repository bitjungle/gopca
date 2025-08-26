#!/usr/bin/env python3
"""
Reference implementation for PCA on meteorological data with missing values.
This helps understand how GoPCA should handle the met_kikut_aarhus.csv dataset.
"""

import pandas as pd
import numpy as np
from sklearn.decomposition import PCA
from sklearn.preprocessing import StandardScaler
from sklearn.impute import SimpleImputer
import json

def analyze_missing_data(df):
    """Analyze missing value patterns in the dataset"""
    print("=" * 60)
    print("MISSING DATA ANALYSIS")
    print("=" * 60)
    
    total_values = df.shape[0] * df.shape[1]
    missing_counts = df.isnull().sum()
    total_missing = missing_counts.sum()
    
    print(f"Dataset shape: {df.shape[0]} rows × {df.shape[1]} columns")
    print(f"Total values: {total_values:,}")
    print(f"Total missing: {total_missing:,} ({100*total_missing/total_values:.1f}%)")
    print()
    
    print("Missing values per column:")
    for col in df.columns:
        count = missing_counts[col]
        pct = 100 * count / len(df)
        print(f"  {col:30s}: {count:6d} ({pct:5.1f}%)")
    
    # Check rows with any missing values
    rows_with_missing = df.isnull().any(axis=1).sum()
    print(f"\nRows with ANY missing values: {rows_with_missing:,} ({100*rows_with_missing/len(df):.1f}%)")
    
    return missing_counts, total_missing

def pca_with_drop_strategy(df_numeric, df_categorical):
    """PCA with drop rows strategy (like GoPCA drop)"""
    print("\n" + "=" * 60)
    print("STRATEGY 1: DROP ROWS WITH MISSING VALUES")
    print("=" * 60)
    
    # Drop rows with any NaN
    print(f"Original shape: {df_numeric.shape}")
    
    # Get indices of rows without NaN
    valid_indices = ~df_numeric.isnull().any(axis=1)
    df_clean = df_numeric[valid_indices]
    
    # CRITICAL: Also filter categorical data to match!
    df_cat_clean = df_categorical[valid_indices] if df_categorical is not None else None
    
    print(f"After dropping NaN rows: {df_clean.shape}")
    print(f"Rows dropped: {len(df_numeric) - len(df_clean):,}")
    
    if len(df_clean) == 0:
        print("ERROR: No data left after dropping NaN rows!")
        return None, None, None
    
    # Standardize
    scaler = StandardScaler()
    scaled_data = scaler.fit_transform(df_clean)
    
    # Run PCA
    pca = PCA(n_components=2)
    scores = pca.fit_transform(scaled_data)
    
    print(f"\nPCA Results:")
    print(f"  Scores shape: {scores.shape}")
    print(f"  Explained variance ratio: {pca.explained_variance_ratio_}")
    print(f"  First 5 scores PC1: {scores[:5, 0]}")
    
    if df_cat_clean is not None:
        print(f"\nCategorical data after filtering:")
        print(f"  Shape: {df_cat_clean.shape}")
        print(f"  Unique values: {df_cat_clean['month'].nunique() if 'month' in df_cat_clean.columns else 'N/A'}")
        print(f"  First 5 values: {df_cat_clean['month'].head().tolist() if 'month' in df_cat_clean.columns else 'N/A'}")
    
    return scores, pca, df_cat_clean

def pca_with_mean_imputation(df_numeric):
    """PCA with mean imputation strategy"""
    print("\n" + "=" * 60)
    print("STRATEGY 2: MEAN IMPUTATION")
    print("=" * 60)
    
    print(f"Original shape: {df_numeric.shape}")
    
    # Impute missing values with column means
    imputer = SimpleImputer(strategy='mean')
    imputed_data = imputer.fit_transform(df_numeric)
    
    print(f"After imputation: {imputed_data.shape}")
    print("Column means used for imputation:")
    for i, col in enumerate(df_numeric.columns):
        if df_numeric[col].isnull().any():
            mean_val = imputer.statistics_[i]
            print(f"  {col}: {mean_val:.3f}")
    
    # Standardize
    scaler = StandardScaler()
    scaled_data = scaler.fit_transform(imputed_data)
    
    # Run PCA
    pca = PCA(n_components=2)
    scores = pca.fit_transform(scaled_data)
    
    print(f"\nPCA Results:")
    print(f"  Scores shape: {scores.shape}")
    print(f"  Explained variance ratio: {pca.explained_variance_ratio_}")
    print(f"  First 5 scores PC1: {scores[:5, 0]}")
    
    return scores, pca

def main():
    # Load data
    print("Loading met_kikut_aarhus.csv...")
    df = pd.read_csv("met_kikut_aarhus.csv", index_col=0, parse_dates=True)
    
    print(f"Full dataset shape: {df.shape}")
    print(f"Columns: {list(df.columns)}")
    
    # Separate numeric and categorical columns
    categorical_cols = ['month']
    numeric_cols = [col for col in df.columns if col not in categorical_cols]
    
    df_numeric = df[numeric_cols]
    df_categorical = df[categorical_cols]
    
    print(f"\nNumeric columns: {numeric_cols}")
    print(f"Categorical columns: {categorical_cols}")
    
    # Analyze missing data
    missing_counts, total_missing = analyze_missing_data(df_numeric)
    
    # Test different strategies
    
    # 1. Drop strategy (what GoPCA desktop does with "drop")
    scores_drop, pca_drop, cat_data_drop = pca_with_drop_strategy(df_numeric, df_categorical)
    
    # 2. Mean imputation strategy
    scores_mean, pca_mean = pca_with_mean_imputation(df_numeric)
    
    # Check what would happen with naive SVD (should fail!)
    print("\n" + "=" * 60)
    print("ATTEMPTING NAIVE SVD (should fail with NaN)")
    print("=" * 60)
    try:
        scaler = StandardScaler()
        # This should fail because StandardScaler can't handle NaN
        scaled_data = scaler.fit_transform(df_numeric)
        pca = PCA(n_components=2)
        scores = pca.fit_transform(scaled_data)
        print("WARNING: SVD succeeded with NaN values! This shouldn't happen!")
        print(f"Scores shape: {scores.shape}")
    except Exception as e:
        print(f"Failed as expected: {e}")
    
    # Save results for comparison
    if scores_drop is not None:
        result = {
            "drop_strategy": {
                "rows_after_drop": len(scores_drop),
                "rows_dropped": len(df_numeric) - len(scores_drop),
                "scores_shape": scores_drop.shape,
                "categorical_shape": cat_data_drop.shape if cat_data_drop is not None else None,
                "explained_variance": pca_drop.explained_variance_ratio_.tolist()
            },
            "mean_strategy": {
                "rows": len(scores_mean),
                "scores_shape": scores_mean.shape,
                "explained_variance": pca_mean.explained_variance_ratio_.tolist()
            }
        }
        
        with open("met_pca_results.json", "w") as f:
            json.dump(result, f, indent=2)
        print(f"\nResults saved to met_pca_results.json")
        
        # Critical insight for the bug
        print("\n" + "=" * 60)
        print("CRITICAL INSIGHT FOR BUG FIX:")
        print("=" * 60)
        print(f"After dropping rows with missing values:")
        print(f"  - Scores matrix has {len(scores_drop)} rows")
        print(f"  - Categorical data has {len(cat_data_drop)} rows")
        print(f"  - These MUST match or indexing will fail!")
        print()
        print("If GoPCA doesn't filter categorical data when dropping rows,")
        print("the frontend will crash when trying to color by category!")

if __name__ == "__main__":
    main()