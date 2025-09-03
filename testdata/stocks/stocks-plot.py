


import pandas as pd
import matplotlib.pyplot as plt

# List of stock tickers (first 16)
tickers = [
    'AAPL', 'MSFT', 'GOOGL', 'AMZN', 'TSLA', 'META', 'NVDA', 'NFLX',
    'INTC', 'AMD', 'IBM', 'ORCL', 'CSCO', 'ADBE', 'PYPL', 'QCOM'
]

fig, axes = plt.subplots(4, 4, figsize=(20, 16), sharex=False)
axes = axes.flatten()

for i, ticker in enumerate(tickers):
    ax1 = axes[i]
    try:
        df = pd.read_csv(f'stock-data-{ticker}.csv', parse_dates=['date'])
    except Exception as e:
        ax1.set_title(f'{ticker} (missing)')
        continue

    # Plot open, high, low, close
    ax1.plot(df['date'], df['open'], label='Open')
    ax1.plot(df['date'], df['high'], label='High')
    ax1.plot(df['date'], df['low'], label='Low')
    ax1.plot(df['date'], df['close'], label='Close')
    ax1.set_ylabel('Price')
    ax1.set_title(f'{ticker}')

    # Secondary y-axis for volume
    ax2 = ax1.twinx()
    ax2.plot(df['date'], df['volume'], color='gray', label='Volume', alpha=0.3)
    ax2.set_ylabel('Volume')

    # Only show legend for the first subplot
    if i == 0:
        lines_1, labels_1 = ax1.get_legend_handles_labels()
        lines_2, labels_2 = ax2.get_legend_handles_labels()
        ax1.legend(lines_1 + lines_2, labels_1 + labels_2, loc='upper left')

# Hide unused subplots if tickers < 16
for j in range(len(tickers), 16):
    fig.delaxes(axes[j])

plt.tight_layout()
plt.savefig('stocks-multiplot.png', format='png', dpi=150)
plt.show()