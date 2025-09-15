# Enriched Stocks Dataset for Advanced PCA Analysis

## Overview

The enriched stocks dataset extends beyond simple stock price data to include market-wide factors that influence individual stock movements. This dataset is designed for advanced PCA analysis to uncover hidden patterns in financial markets.

## Dataset Components

### Primary Focus
- **PYPL (PayPal)**: Individual stock serving as the primary subject of analysis

### Market Factors

#### Volatility Indices
- **VIX (^VIX)**: S&P 500 volatility index ("fear gauge")
- **VXN (^VXN)**: Nasdaq-100 volatility index (tech sector fear)

#### Market Benchmarks
- **QQQ**: Nasdaq-100 ETF (broad tech market proxy)
- **XLK**: Technology Select Sector SPDR Fund (tech sector performance)

#### Interest Rates
- **TNX (^TNX)**: 10-year Treasury yield (long-term rates)
- **FVX (^FVX)**: 5-year Treasury yield (medium-term rates)

#### Currency
- **DXY (DX-Y.NYB)**: US Dollar Index (dollar strength vs basket of currencies)

## Data Pipeline

### Step 1: Data Collection (`stocks_enriched.py`)
Downloads raw historical data for all components using yfinance.

```bash
cd testdata
source .venv/bin/activate
cd stocks
python stocks_enriched.py
```

Output: `stocks_enriched_raw.csv` containing raw price data and volumes.

### Step 2: Data Transformation (`prepare_for_pca.py`)
Transforms raw data into features suitable for PCA:

```bash
python prepare_for_pca.py
```

Outputs:
- `stocks_enriched_for_pca.csv`: Transformed features (not standardized)
- `stocks_enriched_for_pca_standardized.csv`: Z-score standardized features (recommended for PCA)

## Feature Engineering

### Transformations Applied

1. **Returns Calculation**
   - Simple returns: (price_t - price_t-1) / price_t-1
   - Log returns: log(price_t / price_t-1)
   - More stationary than raw prices for PCA

2. **Volume Scaling**
   - Log transformation: log(1 + volume)
   - Reduces skewness in volume data

3. **Volatility Measures**
   - 20-day rolling standard deviation of returns
   - Annualized: σ_daily × √252

4. **Technical Indicators**
   - MA5 ratio: price / 5-day moving average - 1
   - MA20 ratio: price / 20-day moving average - 1
   - Captures momentum and mean reversion

5. **Level vs Changes**
   - Volatility indices: Both levels and daily changes
   - Interest rates: Both levels and daily changes
   - Provides both state and dynamics information

6. **Standardization**
   - Z-score normalization: (x - μ) / σ
   - Essential for PCA due to different scales

## PCA Analysis Guide

### Running PCA with GoPCA

```bash
# Using the pca CLI
cd ../..  # Return to gopca root
./build/pca analyze --components 5 --preprocessing none \
    testdata/stocks/stocks_enriched_for_pca_standardized.csv

# Using GoPCA Desktop
# 1. Load stocks_enriched_for_pca_standardized.csv
# 2. Set preprocessing to "None" (already standardized)
# 3. Choose 5-10 components
# 4. Run analysis
```

### Interpreting Principal Components

#### Typical Patterns in Financial PCA

**PC1 (Market Beta)**
- Usually explains 30-40% of variance
- High loadings on market indices (QQQ, XLK)
- Represents overall market movement
- PYPL loading indicates market sensitivity

**PC2 (Volatility Regime)**
- Often 15-20% of variance
- High loadings on VIX, VXN
- Negative correlation with returns
- Captures risk-on/risk-off dynamics

**PC3 (Interest Rate Factor)**
- Typically 10-15% of variance
- High loadings on TNX, FVX
- May show sector rotation patterns
- Tech stocks often negatively correlated

**PC4 (Currency/International)**
- Usually 5-10% of variance
- High loading on DXY
- Captures international effects
- Important for multinational companies

**PC5+ (Idiosyncratic Factors)**
- Company-specific movements
- Technical patterns (momentum, mean reversion)
- Lower explained variance but potentially actionable

### Visualization Insights

#### Scores Plot
- **Temporal patterns**: Look for clusters by time period
- **Market regimes**: Identify bull/bear market periods
- **Outliers**: Days with unusual market conditions

#### Loadings Plot
- **Variable relationships**: Which factors move together
- **Orthogonal factors**: Independent risk sources
- **PYPL positioning**: How the stock relates to market factors

#### Biplot
- **Combined view**: See both observations and variables
- **Regime identification**: Market conditions driving specific periods

## Example Analysis Workflow

```python
# 1. Generate the enriched dataset
python stocks_enriched.py
python prepare_for_pca.py

# 2. Run PCA (using pca CLI)
../../build/pca analyze \
    --components 7 \
    --preprocessing none \
    --confidence 0.95 \
    --output stocks_pca_results \
    stocks_enriched_for_pca_standardized.csv

# 3. Examine results
# - Check explained variance ratios
# - First 3-4 PCs should explain 60-70% of variance
# - Look for interpretable factor loadings

# 4. Transform new data
../../build/pca transform \
    stocks_pca_results.json \
    new_market_data.csv
```

## Data Dictionary

### Input Features (after transformation)

| Feature | Description | Interpretation |
|---------|-------------|----------------|
| pypl_return | Daily simple return | Stock performance |
| pypl_log_return | Daily log return | Compound growth |
| pypl_volume_log | Log of trading volume | Trading activity |
| pypl_ma5_ratio | Price vs 5-day MA | Short-term momentum |
| pypl_ma20_ratio | Price vs 20-day MA | Medium-term trend |
| pypl_volatility_20d | 20-day rolling volatility | Recent risk level |
| qqq_return | Nasdaq-100 ETF return | Tech market movement |
| xlk_return | Tech sector ETF return | Sector performance |
| vix_level | S&P 500 volatility index | Market fear level |
| vix_change | Daily VIX change | Fear dynamics |
| vxn_level | Nasdaq volatility index | Tech fear level |
| vxn_change | Daily VXN change | Tech fear dynamics |
| tnx_level | 10-year Treasury yield | Long-term rates |
| tnx_change | Daily TNX change | Rate momentum |
| fvx_level | 5-year Treasury yield | Medium-term rates |
| fvx_change | Daily FVX change | Rate dynamics |
| dxy_return | Dollar index return | Currency strength |

## Advanced Usage

### Custom Time Periods

Modify date ranges in `stocks_enriched.py`:
```python
start_date = '2020-01-01'  # COVID period analysis
end_date = '2023-12-31'
```

### Additional Factors

Add more market factors to `MARKET_FACTORS` dictionary:
```python
MARKET_FACTORS = {
    # ... existing factors ...
    'GLD': 'gold',      # Gold ETF
    'TLT': 'bonds',     # Long-term bonds
    'USO': 'oil',       # Oil prices
}
```

### Alternative Preprocessing

Experiment with different transformations:
- Differences instead of returns
- Squared returns for volatility
- Cross-products for interactions
- Lagged features for dynamics

## Troubleshooting

### Missing Data
- Some tickers may have different trading days
- Forward-fill is used for alignment
- Check data completeness summary after running scripts

### Scale Issues
- Always use standardized version for PCA
- Raw version useful for descriptive statistics
- Verify all features have similar scales after standardization

### Interpretation Challenges
- Financial data is noisy
- Patterns may be time-period dependent
- Consider rolling window PCA for dynamic analysis

## References

1. **PCA in Finance**
   - Alexander, C. (2001). "Market Models: A Guide to Financial Data Analysis"
   - Tsay, R. S. (2010). "Analysis of Financial Time Series"

2. **Factor Models**
   - Fama, E. F., & French, K. R. (1993). "Common risk factors in the returns on stocks and bonds"
   - Ross, S. A. (1976). "The arbitrage theory of capital asset pricing"

3. **Market Indicators**
   - CBOE VIX White Paper: Understanding the VIX Index
   - Federal Reserve Economic Data (FRED): Treasury Yield Documentation

## Notes

- This is an exploratory/educational dataset
- Not intended for actual trading decisions
- Patterns found may not persist out-of-sample
- Always validate findings with domain knowledge
- Consider market regime changes when interpreting results

## Future Enhancements

- Add more stocks for sector-wide analysis
- Include fundamental data (P/E ratios, earnings)
- Add macroeconomic indicators (GDP, inflation)
- Implement rolling window PCA
- Create interactive visualizations of factor evolution