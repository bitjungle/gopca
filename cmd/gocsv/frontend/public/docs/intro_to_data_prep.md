# Data Preparation with GoCSV Desktop

## Overview

Real data rarely arrives ready for analysis. It comes with a title block above the table, samples running across the top instead of down the side, a sample ID that hides the batch number inside it, blank cells, and a column that turns out to hold the same value in every row.

GoCSV Desktop is where you sort all of that out. It handles everything that comes *before* PCA — opening awkward files, fixing the shape of the table, filling or removing gaps, reshaping variables — and then hands clean data to GoPCA Desktop.

One division is worth fixing in your mind from the start:

> **GoCSV prepares the data. GoPCA preprocesses it.**
>
> Centering and scaling belong to the analysis, not to the file, because they depend on which samples you are analysing. GoPCA applies them at analysis time and can undo them. Do not apply them here.

| Task | Where |
|------|-------|
| Open CSV, Excel, TSV, Parquet | GoCSV |
| Fix the table's shape (transpose, row names) | GoCSV |
| Handle missing values | GoCSV |
| Choose rows and columns | GoCSV |
| Encode categories, split or combine columns | GoCSV |
| Log / square-root transforms | GoCSV |
| Mark group variables (`#target`) | GoCSV |
| **Mean centering** | **GoPCA** |
| **Scaling (autoscaling, Pareto, SNV…)** | **GoPCA** |
| PCA computation and visualisation | GoPCA |

---

## 1. Getting your data in

### Supported formats

| Format | Extension | Notes |
|--------|-----------|-------|
| CSV | `.csv` | Delimiter and decimal separator are detected automatically |
| TSV | `.tsv` | Tab-separated |
| Excel | `.xlsx`, `.xls` | The first sheet opens directly; use the Import Wizard for another sheet |
| Parquet | `.parquet` | Columnar format used by Kaggle, Hugging Face, Our World in Data and similar sources |

You can save as **CSV** or **Excel**.

**A note on Parquet.** These files have no row index, so GoCSV adds a `Sample_ID` column (1, 2, 3 …) to give every row a unique identifier. String columns arrive marked as `column_name#target`, which makes them available as group variables for colouring plots in GoPCA. Numeric columns come in directly, and nulls become empty cells.

### When a file will not open on its own

Most files open with **Choose File**. Two situations need more control, and **Import with Wizard** handles both.

**The table does not start at the first row.** Spreadsheets are often written for people rather than programs — a report title, a date, a blank row, and only then the real headers. GoCSV recognises this and offers the Import Wizard with the right number of rows already skipped. Check the preview and import.

**You want to choose what comes in.** The wizard lets you pick the sheet, say which row holds the headers, and select only the columns you need.

| Option | Applies to | What it does |
|--------|-----------|--------------|
| Sheet | Excel | Which sheet to read |
| Delimiter | CSV / TSV | Comma, semicolon, tab or pipe |
| First row contains headers | All | Uncheck for files with no header row |
| Header Row | All | Which row holds the column names (0-based) |
| Skip Rows from Top | All | Discard rows above the table — pre-filled when a title block is detected |
| Maximum Rows | All | Read only the first N rows; 0 reads them all |
| Row Names Column | All | Which column holds sample names (−1 for none) |
| Column selection | All | Tick the columns to import, in the preview step |

> **If a spreadsheet refuses to open,** the cause is almost always that the data does not begin at the first row. **Skip Rows from Top** in the Import Wizard is nearly always the answer.

### A file with no numbers in it is still a valid file

You do not need numeric columns to open a file. A table of nothing but text — sample names, sites, categories — opens perfectly well. Preparing data for PCA often *starts* from something that is not numeric yet, and the encoders in section 5 are how you make it numeric.

GoCSV will tell you the file is not ready for PCA yet, which is true and useful. It will not refuse to let you work on it.

---

## 2. Getting the shape right

PCA expects a specific arrangement, and this is worth checking before anything else:

- **Rows are samples** — the things you measured
- **Columns are variables** — the things you measured *about* them

### Your instrument probably disagrees

Spectrometers, chromatographs and sequencers commonly export the other way round: one **column** per sample, one **row** per wavelength or channel. That is the transpose of what PCA needs.

Use **Transpose** in the toolbar. It swaps rows and columns, turning your headers into row names and your row names into headers. It tells you what the result will be before it does anything, and it undoes cleanly if the answer surprises you.

Two things happen that are worth expecting:

- **Column types are recalculated.** A transposed table has entirely different columns, so a row that mixed text and numbers becomes a column that does.
- **Any `#target` marking stops applying.** A target has to be a column; after transposing it is a row.

### Row names: which column identifies your samples

The first column of your file becomes the row-name column. It is shown down the left of the grid, kept out of the numbers, and used to label points in GoPCA's plots.

This happens **whether or not the first column looks like an identifier** — if your file begins with a measurement, that measurement becomes row names. Two fixes, both on the right-click menu of any column header:

- **Use as Row Names** — promote a different column. Whatever was serving as row names comes back into the table, so nothing is lost.
- **Move Row Names into Table** — put the row names back as an ordinary column and leave the table without any.

**Row names must be unique.** GoCSV will not let you promote a column with repeated or empty values, and will tell you which value is the problem. The reason is worth knowing: row names label the points in a scores plot, so two samples sharing a name are indistinguishable exactly where you would most want to tell them apart.

If a file *arrives* with duplicate row names, GoCSV warns you rather than refusing it. The numbers are still analysable; only the labelling is ambiguous, and whether that matters is your call.

---

## 3. Looking before you leap

Open the **Data Quality Report** first, before changing anything. It is quicker to read than the grid and it will often decide what you do next.

**For the dataset:** dimensions, overall missing percentage, duplicate rows, and how many columns are numeric against categorical.

**For each column:** mean, median, standard deviation, quartiles, missing percentage, outlier counts, and a quality score.

Two things it flags are worth acting on:

**Columns with no variation at all.** Every value identical. Such a column contributes exactly nothing to any component — but it is not harmless, because it sits at the origin of every loadings plot where its position can be read as meaningful. Instrument settings, a constant temperature, a batch code that never changed: all common, all worth removing.

**Columns that barely vary.** Reported as a fraction of the column's own level, so the judgement does not depend on whether you recorded metres or kilometres. Below a tenth of a percent, standardisation will scale that column to unit variance anyway — which can turn measurement noise into an apparent component.

Neither is removed for you. Whether a quiet variable matters is a question about your experiment, not about the numbers.

---

## 4. Missing values

**Finding them:** the Data Quality Report gives percentages per column, and empty cells are highlighted in the grid.

**Filling them:** the **Fill Missing Values** dialog works one column at a time.

| Strategy | When to use it |
|----------|----------------|
| Mean | Roughly symmetric numeric data |
| Median | Skewed numeric data, or when outliers are present |
| Mode | The most frequent value — the only sensible choice for categorical columns |
| Forward fill | Time-ordered data; carry the last observation forward |
| Backward fill | Time-ordered data; carry the next observation back |
| Custom value | You know what the gap means — zero, a detection limit, "Unknown" |

Removing rows or columns is the other option, and often the better one: use **Filter Rows** (section 5) or delete the column outright when a variable is mostly empty.

> **You may not need to fill anything.** GoPCA's **NIPALS** algorithm handles moderate missing data directly, without imputation. SVD and Kernel PCA need every value present. If your gaps are few and scattered, choosing NIPALS in GoPCA is more honest than inventing values here.

---

## 5. Choosing and shaping what you analyse

### Choosing rows

**Filter Rows** keeps or removes the rows matching a condition — drop the QC standards, analyse one batch, exclude samples you have decided are unusable.

It shows how many rows match, and how many would remain, *before* you apply it. A filter that would empty the table says so.

One rule is worth knowing because it protects you: **blank cells match only "is empty"**. A negative condition will never sweep up rows for having *no* value in that column. Asking to remove rows where `Region is not Nord` removes the ones you can see are not Nord, not the ones whose region was never recorded. Deciding a sample's fate on a missing value should be something you ask for deliberately, which is what the "is empty" condition is for.

### Averaging replicates

Measuring each sample two or three times is good practice, but those repeats should usually become **one row** before analysis. **Average Replicates** does that: group by the column identifying the sample, and the numeric columns are combined — mean by default, or median, sum, or first value.

It also removes a hazard rather than working around one. Replicates left as separate rows **leak between cross-validation folds**: the same sample lands in both training and validation, and the model looks better than it is. Averaging first prevents that; the alternative is remembering to set `--cv-group` on every PCR run.

Three things it will not do quietly:

- **Missing numbers are skipped, not counted as zero.** Averaging a gap in as zero would drag the result towards zero in proportion to how much data is absent — a silent bias rather than a visible gap. A group with nothing present stays empty.
- **Where a group disagrees on a text value, the cell is cleared** and the count reported. Picking one of the competing values would assert something about the aggregated sample that no row actually said.
- **Rows with no group value stop the operation.** A blank is not a group: averaging the unlabelled rows together would invent a sample, and dropping them would lose data. Remove or label them first — Filter Rows does it in one step.

The grouping column becomes the row-name column afterwards, since the rows it identified no longer exist and the group value is what identifies the new one. Those names are unique by construction, which is exactly what row names need to be — and **Move Row Names into Table** puts them back as a column if you want them there.

> **The grouping column often has to be made first.** If your replicate structure is buried in a sample ID like `B3_S12_r1`, split it on `_` and group by the batch part.

### Choosing columns

- **Delete columns** — remove what you are not analysing: record numbers, timestamps, operator codes
- **Insert Column Before / After** — add an empty column to fill in yourself
- **Rename column** — give variables names you will recognise in a loadings plot
- **Reorder columns** — drag a column header. The move is applied to the data, so it survives export and reaches GoPCA, and Undo reverses it like any other edit
- **Mark as Target Column** — see below

Worth considering for removal: columns with no or almost no variation (section 3), and near-duplicate columns that correlate almost perfectly with another variable.

### Splitting and combining columns

Sample identifiers often carry structure. `B3_S12_r1` means batch 3, sample 12, replicate 1 — three facts stuffed into one string.

**Split Column** divides a column on a delimiter, giving one new column per part. Splitting that ID on `_` gives you the batch as a column of its own, which is exactly what PCR's grouped cross-validation needs and what you would group on to average replicates.

**Combine Columns** does the reverse, joining several into one. Columns join **in the order you tick them**, so `Site` then `Year` gives `Oslo_2024` while `Year` then `Site` gives `2024_Oslo`. The dialog shows the result as you go.

> **The ID you want to split is probably your row-name column,** and row names are not in the selection list. Right-click any header, choose **Move Row Names into Table**, and it becomes a column you can split.

### Making categories numeric

PCA works on numbers. A categorical column has to be encoded before it can take part — and *how* you encode it is a statement about your data, not a formatting choice.

**One-Hot Encode** makes no claim about order. Each category becomes its own column, and PCA treats them as equally distant from one another. This is right for unordered categories: species, site, operator, instrument.

**Ordinal Encode** replaces categories with 0, 1, 2 … in an order you set. Use it only when the categories genuinely form a scale — `lav, middels, høy`, or `never, rarely, sometimes, often, always`. The dialog lists the values with arrows to reorder them, and recognises common scales in English and Norwegian, so `lav / middels / høy` comes up already in the right order.

Both keep the original column by default. Keeping it is usually what you want, because GoPCA colours scores plots by categorical columns — encoding `species` and discarding it costs you a colouring you would probably have wanted.

**The mistake worth avoiding.** Numbering unordered categories tells PCA something untrue. Encoding `species` as setosa = 0, versicolor = 1, virginica = 2 asserts that virginica is three times setosa and that versicolor sits exactly halfway between them. None of that is true, and PCA cannot know — it is a covariance method, so it consumes those invented distances as though they were measurements. The resulting component will look perfectly ordinary. If your categories have no order, reach for one-hot encoding.

> **If you have used scikit-learn's `LabelEncoder`,** it assigns codes alphabetically. For an ordered scale that is usually wrong: `low, medium, high` becomes `high = 0, low = 1, medium = 2`, scrambling the very order the numbers are supposed to carry. Leaving GoCSV's list untouched gives you that same alphabetical result — the arrows exist so you do not have to accept it.

### Target columns

**Mark as Target Column** appends `#target` to a column name. A numeric column marked this way is held back from the PCA itself and offered instead as a reference variable — for colouring a scores plot, or as the response in a PCR model.

Categorical columns are already excluded from the PCA and already available for colouring, so marking one changes its name without changing what it does.

---

## 6. Transformations

Transform in GoCSV when the *distribution* of a variable needs correcting. Leave centering and scaling to GoPCA.

### Which one, and when

Most datasets need none of these. Reach for a transformation when you have a reason, and the reason is usually one of three:

**"One variable has a long tail."** A few large values sit far from the mean, and PCA is a covariance method — it notices distance. That variable will pull a component towards itself for reasons of shape rather than substance. Use **Box-Cox** if every value is above zero, **Yeo-Johnson** if any are zero or negative. Let them fit the exponent; that is what they are for.

**"My columns are parts of a whole."** Percentages, assays, anything summing to 100. Use **CLR**, and read the section below first — this is a correctness problem, not a tidiness one.

**"I know the mechanism."** Log for a process that is multiplicative rather than additive, square root for counts that behave like a Poisson, square for a left skew. These are claims about how your measurement works, and if you can make one, a fixed transform is more defensible than a fitted one because you can say why you chose it.

If none of those applies, **do nothing here.** Centering and scaling in GoPCA handle differences of unit and range, which is what most preparation actually needs.

| Transformation | Use case |
|----------------|----------|
| Log | Right-skewed data — concentrations, counts, incomes |
| Square root | Count data, or moderate skew |
| Square | Left-skewed data |
| Standardisation (z-score) | General scaling — though GoPCA can do this at analysis time |
| Min-max scaling | Scale to [0, 1] or a range you choose |
| Binning | Turn a continuous variable into categories |
| Box-Cox | Right-skewed data, with the strength of the correction fitted to the column |
| Yeo-Johnson | The same, but defined at zero and for negative values |
| Centred log-ratio (CLR) | Compositional data — percentages, assays, parts of a whole |

**A column is transformed completely or not at all.** `log` is undefined at zero and below, and `sqrt` at negatives. If any value in a column is outside the range, GoCSV leaves the whole column untouched and tells you which rows are the problem.

This matters more than it sounds. Transforming the valid values and skipping the rest would leave one variable holding two different scales — some cells in log units, some raw — and nothing downstream could detect it. Zeros in concentration and count data are normal, not exotic, so this is a case you are likely to meet. When you do, decide what the zeros mean before transforming: a true zero, a value below the detection limit, and a missing measurement are three different things.

### The order you do things in changes the answer

This is the part that is easy to get wrong, because every individual step looks correct.

**Filter before you transform.** Box-Cox and Yeo-Johnson fit their exponent to the values present. Remove a batch afterwards and the exponent was fitted partly to rows you have discarded — harmless in most cases, wrong if the batch you removed was the skewed one.

**Decide about replicates before you transform, not after.** Averaging and transforming do not commute, and the difference is not small. Two replicates of a concentration, 10 and 100:

| Order | Result | What it means |
|-------|--------|---------------|
| Average, then log | `log(55.0) = 4.007` | The **arithmetic** mean of the measurements |
| Log, then average | `mean(log) = 3.454` | `log(31.6)` — the **geometric** mean |

Both are defensible; they answer different questions. If your measurement error is additive, average first. If it is multiplicative — as it usually is for concentrations spanning orders of magnitude — transform first, so you are averaging on the scale where the error is symmetric. What you should not do is pick by accident.

**Do not stack distribution transforms.** Log followed by Box-Cox is Box-Cox applied to data that has already been corrected, and the fitted λ will reflect that. Choose one.

**Apply CLR to raw compositional data**, before anything else touches those columns. It works on the ratios between parts; transforming the parts individually first destroys the very relationship it is reading.

### Letting the data choose the transform

`log`, `square root` and `square` all ask you to guess how skewed a variable is. **Box-Cox** and **Yeo-Johnson** fit the exponent instead, by maximum likelihood, so the strength of the correction comes from the column rather than from your judgement about it.

Both belong to one family:

```
                (xᵏ − 1) / k    for k ≠ 0
Box-Cox(x) =
                ln(x)           for k = 0
```

The logarithm is not a special case bolted on — it is the k = 0 member, which the formula approaches smoothly. So "should I take logs, or a square root, or neither?" becomes one question with one answer: what value of k best symmetrises this column?

**Which of the two?** Box-Cox needs every value strictly above zero. **Yeo-Johnson** extends the same idea to zero and negative values, which is precisely where `log` and `sqrt` refuse — so it is the answer for zero-inflated concentration and count data, and the one to reach for when Box-Cox declines your column.

**The fitted λ is reported afterwards**, because a transform whose parameter you cannot see is one you cannot quote in a paper or reproduce anywhere else. You can also supply λ yourself: a validation set should be transformed the same way as its training set, not fitted to its own optimum.

**What it looks like.** A right-skewed concentration column, and what Box-Cox does to it:

| Input | 0.4 | 0.7 | 1.1 | 1.8 | 3.2 | 6.5 | 14.0 | 31.0 |
|-------|-----|-----|-----|-----|-----|-----|------|------|
| **Output** | −0.97 | −0.36 | 0.09 | 0.57 | 1.08 | 1.67 | 2.26 | 2.81 |

> `Box-Cox applied to 'Conc' (λ = −0.1214, fitted by maximum likelihood)`

Read the gaps rather than the numbers. In the input, the last two values are 17 units apart while the first two are 0.3 apart — a ratio of nearly sixty. After transforming, those gaps are 0.55 and 0.60: comparable. The largest sample no longer sits at a distance that would dominate a component on its own.

Note also that λ came out at −0.12, close to zero — which is the logarithm. For data generated by a multiplicative process the estimator tends to find its way there by itself, which is a useful check that it is doing something sensible rather than something arbitrary.

> **Not about normality.** PCA makes no assumption that your variables are normally distributed, so this is not a box to tick before analysis. The reason to reduce skew is more concrete: a variable with a long tail exerts leverage out of proportion to its information, because a handful of large values sit far from the mean and a covariance method notices distance. Reducing the skew reduces that pull. If a variable is not skewed, leave it alone — a fitted λ near 1 is the transform telling you exactly that.

**References:** Box & Cox (1964), *An Analysis of Transformations*, JRSS B 26(2). Yeo & Johnson (2000), *A New Family of Power Transformations to Improve Normality or Symmetry*, Biometrika 87(4).

### Compositional data: parts of a whole

Some measurements only make sense relative to each other. Mineral assays, food composition, soil fractions, percentage breakdowns — each row describes how a whole divides into parts, and the parts add up to 100 (or 1, or a fixed total).

**This breaks PCA in a way that is easy to miss.** If the parts must sum to a constant, then one part rising forces the others to fall, whatever the underlying chemistry. That is not a fact about your samples; it is arithmetic. The covariance matrix becomes singular, correlations between parts come out spuriously negative, and the components you get describe the constant-sum constraint as much as the material.

The fix is over a century old in outline and standard since Aitchison: stop analysing the amounts and start analysing the *ratios* between them.

**Centred Log-Ratio (CLR)** does this. Each part is replaced by the logarithm of its ratio to the geometric mean of the whole row:

```
clr(x)ᵢ = ln( xᵢ / geometric mean of the row )
```

Select **every part of the composition together** — all the oxides, all the percentages — and each becomes a new `_clr` column. Two things follow from the definition and are worth recognising:

- **The values describe ratios, not amounts.** A row of `[2, 3, 5]` and a row of `[2000, 3000, 5000]` transform identically, because they are the same composition measured in different units.
- **Each transformed row sums to zero.** That is inherent to centring on the geometric mean, not a sign of anything wrong.

**What it looks like.** Two samples, three parts, percentages:

| | A | B | C | | A_clr | B_clr | C_clr |
|---|---|---|---|---|---|---|---|
| Sample 1 | 60 | 30 | 10 | → | 0.828 | 0.135 | −0.963 |
| Sample 2 | 20 | 20 | 60 | → | −0.366 | −0.366 | 0.732 |

Sample 2 shows the reading most clearly: A and B are equal, so they take the same value, and it is *negative* — meaning both sit below the geometric mean of that row. C is above it. The numbers say "this part is large or small relative to the rest of this sample", which is the only thing a composition can honestly tell you. And each row sums to zero, as it must.

CLR keeps one output column per input column, so a loadings plot still names your variables — which is why it is the sensible choice here over ILR, whose coordinates are no longer per-variable, or ALR, which needs you to nominate one part as a denominator.

> **Zeros need a decision from you.** The logarithm is undefined at zero, and zeros are routine in trace-element work. GoCSV refuses the transform by default and tells you which rows are affected.
>
> If you want to proceed, give a **replacement value** below your detection limit. The other parts in that row are scaled down so the row total is unchanged, which keeps the ratios among the parts you actually measured intact. This is the standard remedy (Martín-Fernández et al., 2003) — but it invents a measurement that was not made, and "absent" and "below the detection limit" are different claims. That is why GoCSV makes you ask rather than doing it quietly.

You do not need a closed composition. A **subcomposition** — a subset of the parts — is still compositional and is analysed this way routinely. GoCSV tells you which case you are in: whether the columns you selected sum to a constant in every row, or vary. If they vary when you expected them not to, you have probably missed a part.

**References:** Aitchison, J. (1986), *The Statistical Analysis of Compositional Data*, Chapman & Hall, Ch. 4. Egozcue et al. (2003), *Isometric Logratio Transformations for Compositional Data Analysis*, Mathematical Geology 35(3).

---

## 7. Outliers

The Data Quality Report flags unusual values two ways:

- **IQR** — beyond 1.5 × the interquartile range from the quartiles. Robust, and assumes nothing about the distribution.
- **Z-score** — beyond ±3 standard deviations. Assumes the variable is roughly normal.

GoCSV shows you where they are; what to do about them is a judgement it cannot make for you.

- **Correct it** — if you can check the original record and the value is wrong, fix the cell
- **Remove the sample** — if it is confirmed as an error, use Filter Rows or delete the row
- **Transform** — a log or square-root transform reduces the leverage of extreme values without discarding them
- **Keep it** — a genuine extreme value is data, not noise

> Investigate before deleting. In a scores plot an outlier is often the most interesting point on the chart, and "unusual" is not the same as "wrong".

---

## 8. Handing over to GoPCA

**Direct transfer:** click **Open in GoPCA**. The data is validated and passed across without an intermediate file.

**Or export:** CSV keeps `#target` markers and is the most portable; Excel is convenient for sharing.

**Before you hand over:**

- [ ] Rows are samples, columns are variables — transpose if not
- [ ] The row-name column identifies your samples, and its values are unique
- [ ] Missing values dealt with, or NIPALS chosen in GoPCA
- [ ] Columns with no variation removed
- [ ] Categorical variables encoded, if you want them in the analysis
- [ ] Compositional data transformed with CLR, if your columns are parts of a whole
- [ ] Replicates averaged, or `--cv-group` planned for if you are heading to PCR
- [ ] Group and response variables marked with `#target`
- [ ] No duplicate column names

**Validate for GoPCA** checks most of this and explains anything it finds. A warning is not a refusal — it tells you something about your data that you may already know and have a reason for.

---

## Where to go next

- [Introduction to PCA](intro_to_pca.md) — what the analysis actually does, and how to read the plots
- [CLI reference](cli_reference.md) — the same preparation and analysis from the command line
- [Troubleshooting](troubleshooting.md) — when something does not behave as you expect
