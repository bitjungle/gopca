"""Pairplot of the seven NHANES body-measures variables, colored by sex.

Generates body_measures_pairplot.png used by the Body Measures tutorial. Like the
other testdata scripts it reads and writes relative to the current directory, so
run it from testdata/nhanes/:

    cd testdata/nhanes && python make_pairplot.py
"""
import pandas as pd
import seaborn as sns
import matplotlib.pyplot as plt

# Load dataset
df = pd.read_csv('body_measures.csv')

# The seven measurement columns are the PCA features. The remaining columns
# (Sample_ID and the #target columns) must NOT be treated as plot variables.
features = {
    'Weight (kg)': 'Weight',
    'Height (cm)': 'Height',
    'Upper Leg Length (cm)': 'Leg len',
    'Upper Arm Length (cm)': 'Arm len',
    'Arm Circumference (cm)': 'Arm circ',
    'Waist Circumference (cm)': 'Waist',
    'Hip Circumference (cm)': 'Hip',
}

# Colour by sex; give the column a clean name so the legend reads nicely.
df = df.rename(columns={**features, 'Gender#target': 'Sex'})
feature_labels = list(features.values())

# 5096 points overwhelm a 7x7 grid; sample a legible subset (fixed seed = reproducible).
plot_df = df.sample(700, random_state=0)

# Plot
sns.set_theme(style="white")
g = sns.pairplot(
    plot_df,
    vars=feature_labels,          # only the measurements, not Sample_ID / targets
    hue="Sex",
    diag_kind="kde",              # smooth distributions
    palette={'Male': '#2c7fb8', 'Female': '#de8f05'},
    corner=True,                  # lower triangle only — less clutter
    plot_kws=dict(s=10, alpha=0.5, edgecolor='none'),
    height=1.25,
)
g.figure.suptitle('Body measures pairplot (700 adults, colored by sex)', y=1.01)

g.savefig('body_measures_pairplot.png', dpi=110, bbox_inches='tight')
#plt.show()
