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
# Save to CSV
df.to_csv('swiss_roll_color_target.csv', index=True)


# Dicide the color variable into categories from A to H
categories = ['A', 'B', 'C', 'D', 'E', 'F', 'G', 'H']
color_normalized = (color - color.min()) / (color.max() - color.min())  # Normalize to [0, 1]
color_categories = np.array([categories[min(int(c * len(categories)), len(categories) - 1)] for c in color_normalized])    

# Save X and color_categories to a csv file using pandas
df = pd.DataFrame(X, columns=['X', 'Y', 'Z'])
df['color_category'] = color_categories
df.to_csv('swiss_roll.csv', index=True)