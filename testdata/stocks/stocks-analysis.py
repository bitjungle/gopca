# --- SSA (temporal PCA) on PYPL with three plots ---
# Edit these three lines if needed:
CSV_PATH = "stock-data-PYPL.csv"   # or absolute path
PRICE_COLUMN = "Close"             # fallback to "Adj Close"/"close" if needed
L, R = 260, 3                      # window length, retained PCs for MD

import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

def load_series_case_insensitive(csv_path: str, price_col: str):
    df = pd.read_csv(csv_path)
    colmap = {c.lower(): c for c in df.columns}
    # pick a date column
    date_col = next((orig for key, orig in colmap.items() if key.startswith("date")), None)
    if date_col is None:
        raise ValueError("No date-like column found.")
    df[date_col] = pd.to_datetime(df[date_col])
    df = df.sort_values(by=date_col).reset_index(drop=True)
    # pick price column (case-insensitive, with fallbacks)
    wanted = price_col.lower()
    if wanted not in colmap:
        for alt in ("adj close", "close", "price"):
            if alt in colmap:
                wanted = alt; break
        else:
            raise ValueError(f"Requested column '{price_col}' not found. Available: {list(df.columns)}")
    col = colmap[wanted]
    return df[date_col], df[col].astype(float).to_numpy()

def trajectory_matrix(x: np.ndarray, L: int) -> np.ndarray:
    N = len(x)
    if not (1 < L < N): raise ValueError("Require 1 < L < N")
    K = N - L + 1
    return np.column_stack([x[i:i+L] for i in range(K)])

def ssa(x: np.ndarray, L: int):
    x0 = x - np.mean(x)
    X = trajectory_matrix(x0, L)
    U, s, VT = np.linalg.svd(X, full_matrices=False)
    PCs = (VT.T * s)  # K x r_full (scores = V * Σ)
    return X, U, s, VT, PCs

def mahalanobis_squared(Z: np.ndarray, eps: float = 1e-9) -> np.ndarray:
    mu = Z.mean(axis=0, keepdims=True)
    D = Z - mu
    S = np.cov(D, rowvar=False)
    Sinv = np.linalg.pinv(S + eps * np.eye(S.shape[0]))
    return np.einsum("ij,jk,ik->i", D, Sinv, D)

# --- Load data & run SSA ---
dates, price = load_series_case_insensitive(CSV_PATH, PRICE_COLUMN)
x = np.log(price + 1e-12)
X, U, s, VT, PCs = ssa(x, L)
K = PCs.shape[0]
window_dates = dates.iloc[:K].reset_index(drop=True)

# 1) Scores plot: PC1 vs PC2
plt.figure()
t = np.arange(K)
plt.scatter(PCs[:, 0], PCs[:, 1], s=12, c=t)  # default colormap; no explicit colors
plt.xlabel("PC1 score"); plt.ylabel("PC2 score")
plt.title("SSA Scores: PC1 vs PC2 (PYPL)")
plt.tight_layout(); plt.show()

# 2) Loadings plot: first R columns of U vs lag
plt.figure()
lags = np.arange(1, U.shape[0] + 1)
for j in range(min(R, U.shape[1])):
    plt.plot(lags, U[:, j], label=f"U[:, {j+1}]")
plt.xlabel("Lag"); plt.ylabel("Loading")
plt.title(f"SSA Loadings (first {R})")
plt.legend(); plt.tight_layout(); plt.show()

# 3) RSS vs Mahalanobis distance
#    Tail RSS per window = sum of squared tail scores; MD^2 on retained scores
retained = PCs[:, :R]
md2 = mahalanobis_squared(retained)
tail = PCs[:, R:]
tail_rss = np.sum(tail**2, axis=1)

plt.figure()
plt.scatter(md2, tail_rss, s=12)
plt.xlabel("Mahalanobis distance squared (retained PCs)")
plt.ylabel("Per-window tail RSS (sum of squared tail scores)")
plt.title("RSS vs Mahalanobis Distance (PYPL)")
plt.tight_layout(); plt.show()
