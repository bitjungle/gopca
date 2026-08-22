# Exploring Structure in Data: The Corn (NIR) Dataset and PCA

## Background: Spectroscopy, chemistry, and a different kind of data

Near-infrared (NIR) spectroscopy is a rapid, non-destructive technique widely used in food science and agriculture. Instead of breaking down a sample chemically, an NIR instrument shines near-infrared light through the sample and measures how much light is absorbed at each wavelength. The resulting *absorption spectrum* reflects the molecular composition of the sample — water, fat, protein, and starch all absorb NIR light at characteristic wavelengths.

This dataset consists of **80 corn samples**, each measured by an NIR instrument. The instrument recorded absorbance at **700 wavelengths** from **1100 nm to 2498 nm** in steps of 2 nm. For each sample, the true compositional values were also determined by conventional laboratory (wet chemistry) methods:

* `Moisture#target` — moisture content (%)
* `Oil#target` — oil content (%)
* `Protein#target` — protein content (%)
* `Starch#target` — starch content (%)

These four columns are excluded from the PCA automatically and are available for colouring the scores plot — you will use them in Step 4.

The original purpose of such a dataset is **calibration**: build a model that predicts chemical composition from the spectrum alone, so future samples can be analysed in seconds rather than hours. This is a supervised problem — it uses the known laboratory values to train a model.

Here, we will approach the same data with **Principal Component Analysis (PCA)** — an *unsupervised* method that ignores the laboratory values entirely. The question becomes: what structure does PCA find in the spectra on its own? And does that structure relate to chemistry?

The data originates from Cargill and was made publicly available through Eigenvector Research (https://eigenvector.com/data/Corn/). It has been used in food chemistry research, including Engel et al. (2022).

---

## From Wine to Corn: a completely different kind of data

The Iris and Wine datasets consist of **independent measurements** — sepal length, alcohol content, proline concentration. Each variable has its own physical meaning and could in principle be measured on its own.

The Corn NIR dataset is built differently. The 700 variables are not independent measurements — they are **700 consecutive points on a single continuous spectrum**, recorded at wavelengths 2 nm apart. You cannot meaningfully interpret one wavelength in isolation; the information is in the *shape* of the entire curve.

This has two immediate consequences that make Corn unlike anything you have seen before:

**The pairplot is not just impractical — it is inconceivable.** Recall the scaling table from the Iris tutorial: 4 variables give 6 panels, 13 variables give 78. For 700 variables the number of pairplot panels is *700 × 699 / 2* = **244,650 panels**. PCA is not merely convenient here — it is the only realistic tool for exploring this data.

**Adjacent variables are not independent.** In Iris, sepal length and petal length are correlated but measure genuinely different things. In a NIR spectrum, the absorbance at 1500 nm and at 1502 nm are almost identical — they are sampling the same physical absorption feature separated by one instrument step. This extreme collinearity means PCA can compress 700 variables into 2–3 components and lose almost nothing, because most of the 700 variables are carrying nearly the same information.

> Corn NIR is the natural next step: more variables than you can count, all highly correlated. PCA is not one option among several — it is the standard first tool for this kind of data.

---

## First look at the data

Below is a plot showing all 80 spectra overlaid on a single graph.

![Corn NIR spectra](./corn_spectra.png)

Take a few minutes to study this figure carefully.

### Reflect:

* Do all spectra look similar in overall shape?
* Can you see peaks or shoulders at specific wavelengths where absorbance changes sharply?
* Do any spectra stand out from the rest — sitting above or below the main group?

👉 Even though the spectra look broadly similar, small sample-to-sample differences carry the chemical information we are trying to understand. PCA will find and compress this variation.

---

## The challenge

Each sample is described by **700 highly correlated variables**, but there are only **80 samples**.

Working directly in 700 dimensions is impossible. Even looking at individual wavelengths one at a time misses the fact that they are not independent — the entire shape of the spectrum matters.

Your task is to use **Principal Component Analysis (PCA)** to:

> Compress the spectral data into a small number of components while preserving as much of the variation as possible.

Think of PCA as:

* A way to find the **main patterns of spectral variation** across samples
* A way to **detect unusual samples** that do not fit the main pattern
* A first step toward understanding which parts of the spectrum carry chemical information

---

# Your task: Explore the dataset using GoPCA

You will use the application **GoPCA** to explore the dataset.

Do not try to "get the right answer" immediately.
Instead, **experiment, observe, and reflect**.

---

## Step 1: Run PCA with default settings — and diagnose the result

Load the dataset by clicking the **Corn (NIR)** sample-dataset button — if you opened this tutorial from that button, the data is already loaded. Leave all settings at their defaults:

* **Row Preprocessing** → None
* **Column Preprocessing** → Mean Center Only
* **PCA Method** → SVD

Click **Go PCA**.

First, open the **Scree Plot**.

#### Questions:

* How much variance does PC1 alone explain?
* Does this number seem surprisingly high?

Now open the **Loadings Plot** for PC1.

#### Questions:

* Is the loading curve always positive — does it ever cross zero?
* Does it rise smoothly from left to right, or does it show peaks and valleys?
* Does this curve look like it is describing the chemistry of corn?

👉 You should see that PC1 explains an extremely large fraction of the total variance — above 99%. The loading curve is entirely positive and rises monotonically from short to long wavelengths, never crossing zero. **This is not chemistry — this is a baseline artefact.**

NIR spectra are affected by **multiplicative scatter**: differences in particle size, sample packing density, and optical path length between samples cause the entire baseline to shift up or down and tilt from left to right — even when the chemistry is identical. Without correcting for this, PCA finds the direction of maximum variance, which turns out to be the direction in which the baseline slopes vary most. PC1 at 99% is telling you how tilted each sample's baseline is. It is doing exactly what it is designed to do — but the dominant variation is a physical artefact, not a chemical signal.

> Compare this to the Wine tutorial: without standardisation, proline's large numerical range dominated PC1. Here, baseline slope dominates for the same reason — it is the largest source of variance in the raw data. The fix follows the same logic.

---

## Step 2: Apply SNV and compare

NIR spectra need **row-wise scatter correction** before column-wise preprocessing. The standard method in GoPCA is **SNV (Standard Normal Variate)**.

SNV operates on each spectrum individually:

* Subtract the mean of that spectrum's 700 absorbance values
* Divide by its standard deviation

This centres and scales each spectrum *within itself*, removing baseline offset and tilt regardless of scatter differences between samples. It is a physical correction, not a statistical one.

In GoPCA, apply:

* **Row Preprocessing** → **SNV**
* **Column Preprocessing** → **Mean Center**

Click **Go PCA**.

#### Questions:

* How does the explained variance for PC1 change compared to Step 1?
* Open the Loadings Plot for PC1 — does the curve now cross zero and show peaks and valleys?
* Does the shape of the loading curve look more like a spectral feature now?

👉 After SNV, PC1 will typically explain 70–80% of variance rather than 99%. The loading curve will show positive and negative regions corresponding to real absorption features in the NIR spectrum. This is the chemical structure that was hidden underneath the baseline artefact.

> Keep **SNV + Mean Center** active for all remaining steps. This is the standard starting point for NIR PCA.

---

## Step 3: Understand what the loadings mean for spectral data

Open the **Loadings Plot** (with SNV applied).

👉 Important: this plot looks very different from Iris and Wine. For those datasets, the loadings plot showed a small number of isolated arrows — one per variable. Here there are **700 variables**. Each is a point in the loadings plot, ordered by wavelength. Instead of isolated arrows, the 700 points form a **smooth curve** in loading space. The shape of this curve is itself a spectral signature — it tells you which wavelength regions drive each principal component.

#### Questions:

* Which wavelength regions show the strongest positive and negative contributions to PC1?
* How does the PC2 loading curve differ from PC1?
* Do the loading curves show any sharp features, or are they all broad and smooth?

👉 Peaks and valleys in the loading curve correspond to wavelength regions of particular chemical relevance. The smooth shape reflects the high correlation between adjacent wavelengths — the same physical absorption feature spans many consecutive channels.

---

## Step 4: Connect PCA to chemistry

Now lets try to colour the samples by composition.

Look at the **Scores Plot**. Set the **Color by** to `Moisture#target`.

#### Questions:

* Do you see a gradient in the scores — samples ordered by moisture content?
* Does the gradient run along PC1, PC2, or diagonally?

👉 Try switching the **Color by** to `Starch#target`, then `Protein#target`, then `Oil#target`:

* Which compositional property is most strongly captured by PC1?
* Are all four properties captured equally well, or do some require higher components?

👉 Key insight: PCA captures **continuous variation**, not clusters. When scores correlate with a chemical property, it means the spectral variation along that component is driven by that chemistry. This is the foundation of NIR calibration.

---

## Step 5: Look for outliers

Return to the Scores Plot with no **Color by** selected.

#### Questions:

* Are any samples noticeably far from the main group?
* What might cause an outlier in NIR spectral data?

👉 Possible causes of spectral outliers:

* Measurement artefact (instrument problem during that scan)
* Sample contamination or unusual physical properties (e.g. very different particle size)
* A genuinely unusual corn sample

Outlier detection is one of the most practical uses of PCA in spectroscopy. A sample that appears as an outlier in the scores plot should be examined carefully before being included in a calibration model.

---

## Step 6: Push your understanding further

Try these explorations:

* **Exclude the water absorption regions** using GoCSV's column selection before loading into GoPCA

  * The strong water bands at ~1400–1450 nm and ~1900–1960 nm can dominate the spectrum
  * Does removing them change which component captures moisture vs starch?

* **Explore the 3D Scores Plot**

  * Does a third component reveal additional structure not visible in 2D?

* **Compare all four compositional targets**

  * Color by `Moisture#target`, `Oil#target`, `Protein#target`, and `Starch#target` in turn
  * Which is best separated along PC1? Which requires PC2 or PC3?

---

# What you should take away

After completing this exploration, you should be able to:

* Understand PCA as a tool for **high-dimensional spectral data**
* Interpret:

  * Scores plots showing continuous sample variation
  * Loadings as smooth spectral curves rather than isolated arrows
* Explain how PCA can:

  * Compress 700 correlated variables into a few meaningful components
  * Reveal chemical structure without using laboratory reference values
  * Detect unusual samples (outliers)
* Recognise the importance of:

  * **SNV preprocessing** for removing physical scatter effects in spectral data
  * The difference between row-wise (per-spectrum) and column-wise preprocessing

---

## Final reflection

Think about this:

> You started with 700 variables forming a spectrum — far more variables than samples.
> With PCA, you compressed the data into 2–3 components and revealed structure that relates to chemistry.

* How is PCA on spectral data different from PCA on independent measurements like Iris or Wine?
* Why do the loadings appear as smooth curves rather than isolated vectors?
* Why is SNV more appropriate here than column-wise standard scaling?
* The original purpose of this dataset was to build a calibration model predicting composition from spectra. Could you use the PCA scores as input to such a model? What might be the advantage of using scores instead of raw spectra? Hint: search for **principal component regression** on the internet.

---

## References

Eigenvector Research. *Corn dataset*. https://eigenvector.com/data/Corn/

Engel, J., et al. (2022). Towards predicting the quality of food mixtures using NIR spectroscopy.
*Food Chemistry*, 383, 132442. https://doi.org/10.1016/j.foodchem.2022.132442
