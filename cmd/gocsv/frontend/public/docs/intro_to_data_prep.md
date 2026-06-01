# Data Preparation with GoCSV Desktop

## Overview

GoCSV Desktop is the data intake and preparation layer for GoPCA Desktop. Its job is to accept your data in whatever format it arrives — local files, web downloads, spreadsheets, columnar databases — shape it into a clean, well-structured dataset, and hand it off to GoPCA for analysis.

The division of labour is deliberate:

- **GoCSV** handles everything before the analysis: loading formats, cleaning missing values, removing irrelevant columns, detecting outliers, and exporting in a form GoPCA can use.
- **GoPCA** handles the analysis: mean centering, scaling, PCA computation, and visualisation.

Keeping these concerns separate means you can prepare your data once and re-run the analysis with different settings without touching the preparation steps.

---

## Supported File Formats

### Opening files from disk

| Format | Extension | Notes |
|--------|-----------|-------|
| CSV | `.csv` | Comma-separated; auto-detects delimiter and decimal separator |
| TSV | `.tsv` | Tab-separated |
| Excel | `.xlsx`, `.xls` | First sheet loaded; multi-sheet files use the first sheet |
| Parquet | `.parquet` | Columnar format from Kaggle, Hugging Face, OWID, and similar sources |
| ZIP | `.zip` | Archive containing one or more data files (see below) |

**Parquet import:** A `Sample_ID` column (1, 2, 3 …) is added automatically as a unique row identifier. String columns (e.g. country, category, label) are imported as `column_name#target`, making them available as group variables in GoPCA's scores plot.

**ZIP import:** GoCSV lists the data files inside the archive. If there is only one recognisable data file it is imported automatically; if there are several, a file picker lets you choose. Safety checks guard against zip bombs and path traversal attacks.

### Loading data from a URL

Click **Load from URL** to import a file directly from a public web address without downloading it manually first.

1. Paste the direct download URL into the input field and click **Check URL**.
2. GoCSV makes a lightweight head request to verify the file is accessible and identify its format and size — no data is downloaded yet.
3. Review the result, then click **Download and Import** to fetch and load the file.

**Supported URL sources:** any public server that returns a direct file download — data repository pages, GitHub raw file links, OWID catalog URLs, and similar. GitHub *blob* links (the viewable file page) are automatically rewritten to the equivalent raw content URL.

**Not supported:** URLs requiring authentication, login redirects, or JavaScript-driven download flows (e.g. Google Drive share links, Dropbox).

### Saving / exporting

| Format | Notes |
|--------|-------|
| CSV | Recommended for use with GoPCA; preserves all column type markers |
| Excel (.xlsx) | For sharing or further editing in spreadsheet tools |
| TSV | For tab-delimited workflows |

---

## The GoPCA-ready CSV format

When GoCSV exports to CSV for GoPCA, it uses a set of naming conventions that GoPCA recognises automatically:

- **Row identifiers** — if the first column contains non-numeric labels (sample names, IDs), GoPCA uses it as row labels in plots and excludes it from numerical analysis.
- **Numeric columns** — standard values; used directly in PCA.
- **Target / group columns** — column names ending in `#target` (e.g. `species#target`) are treated as group labels for colouring the scores plot. You can mark or unmark any column as a target using the column header menu in GoCSV.

This is plain CSV — no proprietary format, no binary encoding. Any tool that reads CSV can open it.

---

## 1. Data Structure

GoCSV expects data in standard matrix form:
- **Rows** = samples or observations
- **Columns** = variables or measurements

**Column types** detected automatically:
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

## 7. Transfer to GoPCA

**Direct transfer:**
Click **Open in GoPCA Desktop** to validate and pass your data directly to GoPCA without saving an intermediate file.

**Save then open:**
Export as CSV, then open the file in GoPCA Desktop. This gives you a saved copy of the prepared dataset that you can reload later or share with colleagues.

**Pre-transfer checklist:**
- [ ] Rows = samples, columns = variables
- [ ] Missing values addressed
- [ ] Irrelevant columns removed
- [ ] Group / class columns marked with `#target`
- [ ] No duplicate column names

---

## Division of Responsibility

| Task | Tool |
|------|------|
| Load local files (CSV, Excel, Parquet, ZIP) | GoCSV |
| Download and import from a URL | GoCSV |
| Handle missing values | GoCSV |
| Remove irrelevant columns | GoCSV |
| Apply log / root / scaling transforms | GoCSV |
| Mark group variables (`#target`) | GoCSV |
| Mean centering | GoPCA |
| Scaling (autoscaling, Pareto, robust, etc.) | GoPCA |
| PCA computation and visualisation | GoPCA |
