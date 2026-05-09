#!/usr/bin/env python3
"""
draw_cstr_diagram.py  —  v5
Accurate, visually clear P&ID for the CSTR Temporal PCA tutorial.

Usage:  python draw_cstr_diagram.py
Output: cstr_diagram.png (same directory)

Layout philosophy
-----------------
* Feed enters the LEFT side of the reactor at ~83 % height (upper inlet nozzle).
* Product exits the RIGHT side at ~18 % height (lower drain nozzle).
* Motor sits on top of the agitator shaft above the vessel dome.
* Coolant enters jacket bottom-left, exits jacket top-right.
* TIC-201 (PI controller) reads TT-201 and drives TCV-201 on the coolant supply.
* Signal routing is kept away from process streams.
* TCV-201 is a 3-way mixing valve: cold supply (TK-201) + warm bypass (return
  header) → mixed stream at controlled temperature Tc [K] → jacket inlet.
  This is physically consistent with the simulator, which sets Tc directly.
"""

from __future__ import annotations

from pathlib import Path

import numpy as np
import matplotlib.patches as mpatches
import matplotlib.pyplot as plt
from matplotlib.path import Path as MPath
from matplotlib.patches import PathPatch

# ─────────────────────────────────────────────────────────────────────────────
# Palette
# ─────────────────────────────────────────────────────────────────────────────
PROC   = "#1a365d"   # navy  — process lines / vessel
COOL   = "#1e40af"   # blue  — coolant
SIG    = "#c2410c"   # amber — instrument signal (dashed)
JFILL  = "#bfdbfe"   # light blue — jacket annulus fill
VFILL  = "#e8f4fd"   # very light — reactor liquid
INST   = "#fefce8"   # cream — transmitter bubble
CTRL   = "#f0fdf4"   # light green — controller bubble
FEED_C = "#166534"   # green — feed labels
PROD_C = "#6d28d9"   # purple — product labels
GRAY   = "#374151"
LG     = "#6b7280"
STEEL  = "#64748b"   # agitator / shaft

LWP = 3.2
LWC = 3.0
LWS = 1.5
RB  = 2.3   # instrument bubble radius

# ─────────────────────────────────────────────────────────────────────────────
# Drawing primitives
# ─────────────────────────────────────────────────────────────────────────────

def pipe(ax, x0, y0, x1, y1, color=PROC, lw=LWP, z=3):
    ax.plot([x0, x1], [y0, y1], color=color, lw=lw,
            solid_capstyle="butt", zorder=z)


def arrow_end(ax, x0, y0, x1, y1, color=PROC, lw=LWP):
    """Pipe with arrowhead at destination."""
    pipe(ax, x0, y0, x1, y1, color=color, lw=lw)
    dx, dy = x1 - x0, y1 - y0
    length = (dx**2 + dy**2) ** 0.5
    frac = max(0.0, 1.0 - 14.0 / length) if length > 0 else 0.0
    mx, my = x0 + frac * dx, y0 + frac * dy
    ax.annotate("", xy=(x1, y1), xytext=(mx, my),
                arrowprops=dict(arrowstyle="->", color=color,
                                lw=lw * 0.65, mutation_scale=14))


def sig(ax, x0, y0, x1, y1, z=5):
    ax.plot([x0, x1], [y0, y1], color=SIG, lw=LWS, ls="--", zorder=z)


def bubble(ax, cx, cy, top, bot="", ctrl=False, z=7):
    fc = CTRL if ctrl else INST
    ax.add_patch(plt.Circle((cx, cy), RB, fc=fc, ec="#5b21b6", lw=1.5, zorder=z))
    if bot:
        ax.text(cx, cy + 0.58, top, ha="center", va="center",
                fontsize=5.8, fontweight="bold", zorder=z + 1, color="#1e1b4b")
        ax.plot([cx - RB * 0.72, cx + RB * 0.72], [cy, cy],
                color="#5b21b6", lw=0.85, zorder=z + 1)
        ax.text(cx, cy - 0.68, bot, ha="center", va="center",
                fontsize=5.4, zorder=z + 1, color="#1e1b4b")
    else:
        ax.text(cx, cy, top, ha="center", va="center",
                fontsize=6.2, fontweight="bold", zorder=z + 1, color="#1e1b4b")


def valve(ax, cx, cy, label="", color=PROC, size=1.6):
    """Globe valve bow-tie + stem + actuator circle. Returns top-of-actuator y."""
    h = size * 0.55
    stem_top = cy + h * 0.85 + size * 0.95
    act_r = RB * 0.65
    act_cy = stem_top + act_r
    ax.add_patch(plt.Polygon(
        [[cx - h, cy + h * 0.85], [cx - h, cy - h * 0.85], [cx, cy]],
        fc=color, ec=color, zorder=5))
    ax.add_patch(plt.Polygon(
        [[cx + h, cy + h * 0.85], [cx + h, cy - h * 0.85], [cx, cy]],
        fc=color, ec=color, zorder=5))
    ax.plot([cx, cx], [cy + h * 0.85, stem_top], color=color, lw=2.0, zorder=5)
    ax.add_patch(plt.Circle((cx, act_cy), act_r,
                             fc="white", ec=color, lw=1.4, zorder=5))
    ax.text(cx, act_cy, "A", ha="center", va="center",
            fontsize=4.8, color=color, zorder=6)
    if label:
        ax.text(cx, cy - h * 0.85 - 1.3, label, ha="center", va="top",
                fontsize=7.0, color=color)
    return act_cy + act_r


def three_way_valve(ax, cx, cy, color=PROC, size=1.8):
    """3-way mixing valve symbol — actuator on BOTTOM, bypass enters from TOP.

    This ensures the bypass pipe (top) and the actuator (bottom) are on
    opposite sides, so they cannot be visually confused with each other.
    The instrument signal line connects to the actuator from below.

    Stream connections:
      Left  inlet : (cx - h, cy)         — cold supply
      Top   inlet : (cx, cy + h*0.85)   — warm bypass (enters from above)
      Right outlet: (cx + h, cy)         — mixed stream to jacket

    Returns: (actuator_bottom_y, h)
    """
    h = size * 0.55

    # Left inlet triangle (cold supply → centre)
    ax.add_patch(plt.Polygon(
        [[cx - h, cy + h * 0.85], [cx - h, cy - h * 0.85], [cx, cy]],
        fc=color, ec=color, zorder=5))

    # Right outlet triangle (centre → mixed stream)
    ax.add_patch(plt.Polygon(
        [[cx, cy], [cx + h, cy + h * 0.85], [cx + h, cy - h * 0.85]],
        fc=color, ec=color, zorder=5))

    # Top bypass triangle (warm bypass → centre, points DOWN into body)
    ax.add_patch(plt.Polygon(
        [[cx - h * 0.85, cy + h], [cx + h * 0.85, cy + h], [cx, cy]],
        fc=color, ec=color, zorder=5))

    # Actuator on BOTTOM — stem exits bottom of valve body, circle below
    act_r = RB * 0.65
    stem_bot = cy - h * 0.85 - 1.4
    act_cy   = stem_bot - act_r
    ax.plot([cx, cx], [cy - h * 0.85, stem_bot], color=color, lw=2.0, zorder=5)
    ax.add_patch(plt.Circle((cx, act_cy), act_r,
                             fc="white", ec=color, lw=1.4, zorder=5))
    ax.text(cx, act_cy, "A", ha="center", va="center",
            fontsize=4.8, color=color, zorder=6)

    return act_cy - act_r, h  # (actuator_bottom_y, half_size_h)


def vessel_path(cx, cy_bot, cy_top, hw, ht=0.22, hb=0.32):
    """Closed path: cylindrical vessel with elliptic dome top and rounded bottom."""
    a, bt, bb = hw, hw * ht, hw * hb
    th_b = np.linspace(np.pi, 2 * np.pi, 48)
    xb = cx + a * np.cos(th_b)
    yb = cy_bot - bb + bb * (1 + np.sin(th_b))
    th_t = np.linspace(0, np.pi, 48)
    xt = cx + a * np.cos(th_t)
    yt = cy_top + bt * np.sin(th_t)
    xs = np.concatenate([xb, [cx + a, cx + a], xt[::-1], [cx - a, cx - a]])
    ys = np.concatenate([yb, [cy_bot, cy_top], yt[::-1], [cy_top, cy_bot]])
    codes = [MPath.MOVETO] + [MPath.LINETO] * (len(xs) - 2) + [MPath.CLOSEPOLY]
    return MPath(np.column_stack([xs, ys]), codes)


def liquid_path(cx, cy_bot, cy_liq, hw, hb=0.32):
    """Liquid fill: same rounded bottom as vessel, flat top at cy_liq."""
    a, bb = hw, hw * hb
    th_b = np.linspace(np.pi, 2 * np.pi, 48)
    xb = cx + a * np.cos(th_b)
    yb = cy_bot - bb + bb * (1 + np.sin(th_b))
    xs = np.concatenate([xb, [cx + a, cx + a, cx - a, cx - a]])
    ys = np.concatenate([yb, [cy_bot, cy_liq, cy_liq, cy_bot]])
    codes = [MPath.MOVETO] + [MPath.LINETO] * (len(xs) - 2) + [MPath.CLOSEPOLY]
    return MPath(np.column_stack([xs, ys]), codes)


# ─────────────────────────────────────────────────────────────────────────────
# Figure
# ─────────────────────────────────────────────────────────────────────────────
fig, ax = plt.subplots(figsize=(26, 16))
ax.set_xlim(0, 260)
ax.set_ylim(0, 160)
ax.set_aspect("equal")
ax.axis("off")
fig.patch.set_facecolor("white")

# ── Title ────────────────────────────────────────────────────────────────────
ax.text(130, 158.5, "SIMULATED CSTR PROCESS",
        ha="center", fontsize=20, fontweight="bold", color=PROC)
ax.text(130, 154.8,
        "P&ID with Instrumentation and Temperature Control Loop",
        ha="center", fontsize=11, color=GRAY, style="italic")
ax.plot([2, 258], [152.5, 152.5], color=PROC, lw=1.3)

# ─────────────────────────────────────────────────────────────────────────────
# Reactor geometry
# ─────────────────────────────────────────────────────────────────────────────
CX = 128          # vessel / shaft centre x

JHW  = 39         # jacket half-width  → JL=89, JR=167
JBot = 36         # jacket bottom y
JTop = 126        # jacket top y
JL, JR = CX - JHW, CX + JHW

VHW  = 30         # vessel half-width  → VL=98, VR=158
VBot = 44         # vessel bottom y
VTop = 120        # vessel top y
VL,  VR  = CX - VHW, CX + VHW

VLIQ = VBot + int((VTop - VBot) * 0.76)   # liquid level ≈ 76 % fill

# Feed inlet y (upper-left wall, ~83 % height from bottom)
FEED_Y = VBot + int((VTop - VBot) * 0.83)   # ≈ 105

# Product outlet y (lower-right wall, ~18 % height from bottom)
PROD_Y = VBot + int((VTop - VBot) * 0.18)   # ≈ 57

# Coolant inlet x — must be in LEFT jacket annulus (JL=89 to VL=98)
COOL_IN_X  = JL + 5    # = 94  (inside the annular jacket gap, left side)
# Coolant outlet x — must be in RIGHT jacket annulus (VR=158 to JR=167)
COOL_OUT_X = JR - 5    # = 162 (inside the annular jacket gap, right side)

# ── Draw jacket fill ─────────────────────────────────────────────────────────
ax.add_patch(PathPatch(
    vessel_path(CX, JBot, JTop, JHW, ht=0.18, hb=0.28),
    fc=JFILL, ec=COOL, lw=2.8, zorder=1))

# ── Liquid fill ───────────────────────────────────────────────────────────────
ax.add_patch(PathPatch(
    liquid_path(CX, VBot, VLIQ, VHW, hb=0.32),
    fc=VFILL, ec="none", zorder=2))

# ── Wavy liquid surface ───────────────────────────────────────────────────────
xw = np.linspace(VL + 2, VR - 2, 100)
yw = VLIQ + 0.55 * np.sin(np.linspace(0, 5 * np.pi, 100))
ax.plot(xw, yw, color="#60a5fa", lw=1.4, zorder=3)

# ── Vessel outline ────────────────────────────────────────────────────────────
ax.add_patch(PathPatch(
    vessel_path(CX, VBot, VTop, VHW, ht=0.22, hb=0.32),
    fc="none", ec=PROC, lw=2.6, zorder=4))

# Vessel tag
ax.text(CX, JTop + 5.5, "B-101   CSTR", ha="center", fontsize=13,
        fontweight="bold", color=PROC, zorder=5)
ax.text(CX, JTop + 2.0, r"$V = 100\,\mathrm{L}$,   $T_{sp} = 365\,\mathrm{K}$",
        ha="center", fontsize=8.5, color=GRAY, zorder=5)

# Jacket label (outside, left)
ax.text(JL - 4.5, (JBot + JTop) / 2, "COOLING  JACKET",
        ha="center", va="center", fontsize=8.5, color=COOL,
        fontweight="bold", style="italic", rotation=90, zorder=3)

# ── Agitator ─────────────────────────────────────────────────────────────────
SHAFT_TOP = VTop + 2
SHAFT_BOT = VBot + 6
ax.plot([CX, CX], [SHAFT_BOT, SHAFT_TOP], color=STEEL, lw=3.0, zorder=5)

IMP_Y    = VBot + int((VTop - VBot) * 0.30)
BLADE_W  = 14
BLADE_H  = 4.5
for xstart in [CX - BLADE_W, CX]:
    ax.add_patch(mpatches.FancyBboxPatch(
        (xstart, IMP_Y - BLADE_H / 2), BLADE_W, BLADE_H,
        boxstyle="round,pad=0.5", fc="#94a3b8", ec=STEEL, lw=1.3, zorder=5))

# Circulation arrows (indicative)
for sign, xc in [(-1, CX - 14), (+1, CX + 14)]:
    ax.annotate("", xy=(xc, VLIQ - 5), xytext=(xc, IMP_Y + 8),
                arrowprops=dict(arrowstyle="->,head_width=0.35",
                                color="#93c5fd", lw=0.9,
                                connectionstyle=f"arc3,rad={sign * 0.55}"),
                zorder=3)

# ── Motor ─────────────────────────────────────────────────────────────────────
MOT_W, MOT_H = 14, 8
MOT_BOT = SHAFT_TOP + 2
MOT_TOP = MOT_BOT + MOT_H

# Shaft stub above vessel dome
ax.plot([CX, CX], [SHAFT_TOP, MOT_BOT], color=STEEL, lw=3.0, zorder=5)

# Motor box
ax.add_patch(mpatches.FancyBboxPatch(
    (CX - MOT_W / 2, MOT_BOT), MOT_W, MOT_H,
    boxstyle="round,pad=0.6", fc="#e2e8f0", ec=PROC, lw=2.0, zorder=6))
ax.text(CX, MOT_BOT + MOT_H / 2, "M", ha="center", va="center",
        fontsize=11, fontweight="bold", color=PROC, zorder=7)
# Rotation arrow
ax.annotate("", xy=(CX + 2.2, MOT_BOT + MOT_H * 0.73),
            xytext=(CX - 2.2, MOT_BOT + MOT_H * 0.27),
            arrowprops=dict(arrowstyle="->,head_width=0.4", color=GRAY,
                            lw=0.85, connectionstyle="arc3,rad=0.6"), zorder=7)
ax.text(CX, MOT_TOP + 2.5, "Motor", ha="center", fontsize=9.5,
        fontweight="bold", color=GRAY, zorder=7)

# ── Instruments inside vessel ─────────────────────────────────────────────────
# TT-201 reactor temperature (right of shaft, upper reactor)
TT201_X, TT201_Y = CX + 13, VBot + int((VTop - VBot) * 0.65)
bubble(ax, TT201_X, TT201_Y, "TT", "201")
ax.text(TT201_X + RB + 1, TT201_Y, "T", ha="left", fontsize=9.5,
        style="italic", color=PROC, zorder=8)

# AT-201 reactant concentration (right of shaft, mid)
AT201_X, AT201_Y = CX + 13, VBot + int((VTop - VBot) * 0.44)
bubble(ax, AT201_X, AT201_Y, "AT", "201")
ax.text(AT201_X + RB + 1, AT201_Y, r"$C_A$", ha="left", fontsize=8.5,
        color=PROC, zorder=8)

# AT-202 product concentration (right of shaft, lower)
AT202_X, AT202_Y = CX + 13, VBot + int((VTop - VBot) * 0.23)
bubble(ax, AT202_X, AT202_Y, "AT", "202")
ax.text(AT202_X + RB + 1, AT202_Y, r"$C_B$", ha="left", fontsize=8.5,
        color=PROC, zorder=8)

# ─────────────────────────────────────────────────────────────────────────────
# Feed system
#   Feed enters LEFT SIDE of reactor at FEED_Y (~83 % height)
#   Feed pipe comes horizontally from feed tank on the left
# ─────────────────────────────────────────────────────────────────────────────
FEED_PIPE_X_START = 24   # feed tank right wall
FEED_PIPE_X_END   = JL   # jacket left wall

# Feed tank TK-101
FTANK_X = 2
ftank = mpatches.FancyBboxPatch(
    (FTANK_X, 100), 20, 28,
    boxstyle="round,pad=0.5", fc=VFILL, ec=PROC, lw=2.0)
ax.add_patch(ftank)
ax.text(FTANK_X + 10, 123, "FEED TANK", ha="center",
        fontsize=8.5, fontweight="bold", color=PROC)
ax.text(FTANK_X + 10, 120, "TK-101", ha="center", fontsize=8, color=GRAY)
ax.text(FTANK_X + 10, 116.5, r"$C_{A,f},\ T_f$", ha="center",
        fontsize=10, color=FEED_C)

# Feed pipe: horizontal at FEED_Y from x=24 to x=JL, then stub through jacket to VL
pipe(ax, FEED_PIPE_X_START, FEED_Y, FEED_PIPE_X_END, FEED_Y)
# Short stub through jacket annulus into vessel left wall;
# arrowhead at VL pointing RIGHT (into vessel), tail to the left of it
pipe(ax, JL, FEED_Y, VL, FEED_Y)
ax.annotate("", xy=(VL - 0.5, FEED_Y), xytext=(VL - 8, FEED_Y),
            arrowprops=dict(arrowstyle="->", color=PROC,
                            lw=LWP * 0.65, mutation_scale=13))

# Feed label (on pipe, just outside jacket)
ax.text(JL - 2, FEED_Y + 3,
        r"Feed  ($F,\ C_{A,f},\ T_f$)",
        ha="right", fontsize=8.5, color=FEED_C, fontweight="bold")

# Feed instruments on horizontal pipe (left to right)
# TT-101 feed temperature
bubble(ax, 35, FEED_Y, "TT", "101")
sig(ax, 35, FEED_Y - RB, 35, FEED_Y - 5)
ax.text(35, FEED_Y + RB + 2, r"$T_f$", ha="center",
        fontsize=8.5, color=FEED_C, zorder=8)

# AT-101 feed concentration
bubble(ax, 52, FEED_Y, "AT", "101")
sig(ax, 52, FEED_Y - RB, 52, FEED_Y - 5)
ax.text(52, FEED_Y + RB + 2, r"$C_{A,f}$", ha="center",
        fontsize=8.5, color=FEED_C, zorder=8)

# FT-101 feed flow (in-line on pipe, no separate signal needed)
bubble(ax, 70, FEED_Y, "FT", "101")

# ─────────────────────────────────────────────────────────────────────────────
# Product stream
#   Exits RIGHT SIDE of reactor at PROD_Y (~18 % height — lower drain nozzle)
# ─────────────────────────────────────────────────────────────────────────────
PROD_END_X = 222

# Stub from vessel right → through jacket → product header (horizontal)
pipe(ax, VR, PROD_Y, JR, PROD_Y)
arrow_end(ax, JR, PROD_Y, PROD_END_X, PROD_Y)

ax.text(PROD_END_X + 2, PROD_Y + 1,
        "PRODUCT\nSTREAM",
        ha="left", va="center", fontsize=10, fontweight="bold", color=PROD_C)
ax.text(PROD_END_X + 2, PROD_Y - 6,
        r"($C_A,\ C_B$ = reactor contents)",
        ha="left", va="top", fontsize=7.5, color=PROD_C)

# AT-203 product concentration
AT203_X = 205
bubble(ax, AT203_X, PROD_Y, "AT", "203")
ax.text(AT203_X, PROD_Y - RB - 3, r"$C_A,\ C_B$",
        ha="center", fontsize=7.8, color=PROD_C, zorder=8)

# ─────────────────────────────────────────────────────────────────────────────
# Coolant system
#   Supply: enters jacket BOTTOM at COOL_IN_X (left side)
#   Return: exits jacket TOP at COOL_OUT_X (right side) — shows upward annular flow
# ─────────────────────────────────────────────────────────────────────────────
COOL_RETURN_Y = JTop + 13  # return header elevation — above jacket dome apex (~133)

# Coolant source TK-201
ctank = mpatches.FancyBboxPatch(
    (2, 8), 20, 24,
    boxstyle="round,pad=0.5", fc="#eff6ff", ec=COOL, lw=2.0)
ax.add_patch(ctank)
ax.text(12, 28.5, "COOLANT",  ha="center", fontsize=8.5,
        fontweight="bold", color=COOL)
ax.text(12, 25.5, "SOURCE",   ha="center", fontsize=8.5,
        fontweight="bold", color=COOL)
ax.text(12, 22.5, "TK-201",   ha="center", fontsize=7.8, color=GRAY)
ax.text(12, 19.5, r"$T_{c,nom}\!=\!300\,$K",
        ha="center", fontsize=7.2, color=GRAY)
ax.text(12, 16.5, r"$285\!\leq\!T_c\!\leq\!330\,$K",
        ha="center", fontsize=6.8, color=GRAY)

# ── Coolant system layout ────────────────────────────────────────────────────
# The simulator manipulates Tc (coolant inlet temperature in K) directly.
# Physically this is achieved by a 3-way mixing valve TCV-201 that blends:
#   • Cold supply  from TK-201          — enters valve from the LEFT
#   • Warm bypass  from return header   — enters valve from ABOVE
# The mixed stream at controlled temperature Tc exits RIGHT to the jacket.
#
# Actuator sits on the BOTTOM of the valve body so that:
#   (a) the bypass pipe entering from the TOP is visually unambiguous, and
#   (b) the instrument signal line connects to the actuator from below
#       without crossing or merging with the bypass pipe.
#
# The bypass is shown with a line-break stub (standard P&ID off-page symbol)
# rather than routing a pipe across the full diagram height.
MIX_X = 57
MIX_Y = 30   # elevation of main horizontal axis through TCV-201

# ── 3-way mixing valve TCV-201 ───────────────────────────────────────────────
tcv_act_bottom, valve_h = three_way_valve(ax, MIX_X, MIX_Y, color=COOL, size=1.8)

# Valve tag (to the left, since bottom is occupied by actuator and top by bypass)
ax.text(MIX_X - valve_h - 2, MIX_Y + 1.5,
        "TCV-201", ha="right", va="bottom", fontsize=7.0, color=COOL,
        fontweight="bold")
ax.text(MIX_X - valve_h - 2, MIX_Y - 1.5,
        r"MV: $T_c$ [K]",
        ha="right", va="top", fontsize=7.0, color=COOL, style="italic")

# ── Cold supply pipe: TK-201 right wall → valve left inlet ──────────────────
pipe(ax, 22, MIX_Y, MIX_X, MIX_Y, color=COOL, lw=LWC)
ax.text(36, MIX_Y + 2.5, "Cold supply", ha="center",
        fontsize=7.0, color=COOL, style="italic")

# ── Coolant return header and warm bypass routing ────────────────────────────
# The return header runs at COOL_RETURN_Y above the reactor.  At the left end
# of the header (x = MIX_X) a tee taps a warm bypass branch that drops
# straight down to the top port of TCV-201.  The bypass crosses the feed
# pipe at y = FEED_Y; a "hop" crossing arc (standard P&ID convention) shows
# the bypass travels over the feed pipe without connecting to it.
#
# Jacket outlet → tee (at COOL_OUT_X) → RIGHT to "COOLANT RETURN" label
#                                      → LEFT  to MIX_X (bypass branch)
#                                              ↓ down to TCV-201 top port

# Jacket outlet vertical
pipe(ax, COOL_OUT_X, JTop, COOL_OUT_X, COOL_RETURN_Y, color=COOL, lw=LWC)
ax.text(COOL_OUT_X, COOL_RETURN_Y + 2.5, r"$T_{c,out}$ out",
        ha="center", fontsize=8.5, color=COOL, fontweight="bold")

# Tee junction dot
ax.plot(COOL_OUT_X, COOL_RETURN_Y, "o", color=COOL, ms=6.0, zorder=6)

# Right branch: main return header
arrow_end(ax, COOL_OUT_X, COOL_RETURN_Y, PROD_END_X, COOL_RETURN_Y,
          color=COOL, lw=LWC)
ax.text(PROD_END_X + 2, COOL_RETURN_Y, "COOLANT\nRETURN",
        ha="left", va="center", fontsize=10, fontweight="bold", color=COOL)

# Left branch: bypass pipe from tee → MIX_X (runs above reactor jacket)
pipe(ax, COOL_OUT_X, COOL_RETURN_Y, MIX_X, COOL_RETURN_Y, color=COOL, lw=LWC)
ax.text((COOL_OUT_X + MIX_X) / 2, COOL_RETURN_Y + 2.5,
        "Warm bypass", ha="center", fontsize=7.5, color=COOL, style="italic")

# Bypass descends from header level to valve top port, crossing feed pipe
# Upper segment: COOL_RETURN_Y → just above feed pipe
_CR = 2.2   # crossing arc radius
pipe(ax, MIX_X, COOL_RETURN_Y, MIX_X, FEED_Y + _CR, color=COOL, lw=LWC)
# Crossing arc: right-bulging semicircle (bypass passes over feed pipe)
_th = np.linspace(np.pi / 2, -np.pi / 2, 40)
ax.add_patch(PathPatch(
    MPath(np.column_stack([MIX_X + _CR * np.cos(_th),
                           FEED_Y + _CR * np.sin(_th)]),
          [MPath.MOVETO] + [MPath.LINETO] * 38 + [MPath.LINETO]),
    fc="none", ec=COOL, lw=LWC, zorder=6))
# Lower segment: just below feed pipe → valve top port
pipe(ax, MIX_X, FEED_Y - _CR, MIX_X, MIX_Y + valve_h, color=COOL, lw=LWC)

# ── Mixed stream: valve right outlet → COOL_IN_X → jacket bottom ─────────────
pipe(ax, MIX_X + valve_h, MIX_Y, COOL_IN_X, MIX_Y, color=COOL, lw=LWC)
pipe(ax, COOL_IN_X, MIX_Y, COOL_IN_X, JBot, color=COOL, lw=LWC)
ax.annotate("", xy=(COOL_IN_X, JBot + 0.5), xytext=(COOL_IN_X, JBot - 7),
            arrowprops=dict(arrowstyle="->", color=COOL,
                            lw=LWC * 0.65, mutation_scale=13))
ax.text(COOL_IN_X - 1, JBot - 2.5, r"$T_c$ in",
        ha="right", fontsize=8.5, color=COOL, va="top", fontweight="bold")

# TT-202 on coolant return line
TT202_X = 208
bubble(ax, TT202_X, COOL_RETURN_Y, "TT", "202")
ax.text(TT202_X, COOL_RETURN_Y + RB + 2, r"$T_{c,out}$",
        ha="center", fontsize=8, color=COOL, zorder=8)

# ─────────────────────────────────────────────────────────────────────────────
# TIC-201 temperature controller
# ─────────────────────────────────────────────────────────────────────────────
TIC_X, TIC_Y = 232, 90

bubble(ax, TIC_X, TIC_Y, "TIC", "201", ctrl=True)
ax.text(TIC_X, TIC_Y + RB + 2.5, "Temperature\nController",
        ha="center", fontsize=8.5, color="#1e1b4b")
ax.text(TIC_X, TIC_Y - RB - 2.5,
        r"SP: $T_{sp}=365\,$K" + "\nPI  (Kc = 3,  τᵢ = 8 min)",
        ha="center", fontsize=7.5, color=GRAY, va="top")

# ── Signal: TT-201 → TIC-201 ─────────────────────────────────────────────────
# TT-201 is right of shaft; route signal rightward at TT201_Y, then up to TIC_Y
sig(ax, TT201_X + RB, TT201_Y, TIC_X, TT201_Y)
sig(ax, TIC_X, TT201_Y, TIC_X, TIC_Y - RB)

# ── Signal: TIC-201 → TCV-201 actuator (bottom) ──────────────────────────────
# Route: TIC bottom → south to SIG_LOW (below actuator, above info bar)
#        → west to MIX_X → north up to actuator bottom
# Actuator is on the BOTTOM of TCV-201, so the signal arrives from below —
# clearly separated from the bypass pipe which enters the valve from the TOP.
SIG_LOW = tcv_act_bottom - 2.5   # just below actuator bottom

sig(ax, TIC_X, TIC_Y - RB, TIC_X, SIG_LOW)        # down from TIC
sig(ax, TIC_X, SIG_LOW, MIX_X, SIG_LOW)            # west at SIG_LOW elevation
sig(ax, MIX_X, SIG_LOW, MIX_X, tcv_act_bottom)     # north up to actuator

# Disturbance scenarios removed from diagram — documented in the simulator
# script (simulate_cstr_temporal_pca_dataset.py) and in the tutorial text.

# ─────────────────────────────────────────────────────────────────────────────
# Legend (top right)
# ─────────────────────────────────────────────────────────────────────────────
LX, LY = 208, 152
ax.text(LX + 17, LY, "LEGEND", ha="center", fontsize=10,
        fontweight="bold", color=PROC)
# Lines
line_items = [
    (PROC, LWP, False, "Process stream"),
    (COOL, LWC, False, "Coolant stream"),
    (SIG,  LWS, True,  "Instrument signal"),
]
for i, (col, lw, dash, label) in enumerate(line_items):
    yy = LY - 7 - i * 7
    ax.plot([LX, LX + 11], [yy, yy], color=col, lw=lw,
            ls="--" if dash else "-")
    ax.text(LX + 13, yy, label, va="center", fontsize=8.5)

# Transmitter symbol
ax.add_patch(plt.Circle((LX + 2.3, LY - 31), 2.1,
                         fc=INST, ec="#5b21b6", lw=1.4))
ax.text(LX + 2.3, LY - 31, "TT", ha="center", va="center",
        fontsize=6.5, fontweight="bold")
ax.text(LX + 6, LY - 31, "Field transmitter", va="center", fontsize=8.5)

# Controller symbol
ax.add_patch(plt.Circle((LX + 2.3, LY - 39), 2.1,
                         fc=CTRL, ec="#5b21b6", lw=1.4))
ax.text(LX + 2.3, LY - 39, "TIC", ha="center", va="center",
        fontsize=5.8, fontweight="bold")
ax.text(LX + 6, LY - 39, "Indicating controller", va="center", fontsize=8.5)

# Valve symbol
_h = 1.4 * 0.55
ax.add_patch(plt.Polygon(
    [[LX + 0.9, LY - 47 + _h], [LX + 0.9, LY - 47 - _h], [LX + 2.3, LY - 47]],
    fc=PROC, ec=PROC))
ax.add_patch(plt.Polygon(
    [[LX + 3.7, LY - 47 + _h], [LX + 3.7, LY - 47 - _h], [LX + 2.3, LY - 47]],
    fc=PROC, ec=PROC))
ax.text(LX + 6, LY - 47, "3-way mixing valve (actuated)", va="center", fontsize=8.5)

# ─────────────────────────────────────────────────────────────────────────────
# Measured variables — bottom info bar
# ─────────────────────────────────────────────────────────────────────────────
ax.plot([2, 258], [18, 18], color=LG, lw=1.0)
ax.add_patch(mpatches.FancyBboxPatch(
    (2, 1), 255, 17,
    boxstyle="round,pad=0.4", fc="#f9fafb", ec="#9ca3af", lw=1.2))
ax.text(130, 18.6,
        "MEASURED VARIABLES — 12 numeric PCA inputs  "
        "(+3 string columns for coloring: event / regime / fault_active)",
        ha="center", fontsize=8.8, fontweight="bold", color=PROC)
MEAS = [
    ("T_K",                       r"$T$",          "Reactor temperature [K]"),
    ("Tc_out_K",                   r"$T_{c,out}$",  "Coolant outlet temperature [K]"),
    ("Tf_K",                       r"$T_f$",        "Feed temperature [K]"),
    ("CA_mol_L",                   r"$C_A$",        "Reactant concentration [mol/L]"),
    ("CB_mol_L",                   r"$C_B$",        "Product concentration [mol/L]"),
    ("CAf_mol_L",                  r"$C_{A,f}$",    "Feed concentration [mol/L]"),
    ("F_L_min",                    r"$F$",          "Feed flow rate [L/min]"),
    ("cooling_duty_kJ_min",        r"$Q$",          "Cooling duty  UA·(T−Tc) [kJ/min]"),
    ("reaction_rate_mol_L_min",    r"$r_A$",        "Reaction rate  k(T)·CA [mol/L/min]"),
    ("conversion_fraction",        r"$X_A$",        "Fractional conversion (CAf−CA)/CAf"),
    ("heat_transfer_UA_kJ_min_K",  r"$UA$",         "Heat-transfer coeff. [kJ/(min·K)]"),
    ("residence_time_min",         r"$\tau$",       "Hydraulic residence time V/F [min]"),
]
COL_STARTS = [4, 68, 133, 198]
for i, (col_name, sym, desc) in enumerate(MEAS):
    col_i = i // 3
    row_i = i % 3
    x0 = COL_STARTS[col_i]
    y0 = 15.8 - row_i * 4.6
    ax.text(x0,      y0, f"• {col_name}", fontsize=6.2, color="#1a365d",
            va="top", family="monospace")
    ax.text(x0 + 26, y0, sym,            fontsize=6.8, color=GRAY, va="top")
    ax.text(x0 + 33, y0, desc,           fontsize=6.0, color=LG,   va="top")

# ─────────────────────────────────────────────────────────────────────────────
# Save
# ─────────────────────────────────────────────────────────────────────────────
out = Path(__file__).parent / "cstr_diagram.png"
plt.savefig(out, dpi=150, bbox_inches="tight", facecolor="white")
print(f"Saved → {out}  ({out.stat().st_size // 1024} KB)")
