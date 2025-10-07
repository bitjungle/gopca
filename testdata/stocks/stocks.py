"""
stocks.py
Download historical stock data for a list of tickers using yfinance, 
add a few essential stock-specific features, and save each to a CSV file.

Columns in the output CSV:
----------------------------------------------------------
Raw OHLCV Data (direct from Yahoo Finance):
- date   : Trading date
- open   : Price at the start of the trading day
- high   : Highest price reached during the trading day
- low    : Lowest price reached during the trading day
- close  : Price at market close
- volume : Number of shares traded during the day

Engineered Features (calculated from OHLCV):
- log_return     : Daily log return of the stock. Captures relative price change and 
                   is preferred for statistical analysis because it is additive over time.
                   log_return[t] = ln(close[t] / close[t-1])

- ob_volume      : On-Balance Volume. A cumulative indicator that adds volume when price 
                   closes higher and subtracts volume when price closes lower. 
                   Intended to show whether trading volume confirms price trends 
                   (accumulation vs distribution).
----------------------------------------------------------
"""

import yfinance as yf
import pandas as pd
import numpy as np

# List of stock tickers to download
tickers = [
    'AAPL', 'MSFT', 'GOOGL', 'AMZN', 'TSLA', 'META', 'NVDA', 'NFLX', 'INTC', 'AMD',
    'IBM', 'ORCL', 'CSCO', 'ADBE', 'PYPL', 'QCOM', 'TXN', 'AVGO', 'CRM', 'SAP'
]

# Date range for historical data
start_date = '2015-09-01'
end_date = '2025-08-31'


def compute_features(df: pd.DataFrame) -> pd.DataFrame:
    """Compute essential stock-specific features from OHLCV."""

    # Log return: daily % change in logarithmic scale
    df["log_return"] = np.log(df["close"] / df["close"].shift(1))
    df.loc[df.index[0], "log_return"] = 0  # Set the first value to 0 instead of NaN

    # On-Balance Volume (ob_volume)
    ob_volume = [0]
    for i in range(1, len(df)):
        if df.loc[i, "close"] > df.loc[i - 1, "close"]:
            ob_volume.append(ob_volume[-1] + df.loc[i, "volume"])
        elif df.loc[i, "close"] < df.loc[i - 1, "close"]:
            ob_volume.append(ob_volume[-1] - df.loc[i, "volume"])
        else:
            ob_volume.append(ob_volume[-1])
    df["ob_volume"] = ob_volume

    return df


def download_and_format_stock_data(ticker, start_date, end_date, retries=3):
    """Download and format data for a single stock ticker."""
    for i in range(retries):
        try:
            stock_data = yf.download(ticker, start=start_date, end=end_date, auto_adjust=True)
            if not stock_data.empty:
                stock_data.reset_index(inplace=True)
                stock_data = stock_data[['Date', 'Open', 'High', 'Low', 'Close', 'Volume']]
                stock_data.columns = ['date', 'open', 'high', 'low', 'close', 'volume']

                # Add engineered features
                stock_data = compute_features(stock_data)

                return stock_data, None
            else:
                return pd.DataFrame(), f"No data found for {ticker}"
        except Exception as e:
            err_msg = f"Attempt {i+1} failed for {ticker}: {e}"
            if i == retries - 1:
                return pd.DataFrame(), err_msg
    return pd.DataFrame(), f"Unknown error for {ticker}"


failed_tickers = []
error_log = []

for ticker in tickers:
    print(f"Downloading {ticker}...")
    stock_data, error = download_and_format_stock_data(ticker, start_date, end_date)
    if not stock_data.empty:
        file_path = f'stock-data-{ticker}.csv'
        stock_data.to_csv(file_path, index=False)
        print(f"Saved {file_path} ({len(stock_data)} rows, {stock_data.shape[1]} columns)")
    else:
        failed_tickers.append(ticker)
        if error:
            error_log.append(error)
        print(f"{ticker}: FAILED")

if failed_tickers:
    print(f"\nFailed downloads ({len(failed_tickers)}): {failed_tickers}")
    for err in error_log:
        print(f"  {err}")
