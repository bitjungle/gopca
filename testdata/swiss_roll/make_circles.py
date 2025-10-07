# https://scikit-learn.org/stable/modules/generated/sklearn.datasets.make_circles.html
from sklearn.datasets import make_circles
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# Generate data
X, label = make_circles(n_samples=1000, noise=0.05, factor=0.5, random_state=42)

# Save the 2D circle data to a CSV file
# X contains the 2D coordinates, label contains the binary class (0 or 1)
df = pd.DataFrame(X, columns=['X', 'Y'])
df['label'] = label
df.to_csv('circles_label_target.csv', index=True)

# Map labels to categories ('Inner', 'Outer') for categorical visualization
label_categories = np.where(label == 0, 'Inner', 'Outer')

# Save X and label_categories to a CSV file
df = pd.DataFrame(X, columns=['X', 'Y'])
df['label_category'] = label_categories
df.to_csv('circles.csv', index=True)
