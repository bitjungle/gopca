#!/usr/bin/env python3
"""
draw_cstr_diagram.py
Generate a corrected P&ID diagram for the CSTR Temporal PCA tutorial.

Usage:  python draw_cstr_diagram.py
Output: cstr_diagram.png (in the same directory as this script)

Physical model (from simulate_cstr_temporal_pca_dataset.py):
  - Non-isothermal CSTR, V = 100 L
  - Reaction A → B (irreversible, first-order, exothermic)
  - Feed: F, CAf, Tf  →  enters reactor
  - Product: overflows at liquid level  →  exits reactor (same CA, CB as inside)
  - Coolant flows through annular jacket; inlet temperature Tc controlled by PI
  - TIC-201 reads reactor temperature TT-201 and drives TCV-201 on coolant supply
  - No feed-flow feedback control in the model (F is an external disturbance input)
"""

from pathlib import Path

import matplotlib.patches as mpatches
import matplotlib.pyplot as plt

# ── Colour palette ─────────────────────────────────────────────────────────────
COL_PROC = "#1a365d"   # process pipe / vessel border (dark navy)
COL_COOL = "#1d4ed8"   # coolant stream (blue)
COL_SIG  = "#b45309"   # instrument signal lines (amber)
COL_VFIL = "#f0f9ff"   # inner vessel fill (very light blue)
COL_JFIL = "#dbeafe"   # cooling jacket fill (light blue)
COL_INST = "#fefce8"   # transmitter bubble fill (cream)
COL_CTRL = "#f0fdf4"   # controller bubble fill (light green)

LWP = 3.0   # process pipe line width
LWC = 2.8   # coolant pipe line width
LWS = 1.4   # signal line width
R   = 2.3   # instrument bubble radius (data units)


# ── Drawing helpers ─────────────────────────────────────────────────────────────

def pipe(ax, x0, y0, x1, y1, color=COL_PROC, lw=LWP, z=3):
    ax.plot([x0, x1], [y0, y1], color=color, lw=lw,
            solid_capstyle="butt", zorder=z)


def pipe_arrow(ax, x0, y0, x1, y1, color=COL_PROC, lw=LWP):
    """Pipe segment with an arrowhead at (x1, y1)."""
    pipe(ax, x0, y0, x1, y1, color=color, lw=lw)
    ax.annotate(
        "", xy=(x1, y1), xytext=(x0, y0),
        arrowprops=dict(arrowstyle="->", color=color, lw=lw * 0.7,
                        mutation_scale=15),
    )


def signal(ax, x0, y0, x1, y1):
    ax.plot([x0, x1], [y0, y1], color=COL_SIG, lw=LWS, ls="--", zorder=4)


def bubble(ax, cx, cy, top, bot="", ctrl=False):
    """
    Instrument / controller circle (ISA style).
    top = upper tag code, bot = lower tag number (optional).
    ctrl = True → green background (controller).
    """
    fc = COL_CTRL if ctrl else COL_INST
    c = plt.Circle((cx, cy), R, fc=fc, ec="#5b21b6", lw=1.6, zorder=6)
    ax.add_patch(c)
    if bot:
        ax.text(cx, cy + 0.60, top, ha="center", va="center",
                fontsize=6.0, fontweight="bold", zorder=7, color="#1e1b4b")
        ax.plot([cx - R * 0.72, cx + R * 0.72], [cy, cy],
                color="#5b21b6", lw=0.9, zorder=7)
        ax.text(cx, cy - 0.72, bot, ha="center", va="center",
                fontsize=5.5, zorder=7, color="#1e1b4b")
    else:
        ax.text(cx, cy, top, ha="center", va="center",
                fontsize=6.5, fontweight="bold", zorder=7, color="#1e1b4b")


def valve(ax, cx, cy, label="", size=1.5, color=COL_PROC):
    """
    Globe-valve bow-tie symbol with actuator stem above.
    Returns the y-coordinate of the top edge of the actuator circle
    (use for connecting signal lines).
    """
    h        = size * 0.55          # half-height of the triangles
    stem_top = cy + h * 0.85 + size * 0.9
    act_r    = R * 0.62
    act_cy   = stem_top + act_r

    tri_L = plt.Polygon(
        [[cx - h, cy + h * 0.85], [cx - h, cy - h * 0.85], [cx, cy]],
        fc=color, ec=color, zorder=5)
    tri_R = plt.Polygon(
        [[cx + h, cy + h * 0.85], [cx + h, cy - h * 0.85], [cx, cy]],
        fc=color, ec=color, zorder=5)
    ax.add_patch(tri_L)
    ax.add_patch(tri_R)

    ax.plot([cx, cx], [cy + h * 0.85, stem_top], color=color, lw=2.0, zorder=5)
    act = plt.Circle((cx, act_cy), act_r, fc="white", ec=color, lw=1.5, zorder=5)
    ax.add_patch(act)
    ax.text(cx, act_cy, "A", ha="center", va="center",
            fontsize=5, color=color, zorder=6)

    if label:
        ax.text(cx, cy - h * 0.85 - 1.3, label, ha="center", va="top",
                fontsize=7.0, color=color, zorder=6)

    return act_cy + act_r   # top of actuator (signal connection point)


# ── Figure setup ────────────────────────────────────────────────────────────────
fig, ax = plt.subplots(figsize=(24, 15))
ax.set_xlim(0, 240)
ax.set_ylim(0, 150)
ax.set_aspect("equal")
ax.axis("off")
fig.patch.set_facecolor("white")

# ── Title ────────────────────────────────────────────────────────────────────────
ax.text(110, 148.5, "SIMULATED CSTR PROCESS",
        ha="center", va="center", fontsize=20, fontweight="bold", color=COL_PROC)
ax.text(110, 144.8,
        "P&ID with Instrumentation and Control Loops",
        ha="center", va="center", fontsize=12, color="#374151", style="italic")
ax.plot([2, 238], [142.5, 142.5], color=COL_PROC, lw=1.5)

# ── Legend (top right) ──────────────────────────────────────────────────────────
LX, LY = 192, 141
ax.text(LX + 12, LY, "LEGEND", ha="center", fontsize=10,
        fontweight="bold", color=COL_PROC)

ax.plot([LX, LX + 10], [LY - 6,  LY - 6],  color=COL_PROC, lw=LWP)
ax.text(LX + 12, LY - 6,  "Process stream",     va="center", fontsize=8)

ax.plot([LX, LX + 10], [LY - 12, LY - 12], color=COL_COOL, lw=LWC)
ax.text(LX + 12, LY - 12, "Coolant stream",     va="center", fontsize=8)

ax.plot([LX, LX + 10], [LY - 18, LY - 18], color=COL_SIG, lw=LWS, ls="--")
ax.text(LX + 12, LY - 18, "Instrument signal",  va="center", fontsize=8)

b1 = plt.Circle((LX + 2.3, LY - 25), 2.2, fc=COL_INST, ec="#5b21b6", lw=1.5)
ax.add_patch(b1)
ax.text(LX + 2.3, LY - 25, "TT", ha="center", va="center",
        fontsize=7, fontweight="bold")
ax.text(LX + 6, LY - 25, "Field transmitter",  va="center", fontsize=8)

b2 = plt.Circle((LX + 2.3, LY - 32), 2.2, fc=COL_CTRL, ec="#5b21b6", lw=1.5)
ax.add_patch(b2)
ax.text(LX + 2.3, LY - 32, "TIC", ha="center", va="center",
        fontsize=6, fontweight="bold")
ax.text(LX + 6, LY - 32, "Indicating controller", va="center", fontsize=8)

# Mini valve for legend
_h = 1.3 * 0.55
ax.add_patch(plt.Polygon(
    [[LX + 0.9, LY - 39 + _h], [LX + 0.9, LY - 39 - _h], [LX + 2.3, LY - 39]],
    fc=COL_PROC, ec=COL_PROC, zorder=5))
ax.add_patch(plt.Polygon(
    [[LX + 3.7, LY - 39 + _h], [LX + 3.7, LY - 39 - _h], [LX + 2.3, LY - 39]],
    fc=COL_PROC, ec=COL_PROC, zorder=5))
ax.text(LX + 6, LY - 39, "Control valve (actuated)", va="center", fontsize=8)

# ── Disturbances box (top left) ─────────────────────────────────────────────────
dbox = mpatches.FancyBboxPatch(
    (2, 117), 64, 25,
    boxstyle="round,pad=0.5", fc="#fffbeb", ec="#d97706", lw=1.5, zorder=0)
ax.add_patch(dbox)
ax.text(34, 142, "DISTURBANCES", ha="center", fontsize=10,
        fontweight="bold", color="#92400e")
DIST = [
    ("  0 – 120 min", "Normal steady state (baseline)"),
    ("120 – 240 min", "Feed concentration step: C_{A,f} +10 %"),
    ("240 – 360 min", "Feed temperature step: T_f +7 K"),
    ("360 – 520 min", "Feed flow oscillation ±8 %,  T_osc = 40 min"),
    ("520 – 680 min", "Cooling fault: UA drops by 28 %"),
    ("680 – 800 min", "Recovery to nominal conditions"),
]
for i, (t_, d_) in enumerate(DIST):
    yy = 140.0 - i * 3.7
    ax.text( 3.5, yy, f"• {t_}:", fontsize=8, color="#92400e",
            fontweight="bold", va="top")
    ax.text(47,   yy, d_,         fontsize=8, color="#78350f", va="top")

# ──────────────────────────────────────────────────────────────────────────────
# REACTOR GEOMETRY
#
#   Cooling jacket : x ∈ [JL, JR], y ∈ [JB, JT]
#   Inner reactor  : x ∈ [RL, RR], y ∈ [RBot, RTop]
#   Liquid level   : LEVEL_Y  (overflow point = product exit)
# ──────────────────────────────────────────────────────────────────────────────
JL,   JR   = 82,  148   # jacket left / right
JBot, JTop = 38,  122   # jacket bottom / top

RL,   RR   = 88,  142   # reactor inner left / right
RBot, RTop = 44,  117   # reactor inner bottom / top

LEVEL_Y = 106           # liquid level / overflow height (~79 % fill)
CX      = (RL + RR) / 2   # = 115  agitator centre x

# Cooling jacket (drawn first, lowest zorder)
jacket = mpatches.FancyBboxPatch(
    (JL, JBot), JR - JL, JTop - JBot,
    boxstyle="round,pad=0.8",
    fc=COL_JFIL, ec=COL_COOL, lw=2.5, zorder=1)
ax.add_patch(jacket)

# Inner reactor vessel
reactor = mpatches.FancyBboxPatch(
    (RL, RBot), RR - RL, RTop - RBot,
    boxstyle="round,pad=0.5",
    fc=COL_VFIL, ec=COL_PROC, lw=2.2, zorder=2)
ax.add_patch(reactor)

# Reactor tag
ax.text(CX, JTop + 5, "B-101  CSTR", ha="center", fontsize=12,
        fontweight="bold", color=COL_PROC, zorder=3)

# Cooling jacket label — outside the jacket, to the left (rotated)
ax.text(JL - 3.5, (JBot + JTop) / 2, "COOLING  JACKET",
        ha="center", va="center", fontsize=8.5, color=COL_COOL,
        fontweight="bold", style="italic", zorder=3, rotation=90)

# Liquid level dashed line (overflow at liquid level = product exit height)
ax.plot([RL + 1, RR - 1], [LEVEL_Y, LEVEL_Y],
        color="#60a5fa", lw=1.3, ls=":", zorder=4)
ax.text(RR + 1.2, LEVEL_Y, "liquid level / overflow",
        fontsize=6.5, color="#1d4ed8", va="center", zorder=4)

# Agitator
ax.plot([CX, CX], [RBot + 5, RTop - 1],
        color="#4b5563", lw=2.5, zorder=4)
for y_bl in [RBot + 16, RBot + 30]:
    ax.plot([CX - 11, CX + 11], [y_bl, y_bl], color="#4b5563", lw=4.0, zorder=4)

# ──────────────────────────────────────────────────────────────────────────────
# MEASUREMENT INSTRUMENTS INSIDE THE REACTOR
#   TT-201  reactor temperature  (left-centre)
#   AT-201  reactant concentration CA  (left, below TT-201)
#   AT-202  product concentration CB   (left, below AT-201)
# ──────────────────────────────────────────────────────────────────────────────
TT201_X, TT201_Y = 100, 90
bubble(ax, TT201_X, TT201_Y, "TT", "201")
ax.text(TT201_X - R - 1, TT201_Y, "T", ha="right", fontsize=9,
        style="italic", color=COL_PROC, zorder=7)

AT201_X, AT201_Y = 100, 76
bubble(ax, AT201_X, AT201_Y, "AT", "201")
ax.text(AT201_X - R - 1, AT201_Y, r"$C_A$", ha="right", fontsize=8.5,
        color=COL_PROC, zorder=7)

AT202_X, AT202_Y = 100, 62
bubble(ax, AT202_X, AT202_Y, "AT", "202")
ax.text(AT202_X - R - 1, AT202_Y, r"$C_B$", ha="right", fontsize=8.5,
        color=COL_PROC, zorder=7)

# ──────────────────────────────────────────────────────────────────────────────
# FEED SYSTEM
#   Feed tank TK-101 (left, mid-height)
#   Pipe at y = FEED_Y runs from tank → enters LEFT WALL of inner reactor
#   Measurements:  TT-101 (feed temp), AT-101 (feed conc), FT-101 (feed flow)
#   No feedback flow controller in the simulation model
# ──────────────────────────────────────────────────────────────────────────────
FEED_Y = 87    # feed pipe elevation (enters reactor at mid-reactor height)

# Feed tank TK-101
ftank = mpatches.FancyBboxPatch(
    (4, 78), 18, 22,
    boxstyle="round,pad=0.4", fc=COL_VFIL, ec=COL_PROC, lw=2.0)
ax.add_patch(ftank)
ax.text(13, 94, "FEED TANK", ha="center", fontsize=8.5,
        fontweight="bold", color=COL_PROC)
ax.text(13, 91, "TK-101",   ha="center", fontsize=8,   color="#374151")
ax.text(13, 87.8, r"$C_{A,f},\ T_f$", ha="center", fontsize=9,
        color="#14532d")

# Feed pipe:  tank right wall (x=22) → reactor left wall (x=RL=88)
# Instruments on the feed pipe (left to right):
#   TT-101 at x=30, AT-101 at x=42, FT-101 at x=56, then pipe continues to RL
FTANK_RIGHT = 22
pipe(ax, FTANK_RIGHT, FEED_Y, RL, FEED_Y)   # full horizontal pipe

# Arrow just before reactor inlet
ax.annotate("", xy=(RL, FEED_Y), xytext=(RL - 5, FEED_Y),
            arrowprops=dict(arrowstyle="->", color=COL_PROC,
                            lw=LWP * 0.7, mutation_scale=14))

# Feed temperature transmitter TT-101
bubble(ax, 30, FEED_Y + 5, "TT", "101")
signal(ax, 30, FEED_Y + 5 - R, 30, FEED_Y + 0.5)
ax.text(30, FEED_Y + 5 + R + 1, r"$T_f$", ha="center", fontsize=8,
        color="#14532d", zorder=7)

# Feed concentration analyzer AT-101
bubble(ax, 42, FEED_Y + 5, "AT", "101")
signal(ax, 42, FEED_Y + 5 - R, 42, FEED_Y + 0.5)
ax.text(42, FEED_Y + 5 + R + 1, r"$C_{A,f}$", ha="center", fontsize=8,
        color="#14532d", zorder=7)

# Feed flow transmitter FT-101 (ON the pipe)
bubble(ax, 56, FEED_Y, "FT", "101")

# Feed inlet annotation
ax.text(RL - 2, FEED_Y + 2.5,
        r"Feed  ($F,\ C_{A,f},\ T_f$)",
        ha="right", fontsize=8, color="#14532d")

# ──────────────────────────────────────────────────────────────────────────────
# PRODUCT STREAM
#   Exits the right wall of the inner reactor at the liquid level (LEVEL_Y)
#   Passes through jacket annulus and continues right
#   AT-203 measures product concentrations
# ──────────────────────────────────────────────────────────────────────────────
PROD_X_END = 178

# Product pipe:  reactor right wall → through jacket → product header
pipe(ax, RR, LEVEL_Y, JR, LEVEL_Y)             # through jacket annulus
pipe_arrow(ax, JR, LEVEL_Y, PROD_X_END, LEVEL_Y)  # continues right with arrow

# Product concentration analyzer
AT203_X = 160
bubble(ax, AT203_X, LEVEL_Y, "AT", "203")
ax.text(AT203_X, LEVEL_Y - R - 2.5, r"$C_A,\ C_B$",
        ha="center", fontsize=8, color=COL_PROC, zorder=7)

# Product stream label
ax.text(PROD_X_END + 1.5, LEVEL_Y,
        "PRODUCT\nSTREAM",
        ha="left", va="center", fontsize=10, fontweight="bold", color=COL_PROC)

# ──────────────────────────────────────────────────────────────────────────────
# COOLANT SYSTEM
#   Coolant source TK-201 (bottom left)
#   Supply pipe runs at COOL_Y, then turns UP into the jacket BOTTOM
#   TCV-201 on the supply pipe sets the coolant inlet temperature Tc
#   TT-202 on the jacket OUTLET measures Tc,out
#   Coolant outlet: exits jacket TOP → returns to cooling system (right)
# ──────────────────────────────────────────────────────────────────────────────
COOL_Y      = 26      # coolant supply pipe elevation
COOL_IN_X   = 96      # x where coolant pipe rises into jacket bottom
COOL_OUT_X  = 120     # x where coolant pipe exits from jacket top

# Coolant source tank TK-201
ctank = mpatches.FancyBboxPatch(
    (4, 14), 18, 20,
    boxstyle="round,pad=0.4", fc="#eff6ff", ec=COL_COOL, lw=2.0)
ax.add_patch(ctank)
ax.text(13, 30,   "COOLANT",  ha="center", fontsize=8, fontweight="bold", color=COL_COOL)
ax.text(13, 27.2, "SOURCE",   ha="center", fontsize=8, fontweight="bold", color=COL_COOL)
ax.text(13, 24.5, "TK-201",   ha="center", fontsize=7.5, color="#374151")
ax.text(13, 21.8, r"$T_{c,nom} = 300\,\mathrm{K}$",
        ha="center", fontsize=7, color="#374151")
ax.text(13, 19.2, r"$285 \leq T_c \leq 330\,\mathrm{K}$",
        ha="center", fontsize=6.5, color="#374151")

# Coolant supply pipe:  tank right (x=22) → COOL_IN_X (horizontal at COOL_Y)
pipe(ax, 22, COOL_Y, COOL_IN_X, COOL_Y, color=COL_COOL, lw=LWC)
# Turn up into jacket bottom
pipe(ax, COOL_IN_X, COOL_Y, COOL_IN_X, JBot, color=COL_COOL, lw=LWC)
ax.annotate("", xy=(COOL_IN_X, JBot + 0.5), xytext=(COOL_IN_X, JBot - 5),
            arrowprops=dict(arrowstyle="->", color=COL_COOL,
                            lw=LWC * 0.7, mutation_scale=14))

# Coolant inlet label
ax.text(COOL_IN_X, JBot - 2.5, r"$T_c$ in",
        ha="center", fontsize=8, color=COL_COOL, va="top")

# TCV-201 on coolant supply pipe  (at x=60, on the horizontal supply pipe)
TCV_X = 60
tcv_top = valve(ax, TCV_X, COOL_Y, label="TCV-201", size=1.5, color=COL_COOL)

# Coolant outlet: exits jacket TOP at COOL_OUT_X → horizontal → return header
pipe(ax, COOL_OUT_X, JTop, COOL_OUT_X, JTop + 8, color=COL_COOL, lw=LWC)
COOL_RETURN_X_END = 178
pipe_arrow(ax, COOL_OUT_X, JTop + 8, COOL_RETURN_X_END, JTop + 8, color=COL_COOL, lw=LWC)

# Coolant outlet label
ax.text(COOL_RETURN_X_END + 1.5, JTop + 8,
        "COOLANT\nRETURN",
        ha="left", va="center", fontsize=10, fontweight="bold", color=COL_COOL)

# TT-202 coolant outlet temperature transmitter
TT202_X, TT202_Y = 152, JTop + 8
bubble(ax, TT202_X, TT202_Y, "TT", "202")
ax.text(TT202_X, TT202_Y + R + 1.8, r"$T_{c,out}$",
        ha="center", fontsize=8, color=COL_COOL, zorder=7)

# ──────────────────────────────────────────────────────────────────────────────
# CONTROL LOOP  TIC-201
#   TT-201 → TIC-201 :  measured reactor temperature T
#   TIC-201 → TCV-201 :  manipulated variable = coolant supply temperature Tc
#   PI controller:  Tc = Tc_nom − Kc · (e + ∫e dt / τ_I),  Tc clipped [285, 330] K
# ──────────────────────────────────────────────────────────────────────────────
TIC_X, TIC_Y = 188, 78

bubble(ax, TIC_X, TIC_Y, "TIC", "201", ctrl=True)
ax.text(TIC_X, TIC_Y + R + 2.0, "Temperature\nController",
        ha="center", fontsize=8, color="#1e1b4b")
ax.text(TIC_X, TIC_Y - R - 2.0,
        r"SP: $T_{sp} = 365\,\mathrm{K}$" + "\nPI  (Kc=3, τᵢ=8 min)",
        ha="center", fontsize=6.8, color="#374151", va="top")

# Signal:  TT-201 → TIC-201
#   Route: TT-201 (right edge) → east along y=TT201_Y → east to TIC_X
signal(ax, TT201_X + R, TT201_Y, TIC_X - R, TIC_Y)

# Signal:  TIC-201 → TCV-201
#   Route: TIC-201 bottom → south to y=35 (above coolant pipe) →
#           west to TCV_X → south to TCV actuator top
SIG_SOUTH_Y = 35
signal(ax, TIC_X, TIC_Y - R, TIC_X, SIG_SOUTH_Y)
signal(ax, TIC_X, SIG_SOUTH_Y, TCV_X, SIG_SOUTH_Y)
signal(ax, TCV_X, SIG_SOUTH_Y, TCV_X, tcv_top)

# ── Measured variables summary (bottom left) ─────────────────────────────────
ax.plot([2, 238], [14, 14], color="#6b7280", lw=1.0)
meas_box = mpatches.FancyBboxPatch(
    (2, 1), 113, 13,
    boxstyle="round,pad=0.4", fc="#f9fafb", ec="#9ca3af", lw=1.2)
ax.add_patch(meas_box)
ax.text(58.5, 14.5, "MEASURED VARIABLES — 12 PCA inputs", ha="center",
        fontsize=8.5, fontweight="bold", color=COL_PROC)
MEAS = [
    ("T_K",                       "T",         "Reactor temperature [K]"),
    ("Tc_out_K",                   "T_{c,out}", "Coolant outlet temperature [K]"),
    ("Tf_K",                       "T_f",       "Feed temperature [K]"),
    ("CA_mol_L",                   "C_A",       "Reactant concentration in reactor [mol/L]"),
    ("CB_mol_L",                   "C_B",       "Product concentration in reactor [mol/L]"),
    ("CAf_mol_L",                  "C_{A,f}",   "Feed reactant concentration [mol/L]"),
    ("F_L_min",                    "F",         "Feed flow rate [L/min]"),
    ("cooling_duty_kJ_min",        "Q",         "Cooling duty  UA·(T−T_c) [kJ/min]"),
    ("reaction_rate_mol_L_min",    "r_A",       "Reaction rate  k(T)·C_A [mol/(L·min)]"),
    ("conversion_fraction",        "X_A",       "Fractional conversion (C_{A,f}−C_A)/C_{A,f}"),
    ("heat_transfer_UA_kJ_min_K",  "UA",        "Heat-transfer coefficient [kJ/(min·K)]"),
    ("residence_time_min",         "τ",         "Hydraulic residence time  V/F [min]"),
]
half = len(MEAS) // 2
for i, (col_name, sym, desc) in enumerate(MEAS):
    col_x = 3   if i < half else 56
    row_y = 12.5 - (i % half) * 1.95
    ax.text(col_x,      row_y, f"• {col_name}",
            fontsize=6.0, color="#1a365d", va="top", family="monospace")
    ax.text(col_x + 22, row_y, f"({sym})",
            fontsize=6.5, color="#374151", va="top")
    ax.text(col_x + 32, row_y, desc,
            fontsize=6.0, color="#6b7280", va="top")

# ── Typical nominal conditions (bottom right) ────────────────────────────────
cond_box = mpatches.FancyBboxPatch(
    (118, 1), 120, 13,
    boxstyle="round,pad=0.4", fc="#f9fafb", ec="#9ca3af", lw=1.2)
ax.add_patch(cond_box)
ax.text(178, 14.5, "TYPICAL NOMINAL CONDITIONS", ha="center",
        fontsize=8.5, fontweight="bold", color=COL_PROC)
CONDS = [
    ("Reactor volume",         "V = 100 L"),
    ("Feed flow rate",         "F = 100 L/min"),
    ("Residence time",         "τ = V/F = 1 min"),
    ("Feed concentration",     "C_{A,f} = 1.0 mol/L"),
    ("Feed temperature",       "T_f = 350 K"),
    ("Reactor temperature",    "T = 365 K  (setpoint)"),
    ("Coolant temperature",    "T_c = 300 K  (nominal)"),
    ("Heat-transfer coeff.",   "UA = 50 kJ/(min·K)"),
    ("Rate const. at 365 K",   "k ≈ 2.74 min⁻¹"),
    ("Steady-state C_A",       "≈ 0.267 mol/L"),
    ("Steady-state C_B",       "≈ 0.733 mol/L"),
    ("Fractional conversion",  "X_A ≈ 0.73"),
]
for i, (name, val) in enumerate(CONDS):
    col_x = 120 if i < 6 else 178
    row_y = 12.5 - (i % 6) * 1.95
    ax.text(col_x,      row_y, name + ":", fontsize=6.8, color="#374151", va="top")
    ax.text(col_x + 55, row_y, val,        fontsize=6.8, color="#374151",
            ha="right", va="top")

# ── Save ─────────────────────────────────────────────────────────────────────
out = Path(__file__).parent / "cstr_diagram.png"
plt.savefig(out, dpi=150, bbox_inches="tight", facecolor="white")
print(f"Saved → {out}")
print(f"Size:  {out.stat().st_size // 1024} KB")
