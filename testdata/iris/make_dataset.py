from pprint import pprint
import pandas as pd

from sklearn.datasets import load_iris
data = load_iris(as_frame=True)

# Replace the target values with their names
data.frame['species'] = data.target.map(dict(zip(range(len(data.target_names)), data.target_names)))
data.frame.to_csv('iris.csv', index=True)