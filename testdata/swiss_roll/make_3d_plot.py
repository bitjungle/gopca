import pandas as pd
import matplotlib.pyplot as plt
from mpl_toolkits.mplot3d import Axes3D

# Load dataset
df = pd.read_csv("swiss_roll.csv")

# Drop index column if present
if df.columns[0].lower().startswith("unnamed"):
    df = df.drop(columns=df.columns[0])

# Map categories to numeric values for coloring
category_map = {cat: i for i, cat in enumerate(sorted(df["color_category"].unique()))}
colors = df["color_category"].map(category_map)

# Create 3D plot
fig = plt.figure(figsize=(8, 6))
ax = fig.add_subplot(111, projection='3d')

scatter = ax.scatter(
    df["X"], df["Y"], df["Z"],
    c=colors,
    cmap="Spectral",
    s=10,
    alpha=0.8
)

ax.set_xlabel("X")
ax.set_ylabel("Y")
ax.set_zlabel("Z")

# Add colorbar
cbar = plt.colorbar(scatter, ax=ax, pad=0.1)
cbar.set_label("Position along roll (A–H)")

plt.title("Swiss Roll Dataset (3D View)")
plt.tight_layout()

plt.savefig("swiss_roll_3d.png", dpi=300)
plt.show()