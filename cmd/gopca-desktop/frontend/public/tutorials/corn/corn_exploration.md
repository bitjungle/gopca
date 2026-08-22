# Exploring Structure in Data: The Corn (NIR) Dataset and PCA

## Background: Spectroscopy, chemistry, and a different kind of data

Near-infrared (NIR) spectroscopy is a rapid, non-destructive technique widely used in food science and agriculture. Instead of breaking down a sample chemically, an NIR instrument shines near-infrared light through the sample and measures how much light is absorbed at each wavelength. The resulting *absorption spectrum* reflects the molecular composition of the sample — water, fat, protein, and starch all absorb NIR light at characteristic wavelengths.

This dataset consists of **80 corn samples**, each measured by an NIR instrument. The instrument recorded absorbance at **700 wavelengths** from **1100 nm to 2498 nm** in steps of 2 nm. For each sample, the true compositional values were also determined by conventional laboratory (wet chemistry) methods:

* `Moisture#target` — moisture content (%)
* `Oil#target` — oil content (%)
* `Protein#target` — protein content (%)
* `Starch#target` — starch content (%)

These four columns are excluded from the PCA automatically and are available for colouring the scores plot — you will use them in Step 4.

The dataset also carries the same four properties as **categorical** columns — `Moisture`, `Oil`, `Protein` and `Starch`, each binned into `Low` / `Mid` / `High`. These are excluded from the PCA too, and give you a group colouring to set against the continuous `#target` gradients. Both sets appear in the **Color by** dropdown.

The original purpose of such a dataset is **calibration**: build a model that predicts chemical composition from the spectrum alone, so future samples can be analysed in seconds rather than hours. This is a supervised problem — it uses the known laboratory values to train a model.

Here, we will approach the same data with **Principal Component Analysis (PCA)** — an *unsupervised* method that ignores the laboratory values entirely. The question becomes: what structure does PCA find in the spectra on its own? And does that structure relate to chemistry?

The data originates from Cargill and was made publicly available through Eigenvector Research (https://eigenvector.com/data/Corn/). It has been used in food chemistry research, including Fatemi et al. (2022), who searched these same spectra for the wavelength ranges most informative about each constituent.

The full Eigenvector dataset contains the same 80 samples measured on three different NIR instruments. The copy bundled with GoPCA is the **m5** instrument only, so every spectrum you see was recorded on the same machine.

![A mature maize ear on the stalk](./mature_maize_ear_on_a_stalk.jpg)
*A mature maize ear on the stalk — the raw material behind all 80 spectra in this dataset. Photo: Silverije, [CC BY-SA 3.0](https://creativecommons.org/licenses/by-sa/3.0/)*

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

![Corn NIR spectra](./corn_spectra.jpg)

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

> **Settings** — Row-wise: **None** · Column-wise: **Mean Center Only** · Method: **SVD** · Components: 5

Load the dataset by clicking the **Corn (NIR)** sample-dataset button — if you opened this tutorial from that button, the data is already loaded. Leave all settings at their defaults:

* **Step 1: Row-wise Preprocessing** → None
* **Step 2: Column-wise Preprocessing** → Mean Center Only
* **PCA Method** → SVD

Click **Go PCA!**.

First, open the **Scree Plot**.

#### Questions:

* How much variance does PC1 alone explain?
* Does this number seem surprisingly high?

Now open the **Loadings Plot** for PC1.

#### Questions:

* Is the loading curve always positive — does it ever cross zero?
* Does it rise smoothly from left to right, or does it show peaks and valleys?
* Does this curve look like it is describing the chemistry of corn?

👉 You should see that PC1 explains an extremely large fraction of the total variance — **above 99%**. And here is the diagnostic that matters: the loading curve is **entirely positive** — it never crosses zero. It also trends upward across the range, from about 0.007 at 1100 nm to 0.050 at 2498 nm, though not smoothly: you will see distinct shoulders near 1200, 1450 and 1900 nm, because the artefact scales the real absorption features along with everything else.

The sign is the giveaway. A component whose loadings are all the same sign says "every wavelength moves up and down together" — that is a description of the whole spectrum shifting bodily, not of one chemical constituent trading off against another. **This is not chemistry; it is a baseline artefact.**

NIR spectra are affected by **multiplicative scatter**: differences in particle size, sample packing density, and optical path length between samples cause the entire baseline to shift up or down and tilt from left to right — even when the chemistry is identical. Without correcting for this, PCA finds the direction of maximum variance, which turns out to be the direction in which the baseline slopes vary most. PC1 at 99% is telling you how tilted each sample's baseline is. It is doing exactly what it is designed to do — but the dominant variation is a physical artefact, not a chemical signal.

> Compare this to the Wine tutorial: without standardisation, proline's large numerical range dominated PC1. Here, baseline slope dominates for the same reason — it is the largest source of variance in the raw data. The fix follows the same logic.

---

## Step 2: Apply SNV and compare

NIR spectra need **row-wise scatter correction** before column-wise preprocessing. The standard method in GoPCA is **SNV (Standard Normal Variate)**.

SNV operates on each spectrum individually:

* Subtract the mean of that spectrum's 700 absorbance values
* Divide by its standard deviation

This centres and scales each spectrum *within itself*, removing baseline offset and tilt regardless of scatter differences between samples. It is a physical correction, not a statistical one.

> **Settings** — Row-wise: **SNV** · Column-wise: **Mean Center Only** · Method: SVD · Components: 5

In GoPCA, apply:

* **Step 1: Row-wise Preprocessing** → **SNV (Standard Normal Variate)**
* **Step 2: Column-wise Preprocessing** → **Mean Center Only**

Click **Go PCA!**.

#### Questions:

* How does the explained variance for PC1 change compared to Step 1?
* Open the Loadings Plot for PC1 — does the curve now cross zero and show peaks and valleys?
* Does the shape of the loading curve look more like a spectral feature now?

👉 Two things change. PC1 drops from 99% to about **84%**, and — more importantly — the loading curve now **crosses zero about ten times**, with clear positive and negative regions. That sign structure is the improvement: the component is no longer saying "everything moves together", it is contrasting one part of the spectrum against another, which is what a chemical signal looks like.

But be careful about how much credit to give SNV here, because this is where a lot of spectroscopy write-ups overclaim. PC1 still holds 84% of the variance, and it is **still not chemistry**. If you correlate the PC1 scores against the average absorbance of each raw spectrum you get about **0.92** — PC1 is largely still tracking how bright each spectrum is overall. Its two largest loadings sit at the extreme ends of the wavelength range, which is the shape of a residual tilt rather than an absorption band.

That is not a failure of SNV, and it is not a mistake on your part. SNV is a *first-order* scatter correction: it removes the additive offset and the overall scale of each spectrum, but real scatter is wavelength-dependent, so a residual survives. This is precisely why chemometricians also reach for MSC, EMSC and derivative preprocessing. The chemistry is in this data — you will find it in Step 4 — but it is not in PC1.

> Keep **SNV + Mean Center Only** active for all remaining steps. This is the standard starting point for NIR PCA.

---

## Step 3: Understand what the loadings mean for spectral data

> **Settings** — Row-wise: SNV · Column-wise: Mean Center Only · Method: SVD · Components: 5

Nothing to change — open the **Loadings Plot** on the analysis you just ran.

👉 Important: this plot looks very different from Iris and Wine. For those datasets, the loadings plot showed a small number of isolated arrows — one per variable. Here there are **700 variables**. Each is a point in the loadings plot, ordered by wavelength. Instead of isolated arrows, the 700 points form a **smooth curve** in loading space. The shape of this curve is itself a spectral signature — it tells you which wavelength regions drive each principal component.

#### Questions:

* Which wavelength regions show the strongest positive and negative contributions to PC1?
* How does the PC2 loading curve differ from PC1?
* Do the loading curves show any sharp features, or are they all broad and smooth?

👉 Peaks and valleys in the loading curve mark wavelength regions that drive that component. The smoothness is not cosmetic — it is a direct consequence of the collinearity described earlier: adjacent wavelengths correlate at about **0.999998** in this dataset, so a loading curve physically cannot jump from one channel to the next. Any sharp spike you ever see in a spectral loading plot is far more likely to be a bad channel or a dead detector pixel than a chemical feature.

This way of reading loadings — as a *loading spectrum* rather than as a set of arrows — is standard practice in chemometrics (Esbensen et al., 2002, Ch. 3).

One caution carried over from Step 2: the PC1 loading curve is legitimately interesting to look at, but remember what PC1 mostly is on this data. Save your chemical interpretation for PC2, which Step 4 will show carries the composition signal.

---

## Step 4: Connect PCA to chemistry

> **Settings** — Row-wise: SNV · Column-wise: Mean Center Only · Method: SVD · Components: 5

Now let us colour the samples by composition and find out where the chemistry actually lives.

Look at the **Scores Plot**. Set **Color by** to `Moisture#target`.

#### Questions:

* Do you see a gradient in the scores — samples ordered by moisture content?
* Does that gradient run along PC1, along PC2, or diagonally?
* Switch **Color by** to `Protein#target`, then `Starch#target`, then `Oil#target`. Does the gradient run the same way each time?

👉 This is the payoff of the step, and it may not be what you expected.

**The chemistry is in PC2, not PC1.** PC1 carries 84% of the variance and correlates only weakly with any laboratory value — 0.38 with moisture, and just 0.09 with oil. PC2 carries 8.3% of the variance and does far better: about **0.67 with protein** and **0.58 with moisture**.

So the component holding ten times more variance holds much less of the chemistry. That is the same lesson the Swiss Roll tutorial teaches from the opposite direction: **variance measures how much a direction moves, not whether the movement means anything.** PC1 is big because residual scatter is big.

Oil is the awkward one. It shows up best around PC5, and hardly at all in the first two components — which is why a real NIR calibration for oil would use many more components than you would guess from a scree plot.

> **Reality check on "2–3 components is enough".** It is enough to *display* this data — the first three components hold 95% of the variance. It is not enough to *predict* composition from it. Fit a regression of each laboratory value on the first few components and cross-validate: with three components, moisture and protein reach an R² of roughly 0.2, while oil and starch do worse than simply guessing the mean. Around ten components are needed before all four are predicted respectably. Compression and prediction are different jobs, and the scree plot only tells you about the first.

👉 Key insight: PCA captures **continuous variation**, not clusters — no amount of colouring will produce tidy groups here, because corn composition varies smoothly. When scores do correlate with a chemical property, it means the spectral variation along that component is driven by that chemistry. That is the foundation of NIR calibration, and of **principal component regression**, which you will meet in the final reflection.

**Try the categorical colouring too.** Set **Color by** to `Protein` (the `Low`/`Mid`/`High` version rather than `Protein#target`). The same information, binned into three groups, often makes a weak gradient easier to see than a continuous colour ramp does.

---

## Step 5: Look for outliers

> **Settings** — Row-wise: SNV · Column-wise: Mean Center Only · Method: SVD · Components: 5

Return to the **Scores Plot** with **Color by** set to None.

#### Questions:

* Are any samples noticeably far from the main group?
* What might cause an outlier in NIR spectral data?

👉 Two samples sit clearly apart from the rest, both at the high end of PC1.

Now go back to the spectra figure at the top of this tutorial and look again at the top of the band. Those same two spectra are visible there, sitting above the main group across the whole wavelength range. You could have spotted them before running any analysis at all — and since PC1 is largely tracking overall absorbance (Step 2), it is no coincidence that PCA flags exactly the two brightest spectra.

That is worth remembering: **the scores plot did not discover something invisible here, it confirmed something the raw data already showed.** On a dataset with more variables or subtler outliers it would earn its keep more dramatically, but it is always worth checking whether a "discovery" was visible in the raw data all along.

Possible causes of spectral outliers:

* Measurement artefact (instrument problem during that scan)
* Sample contamination or unusual physical properties (e.g. very different particle size)
* A genuinely unusual corn sample

Outlier detection is one of the most practical uses of PCA in spectroscopy. A sample that appears as an outlier in the scores plot should be examined carefully before being included in a calibration model.

---

## Step 6: Push your understanding further

> **Settings** — Row-wise: SNV · Column-wise: Mean Center Only · Method: SVD, except where noted

Three explorations, in increasing order of effort.

* **Compare preprocessing choices directly.** Run the analysis three ways — Mean Center Only, then Standard Scale without SNV, then SNV + Mean Center Only — and note PC1's variance each time. Which correction actually addresses the artefact, and which one barely moves it? (The "Why SNV rather than standard scaling?" section below gives the reasoning; this lets you see it.)

* **Explore the 3D Scores Plot.** With three components covering 95% of the variance, does the third add structure you could not see in 2D, or is it mostly noise?

* **Exclude the water absorption regions.** Water absorbs strongly around **1400–1450 nm** and **1900–1960 nm**, and those bands can dominate a spectrum. Open the dataset in **GoCSV**, delete the wavelength columns in those two ranges (select the columns, then **Delete Column** from the right-click menu — that is 57 of the 700 channels), save, and load the result into GoPCA.

  Before you run it, predict what will happen. Then check: does moisture become harder to see? What happens to **oil**, which was hiding down in PC5? This is the most rewarding of the three, and the result is a genuine chemometric lesson about which parts of a spectrum carry which information.

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
  * That the **largest component is not automatically the meaningful one** — here PC1 remained largely physical even after correction, while the chemistry sat in PC2

---

## Why SNV rather than standard scaling?

This question is worth answering directly, because it is the one most people get wrong when they first meet spectral data.

**Standard scaling works on columns.** It gives every wavelength equal variance, which is the right fix when variables are measured in different units — proline in mg/L against pH, as in the Wine tutorial, where the standard deviations differ by a factor of about 2,500.

That is not this problem. Every one of the 700 wavelengths here is an absorbance, measured on the same instrument in the same units, and their standard deviations differ by only about **6×** across the whole spectrum. There is no unit mismatch to fix. Esbensen et al. (2002, Ch. 4) warn specifically that when variables already share a scale — spectroscopic data being their example — autoscaling mostly amplifies noise in the quiet channels.

**The artefact here runs the other way: between samples, not between wavelengths.** Two spectra of chemically identical corn can sit at different heights because of particle size or packing density. No amount of column scaling repairs that, because the problem is a property of each *row*. SNV normalises each spectrum within itself, which is exactly the right shape of correction.

You can verify this yourself: run the analysis with **Column-wise → Standard Scale** and no SNV. PC1 barely moves, from 99% to about 97% — still the same artefact, still dominating. Then switch to SNV and watch it fall to 84%.

---

## Final reflection

Think about this:

> You started with 700 variables forming a spectrum — far more variables than samples.
> With PCA you compressed them into three components holding 95% of the variance, and found chemistry in the second one. Along the way you saw that the largest component was mostly a physical artefact both before *and* after correction — a reminder that the size of a component says nothing about its meaning.

* How is PCA on spectral data different from PCA on independent measurements like Iris or Wine?
* Why do the loadings appear as smooth curves rather than isolated vectors?
* Why is SNV more appropriate here than column-wise standard scaling? *(Think about which direction the artefact runs in: does the scatter effect differ between wavelengths, or between samples?)*
* The original purpose of this dataset was to build a calibration model predicting composition from spectra. Could you use the PCA scores as input to such a model? What might be the advantage of using scores instead of raw spectra? Hint: search for **principal component regression** on the internet.

---

## References

Eigenvector Research. *Corn dataset*. https://eigenvector.com/data/Corn/

Fatemi, A., Singh, V., & Kamruzzaman, M. (2022). Identification of informative spectral ranges for predicting major chemical constituents in corn using NIR spectroscopy. *Food Chemistry*, 383, 132442. https://doi.org/10.1016/j.foodchem.2022.132442

Barnes, R. J., Dhanoa, M. S., & Lister, S. J. (1989). Standard normal variate transformation and de-trending of near-infrared diffuse reflectance spectra. *Applied Spectroscopy*, 43(5), 772–777. — the original description of SNV, and the definition GoPCA implements.

Esbensen, K. H., et al. (2002). *Multivariate Data Analysis – in Practice*, Chapters 3–4. CAMO. — 1D loading spectra for spectroscopic data, and why autoscaling can be the wrong choice when variables already share a scale.
