# https://scikit-learn.org/stable/modules/generated/sklearn.datasets.make_swiss_roll.html
from sklearn.datasets import make_swiss_roll
import numpy as np
import pandas as pd
import matplotlib.pyplot as plt

# Generate data
X, color = make_swiss_roll(n_samples=1000, noise=0.1, random_state=42)

# Save the Swiss Roll data to a CSV file
# X contains the 3D coordinates, color contains the color values
# Merge X and color into a DataFrame
df = pd.DataFrame(X, columns=['X', 'Y', 'Z'])
df['color'] = color
# Save to CSV (legacy file kept for reference)
df.to_csv('swiss_roll_color_target.csv', index=True)

# Save the main dataset with the GoPCA #target convention:
# columns ending in #target are excluded from PCA and used only for colouring.
df_target = pd.DataFrame(X, columns=['X', 'Y', 'Z'])
df_target['color #target'] = color
df_target.to_csv('swiss_roll.csv', index=True)