#!/usr/bin/env python3
"""
Generate reference Principal Component Regression results using scikit-learn.

The Go implementation is validated against these files by TestValidatePCRAgainstSklearn
in internal/core. Regenerate them with:

    cd testdata && source .venv/bin/activate
    cd validation && python generate_reference_pcr.py

Comparison rules that matter, because getting them wrong produces a spurious
mismatch rather than a silent one:

1. GoPCA standardizes with the sample standard deviation (divisor n-1, via gonum's
   stat.StdDev) while scikit-learn's StandardScaler uses the population divisor n.
   The standardized matrices therefore differ by the single scalar
   c = sqrt(n/(n-1)), which cancels between theta and S^-1: original-scale
   coefficients, the intercept and all predictions agree exactly, while singular
   values and scores differ by c. Only the first group is compared directly.

2. The intercept recovery below omits a -theta . pca.mean_ term. That is legitimate
   only because StandardScaler runs first, leaving pca.mean_ at around 1e-16; the
   formula is wrong for any pipeline where PCA does not receive centered input.

3. Component signs are arbitrary. They flip gamma but not the original-scale
   coefficients or the predictions, so only sign-invariant quantities are stored.

References:
- Massy (1965), Principal Components Regression in Exploratory Statistical
  Research, JASA 60(309), 234-256.
- Jolliffe (1982), A Note on the Use of Principal Components in Regression,
  JRSS C 31(3), 300-303.
"""

import json
import os

import numpy as np
import pandas as pd
from sklearn.decomposition import PCA
from sklearn.linear_model import LinearRegression
from sklearn.model_selection import KFold, cross_val_predict
from sklearn.pipeline import Pipeline
from sklearn.preprocessing import StandardScaler


def generate_pcr_reference(X, y, n_components, preprocessing='standardize'):
    """
    Fit PCR for a single component count and return the quantities GoPCA must match.

    Parameters:
    -----------
    X : array-like, shape (n_samples, n_features)
    y : array-like, shape (n_samples,)
    n_components : int
        Number of leading components retained.
    preprocessing : str
        'standardize' (center and scale) or 'mean_center' (center only).
    """
    X = np.asarray(X, dtype=np.float64)
    y = np.asarray(y, dtype=np.float64)

    if preprocessing == 'standardize':
        steps = [("scale", StandardScaler())]
    elif preprocessing == 'mean_center':
        steps = [("scale", StandardScaler(with_mean=True, with_std=False))]
    else:
        raise ValueError(f"unsupported preprocessing: {preprocessing}")

    steps += [
        ("pca", PCA(n_components=n_components, svd_solver="full")),
        ("regression", LinearRegression()),
    ]
    pipeline = Pipeline(steps=steps).fit(X, y)

    scaler = pipeline.named_steps["scale"]
    pca = pipeline.named_steps["pca"]
    regression = pipeline.named_steps["regression"]

    # Collapse the pipeline into original-scale coefficients. components_ stores
    # loading vectors as rows, so components_.T maps score space to variable space.
    theta = pca.components_.T @ regression.coef_
    scale = scaler.scale_ if scaler.scale_ is not None else np.ones(X.shape[1])
    beta = theta / scale
    intercept = float(regression.intercept_ - scaler.mean_ @ beta)

    fitted = pipeline.predict(X)
    residuals = y - fitted
    rmsec = float(np.sqrt(np.mean(residuals ** 2)))
    ss_res = float(np.sum(residuals ** 2))
    ss_tot = float(np.sum((y - y.mean()) ** 2))

    # Guard the collapsed form against the pipeline it came from, so a broken
    # recovery cannot reach the Go tests as a reference.
    collapsed = intercept + X @ beta
    recovery_error = float(np.max(np.abs(collapsed - fitted)))
    assert recovery_error < 1e-8, f"coefficient recovery disagrees with the pipeline: {recovery_error}"

    return {
        "n_samples": int(X.shape[0]),
        "n_features": int(X.shape[1]),
        "n_components": int(n_components),
        "preprocessing": preprocessing,
        "coefficients": beta.tolist(),
        "intercept": intercept,
        "fitted": fitted.tolist(),
        "rmsec": rmsec,
        "r2c": float(1 - ss_res / ss_tot) if ss_tot > 0 else 0.0,
        "response_mean": float(y.mean()),
        "pca_mean_max_abs": float(np.max(np.abs(pca.mean_))),
        "coefficient_recovery_error": recovery_error,
    }


def generate_cv_reference(X, y, component_counts, n_folds, preprocessing='standardize'):
    """
    Cross-validate PCR over a range of component counts using contiguous folds.

    Contiguous (unshuffled) folds are used deliberately. A shuffled design depends
    on the shuffling algorithm, which differs between scikit-learn and Go, so the
    fold membership could not be reproduced across the two. Unshuffled K-fold cuts
    the rows into blocks in order, which both implementations do identically when
    the row count divides evenly by the fold count. That makes the comparison a
    test of the cross-validation machinery rather than of two random generators.

    RMSECV is pooled: every out-of-fold residual enters one mean before the square
    root is taken, so each observation carries equal weight. Note this is NOT what
    GridSearchCV reports for 'neg_root_mean_squared_error', which averages the
    per-fold RMSE and is a slightly smaller number.
    """
    X = np.asarray(X, dtype=np.float64)
    y = np.asarray(y, dtype=np.float64)

    assert X.shape[0] % n_folds == 0, (
        "the row count must divide evenly by the fold count, or scikit-learn and Go "
        "distribute the remainder differently")

    cv = KFold(n_splits=n_folds, shuffle=False)
    with_std = (preprocessing == 'standardize')

    curve = []
    for k in component_counts:
        if k == 0:
            # scikit-learn cannot fit a pipeline with zero components, but the
            # intercept-only model is a meaningful baseline: each held-out row is
            # predicted by the mean of its training fold. Computing it directly
            # keeps the baseline in the comparison rather than silently omitting
            # the one candidate every other candidate must beat.
            oof = np.empty_like(y)
            for train_idx, test_idx in cv.split(X):
                oof[test_idx] = y[train_idx].mean()
        else:
            pipeline = Pipeline(steps=[
                ("scale", StandardScaler(with_mean=True, with_std=with_std)),
                ("pca", PCA(n_components=k, svd_solver="full")),
                ("regression", LinearRegression()),
            ])
            oof = cross_val_predict(pipeline, X, y, cv=cv)
        residuals = oof - y
        rmsecv = float(np.sqrt(np.mean(residuals ** 2)))
        bias = float(np.mean(residuals))
        sep = float(np.std(residuals, ddof=1))
        ss_tot = float(np.sum((y - y.mean()) ** 2))
        curve.append({
            "n_components": int(k),
            "rmsecv": rmsecv,
            "bias": bias,
            "sep": sep,
            "mae": float(np.mean(np.abs(residuals))),
            "q2": float(1 - np.sum(residuals ** 2) / ss_tot) if ss_tot > 0 else 0.0,
        })
    return {"n_folds": int(n_folds), "curve": curve}


def load_dataset(relative_path, response):
    """Load a dataset, splitting numeric predictors from the named response."""
    root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
    df = pd.read_csv(os.path.join(root, relative_path), index_col=0)

    numeric = df.select_dtypes(include=[np.number])
    feature_cols = [c for c in numeric.columns if not c.endswith('#target')]

    X = df[feature_cols].values.astype(np.float64)
    y = df[response].values.astype(np.float64)

    # Keep only rows with an observed response and complete predictors, matching
    # what the Go test feeds in.
    keep = np.isfinite(y) & np.isfinite(X).all(axis=1)
    return X[keep], y[keep], feature_cols


def main():
    print("Generating PCR reference results...")

    output_dir = os.path.join(os.path.dirname(os.path.abspath(__file__)), 'reference_results')
    os.makedirs(output_dir, exist_ok=True)

    cases = [
        # name,          path,                    response,            components
        ('corn_moisture', 'corn/corn.csv', 'Moisture#target', [1, 3, 7, 15]),
        ('corn_protein', 'corn/corn.csv', 'Protein#target', [3, 10]),
        ('wine_flavanoids', 'wine/wine.csv', None, [1, 3, 5, 10]),
    ]

    for name, path, response, component_counts in cases:
        if response is None:
            # Wine carries no numeric #target column, so regress one measured
            # variable on the others. This exercises a well-conditioned, low
            # dimensional case where n greatly exceeds p.
            root = os.path.dirname(os.path.dirname(os.path.abspath(__file__)))
            df = pd.read_csv(os.path.join(root, path), index_col=0)
            numeric = df.select_dtypes(include=[np.number])
            # Exclude the cultivar label, so this is one chemical measurement
            # predicted from the others rather than from a class code.
            feature_cols = [c for c in numeric.columns
                            if not c.endswith('#target')
                            and c not in ('flavanoids', 'classes')]
            X = df[feature_cols].values.astype(np.float64)
            y = df['flavanoids'].values.astype(np.float64)
            keep = np.isfinite(y) & np.isfinite(X).all(axis=1)
            X, y = X[keep], y[keep]
        else:
            X, y, feature_cols = load_dataset(path, response)

        print(f"\n{name}: X={X.shape}, y={y.shape}")

        for preprocessing in ('standardize', 'mean_center'):
            results = []
            for k in component_counts:
                if k > min(X.shape[0] - 1, X.shape[1]):
                    continue
                results.append(generate_pcr_reference(X, y, k, preprocessing))
                print(f"  - {preprocessing}, k={k}: "
                      f"RMSEC={results[-1]['rmsec']:.6g}, R2C={results[-1]['r2c']:.4f}")

            # Checksums let the Go test verify it loaded the same numbers before
            # comparing any model output. Without them a difference in CSV parsing
            # would surface as an unexplained numerical mismatch deep in the
            # comparison rather than as the loading problem it actually is.
            payload = {
                "dataset": name,
                "source": path,
                "response": response if response else "flavanoids",
                "preprocessing": preprocessing,
                "n_samples": int(X.shape[0]),
                "n_features": int(X.shape[1]),
                "x_checksum": float(np.sum(X)),
                "y_checksum": float(np.sum(y)),
                "first_feature": feature_cols[0],
                "last_feature": feature_cols[-1],
                "fits": results,
            }
            # A cross-validated curve, so the Go tests can check the fold
            # machinery and not only the point fit.
            if X.shape[0] % 5 == 0:
                payload["cross_validation"] = generate_cv_reference(
                    X, y, list(range(0, max(component_counts) + 1)), 5, preprocessing)
                cvc = payload["cross_validation"]["curve"]
                print(f"    CV: RMSECV k=1 {cvc[1]['rmsecv']:.6g}, "
                      f"k={cvc[-1]['n_components']} {cvc[-1]['rmsecv']:.6g}")

            output_file = os.path.join(output_dir, f"{name}_pcr_{preprocessing}.json")
            with open(output_file, 'w') as f:
                json.dump(payload, f, indent=2)
            print(f"    saved {os.path.basename(output_file)}")

    print("\nDone.")


if __name__ == '__main__':
    main()
