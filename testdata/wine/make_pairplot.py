import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt

# Load dataset (as used in GoPCA)
df = pd.read_csv("wine.csv")

# Drop index column if present
if df.columns[0].lower() in ["index", "unnamed: 0"]:
    df = df.drop(columns=df.columns[0])

# Rename target column for clarity
df = df.rename(columns={"classes": "target"})

# Plot style
sns.set(style="white")

# Pairplot
sns.pairplot(
    df,
    hue="target",
    diag_kind="kde",
    palette="deep",
    corner=False
)

plt.savefig("wine_pairplot.png")
plt.show()