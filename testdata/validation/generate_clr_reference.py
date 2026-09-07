#!/usr/bin/env python3
"""
Generate reference centred log-ratio (CLR) values for validating the Go
implementation in pkg/transform.

The CLR transform of a composition x = (x_1 … x_D), all parts strictly
positive, is

    clr(x)_i = ln( x_i / g(x) ),    g(x) = (∏_j x_j)^(1/D)

This script computes it directly from the definition -- forming the geometric
mean as a product and dividing -- rather than by the log-mean identity the Go
code uses. That is deliberate: an independent route to the same numbers is
what makes this a check rather than a restatement. The two agree to within
floating-point tolerance for well-conditioned inputs, which is what the Go
test asserts.

Zero handling is NOT part of the reference. Multiplicative replacement is a
decision about the data, made before the transform, and the reference covers
the transform.

References:
- Aitchison, J. (1986). The Statistical Analysis of Compositional Data.
  Chapman & Hall, Chapter 4.
- Egozcue et al. (2003). Isometric Logratio Transformations for Compositional
  Data Analysis. Mathematical Geology 35(3), 279-300.
"""

import json
import os

import numpy as np


def clr_from_definition(row):
    """CLR computed as ln(x_i / geometric_mean), directly from the definition."""
    row = np.asarray(row, dtype=float)
    geometric_mean = np.prod(row) ** (1.0 / len(row))
    return np.log(row / geometric_mean)


def build_cases():
    """Cases chosen to exercise properties the Go implementation must hold."""
    rng = np.random.default_rng(20260907)

    cases = []

    # A closed composition: percentages summing to 100.
    closed = np.array([
        [50.0, 30.0, 20.0],
        [10.0, 10.0, 80.0],
        [33.4, 33.3, 33.3],
    ])
    cases.append(("closed_percentages", closed))

    # A subcomposition: positive, but the rows do not sum to a constant.
    # CLR is defined here too, and this is the case that would break an
    # implementation that assumed closure.
    subcomposition = np.array([
        [1.0, 2.0, 4.0],
        [10.0, 5.0, 1.0],
        [0.5, 0.25, 8.0],
    ])
    cases.append(("subcomposition", subcomposition))

    # Equal parts: every clr value must be exactly zero, whatever the level.
    cases.append(("equal_parts", np.array([
        [1.0, 1.0, 1.0],
        [25.0, 25.0, 25.0, ],
    ], dtype=object) if False else np.array([
        [1.0, 1.0, 1.0],
        [25.0, 25.0, 25.0],
    ])))

    # Scale invariance: a row and the same row multiplied by a constant must
    # give identical clr values, because clr describes ratios.
    base = np.array([[2.0, 3.0, 5.0]])
    cases.append(("scale_invariance_x1", base))
    cases.append(("scale_invariance_x1000", base * 1000.0))

    # Wide and small-valued, where forming the product first would underflow.
    # Forty parts each around 0.02: the product is ~1e-68, still representable,
    # but this is the shape that makes the naive route fragile.
    wide = rng.uniform(0.01, 0.03, size=(3, 40))
    cases.append(("wide_small_parts", wide))

    # Realistic geochemistry: major oxides, weight percent.
    oxides = np.array([
        [62.1, 15.8, 6.2, 3.1, 5.4, 3.9, 2.1, 1.4],
        [48.3, 17.2, 10.1, 8.4, 9.9, 3.2, 1.6, 1.3],
        [72.4, 13.9, 2.8, 0.9, 1.7, 3.6, 4.2, 0.5],
    ])
    cases.append(("major_oxides", oxides))

    return cases


def main():
    script_dir = os.path.dirname(os.path.abspath(__file__))
    output_dir = os.path.join(script_dir, "reference_results")
    os.makedirs(output_dir, exist_ok=True)

    reference = {}
    for name, matrix in build_cases():
        transformed = np.array([clr_from_definition(row) for row in matrix])
        reference[name] = {
            "input": matrix.tolist(),
            "clr": transformed.tolist(),
            # Every clr row sums to zero by construction. Recorded so the Go
            # test can assert the property as well as the values.
            "row_sums": transformed.sum(axis=1).tolist(),
        }

    out_path = os.path.join(output_dir, "clr_reference.json")
    with open(out_path, "w") as handle:
        json.dump(reference, handle, indent=2)

    print(f"Wrote {len(reference)} CLR reference cases to {out_path}")
    print(f"numpy {np.__version__}")


if __name__ == "__main__":
    main()
