#!/usr/bin/env python3
"""Prepare pollofpolls polling data for downstream analysis."""

from __future__ import annotations

import argparse
import csv
import re
from pathlib import Path

PARENTHESIZED_INTEGER = re.compile(r"\s*\(\s*\d+\s*\)")


def clean_row(row: list[str]) -> list[str]:
    """Remove parenthesized integers from non-date columns."""
    cleaned: list[str] = []
    for idx, cell in enumerate(row):
        text = cell.strip()
        if idx > 0 and text:
            text = PARENTHESIZED_INTEGER.sub("", text).strip()
        cleaned.append(text)
    return cleaned


def prepare(input_path: Path, output_path: Path) -> None:
    with input_path.open("r", encoding="cp1252", newline="") as raw_file:
        reader = csv.reader(raw_file, delimiter=";")
        rows = [clean_row(row) for row in reader]

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8", newline="") as prepared_file:
        writer = csv.writer(prepared_file, delimiter=";")
        writer.writerows(rows)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description="Prepare pollofpolls polling data for analysis."
    )
    parser.add_argument(
        "--input",
        type=Path,
        default=Path(__file__).with_name("pollofpolls_raw.csv"),
        help="Path to the raw pollofpolls CSV file.",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).with_name("pollofpolls.csv"),
        help="Path to write the prepared CSV file.",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    prepare(args.input, args.output)


if __name__ == "__main__":
    main()
