#!/usr/bin/env python3
"""
prepare_for_pca.py
Transform raw enriched stock data into a format suitable for PCA analysis.
This includes calculating returns, handling missing data, and standardization.
"""

import pandas as pd
import numpy as np
from datetime import datetime

def calculate_returns(series, method='simple'):
    """
    Calculate returns from a price series.
    
    Args:
        series: Price series
        method: 'simple' for arithmetic returns, 'log' for log returns
    
    Returns:
        Returns series
    """
    if method == 'log':
        return np.log(series / series.shift(1))
    else:
        return series.pct_change()

def calculate_rolling_volatility(returns, window=20):
    """
    Calculate rolling volatility (standard deviation of returns).
    
    Args:
        returns: Returns series
        window: Rolling window size in days
    
    Returns:
        Rolling volatility series
    """
    return returns.rolling(window=window).std() * np.sqrt(252)  # Annualized

def calculate_moving_average(series, window):
    """
    Calculate simple moving average.
    
    Args:
        series: Input series
        window: Window size in days
    
    Returns:
        Moving average series
    """
    return series.rolling(window=window).mean()

def main():
    """Main function to prepare data for PCA."""
    print("Preparing enriched stock data for PCA analysis")
    print("=" * 60)
    
    # Load raw enriched data
    input_file = 'stocks_enriched_raw.csv'
    try:
        df = pd.read_csv(input_file)
        df['date'] = pd.to_datetime(df['date'])
        df.set_index('date', inplace=True)
        print(f"Loaded {input_file}: {df.shape}")
    except FileNotFoundError:
        print(f"Error: {input_file} not found. Please run stocks_enriched.py first.")
        return
    
    # Create transformed dataframe
    pca_df = pd.DataFrame(index=df.index)
    
    # 1. PYPL returns and volume
    print("\nTransforming PYPL data...")
    pca_df['pypl_return'] = calculate_returns(df['pypl_close'], method='simple')
    pca_df['pypl_log_return'] = calculate_returns(df['pypl_close'], method='log')
    
    # Scale volume using log transformation
    pca_df['pypl_volume_log'] = np.log1p(df['pypl_volume'])
    
    # 2. Market benchmark returns (QQQ, XLK)
    print("Calculating market benchmark returns...")
    if 'qqq_close' in df.columns:
        pca_df['qqq_return'] = calculate_returns(df['qqq_close'])
    if 'xlk_close' in df.columns:
        pca_df['xlk_return'] = calculate_returns(df['xlk_close'])
    
    # 3. Volatility indices (VIX, VXN) - use levels and changes
    print("Processing volatility indices...")
    if 'vix' in df.columns:
        pca_df['vix_level'] = df['vix']
        pca_df['vix_change'] = df['vix'].diff()
    if 'vxn' in df.columns:
        pca_df['vxn_level'] = df['vxn']
        pca_df['vxn_change'] = df['vxn'].diff()
    
    # 4. Interest rates (TNX, FVX) - use changes
    print("Processing interest rates...")
    if 'tnx' in df.columns:
        pca_df['tnx_level'] = df['tnx']
        pca_df['tnx_change'] = df['tnx'].diff()
    if 'fvx' in df.columns:
        pca_df['fvx_level'] = df['fvx']
        pca_df['fvx_change'] = df['fvx'].diff()
    
    # 5. Dollar Index returns
    print("Processing currency data...")
    if 'dxy_close' in df.columns:
        pca_df['dxy_return'] = calculate_returns(df['dxy_close'])
    
    # 6. Add derived features for PYPL
    print("Adding derived features...")
    
    # Moving averages (as ratios to current price)
    if 'pypl_close' in df.columns:
        ma5 = calculate_moving_average(df['pypl_close'], 5)
        ma20 = calculate_moving_average(df['pypl_close'], 20)
        pca_df['pypl_ma5_ratio'] = df['pypl_close'] / ma5 - 1
        pca_df['pypl_ma20_ratio'] = df['pypl_close'] / ma20 - 1
    
    # Rolling volatility
    if 'pypl_return' in pca_df.columns:
        pca_df['pypl_volatility_20d'] = calculate_rolling_volatility(pca_df['pypl_return'], 20)
    
    # 7. Handle missing data
    print("\nHandling missing data...")
    
    # Drop rows with too many missing values (first rows due to returns calculation)
    before_rows = len(pca_df)
    pca_df = pca_df.dropna(thresh=len(pca_df.columns) * 0.5)  # Keep rows with at least 50% data
    
    # Forward fill remaining missing values (for weekends/holidays alignment)
    pca_df = pca_df.fillna(method='ffill')
    
    # Drop any remaining rows with NaN (typically first 20 rows due to rolling calculations)
    pca_df = pca_df.dropna()
    after_rows = len(pca_df)
    print(f"  Rows removed: {before_rows - after_rows}")
    print(f"  Final rows: {after_rows}")
    
    # 8. Standardize all features (z-score normalization)
    print("\nStandardizing features...")
    pca_df_standardized = pca_df.copy()
    
    for col in pca_df_standardized.columns:
        mean = pca_df_standardized[col].mean()
        std = pca_df_standardized[col].std()
        if std > 0:
            pca_df_standardized[col] = (pca_df_standardized[col] - mean) / std
        else:
            print(f"  Warning: {col} has zero variance, keeping as-is")
    
    # 9. Save both versions
    # Reset index to have date as column
    pca_df.reset_index(inplace=True)
    pca_df_standardized.reset_index(inplace=True)
    
    # Save non-standardized version
    output_file_raw = 'stocks_enriched_for_pca.csv'
    pca_df.to_csv(output_file_raw, index=False)
    print(f"\nSaved transformed data to {output_file_raw}")
    
    # Save standardized version
    output_file_std = 'stocks_enriched_for_pca_standardized.csv'
    pca_df_standardized.to_csv(output_file_std, index=False)
    print(f"Saved standardized data to {output_file_std}")
    
    # Print summary statistics
    print("\n" + "=" * 60)
    print("DATA SUMMARY")
    print("=" * 60)
    print(f"Final dataset shape: {pca_df_standardized.shape}")
    print(f"Date range: {pca_df_standardized['date'].min()} to {pca_df_standardized['date'].max()}")
    
    print("\nFeatures for PCA (after standardization):")
    for col in pca_df_standardized.columns:
        if col != 'date':
            mean = pca_df_standardized[col].mean()
            std = pca_df_standardized[col].std()
            print(f"  • {col:25s}: mean={mean:7.4f}, std={std:7.4f}")
    
    print("\nCorrelation with PYPL returns:")
    if 'pypl_return' in pca_df_standardized.columns:
        correlations = pca_df_standardized.drop('date', axis=1).corrwith(
            pca_df_standardized['pypl_return']
        ).sort_values(ascending=False)
        for feature, corr in correlations.head(10).items():
            if feature != 'pypl_return':
                print(f"  • {feature:25s}: {corr:7.4f}")
    
    print("\nDataset is ready for PCA analysis!")
    print("Use 'stocks_enriched_for_pca_standardized.csv' for PCA")

if __name__ == "__main__":
    main()