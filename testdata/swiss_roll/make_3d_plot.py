import pandas as pd
import numpy as np
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d import Axes3D

# Load dataset — use continuous color target matching goPCA display
df = pd.read_csv("swiss_roll_color_target.csv")

if df.columns[0].lower().startswith("unnamed"):
    df = df.drop(columns=df.columns[0])

color_vals = df["color"].values
cmap = "plasma"

fig = plt.figure(figsize=(14, 6), facecolor="white")

# ── Left panel: 3D perspective view ──────────────────────────────────────────
ax1 = fig.add_subplot(121, projection="3d")
ax1.set_facecolor("white")

sc1 = ax1.scatter(
    df["X"], df["Y"], df["Z"],
    c=color_vals, cmap=cmap,
    s=22, alpha=0.9, linewidths=0,
)

for pane in (ax1.xaxis.pane, ax1.yaxis.pane, ax1.zaxis.pane):
    pane.fill = False
    pane.set_edgecolor("#cccccc")

ax1.set_xlabel("X", labelpad=4)
ax1.set_ylabel("Y  (height)", labelpad=4)
ax1.set_zlabel("Z", labelpad=4)
ax1.set_title("3D structure", pad=10, fontsize=12)

# Angle that shows both the spiral layers and the height dimension
ax1.view_init(elev=10, azim=-70)

# ── Right panel: X–Z plane projection (spiral plane, colour = height) ─────────
# This is the key insight plot: remove height (Y) and project onto the
# spiral plane. The concentric rings show immediately why a linear
# projection cannot separate inner from outer layers.
ax2 = fig.add_subplot(122)
ax2.set_facecolor("white")

sc2 = ax2.scatter(
    df["X"], df["Z"],
    c=color_vals, cmap=cmap,
    s=14, alpha=0.85, linewidths=0,
)

ax2.set_xlabel("X", fontsize=11)
ax2.set_ylabel("Z", fontsize=11)
ax2.set_title("Top-down view: X–Z plane\n(height Y removed)", fontsize=12)
ax2.set_aspect("equal")
ax2.tick_params(labelsize=9)
for spine in ax2.spines.values():
    spine.set_edgecolor("#cccccc")

# Annotate inner and outer edges
ax2.annotate("inner edge\n(low t)", xy=(1, 1), xytext=(5, 4),
             fontsize=8, color="#555555",
             arrowprops=dict(arrowstyle="->", color="#555555", lw=0.8))
ax2.annotate("outer edge\n(high t)", xy=(13, -2), xytext=(6, -9),
             fontsize=8, color="#555555",
             arrowprops=dict(arrowstyle="->", color="#555555", lw=0.8))

# ── Shared colorbar ───────────────────────────────────────────────────────────
fig.subplots_adjust(right=0.88, wspace=0.3)
cbar_ax = fig.add_axes([0.91, 0.15, 0.018, 0.7])
cb = fig.colorbar(sc1, cax=cbar_ax)
cb.set_label("Position along roll (t)", fontsize=11)
cb.ax.tick_params(labelsize=9)

fig.suptitle("Swiss Roll Dataset", fontsize=14, y=1.01)

plt.savefig("swiss_roll_3d.png", dpi=150, bbox_inches="tight", facecolor="white")
print("Saved swiss_roll_3d.png")
