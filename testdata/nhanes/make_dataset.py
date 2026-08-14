"""NHANES 2017-2018 Body Measures (BMX_J) -> clean anthropometric PCA dataset.

Source: https://wwwn.cdc.gov/Nchs/Data/Nhanes/Public/2017/DataFiles/BMX_J.htm

Missing-data strategy: the raw file mixes three kinds of columns and only one is
a measurement:
  - BMX*  the actual measurement (weight, waist, ...)          -> features
  - BMI*  a "could-not-obtain" comment code; NaN means the
          measurement WAS taken fine                           -> not features
  - BMD*  status / category flags                              -> not features

So we (1) restrict to adults, which removes the infant-only measures
(BMXRECUM / BMXHEAD are 100% missing here) and age-structured missingness,
(2) keep only the seven whole-population raw measurements and drop the derived
BMI (BMI = weight / (height/100)^2 is an exact function of two features it would
sit next to -> redundant/collinear), and (3) take complete cases. This retains
~92% of adults (5096 rows) with zero missing values, on par with the other
bundled example datasets.

Note: complete-case deletion drops ~8% of adults who skew slightly older and
heavier (waist/hip/leg are harder to measure on older, higher-BMI participants) —
a mild, expected MAR bias; sex balance is preserved.
"""
import numpy as np
import pandas as pd
from pandas_nhanes import get_dataset

# Load body-measures and demographics, merge on participant ID (SEQN)
bmx = get_dataset("BMX_J")
demo = get_dataset("DEMO_J")
df = pd.merge(demo[["SEQN", "RIDAGEYR", "RIAGENDR"]], bmx, on="SEQN")

# 1) Population: adults 18+
adults = df[df["RIDAGEYR"] >= 18].copy()

# 2) Feature set: seven raw anthropometrics collected for all adults.
#    (BMXBMI is intentionally excluded — it is derived from weight and height.)
FEATURES = {
    "BMXWT": "Weight (kg)",
    "BMXHT": "Height (cm)",
    "BMXLEG": "Upper Leg Length (cm)",
    "BMXARML": "Upper Arm Length (cm)",
    "BMXARMC": "Arm Circumference (cm)",
    "BMXWAIST": "Waist Circumference (cm)",
    "BMXHIP": "Hip Circumference (cm)",
}

# 3) Complete-case on the feature set (listwise deletion). No single column is a
#    missingness bottleneck, so we keep all seven and drop the ~8% incomplete rows.
clean = adults.dropna(subset=list(FEATURES)).copy()

# Targets (GoPCA "#target" columns -> used for coloring, not as PCA inputs)
clean["Gender#target"] = clean["RIAGENDR"].map({1: "Male", 2: "Female"})
clean["Age#target"] = clean["RIDAGEYR"].astype(int)
clean["BMI#target"] = (clean["BMXWT"] / (clean["BMXHT"] / 100) ** 2).round(1)
clean["BMI_class#target"] = pd.cut(
    clean["BMI#target"],
    bins=[0, 18.5, 25, 30, np.inf],
    labels=["Underweight", "Normal", "Overweight", "Obese"],
    right=False,
)

# Assemble output: Sample_ID + readable features + targets
out = clean.rename(columns=FEATURES)
cols = list(FEATURES.values()) + [
    "Gender#target", "Age#target", "BMI#target", "BMI_class#target",
]
out = out[["SEQN"] + cols].rename(columns={"SEQN": "Sample_ID"})
out = out.set_index("Sample_ID")
out.index = out.index.astype(int)  # SEQN is promoted to float by the merge

out.to_csv("body_measures.csv", index=True)

# Summary
n_adults = len(adults)
n_features = len(FEATURES)
print(f"adults 18+:           {n_adults}")
print(f"complete-case rows:   {len(out)}  ({100 * len(out) / n_adults:.1f}% retained)")
print(f"missing cells:        {int(out[list(FEATURES.values())].isna().sum().sum())}")
print(f"features:             {n_features} -> {list(FEATURES.values())}")
print("wrote body_measures.csv")
