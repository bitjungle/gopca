"""Generate the two iris CSVs this repository ships.

Both come from scikit-learn, so both carry the corrected values at samples 35
and 38 -- the two rows where the UCI copy of Iris departs from Fisher's original
(#917). They differ only in what they are for:

  iris.csv           A test fixture. Carries species twice -- once as the numeric
                     class code 0/1/2 under "species#target", once as text. The
                     numeric copy exists so that regressing on a class label can
                     be exercised; internal/core/pcr_advisories.go and
                     cmd/gopca-desktop/pcr_parity_test.go both depend on it.

  iris_tutorial.csv  The dataset embedded in GoPCA Desktop and used by the iris
                     tutorial. Species appears once, as text, and rows carry
                     readable identifiers (se_01, ve_12, vi_47) so points in a
                     scores plot can be told apart. No class code, because a
                     reader should not meet a column whose meaning has to be
                     explained before it can be ignored.

Run from this directory:  python make_dataset.py
"""

import pandas as pd
from sklearn.datasets import load_iris

data = load_iris(as_frame=True)

# Both files are derived from these two, so they cannot drift apart. data.frame
# is deliberately not mutated: the fixture used to rename its columns in place,
# which left the tutorial needing a second load_iris() to get a clean frame.
measurements = data.data
species = data.target.map(dict(zip(range(len(data.target_names)), data.target_names)))

# --- the test fixture -------------------------------------------------------
fixture = measurements.copy()
fixture['species#target'] = data.target
fixture['species'] = species
fixture.to_csv('iris.csv', index=True)

# --- the tutorial dataset ---------------------------------------------------
tutorial = measurements.copy()
tutorial['species'] = species

# Two-letter species prefix plus a per-species counter, so a label identifies the
# sample and says what it is: se_01..se_50, ve_01..ve_50, vi_01..vi_50.
counters = {}
labels = []
for name in tutorial['species']:
    counters[name] = counters.get(name, 0) + 1
    labels.append(f"{name[:2]}_{counters[name]:02d}")
tutorial.index = labels
tutorial.index.name = ''

tutorial.to_csv('iris_tutorial.csv', index=True)
