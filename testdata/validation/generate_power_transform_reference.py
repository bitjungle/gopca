#!/usr/bin/env python3
"""
Generate reference values for the Box-Cox and Yeo-Johnson power transforms,
for validating the Go implementation in pkg/transform.

Two things are recorded, and they check different claims:

1. `at_lambda` -- the transform evaluated at a *given* lambda. This checks the
   formula alone, and is compared with a tight tolerance because both sides are
   computing the same closed-form expression.

2. `fitted_lambda` -- the lambda scipy's maximum-likelihood estimator settles
   on. This checks the optimiser, and is compared loosely: scipy uses Brent
   while the Go code uses golden-section search, so the two agree on where the
   maximum is rather than on the arithmetic that found it.

Separating them matters. Comparing only the end result would leave a correct
formula and a mis-converging search indistinguishable from the reverse.

References:
- Box, G.E.P. & Cox, D.R. (1964). An Analysis of Transformations. JRSS B 26(2).
- Yeo, I.-K. & Johnson, R.A. (2000). A New Family of Power Transformations to
  Improve Normality or Symmetry. Biometrika 87(4), 954-959.
- scipy.stats.boxcox, scipy.stats.yeojohnson
"""

import json
import os

import numpy as np
from scipy import stats


def build_cases():
    rng = np.random.default_rng(20260907)

    positive = {
        # Right-skewed, the case Box-Cox is normally reached for.
        "lognormal": np.exp(rng.normal(0.0, 1.0, 60)),
        # Counts.
        "poisson_positive": rng.poisson(5, 60).astype(float) + 1.0,
        # Already close to symmetric: lambda should come out near 1, meaning
        # "barely transform this".
        "near_normal": rng.normal(50.0, 5.0, 60),
        # Concentrations spanning orders of magnitude.
        "wide_range": np.array([0.001, 0.01, 0.1, 1.0, 10.0, 100.0, 1000.0] * 8),
    }

    # Yeo-Johnson also accepts zero and negative values, which is where log and
    # sqrt refuse.
    signed = {
        "with_zeros": np.concatenate([rng.poisson(2, 40).astype(float),
                                      np.zeros(20)]),
        "negative_and_positive": rng.normal(0.0, 3.0, 60),
        "all_negative": -np.abs(rng.normal(5.0, 2.0, 60)),
    }

    return positive, signed


def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    output_dir = os.path.join(script_dir, "reference_results")
    os.makedirs(output_dir, exist_ok=True)

    positive, signed = build_cases()
    reference = {"boxcox": {}, "yeojohnson": {}}

    # Lambdas chosen to cover the special cases (0 for Box-Cox is the log;
    # 0 and 2 are the special cases for Yeo-Johnson) and ordinary values.
    probe_lambdas = [-1.0, -0.5, 0.0, 0.5, 1.0, 1.5, 2.0]

    for name, values in positive.items():
        fitted = stats.boxcox_normmax(values, method="mle")
        reference["boxcox"][name] = {
            "input": values.tolist(),
            "fitted_lambda": float(fitted),
            "at_lambda": {
                f"{lam:g}": stats.boxcox(values, lmbda=lam).tolist()
                for lam in probe_lambdas
            },
        }

    for name, values in {**positive, **signed}.items():
        _, fitted = stats.yeojohnson(values)
        reference["yeojohnson"][name] = {
            "input": values.tolist(),
            "fitted_lambda": float(fitted),
            "at_lambda": {
                f"{lam:g}": stats.yeojohnson(values, lmbda=lam).tolist()
                for lam in probe_lambdas
            },
        }

    out_path = os.path.join(output_dir, "power_transform_reference.json")
    with open(out_path, "w") as handle:
        json.dump(reference, handle, indent=2)

    print(f"Wrote Box-Cox ({len(reference['boxcox'])}) and Yeo-Johnson "
          f"({len(reference['yeojohnson'])}) reference cases to {out_path}")
    print(f"numpy {np.__version__}, scipy {stats.__name__}")


if __name__ == "__main__":
    main()
