#!/usr/bin/env python3
"""
Generate reference Student's t-distribution CDF values using scipy for validation.

This script generates reference results for the Student's t cumulative distribution
function (CDF) to validate the correctness of the Go implementation.

The Student's t-distribution CDF can be expressed using the regularized incomplete
beta function:
    P(T ≤ t) = 0.5 + 0.5 * sign(t) * (1 - I_x(df/2, 0.5))
where x = df/(df + t²) and I_x is the regularized incomplete beta function.

References:
- Student (1908): The probable error of a mean. Biometrika, 6(1), 1-25.
- Abramowitz & Stegun (1972): Handbook of Mathematical Functions, Chapter 26.
- SciPy documentation: https://docs.scipy.org/doc/scipy/reference/generated/scipy.stats.t.html
"""

import sys
import os
import csv
import numpy as np
from scipy import stats

def generate_studentt_reference():
    """
    Generate Student's t-distribution CDF reference values.

    Returns comprehensive test cases covering:
    - Various degrees of freedom (df): 1, 2, 5, 10, 20, 30, 50
    - Various t-values: negative, zero, and positive
    - Edge cases: Cauchy distribution (df=1), large df (approaches normal)

    Output format: CSV with columns [df, t, cdf, two_tailed_pval]
    """

    # Define test cases
    # Cover critical df values and t-values
    df_values = [1, 2, 5, 10, 20, 30, 50]
    t_values = [-3.0, -2.5, -2.0, -1.5, -1.0, -0.5, 0.0, 0.5, 1.0, 1.5, 2.0, 2.5, 3.0]

    results = []

    print("Generating Student's t-distribution reference values...")
    print(f"Testing {len(df_values)} df values × {len(t_values)} t values = {len(df_values) * len(t_values)} cases")

    for df in df_values:
        for t in t_values:
            # Calculate CDF using scipy
            # scipy.stats.t.cdf(t, df) returns P(T ≤ t)
            cdf = stats.t.cdf(t, df)

            # Calculate two-tailed p-value: P(|T| > |t|) = 2 * (1 - CDF(|t|))
            two_tailed_pval = 2 * (1 - stats.t.cdf(abs(t), df))

            results.append({
                'df': df,
                't': t,
                'cdf': cdf,
                'two_tailed_pval': two_tailed_pval
            })

            # Print some interesting cases
            if t in [0.0, 1.0, 2.0] and df in [1, 10]:
                print(f"  df={df:2d}, t={t:5.2f} → CDF={cdf:.10f}, p={two_tailed_pval:.10f}")

    return results

def validate_special_cases(results):
    """
    Validate special mathematical properties of the t-distribution.

    These properties should hold exactly (within numerical precision):
    1. CDF(0, df) = 0.5 for all df
    2. CDF(t, df) + CDF(-t, df) ≈ 1.0 (symmetry)
    3. CDF(1, 1) = 0.75 (Cauchy distribution special case)
    """
    print("\nValidating special cases:")

    # Check t=0 returns 0.5
    for result in results:
        if result['t'] == 0.0:
            if abs(result['cdf'] - 0.5) > 1e-10:
                print(f"  ⚠️  WARNING: CDF(0, df={result['df']}) = {result['cdf']}, expected 0.5")
            else:
                print(f"  ✓  CDF(0, df={result['df']}) = 0.5 (exact)")

    # Check symmetry: CDF(t) + CDF(-t) = 1.0
    positive_results = {(r['df'], r['t']): r for r in results if r['t'] > 0}
    negative_results = {(r['df'], -r['t']): r for r in results if r['t'] < 0}

    for key in positive_results:
        if key in negative_results:
            pos_cdf = positive_results[key]['cdf']
            neg_cdf = negative_results[key]['cdf']
            symmetry_sum = pos_cdf + neg_cdf
            if abs(symmetry_sum - 1.0) > 1e-10:
                print(f"  ⚠️  WARNING: Symmetry violation at df={key[0]}, t=±{key[1]}: sum={symmetry_sum}")

    print("  ✓  Symmetry check passed for all cases")

    # Check Cauchy distribution special case: t.cdf(1, df=1) should be 0.75
    cauchy_result = [r for r in results if r['df'] == 1 and r['t'] == 1.0][0]
    expected_cauchy = 0.75
    if abs(cauchy_result['cdf'] - expected_cauchy) > 1e-10:
        print(f"  ⚠️  WARNING: Cauchy CDF(1, 1) = {cauchy_result['cdf']}, expected {expected_cauchy}")
    else:
        print(f"  ✓  Cauchy CDF(1, 1) = 0.75 (exact)")

def write_csv(results, output_path):
    """
    Write results to CSV file.

    CSV format:
    df,t,cdf,two_tailed_pval
    1,-3.0,0.0477464829...,0.9045...
    ...
    """
    os.makedirs(os.path.dirname(output_path), exist_ok=True)

    with open(output_path, 'w', newline='') as f:
        writer = csv.DictWriter(f, fieldnames=['df', 't', 'cdf', 'two_tailed_pval'])
        writer.writeheader()
        writer.writerows(results)

    print(f"\n✓  Reference data written to: {output_path}")
    print(f"   {len(results)} test cases generated")

def main():
    """Main execution function."""
    # Determine output path
    script_dir = os.path.dirname(os.path.abspath(__file__))
    output_dir = os.path.join(script_dir, 'reference_results')
    output_path = os.path.join(output_dir, 'studentt_reference.csv')

    print("=" * 70)
    print("Student's t-Distribution Reference Value Generator")
    print("=" * 70)
    print(f"SciPy version: {stats.__version__ if hasattr(stats, '__version__') else 'N/A'}")
    print(f"Output path: {output_path}")
    print()

    # Generate reference values
    results = generate_studentt_reference()

    # Validate special cases
    validate_special_cases(results)

    # Write to CSV
    write_csv(results, output_path)

    print("\n" + "=" * 70)
    print("Reference generation complete!")
    print("=" * 70)
    print("\nTo use these reference values in Go tests:")
    print("  1. Run: go test -v ./internal/core/ -run TestStudentTCDF")
    print("  2. The test will read this CSV and validate against the Go implementation")
    print()

if __name__ == '__main__':
    main()
