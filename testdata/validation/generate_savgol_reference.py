#!/usr/bin/env python3
"""
Generate reference values for the Savitzky-Golay filter, for validating the Go
implementation in internal/core/savgol.go against scipy.signal.savgol_filter.

Two groups of cases, checking different things:

1. `synthetic` -- short signals (50 points) chosen to stress the parts of the
   filter that are easy to get subtly wrong. Short signals mean the edge windows
   cover a large fraction of the output, so an error confined to the edges
   cannot hide in a sea of correct interior values. A step and white noise are
   included deliberately: both have strong high-frequency content, which is
   exactly what a derivative amplifies and therefore where a wrong coefficient
   shows up largest.

2. `spectra` -- real near-infrared spectra from testdata/corn, full length.
   These check the filter on the data it exists for, at the window and order a
   spectroscopist would actually reach for.

The Go side asserts agreement to a tight tolerance, because both sides evaluate
the same closed-form linear filter; there is no optimiser here whose path could
legitimately differ.

Edge handling is scipy's default, mode="interp": rather than padding the signal,
the polynomial is fitted to the terminal window and evaluated at the edge
positions. It is recorded here explicitly so the Go implementation is pinned to
that choice rather than agreeing with it by accident.

Spacing is left at scipy's default delta=1.0, matching the Go implementation,
which does not offer a spacing parameter -- it would scale every variable by the
same constant and so cannot change a PCA result.

Reference:
- Savitzky, A. & Golay, M.J.E. (1964). Smoothing and Differentiation of Data by
  Simplified Least Squares Procedures. Analytical Chemistry 36(8), 1627-1639.
- scipy.signal.savgol_filter
"""

import json
import os

import numpy as np
import pandas as pd
from scipy.signal import savgol_filter

# (window_length, polyorder, deriv)
CONFIGS = [
    (5, 2, 0),
    (5, 2, 1),
    (11, 2, 0),
    (11, 2, 1),
    (11, 2, 2),
    (15, 3, 1),
    (21, 3, 2),
    (7, 6, 0),   # polyorder = window-1: the fit interpolates, so this is the identity
    (31, 4, 1),  # wide window, higher order -- the worst-conditioned case here
]

# The real spectra are 700 variables wide, so recording every configuration for
# every sample would dominate the file for little extra evidence: the synthetic
# cases already cover the configuration space, and these exist to confirm the
# filter behaves on real, correlated, noisy data at the settings actually used.
SPECTRA_CONFIGS = [(11, 2, 0), (11, 2, 1), (11, 2, 2), (15, 3, 1)]


def synthetic_signals():
    rng = np.random.default_rng(20260907)
    n = 50
    t = np.arange(n, dtype=float)
    return {
        # Pure noise: no smooth structure at all, so every output value is
        # decided entirely by the coefficients.
        "white_noise": rng.normal(0.0, 1.0, n),
        # A discontinuity is the hardest thing for a polynomial window to
        # follow, and the ringing around it is highly sensitive to the weights.
        "step": np.where(t < n // 2, 0.0, 1.0),
        # Smooth and band-limited, with a known analytic derivative.
        "sine": np.sin(2 * np.pi * t / 17.0),
        # A cubic is reproduced exactly by any fit of order >= 3, and its
        # derivatives are exact too -- so this case pins the scaling.
        "cubic": 1.0 + 2.0 * t - 0.03 * t**2 + 0.0004 * t**3,
        # A narrow peak on a sloping baseline: the shape derivative
        # preprocessing exists to isolate.
        "peak_on_slope": 0.5 + 0.02 * t + 3.0 * np.exp(-((t - 23.0) ** 2) / 8.0),
    }


def corn_spectra(n_rows=2):
    here = os.path.dirname(os.path.abspath(__file__))
    path = os.path.join(here, "..", "corn", "corn.csv")
    frame = pd.read_csv(path, index_col=0)
    wavelengths = [c for c in frame.columns if "#" not in c and c.replace(".", "").isdigit()]
    matrix = frame[wavelengths].to_numpy(float)
    return wavelengths, matrix[:n_rows]


def main():
    reference = {
        "description": "scipy.signal.savgol_filter reference values, mode='interp', delta=1.0",
        "scipy_mode": "interp",
        "delta": 1.0,
        "configs": [
            {"window_length": w, "polyorder": p, "deriv": d} for w, p, d in CONFIGS
        ],
        "spectra_configs": [
            {"window_length": w, "polyorder": p, "deriv": d} for w, p, d in SPECTRA_CONFIGS
        ],
        "synthetic": {},
        "spectra": {},
    }

    for name, signal in synthetic_signals().items():
        entry = {"input": signal.tolist(), "outputs": {}}
        for w, p, d in CONFIGS:
            if w > signal.size:
                continue
            key = f"w{w}_p{p}_d{d}"
            entry["outputs"][key] = savgol_filter(
                signal, w, p, deriv=d, mode="interp"
            ).tolist()
        reference["synthetic"][name] = entry

    wavelengths, spectra = corn_spectra()
    reference["spectra"]["wavelengths"] = wavelengths
    reference["spectra"]["samples"] = []
    for row in spectra:
        entry = {"input": row.tolist(), "outputs": {}}
        for w, p, d in SPECTRA_CONFIGS:
            key = f"w{w}_p{p}_d{d}"
            entry["outputs"][key] = savgol_filter(
                row, w, p, deriv=d, mode="interp"
            ).tolist()
        reference["spectra"]["samples"].append(entry)

    here = os.path.dirname(os.path.abspath(__file__))
    output_dir = os.path.join(here, "reference_results")
    os.makedirs(output_dir, exist_ok=True)
    out = os.path.join(output_dir, "savgol_reference.json")
    with open(out, "w") as handle:
        json.dump(reference, handle, indent=2)

    n_cases = sum(len(v["outputs"]) for v in reference["synthetic"].values())
    n_cases += sum(len(s["outputs"]) for s in reference["spectra"]["samples"])
    print(f"Wrote {out} ({n_cases} filtered signals)")


if __name__ == "__main__":
    main()
