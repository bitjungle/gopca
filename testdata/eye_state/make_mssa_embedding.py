#!/usr/bin/env python3
"""
Create an MSSA-style embedded (trajectory) matrix from the GoPCA EEG Eye State CSV.

Input format expected:
    time, AF3, F7, ..., AF4, eye_state

Output format:
    window_start_time, AF3_lag0, F7_lag0, ..., AF4_lag0,
                       AF3_lag1, F7_lag1, ..., AF4_lag1,
                       ...,
                       AF3_lag<L-1>, ..., AF4_lag<L-1>, eye_state

Rows are overlapping sliding windows. The target is taken from the final time point
in each window by default, which is often the most natural convention for a
window ending at that observation.
"""

from __future__ import annotations

import argparse
from pathlib import Path
from typing import Literal

import numpy as np
import pandas as pd


TargetMode = Literal["last", "first", "majority", "drop"]


def majority_label(values: np.ndarray) -> object:
    """Return the most frequent label in a 1D array. Ties are resolved by first occurrence."""
    counts: dict[object, int] = {}
    first_pos: dict[object, int] = {}
    for i, value in enumerate(values):
        counts[value] = counts.get(value, 0) + 1
        first_pos.setdefault(value, i)
    return max(counts, key=lambda value: (counts[value], -first_pos[value]))


def make_mssa_embedding(
    df: pd.DataFrame,
    window_length: int,
    time_col: str = "time",
    target_col: str = "eye_state",
    target_mode: TargetMode = "last",
) -> pd.DataFrame:
    """Create an MSSA-style lagged embedding from a multivariate time series."""
    if window_length < 2:
        raise ValueError("window_length must be at least 2")

    if time_col not in df.columns:
        raise ValueError(f"Time column {time_col!r} not found in input CSV")

    has_target = target_col in df.columns

    excluded = {time_col}
    if has_target:
        excluded.add(target_col)

    signal_cols = [col for col in df.columns if col not in excluded]
    if not signal_cols:
        raise ValueError("No signal columns found after excluding time and target columns")

    n_rows = len(df)
    n_windows = n_rows - window_length + 1
    if n_windows <= 0:
        raise ValueError(
            f"window_length={window_length} is too large for {n_rows} rows"
        )

    X = df[signal_cols].to_numpy(dtype=float)

    # Build lagged blocks: lag0, lag1, ..., lag(window_length - 1)
    blocks = []
    out_cols = []
    for lag in range(window_length):
        blocks.append(X[lag : lag + n_windows, :])
        out_cols.extend([f"{col}_lag{lag}" for col in signal_cols])

    embedded = np.hstack(blocks)
    out = pd.DataFrame(embedded, columns=out_cols)

    # Use window start time as the first column / observation identifier.
    # Do not include time as a PCA variable.
    out.insert(0, "window_start_time", df[time_col].iloc[:n_windows].to_numpy())

    if has_target and target_mode != "drop":
        target_values = df[target_col].to_numpy()
        if target_mode == "first":
            out[target_col] = target_values[:n_windows]
        elif target_mode == "last":
            out[target_col] = target_values[window_length - 1 :]
        elif target_mode == "majority":
            out[target_col] = [
                majority_label(target_values[i : i + window_length])
                for i in range(n_windows)
            ]
        else:
            raise ValueError(f"Unknown target_mode: {target_mode}")

    return out


def main() -> None:
    parser = argparse.ArgumentParser(
        description="Convert raw EEG Eye State CSV into an MSSA-style embedded trajectory matrix."
    )
    parser.add_argument(
        "--input",
        default="eeg_eye_state.csv",
        help="Input CSV file. Default: eeg_eye_state.csv",
    )
    parser.add_argument(
        "--output",
        default=None,
        help="Output CSV file. Default: eeg_eye_state_mssa_L<window_length>.csv",
    )
    parser.add_argument(
        "--window-length",
        type=int,
        default=16,
        help="Number of consecutive time points per embedded window. Default: 16",
    )
    parser.add_argument(
        "--time-col",
        default="time",
        help="Name of the time column. Default: time",
    )
    parser.add_argument(
        "--target-col",
        default="eye_state",
        help="Name of the target column. Default: eye_state",
    )
    parser.add_argument(
        "--target-mode",
        choices=["last", "first", "majority", "drop"],
        default="last",
        help=(
            "How to assign target labels to windows: last, first, majority, or drop. "
            "Default: last"
        ),
    )

    args = parser.parse_args()

    input_path = Path(args.input)
    output_path = Path(args.output) if args.output else Path(
        f"eeg_eye_state_mssa_L{args.window_length}.csv"
    )

    df = pd.read_csv(input_path)
    embedded = make_mssa_embedding(
        df,
        window_length=args.window_length,
        time_col=args.time_col,
        target_col=args.target_col,
        target_mode=args.target_mode,
    )
    embedded.to_csv(output_path, index=False)

    signal_cols = [
        col for col in df.columns if col not in {args.time_col, args.target_col}
    ]
    print(f"Input rows: {len(df)}")
    print(f"Signal channels: {len(signal_cols)}")
    print(f"Window length: {args.window_length}")
    print(f"Output shape: {embedded.shape[0]} rows x {embedded.shape[1]} columns")
    print(f"Wrote: {output_path}")


if __name__ == "__main__":
    main()
