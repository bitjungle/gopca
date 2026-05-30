# Data Preparation with GoCSV Desktop

## Overview

GoCSV Desktop prepares your data for analysis. It handles the tasks that come before PCA: loading different file formats, cleaning missing values, removing irrelevant columns, detecting outliers, and transferring clean data to GoPCA Desktop. GoPCA handles PCA-specific preprocessing (centering, scaling) — GoCSV focuses on everything before that.

---

## Supported File Formats

GoCSV can open and save the following formats:

| Format | Extension | Notes |
|--------|-----------|-------|
| CSV | `.csv` | Comma-separated; auto-detects delimiter and decimal separator |
| TSV | `.tsv` | Tab-separated |
| Excel | `.xlsx`, `.xls` | First sheet loaded; multi-sheet files use the first sheet |
| Parquet | `.parquet` | Columnar format used by Kaggle, Hugging Face, OWID, and similar data sources |

**Parquet import details:**
- A `Sample_ID` column (1, 2, 3 …) is added automatically as a unique row identifier, since Parquet files have no built-in row index
- String columns (e.g. country, category, label) are imported as `column_name#target` — making them available as group variables in GoPCA's scores plot
- Numeric columns (float, integer) import directly into the data grid
- Null values become empty cells

**Export formats:** CSV, Excel (.xlsx), TSV

---

## 1. Data Structure

GoCSV expects data in standard matrix form:
- **Rows** = samples or observations
- **Columns** = variables or measurements

**Row identifiers:** If the first column contains non-numeric labels (sample names, IDs), GoCSV detects it automatically and uses it as the row name — it is shown in the grid but excluded from numerical analysis. For Parquet files, the auto-generated `Sample_ID` column serves this purpose.

**Column types** are detected automatically:
- **Numeric** — values that parse as numbers (used in PCA)
- **Categorical** — repeated string values; available as a group variable in GoPCA
- **Target (`#target`)** — a column marked for use as a class label or group variable in GoPCA

You can toggle any column between numeric and target using the column header menu.

---

## 2. Missing Data

**Detection:** The Data Quality Dashboard shows missing percentages per column and highlights empty cells in the grid.

**Strategies:**

| Strategy | When to use |
|----------|-------------|
| Row deletion | Missing completely at random; data is plentiful |
| Column deletion | Variable has high missing percentage or is not essential |
| Mean / median imputation | Random missingness; preserves sample count |
| Forward / backward fill | Time-ordered or sequential data |
| Custom value | You know what the missing value represents (e.g. zero, detection limit) |

> **Note:** GoPCA's NIPALS algorithm can handle moderate missing data without imputation. For SVD or Kernel PCA, all values must be present.

---

## 3. Data Quality Dashboard

Run this first — it gives you an overview before making any changes.

**Dataset level:**
- Dimensions (rows × columns)
- Overall missing percentage
- Duplicate row count
- Numeric vs. categorical column counts

**Per-column:**
- Mean, median, std dev, quartiles
- Missing percentage
- Outlier count (IQR and z-score methods)
- Quality score (0–100) based on completeness and consistency

---

## 4. Transformations

Apply transformations in GoCSV when the raw data distribution needs correction before PCA. Do **not** apply mean centering here — let GoPCA handle that.

| Transformation | Use case |
|----------------|----------|
| Log | Right-skewed data (concentrations, counts, income) |
| Square root | Count data or moderate skew |
| Square | Left-skewed data |
| Standardization (z-score) | General scaling; note that GoPCA can do this too |
| Min-max scaling | Scale to [0, 1] or a custom range |
| Binning | Discretize continuous variables into categories |
| One-hot encoding | Expand a categorical column into binary columns |

---

## 5. Column Management

- **Delete columns** — remove irrelevant variables (IDs, timestamps, metadata)
- **Insert columns** — add derived variables
- **Rename columns** — use consistent, descriptive names
- **Toggle target** — mark/unmark a column as `#target` for use as a group label in GoPCA
- **Reorder columns** — drag to rearrange

**Variables to consider removing:**
- Zero or near-zero variance (flagged by the Data Quality Dashboard)
- Near-duplicate columns (high correlation with another variable)
- Administrative fields (record numbers, timestamps, operator codes)

---

## 6. Outlier Detection and Treatment

GoCSV flags outliers using two methods:
- **IQR method** — beyond 1.5 × IQR from the quartiles (robust, distribution-free)
- **Z-score method** — beyond ±3 standard deviations (assumes approximate normality)

**Handling options:**
- **Correct** — if you can verify the true value
- **Delete row** — if confirmed as an error
- **Transform** — log or root transform reduces outlier leverage
- **Keep** — valid extreme values should not be removed automatically

> Investigate before deleting. Outliers are sometimes the most scientifically interesting observations.

---

## 7. Export and Transfer to GoPCA

**Direct transfer:**
Click **Open in GoPCA Desktop** to validate and pass your data directly to GoPCA without an intermediate file.

**Manual export:**
- CSV — most compatible format; preserves `#target` column markers
- Excel (.xlsx) — for sharing or further editing in spreadsheet tools
- TSV — tab-delimited for other tools

**Pre-export checklist:**
- [ ] Rows = samples, columns = variables
- [ ] Missing values addressed
- [ ] Irrelevant columns removed
- [ ] Target / group columns marked with `#target`
- [ ] No duplicate column names

---

## Division of Responsibility

| Task | Where to do it |
|------|---------------|
| Load files (CSV, Excel, Parquet) | GoCSV |
| Handle missing values | GoCSV |
| Remove irrelevant columns | GoCSV |
| Apply log / root transforms | GoCSV |
| Mark group variables (#target) | GoCSV |
| Mean centering | GoPCA |
| Scaling (autoscaling, Pareto, etc.) | GoPCA |
| PCA computation and visualization | GoPCA |
