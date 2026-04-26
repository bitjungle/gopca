# Exploring Structure in Data: EEG Eye State Dataset and Temporal PCA

## Background: Brain signals, time series, and a different kind of data

Electroencephalography (EEG) measures electrical activity at the scalp by placing small electrodes at standardised positions across the head. Each electrode records a voltage signal reflecting the summed electrical activity of neurons underneath. EEG is fast — it captures changes that occur within milliseconds — and it is widely used in neuroscience and brain–computer interface research.

This dataset contains a recording from **one subject** using the **Emotiv EEG Neuroheadset**, with 14 electrodes placed according to the international 10–20 system:

* Frontal: `AF3`, `F7`, `F3`, `F4`, `F8`, `AF4`
* Central/temporal: `FC5`, `T7`, `T8`, `FC6`
* Parietal/occipital: `P7`, `O1`, `O2`, `P8`

The recording lasted **117 seconds** at a sampling rate of **128 Hz** — giving one measurement every 7.8 ms and approximately **14,980 rows** in total. During the recording, the subject alternately opened and closed their eyes. Eye state was determined from a video camera and added manually as a label:

* `eye_state = open` — eyes open
* `eye_state = closed` — eyes closed

The original research motivation was classification: can the EEG signal alone predict whether the eyes are open or closed? Here, we approach the same data using **Principal Component Analysis (PCA)** — an *unsupervised* method that ignores the labels entirely. The question becomes: what structure does PCA find in the EEG channels on its own, and does that structure relate to eye state?

👉 This dataset is fundamentally different from Iris, Wine, Corn, and Swiss Roll:

* Each row is **one time point**, not one independent sample — neighbouring rows are consecutive observations separated by 7.8 ms
* The 14 EEG channels are measured **simultaneously** and are correlated in space and in time
* There are no predefined sample classes in the static sense — variation is both spatial (across channels) and temporal (across time)

This is a **multivariate time series**. Standard PCA can find spatial correlation patterns between channels, but it treats each time point as an independent observation and completely ignores the order of the rows. To make PCA sensitive to temporal dynamics, we need to first transform the data — and GoPCA's built-in **Temporal PCA** method does exactly this.

---

## The challenge

EEG data has two kinds of structure simultaneously:

* **Spatial structure**: the 14 channels represent different scalp locations; nearby electrodes tend to respond similarly
* **Temporal structure**: the signal evolves over time; oscillations, bursts, and waveform shapes carry the most meaningful information in EEG

Standard PCA on the raw EEG table reveals only the spatial structure — the correlation between channels. It cannot detect oscillations, repeated waveforms, or delayed relationships between channels, because it does not know which row came before another.

> Standard PCA treats the dataset as a cloud of points. Shuffling the rows would give exactly the same PCA result. It ignores time entirely.

This tutorial therefore has two parts:

* First, run standard SVD PCA on the raw EEG table and examine what it finds
* Then, use GoPCA's **Temporal PCA** method to reveal the temporal structure hidden inside the signals

---

## Step 1: Load the dataset

Click the **EEG Eye State** sample dataset button to load the data.

GoPCA will load the dataset automatically. The `time` column is used as row identifiers and is excluded from the PCA variables. The 14 EEG channel columns are the input variables. The `eye_state` column contains text labels (`open`/`closed`) and is automatically recognised as a categorical column.

In the PCA configuration panel:

* **Target Column** → `eye_state`
* **PCA Method** → SVD

#### Questions:

* How many rows and columns does the dataset have?
* What does a single row represent — a single brain measurement, or something else?
* Do you expect neighbouring rows (consecutive time points) to be similar or very different?

👉 Hint: at 128 Hz, two consecutive rows are only 7.8 ms apart. Brain signals change slowly relative to the sampling rate, so neighbouring rows are very similar.

---

## Step 2: Run standard PCA and examine the scores

Click **Go PCA** to run the analysis.

Open the **Scores Plot (PC1 vs PC2)** and colour by `eye_state`.

#### Questions:

* Do `open` and `closed` samples separate in the scores plot?
* Is the separation clear, partial, or weak?
* Do you see distinct clusters, or a continuous mixture?

👉 Standard PCA may show some separation between eye states — there is a well-known alpha-wave suppression effect when the eyes open. But the separation may be incomplete or noisy, because PCA is working only with the spatial correlations between channels at each instant, not with the temporal dynamics.

Now open the **Loadings Plot**.

#### Questions:

* Which EEG channels contribute most to PC1?
* Which channels contribute most to PC2?
* Do occipital electrodes (`O1`, `O2`) stand out?

👉 Hint: eye closure often triggers a strong increase in alpha-band power (~8–13 Hz) at occipital sites. If occipital electrodes load heavily on a component that separates the eye states, this is consistent with that physiology.

---

## Step 3: What standard PCA misses

Standard PCA computes correlations between channels across all rows simultaneously. What it cannot see:

* **Oscillations**: a 10 Hz alpha wave completes one full cycle in 100 ms — about 13 consecutive rows. Standard PCA has no way to detect this periodic structure.
* **Delayed relationships**: one channel's activity may predict another channel's activity a few time steps later. Standard PCA ignores this.
* **Waveform shape**: the characteristic shape of an EEG event (e.g., a slow wave during eye closure) spans many consecutive time points. Standard PCA sees only one snapshot at a time.

#### Questions:

* If you shuffled the 14,980 rows in a random order before running PCA, would the result change?
* What does this tell you about what standard PCA can and cannot detect in time-series data?

👉 Key insight: standard PCA is completely insensitive to the order of rows. A shuffled EEG dataset gives exactly the same PCA scores and loadings. All temporal structure is invisible to it.

---

## Step 4: Temporal structure and the trajectory matrix

To make PCA sensitive to time, we first transform the time series using a **sliding window**. This is the central idea of **Singular Spectrum Analysis (SSA)** and its multivariate extension **MSSA**.

### The embedding step

Instead of describing each time point as one vector of 14 channel values, we describe it as a short sequence of *L* consecutive time points — a **window** of length *L*.

For a single channel, a window of length *L* = 4 starting at time *t* would look like:

```
x(t), x(t+1), x(t+2), x(t+3)
```

For all 14 channels simultaneously (MSSA), the window captures all channels over *L* consecutive time steps. Each window is represented as a row in the **trajectory matrix** — a Hankel-block-Hankel structure where each channel contributes a Hankel block (Golyandina et al., 2015). The trajectory matrix has:

* One row per window position (approximately *N − L + 1* rows, where *N* is the total number of time points)
* One column per channel per lag: 14 channels × *L* lags = 14*L* columns

### The decomposition step

SVD is then applied to this trajectory matrix. The resulting components are no longer simple spatial patterns — they are **spatiotemporal patterns**, capturing which channels co-vary and *how that co-variation evolves over the window*.

### The four SSA steps

The full SSA algorithm consists of four steps (Golyandina et al., 2015):

1. **Embedding** — construct the trajectory (Hankel) matrix using sliding windows
2. **Decomposition** — apply SVD to decompose the trajectory matrix
3. **Grouping** — combine components that represent related structures (e.g., a periodic pair)
4. **Reconstruction** — transform selected components back into the time domain

GoPCA's **Temporal PCA** method implements steps 1 and 2. This is the core of SSA: once the data is embedded, standard linear PCA/SVD on the trajectory matrix reveals the temporal structure.

### Choosing the window length

The window length *L* determines what temporal scale the analysis can resolve.

For this EEG dataset (128 Hz sampling rate):

* *L* = 8 → 62.5 ms window — captures fast beta activity (~20 Hz)
* *L* = 16 → 125 ms window — roughly one alpha cycle (~10 Hz, period ≈ 100 ms)
* *L* = 32 → 250 ms window — covers theta rhythms (~4–8 Hz)
* *L* = 64 → 500 ms window — slower dynamics, slow cortical potentials

A general guideline (Golyandina, 2020): for oscillatory components, *L* should be at least as long as one full period of the oscillation of interest. For exploratory analysis, starting with *L* = 16 is reasonable for EEG at 128 Hz.

---

## Step 5: Switch to Temporal PCA in GoPCA

Change the **PCA Method** to **Temporal PCA**.

A new option appears: **Number of Time Lags**. Set it to **16**.

Keep:

* **Target Column** → `eye_state`

Click **Go PCA**.

GoPCA builds the trajectory matrix internally from the 14 EEG channels: each row of the original data is expanded into a window of 16 consecutive time steps, producing a trajectory matrix with 14 × 16 = 224 columns. SVD is then applied to this matrix. The result has approximately *N* − *L* + 1 = 14,965 score rows — one per window position.

Open the **Scores Plot** and colour by `eye_state`.

#### Questions:

* Does the separation between `open` and `closed` improve compared to standard SVD PCA?
* Do you see tighter clusters or a clearer gradient?
* Are there time regions where the label switches — can you see transitions in the scores?

👉 Temporal PCA gives each observation access to a 16-step context window (125 ms). The components can now represent oscillatory and dynamic patterns, not just instantaneous spatial correlations.

**Note on available plots**: some plots available for SVD PCA are not available for Temporal PCA — specifically the **Loadings Plot**, **Biplot**, **3D Biplot**, **Circle of Correlations**, and **Diagnostic Plot**. These require loadings in the original variable space, which Temporal PCA does not produce directly (loadings live in the higher-dimensional trajectory space). Instead, two dedicated plots are available: **Temporal Loadings** and **Variable Importance**.

---

## Step 6: Interpret the Temporal Loadings Plot

Open the **Temporal Loadings** plot.

This plot is unique to Temporal PCA. Unlike the standard Loadings Plot — which shows one point per variable — the Temporal Loadings Plot shows how each component's loading evolves **across time lags** for each channel.

For each principal component, you will see curves for each of the 14 EEG channels, plotted across lag positions 0 to *L*−1.

#### Questions:

* Does the loading curve for any channel show a smooth oscillatory shape — rising and falling as the lag increases?
* Do different channels show similar loading curves, or very different ones?
* Which channels have the largest amplitude loadings across all lags?

👉 Key insight: a sinusoidal shape in the temporal loading curve means that component captures an oscillation at a particular frequency. A smooth, flat loading means the component captures a slow drift or mean level. The Temporal Loadings plot is the SSA equivalent of a spectral decomposition — each component corresponds to a structured temporal pattern rather than a static spatial direction.

---

## Step 7: Examine the Scree Plot and Variable Importance

Open the **Scree Plot**.

#### Questions:

* How many components are needed to explain the bulk of the variance?
* Does the Temporal PCA scree plot look different from the SVD scree plot for the same data?

Now open the **Variable Importance** plot.

This plot shows the aggregated contribution of each original EEG channel across all time lags, for each temporal principal component. Contributions are computed using root mean square (RMS) aggregation across lags.

#### Questions:

* Which EEG channels contribute most strongly to the first temporal component?
* Do occipital electrodes (`O1`, `O2`) appear more or less important here than in standard PCA?
* Is the pattern of variable importance similar across components, or does each component emphasise different channels?

👉 Variable Importance removes the lag dimension and gives a compact answer to: which *channels* (regardless of timing within the window) drive each component? Compare this to the raw loadings from standard SVD PCA — you may find the ranking changes significantly.

---

## Step 8: Experiment with the Number of Time Lags

Change the **Number of Time Lags** and observe how the results shift. Try:

* **8 lags** → 62.5 ms — a short window, sensitive to fast dynamics
* **16 lags** → 125 ms — roughly one alpha cycle
* **32 lags** → 250 ms — covers theta and slow alpha

After each change, click **Go PCA** and examine the **Scores Plot**, **Temporal Loadings**, and **Scree Plot**.

#### Questions:

* Does the separation between `open` and `closed` change as *L* increases?
* Do the Temporal Loading curves become smoother or more complex with longer windows?
* Does the number of components needed to explain most variance increase or decrease?
* Is there a window length where the eye-state separation appears clearest?

👉 Hint: too short a window (small *L*) means the analysis cannot see a full oscillation cycle — components will be dominated by adjacent-sample correlations rather than meaningful rhythms. Too long a window risks mixing patterns from different brain states within a single window, and creates a very wide trajectory matrix with many columns.

---

# What you should take away

After this exploration, you should be able to:

* Explain why EEG data is a multivariate time series and why rows are not independent
* Understand why standard PCA ignores temporal order — and what this means in practice
* Describe the SSA embedding step: sliding windows, trajectory matrix, window length *L*
* Interpret **Temporal Loadings** as spatiotemporal patterns across channels and lags
* Use **Variable Importance** to identify which channels drive each temporal component
* Make an informed choice of window length based on the sampling rate and the dynamics of interest

You should also recognise the trade-off:

* Standard SVD PCA is simple, fast, and easy to interpret — each loading is one channel
* Temporal PCA reveals rhythmic and dynamic structure invisible to standard PCA, but the trajectory matrix is much larger and the components require the Temporal Loadings plot to interpret

---

## Final Reflection

> Standard PCA treats the EEG table as a collection of independent snapshots. Temporal PCA, by embedding the data into sliding windows, gives PCA access to *sequences* — and the resulting components can represent oscillations and temporal dynamics rather than just spatial correlations. The same SVD algorithm is used in both cases; the embedding step is what makes the difference.

#### Questions:

* Why does shuffling the rows of an EEG table leave standard PCA unchanged, but break Temporal PCA?
* The SSA algorithm has four steps: embedding, decomposition, grouping, and reconstruction. GoPCA implements the first two. What would you gain from the grouping and reconstruction steps — and what tasks would those enable?
* A window length of 16 at 128 Hz covers 125 ms. Alpha oscillations have a period of roughly 100 ms. What window length would you choose if you were primarily interested in theta band activity (4–8 Hz, period 125–250 ms)?
* Could the Temporal PCA scores be used as input features to a classifier predicting `eye_state`? What might be the advantage compared to using the raw EEG values?
* How does the trajectory matrix interpretation differ between a single-channel SSA and a 14-channel MSSA?

---

## References

Golyandina, N., Korobeynikov, A., Shlemov, A., & Usevich, K. (2015). Multivariate and 2D extensions of singular spectrum analysis with the Rssa package. *Journal of Statistical Software*, 67(2), 1–78. https://doi.org/10.18637/jss.v067.i02

Golyandina, N. (2020). Particularities and commonalities of singular spectrum analysis as a method of time series analysis and signal processing. *WIREs Computational Statistics*, 12(4), e1487. https://doi.org/10.1002/wics.1487

Roesler, O., & Suendermann, D. (2013). A first step towards eye state prediction using EEG. In *Proceedings of the AIHLS 2013*. UCI Machine Learning Repository. https://archive.ics.uci.edu/dataset/264/eeg+eye+state
