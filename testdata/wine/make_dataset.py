from pprint import pprint
import pandas as pd

from sklearn.datasets import load_wine
data = load_wine(as_frame=True)

# Replace the target values with their names
data.frame['target'] = data.target.map(dict(zip(range(len(data.target_names)), data.target_names)))

# Rename the target column to 'classes'
data.frame = data.frame.rename(columns={'target': 'classes'})

# Save
data.frame.to_csv('wine.csv', index=True)
