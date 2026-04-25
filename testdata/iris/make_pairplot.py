import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt
from sklearn.datasets import load_iris

# Load dataset
iris = load_iris()

# Create DataFrame
df = pd.DataFrame(iris.data, columns=iris.feature_names)

# Add target labels
df["target"] = iris.target
df["target"] = df["target"].map({
    0: "setosa",
    1: "versicolor",
    2: "virginica"
})

# Clean column names (optional, but matches your plot style)
df.columns = [col.replace(" (cm)", "") for col in df.columns]

# Plot
sns.set(style="white")
sns.pairplot(
    df,
    hue="target",
    diag_kind="kde",   # smooth distributions like in your image
    palette="deep"
)

plt.savefig('iris_pairplot.png')
plt.show()
