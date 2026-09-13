"""Generate iris.csv, the Iris test fixture.

This repository ships two iris files, deliberately (#915):

  iris.csv           Generated here, from scikit-learn. Carries species twice --
                     once as the numeric class code 0/1/2 under "species#target",
                     once as text. The numeric copy exists so that regressing on
                     a class label can be exercised; internal/core/pcr_advisories.go
                     and cmd/gopca-desktop/pcr_parity_test.go both depend on it.

  iris_tutorial.csv  The dataset embedded in GoPCA Desktop and used by the iris
                     tutorial. Species appears once, as text, and rows carry
                     readable identifiers (se_01, ve_12, vi_47) so points in a
                     scores plot can be told apart.

                     NOT generated here, and not regenerable from scikit-learn:
                     it differs at samples 35 and 38, the two rows where the UCI
                     copy of Iris is known to depart from Fisher's original.
                     scikit-learn ships the corrected values (35: 4.9,3.1,1.5,0.2
                     and 38: 4.9,3.6,1.4,0.1); the tutorial file has the UCI ones.
                     Emitting it from here would silently rewrite the tutorial's
                     data, so it is a committed artifact instead.

Run from this directory:  python make_dataset.py
"""

import pandas as pd
from sklearn.datasets import load_iris

data = load_iris(as_frame=True)

# Replace the target values with their names
data.frame['species'] = data.target.map(
    dict(zip(range(len(data.target_names)), data.target_names)))

# Rename the target column to species#target
data.frame = data.frame.rename(columns={'target': 'species#target'})

data.frame.to_csv('iris.csv', index=True)
