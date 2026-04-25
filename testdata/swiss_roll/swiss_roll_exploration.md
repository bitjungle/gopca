# Exploring Structure in Data: The Swiss Roll Dataset and Kernel PCA

## Background: Manifolds, geometry, and the limits of linear methods

The Swiss Roll is a **synthetic benchmark** — a dataset constructed mathematically rather than measured in the real world. Synthetic datasets like this one play a special role in data science: they let us test what an algorithm *should* find, because we already know the true structure.

The dataset consists of **1,000 samples** in three variables:

* `X`, `Y`, `Z` — the 3D coordinates of each point
* `color_category` — a label (A through H) encoding position along the roll

The data is generated from two parameters:

* *t* — controls position along the length of the roll (the "unrolled" coordinate)
* *h* — a random height offset adding thickness

The 3D coordinates follow from:

* x = t · cos(t)
* y = h
* z = t · sin(t)

The `color_category` column divides *t* into eight equal bins labelled A (inner edge) to H (outer edge). PCA does not use these labels during computation — they are used only for colouring the scores plot afterwards.

👉 This dataset is fundamentally different from Iris, Wine, and Corn:

* It has only **3 variables** — no dimensionality problem in the usual sense
* The challenge is entirely **geometric**: the 2D structure is *curved* in 3D space
* The goal is not compression — it is **recovering the true flat 2D sheet** hidden inside the roll

This type of structure is called a **manifold** — a surface that is locally flat but globally curved. A piece of paper is flat; roll it up and it becomes a manifold embedded in 3D. The Swiss Roll is exactly this: a flat 2D sheet wound into a helix.

The central question of this tutorial is:

> Can PCA find and unroll this manifold — or does it fail?

The answer depends on *which kind* of PCA you use.

---

## Step 1: Load the data and visualise the raw structure

Load the dataset by clicking the **Swiss Roll** sample dataset button.

Do not set a target column yet — explore the raw geometry first.

Below is a 3D visualisation of the data, coloured by position along the roll:

![Swiss Roll 3D](./swiss_roll_3d.png)

Study the figure carefully.

#### Questions:

* What overall shape do the data points form?
* Can you see the "layers" of the roll — the colour gradient progressing from A (inner) to H (outer)?
* If you could cut the roll along one edge and flatten it, what shape would it become?

👉 Hint: a perfectly flattened Swiss Roll would be a rectangle, with one axis representing *t* (position along the roll) and the other representing *h* (height). The goal of dimensionality reduction here is to recover exactly that rectangle.

---

## Step 2: Run linear PCA and observe the result

Set:

* **PCA Method** → SVD (or NIPALS)
* **Target Column** → `color_category`

Open the **Scores Plot (PC1 vs PC2)**.

#### Questions:

* Is the colour gradient ordered cleanly from A to H?
* Are points that are close in the gradient (e.g. B and C) also close together in the scores plot?
* Does the scores plot look like a flat unrolled sheet — or is the colour pattern mixed?

👉 You will likely see that colours are **jumbled**. Points from opposite sides of the roll — far apart along the manifold — are projected on top of each other.

Now open the **Scree Plot**.

#### Questions:

* How much variance is explained by PC1 and PC2 together?
* Does the high explained variance mean the structure has been correctly recovered?

👉 Key insight: linear PCA can explain a large fraction of the total variance and still fail to reveal the true structure. Variance is a property of the coordinate system — not of the manifold geometry.

---

## Step 3: Why does linear PCA fail, and what does Kernel PCA do differently?

### Why linear PCA fails

Linear PCA finds directions of **maximum variance** using straight-line projections. For the Swiss Roll, the direction of greatest variance runs *across* the roll — from the inner to the outer layers — not *along* its surface.

This means points that are far apart along the manifold (e.g. at opposite ends of the unrolled sheet) end up nearby in the linear PCA scores plot, because they are close in straight-line distance through 3D space. The rolled geometry is lost.

Analogy: draw a straight line across a rolled-up map. When you unroll the map, that line is no longer straight — it zig-zags across the sheet. Linear projection cannot follow a curved surface.

### How Kernel PCA works

Kernel PCA addresses this by replacing the linear projection with a **nonlinear one**, using two ideas:

**1. Mapping to a feature space.** Each data point x is implicitly mapped to a point Φ(x) in a higher-dimensional (possibly infinite-dimensional) feature space **F** via a nonlinear function Φ. Standard linear PCA is then performed in **F**. Directions that are *nonlinear* in the original 3D space can be *linear* in **F**.

**2. The kernel trick.** Computing Φ(x) explicitly for a high-dimensional **F** would be expensive or impossible. The kernel trick avoids this entirely: instead of computing Φ(x), we compute only pairwise *similarities* between data points. The **RBF (Gaussian) kernel** used here is:

> k(x, y) = exp(−γ · ‖x − y‖²)

where ‖x − y‖² is the squared Euclidean distance between x and y in the original 3D space, and γ (gamma) is a free parameter controlling the shape of the kernel.

All the information needed to perform PCA in **F** can be derived from this n × n matrix of pairwise kernel values — here a 1,000 × 1,000 matrix. The data is never explicitly mapped to the high-dimensional space.

**What this means geometrically**: the RBF kernel assigns high similarity to nearby points and low similarity to distant ones. Points connected along the surface of the roll are close in terms of path length along the manifold, even if they are far in straight-line 3D distance. With the right choice of γ, the kernel can distinguish these cases — and PCA in the resulting feature space recovers the unrolled structure.

👉 Kernel PCA was introduced by Schölkopf, Smola & Müller (1998). Its properties and applications to de-noising are developed further in Mika et al. (1998). See the References section.

---

## Step 4: Switch to Kernel PCA and unroll the manifold

Change:

* **PCA Method** → Kernel PCA
* **Kernel Type** → RBF
* **Gamma** → 0.1

Look at the **Scores Plot**.

#### Questions:

* Does the colour gradient now run more smoothly from A to H?
* Can you see a more rectangular structure?
* Does the scores plot resemble the unrolled sheet you imagined in Step 1?

👉 This is the key moment: with the right kernel and a reasonable gamma, Kernel PCA recovers the flat 2D manifold that linear PCA could not find.

**Preprocessing note**: when Kernel PCA is selected, GoPCA restricts the column-wise preprocessing options. Mean centering and standard scaling are unavailable — this is because Kernel PCA performs its own centering in kernel space. You may apply **Variance Scale** if your features have different units, or leave preprocessing as **None**.

**Visualisation note**: several plots available for linear PCA are not available for Kernel PCA — specifically the **Loadings Plot**, **Biplot**, **Biplot3D**, **Correlations Plot**, and **Diagnostic Scatter Plot**. These require loadings in the original variable space, which Kernel PCA does not produce (its components live in the high-dimensional feature space **F**). The **Scores Plot**, **Scree Plot**, and **Kernel Matrix Heatmap** remain available.

---

## Step 5: Explore the effect of the gamma parameter

The gamma parameter directly controls the shape of the RBF kernel:

> k(x, y) = exp(−γ · ‖x − y‖²)

* **Large γ** → the exponential decays quickly with distance → only very close points have high similarity → the kernel is *local*, sensitive to fine neighbourhood structure
* **Small γ** → the exponential decays slowly → even distant points have similar kernel values → the kernel is *global*, insensitive to local geometry

Try these values:

* Gamma = 0.33
* Gamma = 0.1
* Gamma = 0.05
* Gamma = 0.01

#### Questions:

* Which gamma gives the cleanest gradient from A to H in the scores plot?
* What happens when gamma is too large — can you see the structure fragmenting?
* What happens when gamma is too small — does the unrolled sheet collapse or distort?
* Is there a clear "best" gamma, or is the result relatively stable across a range?

👉 Hint: the best gamma is the one that matches the characteristic length scale of the manifold — roughly the typical distance between neighbouring points along the surface. Too local and the kernel cannot see the manifold; too global and it cannot distinguish the layers.

---

## Step 6: Compare linear PCA and Kernel PCA directly

Switch back and forth between:

* **SVD** (linear PCA)
* **Kernel PCA** with **Kernel Type** → RBF and your best gamma from Step 5

Compare the **Scores Plot** each time.

#### Questions:

* Is the difference between the two methods subtle or dramatic?
* Which method reveals the true 2D structure of the data?
* Can you describe in your own words why one method succeeds and the other does not?

👉 This comparison is the central lesson of the Swiss Roll dataset. A method that captures nearly all of the total variance (linear PCA) can nonetheless completely fail to recover the meaningful structure — while a method that operates via a seemingly indirect route (kernel feature mapping) finds it immediately.

---

## Step 7: Limitations and real-world relevance

The Swiss Roll is a clean, noise-free example designed to make the comparison vivid. Real-world data is messier:

* Noise blurs the manifold boundaries
* The true dimensionality of the manifold is rarely known
* The manifold structure may only exist locally, not globally

Kernel PCA introduces challenges that linear PCA does not have:

* You must choose a **kernel function** — RBF is most versatile, but polynomial and other kernels exist
* You must tune **gamma** — without knowing the true structure, this requires cross-validation or domain knowledge
* The kernel matrix scales as n × n in memory and computation — for very large datasets, this can be prohibitive

In many real datasets — including Iris and Wine — linear PCA is entirely sufficient. However, nonlinear manifold structure appears in genuine scientific data:

* **Single-cell RNA sequencing**: cells form curved trajectories through gene expression space as they differentiate
* **Image datasets**: images of a rotating object lie on a curved manifold parameterised by rotation angle
* **Sensor data**: physical processes with nonlinear dynamics produce data on curved manifolds in measurement space
* **Spectroscopic data**: nonlinear mixture effects can create curved structure in spectral space

In these cases, Kernel PCA or related methods (UMAP, t-SNE, Isomap) may reveal structure that linear PCA cannot.

---

## Final Reflection

> You started with data that lives in 3D space — seemingly straightforward. Yet linear PCA, which handles hundred-variable spectral datasets successfully, completely failed to find the structure. Kernel PCA, by implicitly mapping the data into an infinite-dimensional space, recovered it with ease. The lesson is not that one method is always better: it is that the right method depends on the geometry of the data.

#### Questions:

* What is the difference between a linear and a nonlinear manifold? Can you give an example of each from the datasets you have explored?
* Why does Kernel PCA succeed where linear PCA fails on the Swiss Roll?
* The RBF kernel has one key parameter, gamma. How would you choose gamma in a real-world problem where you do not know the true structure?
* Kernel PCA's components live in an infinite-dimensional feature space and have no direct interpretation in the original variables. Is this a problem? What can you still learn from the scores plot alone?
* Could you use Kernel PCA scores as input features for a predictive model? What might be the advantage or disadvantage compared to using the raw variables?

---

## References

Schölkopf, B., Smola, A., & Müller, K.-R. (1998). Nonlinear component analysis as a kernel eigenvalue problem. *Neural Computation*, 10(5), 1299–1319.

Mika, S., Schölkopf, B., Smola, A., Müller, K.-R., Scholz, M., & Rätsch, G. (1998). Kernel PCA and de-noising in feature spaces. In M. Kearns, S. Solla, & D. Cohn (Eds.), *Advances in Neural Information Processing Systems* (Vol. 11). MIT Press.

Marsland, S. (2014). *Machine Learning: An Algorithmic Perspective* (2nd ed.). CRC Press.
