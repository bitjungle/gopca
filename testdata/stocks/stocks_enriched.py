#!/usr/bin/env python3
"""
stocks_enriched.py
Download historical stock data for PYPL along with market factors using yfinance.
This creates an enriched dataset for advanced PCA analysis of financial markets.
"""

import yfinance as yf
import pandas as pd
from datetime import datetime

# Define tickers and market factors
PRIMARY_TICKER = 'PYPL'  # PayPal as primary focus

MARKET_FACTORS = {
    # Volatility indices
    '^VIX': 'vix',      # S&P 500 volatility
    '^VXN': 'vxn',      # Nasdaq-100 volatility
    
    # Market/Sector benchmarks  
    'QQQ': 'qqq',       # Nasdaq-100 ETF
    'XLK': 'xlk',       # Technology sector ETF
    
    # Interest rates
    '^TNX': 'tnx',      # 10-year Treasury yield
    '^FVX': 'fvx',      # 5-year Treasury yield
    
    # Currency
    'DX-Y.NYB': 'dxy'   # Dollar Index
}

# Time range (same as stocks.py)
start_date = '2015-09-01'
end_date = '2025-08-31'

def download_ticker_data(ticker, start_date, end_date, retries=3):
    """Download data for a single ticker with retry logic."""
    for i in range(retries):
        try:
            data = yf.download(ticker, start=start_date, end=end_date, auto_adjust=True, progress=False)
            if not data.empty:
                return data, None
            else:
                return pd.DataFrame(), f"No data found for {ticker}"
        except Exception as e:
            err_msg = f"Attempt {i+1} failed for {ticker}: {e}"
            if i == retries - 1:
                return pd.DataFrame(), err_msg
    return pd.DataFrame(), f"Unknown error for {ticker}"

def main():
    """Main function to download and save enriched stock data."""
    print(f"Downloading enriched stock data from {start_date} to {end_date}")
    print("=" * 60)
    
    # Dictionary to store all data
    all_data = {}
    failed_downloads = []
    
    # Download primary ticker (PYPL)
    print(f"Downloading primary ticker: {PRIMARY_TICKER}")
    pypl_data, error = download_ticker_data(PRIMARY_TICKER, start_date, end_date)
    if not pypl_data.empty:
        all_data[PRIMARY_TICKER] = pypl_data
        print(f"  ✓ {PRIMARY_TICKER}: {len(pypl_data)} rows")
    else:
        print(f"  ✗ {PRIMARY_TICKER}: FAILED - {error}")
        print("Cannot proceed without primary ticker data")
        return
    
    # Download market factors
    print("\nDownloading market factors:")
    for ticker, short_name in MARKET_FACTORS.items():
        print(f"  Downloading {ticker} ({short_name})...", end=" ")
        data, error = download_ticker_data(ticker, start_date, end_date)
        if not data.empty:
            all_data[ticker] = data
            print(f"✓ {len(data)} rows")
        else:
            failed_downloads.append((ticker, error))
            print(f"✗ FAILED")
    
    # Create combined dataframe with all raw data
    print("\nCombining data...")
    
    # Start with PYPL data
    combined_df = pypl_data[['Close', 'Volume']].copy()
    combined_df.columns = ['pypl_close', 'pypl_volume']
    
    # Add market factors
    for ticker, short_name in MARKET_FACTORS.items():
        if ticker in all_data:
            factor_data = all_data[ticker]
            if ticker in ['^VIX', '^VXN', '^TNX', '^FVX']:
                # For indices and yields, use Close value directly
                combined_df[short_name] = factor_data['Close']
            else:
                # For ETFs and currency, use Close price
                combined_df[f'{short_name}_close'] = factor_data['Close']
    
    # Reset index to have date as a column
    combined_df.reset_index(inplace=True)
    combined_df.rename(columns={'Date': 'date'}, inplace=True)
    
    # Save raw enriched data
    output_file = 'stocks_enriched_raw.csv'
    combined_df.to_csv(output_file, index=False)
    print(f"\nSaved raw enriched data to {output_file}")
    print(f"Shape: {combined_df.shape}")
    print(f"Date range: {combined_df['date'].min()} to {combined_df['date'].max()}")
    
    # Print summary
    print("\n" + "=" * 60)
    print("SUMMARY")
    print("=" * 60)
    print(f"Successfully downloaded: {len(all_data)} / {len(MARKET_FACTORS) + 1} tickers")
    
    if failed_downloads:
        print(f"\nFailed downloads ({len(failed_downloads)}):")
        for ticker, error in failed_downloads:
            print(f"  • {ticker}: {error}")
    
    print("\nColumns in enriched dataset:")
    for col in combined_df.columns:
        if col != 'date':
            non_null = combined_df[col].notna().sum()
            pct_complete = (non_null / len(combined_df)) * 100
            print(f"  • {col}: {non_null}/{len(combined_df)} rows ({pct_complete:.1f}% complete)")

if __name__ == "__main__":
    main()