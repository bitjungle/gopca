#!/usr/bin/env python3
"""
simulate_cstr_temporal_pca_dataset.py

Generate a teaching-oriented dynamic CSTR dataset for PCA / Temporal PCA (DPCA).

Model:
    Non-isothermal continuous stirred-tank reactor (CSTR)
    Irreversible first-order exothermic reaction: A → B

Balances:
    dC_A/dt = (F/V) * (C_Af - C_A) - k(T) * C_A

    dC_B/dt = (F/V) * (0 - C_B) + k(T) * C_A        (no B in feed)

    dT/dt   = (F/V) * (T_f - T)
              + (-ΔH / (ρ · Cp)) · k(T) · C_A
              - (UA / (ρ · Cp · V)) · (T - T_c)

    k(T) = k₀ · exp(-E/R / T)                        (Arrhenius)

Coolant temperature controller:
    A PI controller manipulates T_c to keep reactor temperature T near setpoint.
    Anti-windup logic prevents integrator wind-up at actuator limits.

Simulated operating scenarios (designed for Temporal PCA teaching):
    0  – 120 min   normal         Reactor starts and reaches steady state.
                                  Baseline period — establish PCA model here.
    120 – 240 min  feed_conc_high Feed concentration step +10 %.
                                  Disturbance in C_A and T; controller corrects T.
    240 – 360 min  feed_temp_high Feed temperature step +7 K.
                                  Thermal disturbance; controller compensates.
    360 – 520 min  flow_osc       Feed flow oscillates ±8 % at 40-min period.
                                  Creates periodic variation → SSA oscillatory pairs.
    520 – 680 min  cooling_fault  Heat-transfer coefficient drops 28 %.
                                  T drifts; controller saturates; fault detectable
                                  in cooling-duty and temperature variables.
    680 – 800 min  recovery       Return to nominal conditions.

Suggested GoPCA Temporal PCA settings:
    Target column : regime   (for coloring scores by operating state)
    Preprocessing : Standard Scaling
    Lags          : start with L = 10 (10 min ≈ 1.25 × τ_I)
                    try L = 5 (≈ 5 residence times) and L = 20 for comparison
                    Note: the 40-min flow oscillation is cleanly resolved at L = 40 (one full
                    window per period; verified sine/cosine pair, R^2 = 1.00);
                    at L = 10–20 it will appear as low-frequency paired components
    Components    : 5–10 to see both slow trends and the oscillatory pair

Process time constants (for lag selection):
    Residence time τ = V/F = 1 min  (nominal; varies during flow oscillation)
    PI integral time τ_I = 8 min    → L = 8–16 captures one controller action
    Flow oscillation period = 40 min → L = 40 resolves it (L up to 80 also works)

CSV format and column names (matching P&ID diagram — cstr_diagram_v2.png):
    time_min     — observation identifier / time axis (not a PCA variable)

    PCA variables (12 sensor / derived measurements):
        T_K           T       Reactor temperature                  [K]
        Tc_out_K      T_c,out Coolant outlet temperature           [K]
        Tf_K          T_f     Feed temperature                     [K]
        CA_mol_L      C_A     Reactant concentration in reactor    [mol/L]
        CB_mol_L      C_B     Product concentration in reactor     [mol/L]
        CAf_mol_L     C_A,f   Feed reactant concentration          [mol/L]
        F_L_min       F       Feed flow rate                       [L/min]
        cooling_duty_kJ_min   Q   Cooling duty                     [kJ/min]
        reaction_rate_mol_L_min   r_A  Reaction rate               [mol/L/min]
        conversion_fraction   X_A  Fractional conversion of A      [—]
        heat_transfer_UA_kJ_min_K UA  Heat-transfer coefficient    [kJ/(min·K)]
        residence_time_min    tau  Hydraulic residence time        [min]

    String columns (auto-detected by GoPCA, use for coloring): event, regime
    String column : fault_active ("no" = normal/recovery, "yes" = disturbance/fault)
                   String values cause GoPCA to treat this as categorical,
                   excluding it from PCA automatically.

Physical correctness notes:
    - C_B is integrated as an ODE state, NOT derived from C_Af − C_A.
      The latter is only valid at steady state; during transients C_B has
      its own dynamics driven by the residence-time distribution.
    - At steady state with F = 100 L/min, V = 100 L, T = 365 K:
      k ≈ 2.74 min⁻¹, C_A ≈ 0.267 mol/L, C_B ≈ 0.733 mol/L, T_c ≈ 299 K.
      The simulation starts at C_A = 0.27, C_B = 0.73, T = 365 K, I = 0,
      which is close to steady state so the baseline period is clean.

References:
    Seborg, D. E., Edgar, T. F., Mellichamp, D. A., & Doyle, F. J. III.
    Process Dynamics and Control. Wiley.

    Kantor, J. C. CBE30338: Simulation of an Exothermic CSTR.
    https://jckantor.github.io/CBE30338/07.04-Simulation-of-an-Exothermic-CSTR.html

    MathWorks. CSTR Model.
    https://www.mathworks.com/help/mpc/gs/cstr-model.html

License:
    You may use, modify, and redistribute the generated dataset and this script
    as part of your teaching material. If publishing externally, cite the model
    references above and clearly state that the data are simulated.
"""

from __future__ import annotations

from dataclasses import dataclass
from pathlib import Path
import argparse
import numpy as np
import pandas as pd
from scipy.integrate import solve_ivp


@dataclass(frozen=True)
class CSTRParameters:
    """Physical and kinetic parameters for the CSTR model."""

    # Reactor and fluid
    V: float = 100.0              # L
    rho: float = 1000.0           # g/L
    Cp: float = 0.239             # J/(g·K)

    # Heat transfer
    UA: float = 5.0e4             # J/(min·K)

    # Reaction A → B (exothermic)
    deltaH: float = -5.0e4        # J/mol  (negative = exothermic)
    k0: float = 7.2e10            # 1/min
    E_over_R: float = 8750.0      # K  (= Ea/R)

    # Nominal operating conditions
    F_nom: float = 100.0          # L/min
    CAf_nom: float = 1.0          # mol/L
    Tf_nom: float = 350.0         # K
    T_sp: float = 365.0           # K  (temperature setpoint)

    # Coolant actuator limits
    Tc_min: float = 285.0         # K
    Tc_max: float = 330.0         # K
    Tc_nom: float = 300.0         # K  (bias / nominal coolant temperature)

    # PI controller: positive T error → reduce Tc (more cooling)
    Kc: float = 3.0               # K coolant / K reactor error
    tau_I: float = 8.0            # min


def reaction_rate_constant(T: float | np.ndarray, p: CSTRParameters) -> float | np.ndarray:
    """Arrhenius first-order rate constant k(T) = k₀ · exp(−E/R / T)."""
    return p.k0 * np.exp(-p.E_over_R / T)


def scenario_inputs(t: float, p: CSTRParameters) -> dict[str, float | str]:
    """
    Return time-varying inputs for the given simulation time t [min].

    Each scenario is designed to produce a distinct and physically interpretable
    multivariate signature for Temporal PCA teaching.
    """
    F: float = p.F_nom
    CAf: float = p.CAf_nom
    Tf: float = p.Tf_nom
    UA_factor: float = 1.0
    event: str = "normal"

    if 120.0 <= t < 240.0:
        CAf = 1.10
        event = "feed_conc_high"
    elif 240.0 <= t < 360.0:
        Tf = 357.0
        event = "feed_temp_high"
    elif 360.0 <= t < 520.0:
        F = p.F_nom * (1.0 + 0.08 * float(np.sin(2.0 * np.pi * (t - 360.0) / 40.0)))
        event = "flow_oscillation"
    elif 520.0 <= t < 680.0:
        UA_factor = 0.72
        event = "cooling_fault"
    elif 680.0 <= t < 700.0:
        # Ramp UA back to nominal over 20 min — avoids the one-sample outlier
        # that an instantaneous step would create (reactor still at fault
        # temperature when UA jumps, producing an extreme cooling_duty spike).
        UA_factor = 0.72 + 0.28 * (t - 680.0) / 20.0
        event = "recovery"
    elif t >= 700.0:
        event = "recovery"

    return {"F": F, "CAf": CAf, "Tf": Tf, "UA_factor": UA_factor, "event": event}


def cstr_rhs(t: float, y: np.ndarray, p: CSTRParameters) -> np.ndarray:
    """
    Right-hand side of the augmented CSTR model.

    ODE states:
        y[0] = C_A   reactant concentration in reactor [mol/L]
        y[1] = C_B   product  concentration in reactor [mol/L]
        y[2] = T     reactor temperature [K]
        y[3] = I     PI controller integral state [K·min]

    Note: C_B is an independent ODE state.  Using C_B = C_Af − C_A is only
    valid at steady state; during transients C_B has its own residence-time
    driven dynamics and must be integrated separately.
    """
    CA, CB, T, I = y
    u = scenario_inputs(t, p)

    F       = float(u["F"])
    CAf     = float(u["CAf"])
    Tf      = float(u["Tf"])
    UA_eff  = p.UA * float(u["UA_factor"])

    # PI controller (positive error = reactor too hot → cool more → lower Tc)
    error = T - p.T_sp
    Tc_unsat = p.Tc_nom - p.Kc * (error + I / p.tau_I)
    Tc = float(np.clip(Tc_unsat, p.Tc_min, p.Tc_max))

    k  = reaction_rate_constant(T, p)
    rA = k * CA                         # reaction rate [mol/L/min]

    dCA_dt = (F / p.V) * (CAf - CA) - rA
    dCB_dt = (F / p.V) * (0.0 - CB) + rA   # no B in feed

    dT_dt = (
        (F / p.V) * (Tf - T)
        + (-p.deltaH / (p.rho * p.Cp)) * rA
        - (UA_eff / (p.rho * p.Cp * p.V)) * (T - Tc)
    )

    # Anti-windup: freeze integrator when actuator is saturated in the
    # direction that would worsen wind-up.
    saturated_low  = bool(Tc_unsat < p.Tc_min) and bool(error > 0)
    saturated_high = bool(Tc_unsat > p.Tc_max) and bool(error < 0)
    dI_dt = 0.0 if (saturated_low or saturated_high) else error

    return np.array([dCA_dt, dCB_dt, dT_dt, dI_dt])


def simulate(
    duration_min: float = 800.0,
    sample_time_min: float = 1.0,
    seed: int = 42,
    noise_level: float = 1.0,
) -> pd.DataFrame:
    """
    Simulate the CSTR and return a GoPCA-ready DataFrame.

    Column layout:
        time_min         — observation identifier / time axis (not a PCA variable)
        <numeric vars>   — process measurements with realistic sensor noise
        event            — fine-grained operating scenario label (string)
        regime           — coarse regime for GoPCA coloring (string)
        fault_active     — "yes" during disturbance/fault periods, "no" during normal/recovery
    """
    rng = np.random.default_rng(seed)
    p   = CSTRParameters()

    t_eval = np.arange(0.0, duration_min + sample_time_min, sample_time_min)

    # Initial condition: close to nominal steady state
    # At T=365 K: k ≈ 2.74 min⁻¹ → CA_ss ≈ 0.267, CB_ss ≈ 0.733 mol/L
    y0 = np.array([0.27, 0.73, 365.0, 0.0])

    sol = solve_ivp(
        fun=lambda t, y: cstr_rhs(t, y, p),
        t_span=(t_eval[0], t_eval[-1]),
        y0=y0,
        t_eval=t_eval,
        method="LSODA",
        rtol=1e-8,
        atol=1e-10,
    )

    if not sol.success:
        raise RuntimeError(f"ODE integration failed: {sol.message}")

    CA   = sol.y[0]
    CB   = sol.y[1]
    T    = sol.y[2]
    I    = sol.y[3]
    time = sol.t

    rows: list[dict] = []

    for i, t in enumerate(time):
        u = scenario_inputs(float(t), p)

        F        = float(u["F"])
        CAf      = float(u["CAf"])
        Tf       = float(u["Tf"])
        UA_eff   = p.UA * float(u["UA_factor"])
        event    = str(u["event"])

        k        = float(reaction_rate_constant(T[i], p))
        rA       = k * CA[i]
        conversion = (CAf - CA[i]) / CAf

        error    = T[i] - p.T_sp
        Tc_unsat = p.Tc_nom - p.Kc * (error + I[i] / p.tau_I)
        Tc       = float(np.clip(Tc_unsat, p.Tc_min, p.Tc_max))

        cooling_duty = UA_eff * (T[i] - Tc) / 1000.0          # kJ/min
        residence_time = p.V / F                                # min

        if event in {"normal", "recovery"}:
            regime = "normal"
        elif event == "cooling_fault":
            regime = "cooling_fault"
        elif event == "flow_oscillation":
            regime = "oscillation"
        else:
            regime = "feed_disturbance"

        n = noise_level
        rows.append({
            "time_min":                      round(float(t), 4),
            "T_K":                           T[i]    + rng.normal(0, 0.08   * n),
            "Tc_out_K":                      Tc      + rng.normal(0, 0.05   * n),
            "Tf_K":                          Tf      + rng.normal(0, 0.05   * n),
            "CA_mol_L":                      CA[i]   + rng.normal(0, 0.0015 * n),
            "CB_mol_L":                      CB[i]   + rng.normal(0, 0.0015 * n),
            "CAf_mol_L":                     CAf     + rng.normal(0, 0.0010 * n),
            "F_L_min":                       F       + rng.normal(0, 0.15   * n),
            "cooling_duty_kJ_min":           cooling_duty + rng.normal(0, 0.25 * n),
            "reaction_rate_mol_L_min":       rA      + rng.normal(0, 0.0004 * n),
            "conversion_fraction":           conversion + rng.normal(0, 0.001 * n),
            "heat_transfer_UA_kJ_min_K":     UA_eff / 1000.0 + rng.normal(0, 0.05 * n),
            "residence_time_min":            residence_time + rng.normal(0, 0.001 * n),
            "event":                         event,
            "regime":                        regime,
            "fault_active":                  "yes" if regime != "normal" else "no",
        })

    df = pd.DataFrame(rows)

    # Place string/categorical columns last so GoPCA numeric detection works cleanly
    target_cols  = ["event", "regime", "fault_active"]
    feature_cols = [c for c in df.columns if c not in target_cols]
    return df[feature_cols + target_cols]


def main() -> None:
    parser = argparse.ArgumentParser(
        description=(
            "Generate a simulated non-isothermal CSTR dataset for PCA / Temporal PCA teaching."
        )
    )
    parser.add_argument("--output",       type=Path,  default=Path("cstr_temporal_pca.csv"))
    parser.add_argument("--duration",     type=float, default=800.0,
                        help="Simulation duration [min] (default: 800)")
    parser.add_argument("--sample-time",  type=float, default=1.0,
                        help="Sampling interval [min] (default: 1.0)")
    parser.add_argument("--seed",         type=int,   default=42)
    parser.add_argument("--noise-level",  type=float, default=1.0,
                        help="Noise scale factor: 0 = noiseless, 1 = realistic (default: 1.0)")
    args = parser.parse_args()

    df = simulate(
        duration_min=args.duration,
        sample_time_min=args.sample_time,
        seed=args.seed,
        noise_level=args.noise_level,
    )

    df.to_csv(args.output, index=False)

    print(f"Wrote {args.output}")
    print(f"Rows    : {len(df)}")
    print(f"Columns : {len(df.columns)}")
    print()
    print("Suggested GoPCA Temporal PCA configuration:")
    print("  Time column   : time_min")
    print("  Target column : regime    (for coloring)")
    print("  Preprocessing : Standard Scaling")
    print("  Lags (L)      : start with 10, compare 5 and 20")
    print("                  residence time τ = V/F = 1 min (nominal)")
    print("                  PI integral time τ_I = 8 min")
    print("                  flow oscillation period = 40 min")
    print("  Components    : 5–10")
    print()
    print("Event schedule:")
    print("  0   – 120 min  normal             (steady-state baseline)")
    print("  120 – 240 min  feed_conc_high     (+10 % feed concentration step)")
    print("  240 – 360 min  feed_temp_high     (+7 K feed temperature step)")
    print("  360 – 520 min  flow_oscillation   (±8 % flow, 40-min period)")
    print("  520 – 680 min  cooling_fault      (−28 % heat transfer coefficient)")
    print("  680 – 800 min  recovery           (return to nominal)")


if __name__ == "__main__":
    main()
