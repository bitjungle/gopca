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
from sklearn.model_selection import (KFold, LeaveOneGroupOut, LeaveOneOut,
                                     cross_val_predict)
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


def _fold_curve(X, y, component_counts, splitter, groups, preprocessing):
    """Cross-validate over any deterministic sklearn splitter.

    One decomposition per fold serves every candidate count: PCA's leading k
    components do not depend on how many were requested, so the first k score
    columns of a full fit are exactly what PCA(n_components=k) would give. This
    mirrors what GoPCA does internally and makes leave-one-out affordable.

    Returns per-candidate pooled and per-fold statistics. The per-fold ones are
    the point of this function: GoPCA reports RMSECVMean and RMSECVSE alongside
    the pooled RMSECV, they feed the one-standard-error selection rule, and
    nothing compared them against a reference until now.
    """
    X = np.asarray(X, dtype=np.float64)
    y = np.asarray(y, dtype=np.float64)
    with_std = (preprocessing == 'standardize')
    kmax = max(component_counts)

    oof = {k: np.empty_like(y) for k in component_counts}
    fold_rmse = {k: [] for k in component_counts}
    fold_mae = {k: [] for k in component_counts}

    for train_idx, test_idx in splitter.split(X, y, groups):
        scaler = StandardScaler(with_mean=True, with_std=with_std).fit(X[train_idx])
        Xtr, Xte = scaler.transform(X[train_idx]), scaler.transform(X[test_idx])
        pca = PCA(n_components=min(kmax, len(train_idx) - 1, X.shape[1]),
                  svd_solver='full').fit(Xtr)
        Ttr, Tte = pca.transform(Xtr), pca.transform(Xte)

        for k in component_counts:
            if k == 0:
                pred = np.full(len(test_idx), y[train_idx].mean())
            else:
                kk = min(k, Ttr.shape[1])
                pred = LinearRegression().fit(Ttr[:, :kk], y[train_idx]).predict(Tte[:, :kk])
            oof[k][test_idx] = pred
            resid = pred - y[test_idx]
            fold_rmse[k].append(float(np.sqrt(np.mean(resid ** 2))))
            fold_mae[k].append(float(np.mean(np.abs(resid))))

    ss_tot = float(np.sum((y - y.mean()) ** 2))
    curve = []
    for k in component_counts:
        resid = oof[k] - y
        fr = np.asarray(fold_rmse[k]); fm = np.asarray(fold_mae[k])
        # Standard error as GoPCA computes it: the sample standard deviation of
        # the per-fold values (ddof=1) divided by the square root of the count.
        se = lambda a: float(np.std(a, ddof=1) / np.sqrt(len(a))) if len(a) > 1 else 0.0
        curve.append({
            "n_components": int(k),
            "rmsecv": float(np.sqrt(np.mean(resid ** 2))),
            "rmsecv_mean": float(np.mean(fr)),
            "rmsecv_se": se(fr),
            "bias": float(np.mean(resid)),
            "sep": float(np.std(resid, ddof=1)),
            "mae": float(np.mean(np.abs(resid))),
            "mae_se": se(fm),
            "q2": float(1 - np.sum(resid ** 2) / ss_tot) if ss_tot > 0 else 0.0,
        })
    return curve


def generate_cv_designs(X, y, component_counts):
    """Reference curves for the fold layouts GoPCA supports beyond plain K-fold.

    Only contiguous K-fold was ever validated against scikit-learn. GroupKFold is
    one code path in GoPCA serving K-fold, leave-one-out, grouped K-fold and
    leave-one-group-out, so a fold assignment that disagreed with scikit-learn
    would have shown up in none of the existing assertions.

    Both designs here are chosen because their partition is *unique*. Shuffled
    K-fold and balanced GroupKFold both depend on assignment heuristics that
    differ between implementations, so a disagreement would say nothing about
    the machinery. With one row per fold, or one group per fold, there is only
    one possible answer and any difference is a real one.
    """
    n = X.shape[0]
    designs = []

    designs.append({
        "design": "leave_one_out",
        "n_folds": int(n),
        "curve": _fold_curve(X, y, component_counts, LeaveOneOut(), None, 'standardize'),
    })

    # Interleaved groups, deliberately not contiguous blocks: a grouping that
    # happens to align with row order would be indistinguishable from ordinary
    # contiguous K-fold and would test nothing about keeping groups together.
    groups = np.arange(n) % 8
    designs.append({
        "design": "leave_one_group_out",
        "n_folds": 8,
        "groups": groups.tolist(),
        "curve": _fold_curve(X, y, component_counts, LeaveOneGroupOut(), groups, 'standardize'),
    })

    return designs


def generate_semisupervised_reference(X, y, n_components, unlabelled_every=3):
    """Reference for a fit where only some rows carry a response.

    This is the path bronir2 exercises and the one where a leak makes the result
    *better*, so no ordinary assertion fails when it is present. GoPCA fits the
    decomposition on the labelled rows plus every unlabelled row -- the
    predictors of an unmeasured sample still carry usable structure, and PCA
    never sees the response -- and fits the regression on the labelled rows
    alone.

    Reproduced here explicitly: scaler and PCA on all rows, LinearRegression on
    the labelled subset only.
    """
    X = np.asarray(X, dtype=np.float64)
    y = np.asarray(y, dtype=np.float64)

    labelled = np.ones(len(y), dtype=bool)
    labelled[::unlabelled_every] = False   # deterministic, no RNG to agree about

    scaler = StandardScaler().fit(X)
    Xs = scaler.transform(X)
    pca = PCA(n_components=n_components, svd_solver='full').fit(Xs)
    scores = pca.transform(Xs)

    reg = LinearRegression().fit(scores[labelled], y[labelled])
    fitted = reg.predict(scores[labelled])
    resid = fitted - y[labelled]

    theta = pca.components_.T @ reg.coef_
    beta = theta / scaler.scale_
    intercept = float(reg.intercept_ - scaler.mean_ @ beta)

    ss_tot = float(np.sum((y[labelled] - y[labelled].mean()) ** 2))
    return {
        "n_components": int(n_components),
        "labelled_rows": [int(i) for i in np.flatnonzero(labelled)],
        "n_labelled": int(labelled.sum()),
        "n_decomposition_rows": int(len(y)),
        "coefficients": beta.tolist(),
        "intercept": intercept,
        "fitted": fitted.tolist(),
        "rmsec": float(np.sqrt(np.mean(resid ** 2))),
        "r2c": float(1 - np.sum(resid ** 2) / ss_tot) if ss_tot > 0 else 0.0,
    }


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
                ks = list(range(0, max(component_counts) + 1))
                payload["cross_validation"] = generate_cv_reference(
                    X, y, ks, 5, preprocessing)
                cvc = payload["cross_validation"]["curve"]
                print(f"    CV: RMSECV k=1 {cvc[1]['rmsecv']:.6g}, "
                      f"k={cvc[-1]['n_components']} {cvc[-1]['rmsecv']:.6g}")

                # The per-fold statistics, computed by the same routine the new
                # designs use. Cross-checked against the pooled figures above,
                # which come from an independent cross_val_predict path: if the
                # two disagree the reference itself is wrong, and that must not
                # be discovered later as an unexplained Go/Python mismatch.
                fast = _fold_curve(X, y, ks, KFold(n_splits=5, shuffle=False),
                                   None, preprocessing)
                for a, b in zip(cvc, fast):
                    assert abs(a["rmsecv"] - b["rmsecv"]) < 1e-10, (
                        f"the two reference paths disagree at k={a['n_components']}: "
                        f"{a['rmsecv']} vs {b['rmsecv']}")
                    a["rmsecv_mean"] = b["rmsecv_mean"]
                    a["rmsecv_se"] = b["rmsecv_se"]
                    a["mae_se"] = b["mae_se"]

            # Alternative fold layouts, and the partially-labelled path. Only
            # for the standardized variant: these check the fold machinery and
            # the row filtering, neither of which depends on the scaling choice,
            # and leave-one-out on 80 spectra is not cheap.
            if preprocessing == 'standardize':
                design_ks = [k for k in (0, 1, 3, 7) if k <= min(X.shape[0] - 2, X.shape[1])]
                payload["cv_designs"] = generate_cv_designs(X, y, design_ks)
                for d in payload["cv_designs"]:
                    last = d["curve"][-1]
                    print(f"    {d['design']}: RMSECV k={last['n_components']} "
                          f"{last['rmsecv']:.6g} (mean-of-folds {last['rmsecv_mean']:.6g})")

                semi_k = min(5, X.shape[1], X.shape[0] // 2)
                payload["semi_supervised"] = generate_semisupervised_reference(X, y, semi_k)
                ss = payload["semi_supervised"]
                print(f"    semi-supervised: {ss['n_labelled']} labelled of "
                      f"{ss['n_decomposition_rows']}, RMSEC={ss['rmsec']:.6g}")

            output_file = os.path.join(output_dir, f"{name}_pcr_{preprocessing}.json")
            with open(output_file, 'w') as f:
                json.dump(payload, f, indent=2)
            print(f"    saved {os.path.basename(output_file)}")

    print("\nDone.")


if __name__ == '__main__':
    main()
