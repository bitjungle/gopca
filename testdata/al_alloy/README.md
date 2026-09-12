## Aluminium Alloy Dataset

The Aluminium Alloy Dataset for Unsupervised Learning is a publicly available materials dataset containing 1,154 aluminium alloy instances characterized by their elemental compositions and processing conditions. The dataset includes concentrations of 25 alloying elements together with categorical information describing heat treatment and mechanical processing states, making it suitable for exploring relationships between composition, processing history, and alloy classification. Originally compiled from published literature and experimental sources, the dataset was developed to support unsupervised machine-learning approaches for discovering natural groupings and patterns within aluminium alloy systems. Its relatively diverse chemical and processing information makes it a useful benchmark for clustering, dimensionality reduction, anomaly detection, and other materials-informatics studies. 

* [Data source](https://data.mendeley.com/datasets/tvtg7gs59p/1)

* Research article: Bhat, N., Barnard, A. S., & Birbilis, N. (2023). Unsupervised machine learning discovers classes in aluminium alloys. Royal Society Open Science, 10(2), 220360.

## The file

`unsupervised_learning_dataset.xlsx` is the dataset exactly as downloaded from
Mendeley Data: a single sheet named `in`, with 1155 data rows.

That count is one higher than the 1154 alloys the paper describes, and the
difference is a malformed row. Sheet row 42 holds `compression test` in column C
and nothing else, while the record above it ends with a dangling comma:

```
row 41  Condition: "Cast material, quenched from 648C, aged 300C/5 h,"
row 42  Condition: "compression test"        (and nothing in any other column)
```

Three neighbouring records from the same study end `..., compression test`, so a
spreadsheet line break that stranded the tail of row 41 is the likely
explanation. Only the free-text `Condition` is affected; the columns that carry
the processing information for modelling, `Condition augmented` and `proc_num`,
are intact and agree with those neighbours.

The row is left exactly as published rather than corrected here, so this copy
does not disagree with the source. GoCSV's data quality report flags it.
