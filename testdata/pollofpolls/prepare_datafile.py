#!/usr/bin/env python3
"""Prepare pollofpolls polling data for downstream analysis."""

from __future__ import annotations

import argparse
import csv
import re
from pathlib import Path
from typing import NamedTuple

PARENTHESIZED_INTEGER = re.compile(r"\s*\(\s*\d+\s*\)")

class GovernmentPeriod(NamedTuple):
    """Represents a government period with start date, end date, and parties."""

    fra: str  # Start date (YYYY-MM-DD)
    til: str  # End date (YYYY-MM-DD) or empty string for current
    regjeringspartier: str  # Government parties


# Mapping of Norwegian month names to month numbers
NORWEGIAN_MONTHS = {
    "januar": "01",
    "februar": "02",
    "mars": "03",
    "april": "04",
    "mai": "05",
    "juni": "06",
    "juli": "07",
    "august": "08",
    "september": "09",
    "oktober": "10",
    "november": "11",
    "desember": "12",
}


def convert_date(date_str: str) -> str:
    """Convert Norwegian date format to ISO format (YYYY-MM).

    Examples:
        Juli '24 -> 2024-07
        Februar '23 -> 2023-02
        Desember '19 -> 2019-12
        Mai '10 -> 2010-05
    """
    date_str = date_str.strip()
    if not date_str:
        return date_str

    # Match pattern: Month 'YY
    match = re.match(r"(\w+)\s+'(\d{2})", date_str)
    if not match:
        return date_str  # Return unchanged if pattern doesn't match

    month_name, year_abbrev = match.groups()
    month_name_lower = month_name.lower()

    if month_name_lower not in NORWEGIAN_MONTHS:
        return date_str  # Return unchanged if month not recognized

    month_num = NORWEGIAN_MONTHS[month_name_lower]

    # Convert abbreviated year to full year
    # Assume years 00-50 are 2000-2050, years 51-99 are 1951-1999
    year_num = int(year_abbrev)
    if year_num <= 50:
        full_year = 2000 + year_num
    else:
        full_year = 1900 + year_num

    return f"{full_year}-{month_num}"


def load_government_periods(regjeringer_path: Path) -> list[GovernmentPeriod]:
    """Load government periods from regjeringer.csv."""
    periods: list[GovernmentPeriod] = []
    with regjeringer_path.open("r", encoding="utf-8", newline="") as csvfile:
        reader = csv.DictReader(csvfile)
        for row in reader:
            periods.append(
                GovernmentPeriod(
                    fra=row["fra"].strip(),
                    til=row["til"].strip(),
                    regjeringspartier=row["regjeringspartier"].strip(),
                )
            )
    return periods


def match_government(date_str: str, periods: list[GovernmentPeriod]) -> str:
    """Match a date (YYYY-MM) to a government period.

    Args:
        date_str: Date in YYYY-MM format
        periods: List of government periods

    Returns:
        The regjeringspartier value for the matching period, or empty string if no match
    """
    if not date_str:
        return ""

    # Convert YYYY-MM to comparable format by adding -01 for day
    date_with_day = f"{date_str}-01"

    for period in periods:
        # Check if date is within this period
        # Date must be >= fra
        if date_with_day < period.fra:
            continue

        # Date must be < til (or til is empty for current government)
        if period.til and date_with_day >= period.til:
            continue

        return period.regjeringspartier

    return ""  # No match found


def clean_row(row: list[str]) -> list[str]:
    """Remove parenthesized integers from non-date columns and convert dates."""
    cleaned: list[str] = []
    for idx, cell in enumerate(row):
        text = cell.strip()
        if idx == 0:
            # First column: convert date format
            text = convert_date(text)
        elif text:
            # Other columns: remove parenthesized integers
            text = PARENTHESIZED_INTEGER.sub("", text).strip()
        cleaned.append(text)
    return cleaned


def prepare(
    input_path: Path, output_path: Path, regjeringer_path: Path | None = None
) -> None:
    """Prepare pollofpolls data with optional government period annotation.

    Args:
        input_path: Path to raw pollofpolls CSV file
        output_path: Path to write prepared CSV file
        regjeringer_path: Optional path to regjeringer.csv for government periods
    """
    # Load government periods if path provided
    periods: list[GovernmentPeriod] = []
    if regjeringer_path and regjeringer_path.exists():
        periods = load_government_periods(regjeringer_path)

    # Clean the raw data
    with input_path.open("r", encoding="cp1252", newline="") as raw_file:
        reader = csv.reader(raw_file, delimiter=";")
        rows = [clean_row(row) for row in reader]

    # Add government column if periods are available
    if periods:
        # Add "regjering" to header row
        if rows:
            rows[0].append("regjering")

        # Add government value to each data row
        for row in rows[1:]:
            if row:  # Skip empty rows
                date_str = row[0] if row else ""
                government = match_government(date_str, periods)
                row.append(government)

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
    parser.add_argument(
        "--regjeringer",
        type=Path,
        default=Path(__file__).with_name("regjeringer.csv"),
        help="Path to the regjeringer CSV file with government periods.",
    )
    return parser.parse_args()


def main() -> None:
    args = parse_args()
    prepare(args.input, args.output, args.regjeringer)


if __name__ == "__main__":
    main()
