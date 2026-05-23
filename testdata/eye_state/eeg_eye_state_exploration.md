# Exploring Structure in Data: EEG Eye State Dataset and Temporal PCA

## Background

Electroencephalography (EEG) measures electrical activity at the scalp via small electrodes. Each electrode records a voltage signal reflecting the activity of neurons underneath, sampled many times per second.

This dataset contains a 117-second recording from **one subject** using the **Emotiv EEG Neuroheadset**, with 14 electrodes placed according to the international 10–20 system:

* Frontal: `AF3`, `F7`, `F3`, `F4`, `F8`, `AF4`
* Central/temporal: `FC5`, `T7`, `T8`, `FC6`
* Parietal/occipital: `P7`, `O1`, `O2`, `P8`

The figure below shows the scalp positions of all 14 electrodes, viewed from above (nose pointing up). The occipital electrodes `O1` and `O2` sit at the bottom — the region most sensitive to visual and alpha-band activity.

![EEG electrode positions](./eeg_electrode_map.png)

The sampling rate is **128 Hz** — one measurement every 7.8 ms, giving approximately **14,980 rows** in total. During the recording the subject alternately opened and closed their eyes. Eye state was determined from video and added as a label:

* `eye_state = open` — eyes open
* `eye_state = closed` — eyes closed

> **Note on data quality**: the dataset contains four isolated time points with extreme values — at approximately t = 7 s, 81 s, 90 s, and 103 s — where one or more channels reach values 70–150× the normal signal range (almost certainly brief electrode artifacts). These will appear as isolated extreme points far from the main cluster in the scores plot. Keep them in mind when interpreting unexpected structure.

The original research motivation was classification: can EEG alone predict eye state? Here we use PCA — an *unsupervised* method that ignores the labels entirely — to ask: what structure does PCA find on its own, and does it relate to eye state?

---

## This dataset versus the others

In the previous tutorial, the Swiss Roll showed us that the *shape* of the data matters. Here, the challenge is different: the structure is about **time**.

| | Swiss Roll | EEG Eye State |
|---|---|---|
| **What each row is** | One independent data point | One snapshot of an ongoing signal |
| **What PCA sees first** | Geometric shape of the data cloud | Spatial correlations between 14 channels |
| **What PCA misses** | Curved manifold structure | Temporal dynamics and oscillations |
| **The solution** | Kernel PCA | Temporal PCA — give PCA a memory by embedding time |
| **Why standard PCA fails** | Data lies on a curved surface | Shuffling the rows gives *identical* PCA results |

> **The key analogy**: Kernel PCA transforms *space*. Temporal PCA transforms *time*. Both use the same SVD algorithm on a larger, restructured matrix.

Standard PCA treats each row as an independent observation and completely ignores row order. A shuffled EEG dataset gives exactly the same PCA result. This tutorial has two parts: first run standard PCA to see what it finds (and what it misses), then switch to Temporal PCA to reveal the temporal structure.

---

## Step 1: Load the dataset

Click the **EEG Eye State** sample dataset button to load the data.

The `time` column is used as row identifiers and excluded from PCA. The 14 EEG channel columns are the input variables. The `eye_state` column is automatically recognised as categorical and available for plot colouring.

In the PCA configuration panel, set:

* **PCA Method** → SVD
* **Preprocessing** → leave at the default (mean centering only) for now

#### Questions:

* How many rows and columns does the dataset have?
* At 128 Hz, two consecutive rows are only 7.8 ms apart. Do you expect neighbouring rows to be similar or very different?

---

## Step 2: Run standard PCA — scale problem and outliers

Click **Go PCA**.

Before opening any plots, read the GoPCA warning banner:

> *"Variables have very different scales (40000× difference). Consider standardisation unless this is intentional."*

All 14 channels are EEG voltages in microvolts — they should be comparable. The large scale difference comes from brief electrode artifacts and slightly different impedance conditions between channels.

Open the **Loadings Plot (PC1)**.

#### Questions:

* Which channels dominate PC1? Do they look like a genuine signal pattern, or could a few high-variance channels simply be drowning out the rest?

👉 Without scaling, PCA measures *covariance* — channels with higher variance automatically dominate regardless of whether that variance is signal or noise.

### Re-run with Standard Scaling

Change **Preprocessing** to **Standard Scaling**. Click **Go PCA** again and re-open the **Loadings Plot (PC1)**.

#### Questions:

* How does the loading distribution change? Do more channels now contribute meaningfully?
* Do occipital electrodes (`O1`, `O2`) stand out more?

👉 Standard scaling makes PCA measure *correlation* — all channels are treated equally regardless of raw amplitude. For EEG with known scale imbalances this is generally the more informative starting point. The GoPCA warning is your cue to make this choice deliberately.

Now open the **Scores Plot (PC1 vs PC2)** and colour by `eye_state`.

#### Questions:

* Do you see most samples clustered near the origin with only a handful of extreme outliers far away?
* Hover over the extreme outliers — what time labels do they carry? Do they match the artifact times mentioned above (~7 s, 81 s, 90 s, 103 s)?

👉 This is **outlier domination**: standard scaling equalises column variances but does not protect against extreme rows. The four artifact rows have values 70–150× the normal range, so PCA points both PC1 and PC2 towards them — collapsing the remaining ~14,976 normal time points into a tiny cluster near the origin.

| Problem | What it is | Fix |
|---|---|---|
| **Scale imbalance** | Different *variables* have very different variance | Standard scaling (column-wise) |
| **Outlier domination** | A few extreme *observations* dominate the components | Identify with Diagnostic Plot, remove with lasso |

### Identify outliers with the Diagnostic Plot

Open the **Diagnostic Plot**. Two axes:

* **Horizontal — Hotelling's T²**: how far an observation is from the centroid within the PCA model space
* **Vertical — Q-statistic**: how well the model fits the observation (large Q = large residuals)

| Region | T² | Q | Interpretation |
|---|---|---|---|
| Bottom-left | Low | Low | Regular observations |
| Top-left | Low | High | Orthogonal outliers — structure the model cannot represent |
| Bottom-right | High | Low | Good leverage — extreme but on-model |
| Top-right | High | High | Bad outliers — extreme and poorly fitted |

The artifact points should appear in the top-right or far right. The normal ~14,976 time points should cluster in the bottom-left inside both dashed limits.

### Remove outliers with the lasso tool

1. Click the **lasso icon** in the plot toolbar
2. Draw a selection around the extreme outlier points
3. GoPCA will offer to **exclude** them
4. Click **Go PCA** again

Repeat until all artifact time points are excluded.

#### Questions (after removing outliers):

* How does the Scores Plot change?
* Do `open` and `closed` samples show any visible separation now?
* Is the separation clear, or do the two states still overlap considerably?

👉 Some separation is expected, but with considerable overlap. Standard PCA works only with spatial correlations between channels at each instant — not with temporal dynamics. The main lesson: even clean, well-scaled data may not yield clean class separation when the relevant structure is temporal rather than spatial.

---

## Step 3: What standard PCA cannot see

Standard PCA computes correlations between channels across all rows simultaneously. What it cannot detect:

* **Oscillations**: a 10 Hz alpha wave completes one full cycle in 100 ms (~13 consecutive rows). Standard PCA has no way to detect this periodic structure.
* **Delayed relationships**: one channel's activity may predict another's activity a few time steps later. Standard PCA ignores row order entirely.
* **Waveform shape**: EEG events span many consecutive time points. Standard PCA sees only one snapshot at a time.

#### Question:

* If you shuffled the 14,980 rows in random order before running PCA, would the result change? What does this tell you about what standard PCA can and cannot detect?

---

## Step 4: The trajectory matrix

To make PCA sensitive to time, we transform the data using a **sliding window** — the central idea of **Singular Spectrum Analysis (SSA)** and its multivariate extension MSSA.

Instead of describing each time point as one vector of 14 channel values, we describe it as a short sequence of *L* consecutive time points — a **window**. Each window becomes one row in a **trajectory matrix**:

* Rows: approximately *T* − *L* + 1 (one per window position)
* Columns: 14 channels × *L* lags = 14*L* columns

With *L* = 32 and *T* = 14,980: approximately 14,949 rows and 448 columns. SVD on this larger matrix finds **spatiotemporal patterns** — capturing which channels co-vary and how that co-variation evolves across the window.

> GoPCA implements the first two SSA steps — **embedding** (build the trajectory matrix) and **decomposition** (SVD). The full SSA algorithm also includes grouping and reconstruction, which transform selected components back into the time domain.

### Choosing the window length *L*

The window should cover 2–3 full periods of the oscillation you want to detect. For this dataset, eye closure triggers **alpha activity at ~8–12 Hz** (period 83–125 ms). At 128 Hz, one alpha period is about 11–16 samples:

| Lags | Duration | Notes |
|-----:|--------:|-------|
|    8 |    63 ms | Too short — less than one alpha cycle |
|   16 |   125 ms | Borderline — one short alpha cycle |
|   32 |   250 ms | **Recommended** — 2–3 alpha cycles, readable loadings |
|   64 |   500 ms | Fine, but loadings plot becomes crowded |

> **General rule**: convert the period of the oscillation you want to detect into samples (period in seconds × sampling rate). Set *L* to 2–4 times that value.

> Increasing the lag gives PCA memory. But too much memory can make the model harder to interpret.

---

## Step 5: Switch to Temporal PCA

Change **PCA Method** to **Temporal PCA**. Set **Number of Time Lags** to **32**. Keep **Preprocessing** at **Standard Scaling** — the same scale imbalance affects Temporal PCA, since preprocessing is applied to the original time series before the trajectory matrix is built.

Click **Go PCA**.

GoPCA builds the trajectory matrix (14 × 32 = 448 columns) and applies SVD. The result has approximately 14,949 score rows — one per window position.

Open the **Scores Plot** and colour by `eye_state`.

### Reading the scores plot as a trajectory

This is the most important conceptual shift from standard PCA to Temporal PCA.

In standard PCA, each point is an **independent sample** — you look for clusters and class separation. In Temporal PCA, each point is a **moment in time**. Consecutive points are 7.8 ms apart. You are watching the brain's state move through a 2D projection of a high-dimensional state space. **The path is the story.**

**How to read a phase-space trajectory:**

1. **Tight loops** — the trajectory traces small repeated loops: these are oscillations
2. **Long sweeping arms** — the trajectory moves far from the central cluster: these are state transitions
3. **Dense regions** — where the trajectory spends most time: stable states

**What to expect in this scores plot:**

* A **dense central cluster** — mostly closed-eye periods (the resting state)
* **Long sweeping arms** extending to very negative PC1 — eyes-open periods. When the eyes open, alpha-band activity drops sharply across the scalp (**alpha suppression**), which appears as a large negative PC1 score
* The arm shape shows the brain taking several hundred milliseconds to transition fully into and out of the eyes-open state

#### Questions:

* Can you identify the dense central cluster (eyes closed) and the sweeping arms (eyes open)?
* Inside the dense cluster, can you find small repeated loops? Those are oscillations — the brain's idle rhythms while the eyes are closed.
* Which eye state produces more negative PC1 scores?

> **Note on available plots**: the **Loadings Plot**, **Biplot**, **Circle of Correlations**, and **Diagnostic Plot** are not available for Temporal PCA — they require loadings in the original variable space, which Temporal PCA does not produce directly. The dedicated plots are **Temporal Loadings** and **Variable Importance**.

---

## Step 6: The Temporal Loadings Plot

Open the **Temporal Loadings** plot with **5 components** first, then increase to **15–20**.

Each curve corresponds to one principal component. The horizontal axis is lag (0 to *L*−1); the vertical axis shows the loading of the most influential channel for that component across the window.

> Each curve is one line per *component* — not one line per EEG channel. For each component, GoPCA finds the single channel with the highest overall loading and plots its signed values across all lags. This preserves sign information, which is essential for reading temporal structure.

**Three patterns to look for:**

| Curve shape | Interpretation |
|---|---|
| **Nearly flat** | Global mean-shift — equal loading at all lags |
| **Monotone ramp or S-shape** | Slow trend or step-response dynamics |
| **Sinusoidal (multiple zero-crossings)** | Oscillatory rhythm |

**With only 5 components, do not expect oscillations.** The top components are dominated by slow eye-state modulation, which generates far more variance than fast rhythms:

* **PC1** (~53% variance): nearly flat — a global mean-shift component
* **PC2–PC4**: gently sloped or arched — slow trend components
* **PC5**: may show an S-shape — slow modulation, not yet a true oscillation

> **Why does slow structure dominate?** The eye-state shift lasts several seconds and affects all 14 channels simultaneously — this generates large variance. Alpha oscillations at 10 Hz are faster, more localised, and lower in variance. They are present, but outranked.

Increase **Components** to **15 or 20** and click **Go PCA** again.

#### Frequency from zero-crossings

Count the number of times a curve crosses zero. Each pair of zero-crossings = one complete cycle:

* 2 zero-crossings → 1 cycle → **4 Hz** (theta)
* 4 zero-crossings → 2 cycles → **8 Hz** (alpha lower bound)
* 5–6 zero-crossings → ~2.5 cycles → **~10 Hz** (alpha)
* 10 zero-crossings → 5 cycles → **~20 Hz** (beta)

Curves with many zero-crossings look angular rather than smooth — this is normal at high frequencies where only a few samples exist per cycle.

> **Practical tip**: when a 20-component plot looks overwhelming, use the Plotly legend. **Click a component name to hide it; double-click to show only that one.** Add others back one by one.

#### Questions:

* Does PC1 appear nearly flat?
* Looking at components 11–15, can you identify any that cross zero multiple times?
* Count the zero-crossings on the most rapidly oscillating curve — what frequency does this imply?

---

## Step 7: Paired components and the Scree Plot

Open the **Scree Plot** alongside the Temporal Loadings plot.

A fundamental property of SSA is that **oscillatory signals produce pairs of components** (Vautard & Ghil, 1989). SSA extracts two components per oscillation: one resembling a sine wave and one a cosine wave, offset by exactly one quarter of the period (90°).

**Two ways to identify a pair:**

1. **Scree Plot**: two adjacent bars of nearly equal height (e.g. both ~0.8%)
2. **Temporal Loadings**: two curves at the same frequency, phase-shifted by ~90°

Neither component alone gives the full picture. Together they encode one complete oscillation.

**What to look for with 20 components:**

* **PC1** (53%): flat — global mean
* **PC2–PC10** (declining): slow trend and modulation components
* **PC11–PC14** (~0.8% each, nearly equal): the first oscillatory pairs — look for two adjacent components where one curve is a sine and the other is a cosine at the same frequency
* **PC15–PC20** (~0.4–0.5%): higher-frequency pairs or noise

**Definitive test**: always check the shape of the temporal loading curves. Equal variance alone is not sufficient — two unrelated components can share variance by coincidence.

#### Questions:

* Do you see adjacent pairs of components with nearly equal explained variance in the Scree Plot?
* Identify the first oscillatory pair: which component numbers, what % variance, and what frequency (from zero-crossings)?
* Are the two curves approximately 90° phase-shifted — does one peak where the other crosses zero?

👉 A 10 Hz alpha wave has a period of ~12.8 samples at 128 Hz. Over 32 lags you see ~2.5 complete cycles — enough for the sinusoidal pattern to be clearly visible.

---

## Step 8: Variable Importance

Open the **Variable Importance** plot.

This heatmap shows the RMS loading of each EEG channel aggregated across all lags, for each component. It answers the question the Temporal Loadings plot cannot: *which channels* drive each component.

> **Temporal Loadings** tells you the *temporal shape* of each component. **Variable Importance** tells you *where on the scalp* it originates. Together they give the full spatiotemporal picture.

#### Questions:

* Which channels contribute most strongly to PC1?
* For the oscillatory pair from Step 7: do the two paired components show the same spatial pattern of channel importance? (They should — they represent the same oscillation, just phase-shifted in time.)
* Are occipital channels (`O1`, `O2`) prominent in the alpha-band components? Does that make physiological sense?

---

## Step 9: Experiment with window length

Change **Number of Time Lags** and observe how the results shift. Try these values:

* **8 lags** → 63 ms — less than one alpha cycle
* **16 lags** → 125 ms — about one alpha cycle
* **32 lags** → 250 ms — recommended: 2–3 alpha cycles
* **64 lags** → 500 ms — more temporal context, wider trajectory matrix

After each change, click **Go PCA** and compare the **Scores Plot**, **Temporal Loadings**, and **Scree Plot**.

#### Questions:

* Does the eye-state separation in the scores plot change as *L* increases?
* Do the temporal loading curves become smoother and more oscillatory with longer windows?
* At 16 lags (125 ms), is one full alpha cycle visible — or is the window too short?
* At what lag does interpretation start to become difficult?

👉 A short window cannot see a full oscillation cycle — components reflect adjacent-sample correlations rather than meaningful rhythms. A very long window adds temporal context but makes the trajectory matrix wider and components harder to interpret.

---

# What you should take away

After this exploration, you should be able to:

* Explain why EEG is a multivariate time series and why rows are not independent samples
* Diagnose a scale problem from the GoPCA warning and fix it with Standard Scaling
* Identify and remove extreme outliers using the Diagnostic Plot and lasso tool
* Explain the SSA embedding step: sliding windows, trajectory matrix, window length *L*
* Interpret the Temporal Loadings plot: one curve per component showing the dominant channel's signed temporal eigenvector — not one curve per channel
* Recognise the three curve types: flat (global), monotone (slow trend), sinusoidal (oscillation)
* Estimate oscillation frequency from the number of zero-crossings
* Identify **paired oscillatory components** from the Scree Plot (equal variance) and Temporal Loadings (90° phase shift)
* Use **Variable Importance** to identify which channels drive each component, and verify that paired components share the same spatial pattern

---

## Final Reflection

> Standard PCA treats the EEG table as a collection of independent snapshots. Temporal PCA, by embedding the data into sliding windows, gives PCA access to *sequences* — and the resulting components represent oscillations and temporal dynamics rather than just spatial correlations. The same SVD algorithm is used in both cases; the embedding step is what makes the difference.
>
> In SSA, oscillatory signals leave a characteristic fingerprint: a pair of components with equal singular values, 90°-phase-shifted temporal eigenvectors, and the same spatial pattern of channel importance. Learning to recognise this fingerprint is the core skill of temporal dimensionality reduction.

#### Questions:

* The SSA algorithm has four steps: embedding, decomposition, grouping, and reconstruction. GoPCA implements the first two. What would you gain from grouping and reconstruction — and what tasks would those enable?
* Could the Temporal PCA scores be used as input features to a classifier predicting `eye_state`? What might be the advantage compared to using the raw EEG values?

---

## References

Broomhead, D. S., & King, G. P. (1986). Extracting qualitative dynamics from experimental data. *Physica D: Nonlinear Phenomena*, 20(2–3), 217–236. https://doi.org/10.1016/0167-2789(86)90031-X

Vautard, R., & Ghil, M. (1989). Singular spectrum analysis in nonlinear dynamics, with applications to paleoclimatic time series. *Physica D: Nonlinear Phenomena*, 35(3), 395–424. https://doi.org/10.1016/0167-2789(89)90077-8

Ghil, M., Allen, M. R., Dettinger, M. D., Ide, K., Kondrashov, D., Mann, M. E., Robertson, A. W., Saunders, A., Tian, Y., Varadi, F., & Yiou, P. (2002). Advanced spectral methods for climatic time series. *Reviews of Geophysics*, 40(1), 1003. https://doi.org/10.1029/2000RG000092

Golyandina, N., Korobeynikov, A., Shlemov, A., & Usevich, K. (2015). Multivariate and 2D extensions of singular spectrum analysis with the Rssa package. *Journal of Statistical Software*, 67(2), 1–78. https://doi.org/10.18637/jss.v067.i02

Golyandina, N. (2020). Particularities and commonalities of singular spectrum analysis as a method of time series analysis and signal processing. *WIREs Computational Statistics*, 12(4), e1487. https://doi.org/10.1002/wics.1487

Roesler, O., & Suendermann, D. (2013). A first step towards eye state prediction using EEG. In *Proceedings of the AIHLS 2013*. UCI Machine Learning Repository. https://archive.ics.uci.edu/dataset/264/eeg+eye+state
