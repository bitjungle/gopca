#!/usr/bin/env python3
"""Convert pollofpolls CSV to use decimal dots and comma field separators."""

from __future__ import annotations

import argparse
import csv
from pathlib import Path


def normalize(input_path: Path, output_path: Path) -> None:
    with input_path.open("r", encoding="utf-8", newline="") as source:
        reader = csv.reader(source, delimiter=";")
        rows: list[list[str]] = []
        for row_idx, row in enumerate(reader):
            normalized: list[str] = []
            for col_idx, cell in enumerate(row):
                text = cell.strip()
                if row_idx > 0 and col_idx > 0 and text:
                    text = text.replace(",", ".")
                normalized.append(text)
            rows.append(normalized)

    output_path.parent.mkdir(parents=True, exist_ok=True)
    with output_path.open("w", encoding="utf-8", newline="") as target:
        writer = csv.writer(target, delimiter=",")
        writer.writerows(rows)


def parse_args() -> argparse.Namespace:
    parser = argparse.ArgumentParser(
        description=(
            "Switch pollofpolls data to dot decimals and comma field separators."
        )
    )
    parser.add_argument(
        "--input",
        type=Path,
        default=Path(__file__).with_name("pollofpolls.csv"),
        help="Path to the semicolon separated pollofpolls CSV file.",
    )
    parser.add_argument(
        "--output",
        type=Path,
        default=Path(__file__).with_name("pollofpolls_commas.csv"),
        help="Path to write the converted CSV file.",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    normalize(args.input, args.output)


if __name__ == "__main__":
    main()
