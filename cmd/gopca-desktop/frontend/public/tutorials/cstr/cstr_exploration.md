# Monitoring a Chemical Reactor: Process Data and Temporal PCA

## Background: Why reactors produce time-series data

In a chemical plant, sensors log dozens of measurements every minute — temperatures, flow rates, concentrations, pressures, valve positions. The goal of **process monitoring** is to detect abnormal behaviour early, understand the relationship between variables, and distinguish faults from ordinary process disturbances.

This dataset was simulated from a **non-isothermal Continuous Stirred-Tank Reactor (CSTR)** running an irreversible first-order exothermic reaction:

$$\text{A} \rightarrow \text{B} \quad (\text{exothermic})$$

A CSTR is one of the most studied systems in chemical engineering. Reactant A flows in continuously, the reaction occurs inside the tank, and product B flows out. Because the reaction releases heat, a coolant circuit controls the reactor temperature.

### The PI temperature controller

A proportional-integral (PI) controller manipulates the coolant temperature T_c to keep the reactor temperature T near the setpoint (365 K). When T rises above setpoint, the controller requests colder coolant; when it falls, warmer coolant. Anti-windup logic prevents the controller from accumulating a large integral error when the coolant valve is fully open or closed.

This controller introduces **closed-loop dynamics** into the data — disturbances cause transients, the controller responds, and variables return to steady state. Temporal PCA can capture all of this.

![CSTR P&ID — process variables and instrumentation](./cstr_diagram.png)
*Simplified P&ID showing the measured variables, instrumentation, and PI temperature control loop for the simulated CSTR process.*

### The dataset

The simulation runs for **800 minutes** at 1-minute sampling, producing **801 observations** across **12 process variables**:

| Column | Symbol | Units | Description |
|---|---|---|---|
| `T_K` | *T* | K | Reactor temperature (controlled variable) |
| `Tc_out_K` | *T*_c,out | K | Coolant outlet temperature (manipulated via PI) |
| `Tf_K` | *T*_f | K | Feed temperature |
| `CA_mol_L` | *C*_A | mol/L | Reactant A concentration in reactor |
| `CB_mol_L` | *C*_B | mol/L | Product B concentration in reactor |
| `CAf_mol_L` | *C*_A,f | mol/L | Reactant A concentration in feed |
| `F_L_min` | *F* | L/min | Volumetric feed flow rate |
| `cooling_duty_kJ_min` | *Q* | kJ/min | Heat removed by coolant system |
| `reaction_rate_mol_L_min` | *r*_A | mol/L/min | Rate of reaction A→B |
| `conversion_fraction` | *X*_A | — | Fraction of A converted to B |
| `heat_transfer_UA_kJ_min_K` | *UA* | kJ/(min·K) | Overall heat-transfer coefficient |
| `residence_time_min` | *τ* | min | Hydraulic residence time (V/F) |

The dataset also includes string columns (`event`, `regime`) and a binary flag (`fault_active`) for coloring and interpretation. **Do not include these as PCA variables** — use `regime` as the color target.

### Operating scenarios

The simulation passes through six distinct operating phases designed to produce recognisable signatures in Temporal PCA:

| Time (min) | Scenario | What happens |
|---|---|---|
| 0 – 120 | **Normal** | Reactor near steady state. Use this period to establish your mental baseline. |
| 120 – 240 | **Feed concentration step** | Feed concentration rises +10 %. Reactant A rises, more heat is released, controller corrects temperature. |
| 240 – 360 | **Feed temperature step** | Feed enters +7 K hotter. Thermal disturbance; controller compensates with colder coolant. |
| 360 – 520 | **Flow oscillation** | Feed flow oscillates ±8 % at a **40-minute period**. Creates periodic variation across nearly all variables. |
| 520 – 680 | **Cooling fault** | Heat-transfer coefficient drops 28 % — partial fouling or cooling water problem. Reactor temperature drifts upward; controller saturates. |
| 680 – 800 | **Recovery** | Nominal conditions restored. |

---

## Key process time constants — choose your lag L wisely

Before running Temporal PCA, you need to decide how many lags L to include. The right choice depends on the dynamics you want to capture:

| Dynamic | Time constant | Recommended L |
|---|---|---|
| Residence time τ = V/F | ≈ 1 min | L = 2–5 |
| PI integral time τ_I | 8 min | L = 8–16 (captures one controller action) |
| Step response settling | ≈ 20–30 min | L = 20–30 |
| Flow oscillation period | 40 min | L = 40–80 (to fully resolve) |

Start with **L = 10** to see the controller dynamics clearly. Then try L = 5 (faster dynamics) and L = 20 (slower trends). Note: at L = 10–20 the 40-minute oscillation will appear as a low-frequency oscillatory pair of components — which is exactly what you want to identify.

---

# Your task: Monitor the reactor using Temporal PCA

Work through the steps in order. Each step builds on the previous one.

---

## Step 1: Load the data and configure GoPCA

Load `cstr_temporal_pca.csv` into GoPCA. Set:

* **Time column** → `time_min`
* **Target column** → `regime` (for coloring)
* **Preprocessing** → **Standard Scale**
* **Lags (L)** → **10**
* **Number of Components** → **8**
* **PCA Method** → SVD (default)

Click **Go Temporal PCA**.

> **Why Standard Scale?** The 12 variables have completely different units and magnitudes — temperatures in the hundreds of Kelvin, flow rates in hundreds of L/min, concentrations below 1 mol/L, heat duty in thousands of kJ/min. Without scaling, PCA would be dominated by whichever variable has the largest numerical variance, not the most process-relevant variation.

---

## Step 2: Read the Scree Plot — how many components carry information?

Open the **Scree Plot**.

#### Questions:

* Where is the elbow — after the first component, the second, or later?
* How many components are needed to reach 80% cumulative variance?
* How does the Scree Plot for this dataset compare to Iris (4 variables, clean separation) or Wine (13 variables, ~55% in 2D)?

👉 A CSTR at steady state has only a few degrees of freedom (temperature, concentration, flow are tightly coupled through the energy and mass balances). During disturbances and faults, additional variation enters from different directions. The Scree Plot reflects this: a few large components capturing steady-state variation, then smaller components capturing the specific disturbance signatures.

---

## Step 3: The scores plot — reading a process trajectory

Open the **Scores Plot (PC1 vs PC2)** and color by `regime`.

For time-series data, the scores form a **time-ordered trajectory** through PC space. This is different from a static dataset like Iris or Wine, where each point is an independent sample. Here, consecutive points are connected in time — the path the reactor traces through multivariate space.

**How to read a process trajectory:**

* **Tight cluster** → reactor operating in a stable, repeating condition
* **Slow drift away from cluster** → gradual process change (e.g., fouling, catalyst deactivation)
* **Sharp jump** → step disturbance (e.g., feed change)
* **Closed loop or ellipse** → periodic oscillation (the 40-minute flow cycle will appear this way)
* **Failure to return** → fault that the controller cannot correct

#### Questions:

* Can you identify which region of the plot corresponds to normal operation?
* Where does the cooling fault period appear — close to normal operation or far from it?
* Does the flow oscillation period trace a recognisable loop? How large is it compared to the steady-state cluster?
* Does the reactor return to the normal region during the recovery period?

> **Chemical engineering connection:** This is the same logic as a yield–selectivity map in a reaction engineering course. When you plot yield vs. selectivity over time during a runaway, the trajectory moves away from the optimal operating point. The scores plot is a multivariate generalisation of that idea — it shows where the process is in a compressed, multi-dimensional operating space.

---

## Step 4: The Temporal Loadings — what dynamics are in each component?

Open the **Temporal Loadings Plot**. Display at least 8 components.

Each panel shows how one component's loading evolves across lags 0 to L. The x-axis is lag (minutes into the past); the y-axis shows the loading of the dominant channel for that component.

**Three patterns to look for:**

| Shape | Interpretation |
|---|---|
| **Flat / near-zero** | No temporal structure — component captures instantaneous variance |
| **Monotone decay** | Exponential response — controller or first-order process dynamics |
| **Sinusoidal oscillation** | Periodic variation — oscillatory process behaviour |

#### Questions:

* Which components show a monotone decaying shape? What time constant does the decay suggest (how many lags before it flattens)?
* Can you find any components with a sinusoidal shape? These represent the 40-minute flow oscillation.
* Compare the decay time constant to the PI integral time τ_I = 8 minutes. Do they match?

> **Hint:** A sinusoidal temporal loading pattern means that component is tracking a variable that oscillates in time. The period of the oscillation can be estimated from the zero-crossings: if you count k zero-crossings over L lags, the oscillation period ≈ 2L/k minutes. For a 40-minute oscillation with L = 10 lags, you will see only about half a cycle — try L = 40 to resolve the full period.

---

## Step 5: Identify the oscillatory pair

Temporal PCA (SSA) represents a single oscillation as a **pair of components** with:

1. Nearly equal % variance (within a few percent of each other)
2. Sinusoidal temporal loadings — shifted by approximately 90° relative to each other

This pairing occurs because a sine wave requires both a sine and a cosine component to be fully represented.

#### Questions:

* Looking at the Scree Plot, can you find two adjacent components with similar % variance? Which components are they?
* Open the Temporal Loadings Plot for those two components. Are their curves sinusoidal and approximately 90° shifted from each other?
* Now look at which variable has the highest loading for this pair — which process variable is driving the oscillation? Does this make physical sense given that the flow rate oscillates?

> **Try L = 40:** With L = 10, the 40-minute oscillation completes only one quarter of a cycle across the lag window. With L = 40, you will see a complete sine and cosine pair. Run Temporal PCA again with L = 40 and 10 components, then re-examine the oscillatory pair.

---

## Step 6: Fault detection — does Temporal PCA see the cooling fault?

The cooling fault (minutes 520–680) reduces the heat-transfer coefficient by 28 %. The controller tries to compensate by lowering T_c, but eventually saturates at its minimum value.

Return to the Scores Plot.

#### Questions:

* Is the cooling fault period clearly separated from the normal operation cluster in PC1–PC2 space?
* If not, try PC1 vs PC3, or PC2 vs PC3 — the fault may project most strongly onto a higher component
* Open the Loadings Plot for the component(s) that best separate the fault period. Which variables have the largest loadings? Do they include `cooling_duty_kJ_min`, `heat_transfer_UA_kJ_min_K`, or `T_K`?

👉 A 28 % drop in UA is a significant fault. In a real plant, this would appear gradually over hours or days. The key question is: which combination of variables carries the fault signature, and how early in the fault period does it become visible in the scores plot?

#### Extension:

* In industrial practice, control charts are placed on the scores of a PCA model built on normal data only. Any new sample whose scores fall outside the 95% confidence ellipse is flagged as abnormal. Does the cooling fault period fall outside the 95% confidence ellipse for normal operation?

---

## Step 7: Compare lag settings — L = 5, L = 10, L = 20

Run Temporal PCA three times with different lag values: **L = 5**, **L = 10**, **L = 20** (keep 8 components, Standard Scale). Compare the Scree Plots and Temporal Loadings.

#### Questions:

* Does increasing L reveal more temporal structure (more sinusoidal components)?
* At L = 5, does the Scree Plot still show a clear elbow?
* At L = 20, how many components do you need to explain 80% of the variance? Why does the number increase with L?
* Which lag setting gives you the most useful separation between normal operation and the cooling fault in the scores plot?

> **Rule of thumb:** L should be at least as large as the longest process time constant you want to capture. For controller dynamics (τ_I = 8 min), L ≥ 10 is sensible. For the full flow oscillation (40 min), you need L ≥ 40. Larger L improves frequency resolution but increases the size of the trajectory matrix — for 801 observations and L = 40, each "augmented observation" spans 41 time steps, leaving 761 usable rows.

---

# What you should take away

After completing this exploration, you should be able to:

* Explain why **Standard Scale is essential** for process datasets with mixed units
* Read a **process trajectory** in a scores plot — identifying stable operation, disturbances, oscillations, and faults
* Recognise the **time constant** of process dynamics from the shape of temporal loading curves
* Identify a **paired oscillatory component** in the Scree Plot and Temporal Loadings
* Select an appropriate **lag parameter L** based on the process time constants you want to capture
* Describe how **Temporal PCA could be used for fault detection** in a real plant

---

## Final reflection

> You started with 12 process variables measured every minute for 800 minutes. A conventional pairwise analysis would give you 66 panels and no information about how the variables evolve together over time.

Think about these questions:

* The PI controller reduces temperature variation — it actively fights the disturbances. Does this make Temporal PCA harder or easier? What would the scores plot look like without any controller?
* The cooling fault changes `heat_transfer_UA_kJ_min_K` directly. But which other variables are affected indirectly, and how quickly? Can you trace the fault propagation through the scores trajectory?
* In a real plant, you would build the Temporal PCA model on normal data only, then apply it to new data in real time. What would you need to store — the loadings? The mean and standard deviation for scaling? The lag structure?
* The flow oscillation was designed as a process disturbance. Could it also be a deliberate sinusoidal test signal, used to identify the process frequency response? How would Temporal PCA help you characterise the plant dynamics?

---

## References

Seborg, D. E., Edgar, T. F., Mellichamp, D. A., & Doyle, F. J. III. *Process Dynamics and Control* (4th ed.). Wiley. — Standard chemical engineering reference for CSTR modelling and PI control.

Vautard, R., & Ghil, M. (1989). Singular spectrum analysis in nonlinear dynamics, with applications to paleoclimatic time series. *Physica D*, 35(3), 395–424. — Foundational paper for SSA (the mathematical basis of Temporal PCA).

Qin, S. J. (2003). Statistical process monitoring: basics and beyond. *Journal of Chemometrics*, 17(8–9), 480–502. — PCA-based fault detection in process industries.

Kantor, J. C. *CBE30338: Simulation of an Exothermic CSTR*. — Open-source chemical engineering course material used as reference for this simulation.
