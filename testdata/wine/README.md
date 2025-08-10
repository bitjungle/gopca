# The Wine Dataset

Wine dataset in scikit-learn is a well-known classification dataset originally from the UCI Machine Learning Repository, and it’s often used for demonstrating multivariate analysis, classification algorithms, and PCA.
Background

## Source

Originally hosted in the UCI Machine Learning Repository:
https://archive.ics.uci.edu/ml/datasets/wine (DOI: https://doi.org/10.24432/C5PC7J)

The data in this folder is a copy from scikit-learn, using `from sklearn.datasets import load_wine`.

## Purpose

The goal is to classify wines into one of the three cultivars based on the results of a chemical analysis.

## Samples

* Total: 178 samples
* Classes: 3 wine cultivars
* Class distribution: [59, 71, 48]

## Features

The 13 chemical analysis results include:

* Alcohol
* Malic acid
* Ash
* Alcalinity of ash
* Magnesium
* Total phenols
* Flavanoids
* Nonflavanoid phenols
* Proanthocyanins
* Color intensity
* Hue
* OD280/OD315 of diluted wines
* Proline

## Target

Integer labels (0, 1, 2) corresponding to the three cultivars.

