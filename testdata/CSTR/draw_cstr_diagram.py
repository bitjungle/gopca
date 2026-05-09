#!/usr/bin/env python3
"""
draw_cstr_diagram.py  —  v4
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
COOL_PIPE_Y   = 22    # horizontal coolant supply pipe elevation
COOL_RETURN_Y = JTop + 9   # coolant return header elevation

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

# ── Coolant temperature manipulation: 3-way mixing valve ─────────────────────
# The simulator manipulates Tc (coolant inlet temperature, K) directly.
# The standard engineering implementation is a 3-way mixing valve that blends:
#   • cold fresh coolant from TK-201 (supply)
#   • warm coolant bypassed from the jacket return
# The mixing ratio — set by TCV-201 — determines Tc.
#
# MIX_X is the mixing point on the supply pipe.
MIX_X = 57   # x-coordinate of the 3-way mixing valve

# Cold supply pipe: tank right → MIX_X
pipe(ax, 22, COOL_PIPE_Y, MIX_X, COOL_PIPE_Y, color=COOL, lw=LWC)

# 3-way mixing valve symbol at MIX_X:
#   left triangle (cold supply from left) + right triangle (warm bypass from above)
#   + right-hand triangle (mixed outlet to right) — drawn as two bow-ties
MV_H = 1.6 * 0.55   # half-height of valve triangle
# horizontal pass-through bow-tie (cold in → mixed out)
ax.add_patch(plt.Polygon(
    [[MIX_X - MV_H, COOL_PIPE_Y + MV_H * 0.85],
     [MIX_X - MV_H, COOL_PIPE_Y - MV_H * 0.85],
     [MIX_X,        COOL_PIPE_Y]],
    fc=COOL, ec=COOL, zorder=5))
ax.add_patch(plt.Polygon(
    [[MIX_X + MV_H, COOL_PIPE_Y + MV_H * 0.85],
     [MIX_X + MV_H, COOL_PIPE_Y - MV_H * 0.85],
     [MIX_X,        COOL_PIPE_Y]],
    fc=COOL, ec=COOL, zorder=5))
# vertical bypass inlet triangle (warm bypass comes down from above)
ax.add_patch(plt.Polygon(
    [[MIX_X - MV_H * 0.85, COOL_PIPE_Y + MV_H],
     [MIX_X + MV_H * 0.85, COOL_PIPE_Y + MV_H],
     [MIX_X,                COOL_PIPE_Y]],
    fc=COOL, ec=COOL, zorder=5))

# Actuator stem and circle above the vertical inlet
MV_STEM_TOP = COOL_PIPE_Y + MV_H + 2.5
MV_ACT_R    = RB * 0.65
MV_ACT_CY   = MV_STEM_TOP + MV_ACT_R
ax.plot([MIX_X, MIX_X], [COOL_PIPE_Y + MV_H, MV_STEM_TOP], color=COOL, lw=2.0, zorder=5)
ax.add_patch(plt.Circle((MIX_X, MV_ACT_CY), MV_ACT_R,
                         fc="white", ec=COOL, lw=1.4, zorder=5))
ax.text(MIX_X, MV_ACT_CY, "A", ha="center", va="center",
        fontsize=4.8, color=COOL, zorder=6)
ax.text(MIX_X, COOL_PIPE_Y - MV_H * 0.85 - 1.3, "TCV-201",
        ha="center", va="top", fontsize=7.0, color=COOL)

tcv_top = MV_ACT_CY + MV_ACT_R   # top of actuator (signal connection point)

# (Warm bypass circuit not drawn to avoid clutter — see MV annotation below.)

# Mixed outlet pipe: MIX_X → COOL_IN_X (horizontal), then up into jacket bottom
pipe(ax, MIX_X + MV_H, COOL_PIPE_Y, COOL_IN_X, COOL_PIPE_Y, color=COOL, lw=LWC)
pipe(ax, COOL_IN_X, COOL_PIPE_Y, COOL_IN_X, JBot, color=COOL, lw=LWC)
ax.annotate("", xy=(COOL_IN_X, JBot + 0.5), xytext=(COOL_IN_X, JBot - 7),
            arrowprops=dict(arrowstyle="->", color=COOL,
                            lw=LWC * 0.65, mutation_scale=13))
ax.text(COOL_IN_X - 1, JBot - 2.5, r"$T_c$ in",
        ha="right", fontsize=8.5, color=COOL, va="top", fontweight="bold")

# MV = manipulated variable annotation (above the pipe, right of valve)
ax.text(MIX_X + MV_H + 2, COOL_PIPE_Y + 2.5,
        r"MV: $T_c$ [K]  (3-way mixing valve)",
        ha="left", fontsize=7.5, color=COOL, style="italic")

# Coolant return: exits jacket top → up → right
pipe(ax, COOL_OUT_X, JTop, COOL_OUT_X, COOL_RETURN_Y, color=COOL, lw=LWC)
arrow_end(ax, COOL_OUT_X, COOL_RETURN_Y, PROD_END_X, COOL_RETURN_Y,
          color=COOL, lw=LWC)
ax.text(PROD_END_X + 2, COOL_RETURN_Y, "COOLANT\nRETURN",
        ha="left", va="center", fontsize=10, fontweight="bold", color=COOL)
ax.text(COOL_OUT_X, COOL_RETURN_Y + 2.5, r"$T_{c,out}$ out",
        ha="center", fontsize=8.5, color=COOL, fontweight="bold")

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

# ── Signal: TIC-201 → TCV-201 ────────────────────────────────────────────────
# Route: TIC bottom → south to SIG_LOW (between coolant return and feed pipe)
#        → west to TCV_X → down to TCV actuator top
# Route signal below the coolant return pipe to avoid crossing it
SIG_LOW = COOL_PIPE_Y + 6    # between coolant supply pipe and bottom info box

sig(ax, TIC_X, TIC_Y - RB, TIC_X, SIG_LOW)      # down from TIC
sig(ax, TIC_X, SIG_LOW, MIX_X, SIG_LOW)          # west at SIG_LOW elevation
sig(ax, MIX_X, SIG_LOW, MIX_X, tcv_top)          # down to TCV actuator

# ─────────────────────────────────────────────────────────────────────────────
# Disturbances box (bottom-left area — below feed tank)
# ─────────────────────────────────────────────────────────────────────────────
dbox = mpatches.FancyBboxPatch(
    (2, 38), 80, 52,
    boxstyle="round,pad=0.6", fc="#fffbeb", ec="#d97706", lw=1.8, zorder=0)
ax.add_patch(dbox)
ax.text(42, 90, "DISTURBANCE SCENARIOS", ha="center", fontsize=10,
        fontweight="bold", color="#92400e")
DIST = [
    ("  0 – 120 min", "Normal steady state (baseline)"),
    ("120 – 240 min", r"Feed conc. step: $C_{A,f}$ +10 %"),
    ("240 – 360 min", r"Feed temp. step: $T_f$ +7 K"),
    ("360 – 520 min", "Feed flow osc. ±8 %,  period 40 min"),
    ("520 – 680 min", "Cooling fault: UA drops 28 %"),
    ("680 – 800 min", "Recovery to nominal conditions"),
]
for i, (t_, d_) in enumerate(DIST):
    yy = 87 - i * 7.6
    ax.text(4,  yy, f"• {t_}:", fontsize=8, color="#92400e",
            fontweight="bold", va="top")
    ax.text(51, yy, d_,        fontsize=8, color="#78350f", va="top")

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
ax.text(LX + 6, LY - 47, "Control valve (actuated)", va="center", fontsize=8.5)

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
