# The Body Measures Dataset

Anthropometric measurements of US adults from the **National Health and Nutrition
Examination Survey (NHANES) 2017–2018**, Body Measures examination file (`BMX_J`),
merged with the demographics file (`DEMO_J`). NHANES is run by the US National
Center for Health Statistics (NCHS/CDC).

Unlike the curated benchmark datasets (Iris, Wine), this is a **real population
sample** — 5096 adults with seven standardized body measurements. Because all
body measurements grow together with overall body size, it is an ideal dataset
for seeing how PCA separates an interpretable **size** factor from a **shape**
factor.

## Source

* Body Measures (BMX_J): https://wwwn.cdc.gov/Nchs/Data/Nhanes/Public/2017/DataFiles/BMX_J.htm
* Demographics (DEMO_J): https://wwwn.cdc.gov/Nchs/Data/Nhanes/Public/2017/DataFiles/DEMO_J.htm
* **License:** Public domain (US federal data; not subject to copyright).

See `testdata/nhanes/README.md` in the repository for how the CSV is prepared,
including the missing-data handling.

## Samples

* 5096 adults (age 18+), one row per participant (`Sample_ID` = NHANES `SEQN`).

## Features (PCA inputs)

Seven raw anthropometric measurements: weight (kg); height, upper leg length,
upper arm length, arm circumference, waist circumference, hip circumference (cm).

## Targets (for coloring, not PCA inputs)

`Gender#target` (Male/Female), `Age#target` (years), `BMI#target` (kg/m²),
`BMI_class#target` (Underweight/Normal/Overweight/Obese).
