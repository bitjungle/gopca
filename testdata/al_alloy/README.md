## Aluminium Alloy Dataset

The Aluminium Alloy Dataset for Unsupervised Learning is a publicly available materials dataset containing 1,154 aluminium alloy instances characterized by their elemental compositions and processing conditions. The dataset includes concentrations of 25 alloying elements together with categorical information describing heat treatment and mechanical processing states, making it suitable for exploring relationships between composition, processing history, and alloy classification. Originally compiled from published literature and experimental sources, the dataset was developed to support unsupervised machine-learning approaches for discovering natural groupings and patterns within aluminium alloy systems. Its relatively diverse chemical and processing information makes it a useful benchmark for clustering, dimensionality reduction, anomaly detection, and other materials-informatics studies. 

* [Data source url](https://data.mendeley.com/datasets/tvtg7gs59p/1)

* Research article: Bhat, N., Barnard, A. S., & Birbilis, N. (2023). Unsupervised machine learning discovers classes in aluminium alloys. Royal Society Open Science, 10(2), 220360.

## Files

| File | Contents |
|------|----------|
| `unsupervised_learning_dataset.xlsx` | The dataset exactly as published, one sheet (`in`), 1155 data rows |
| `unsupervised_learning_dataset_fixed.xlsx` | Two sheets: `orig` is the published sheet unchanged, `fixed` has the stray row removed and holds 1154 records |

The published sheet contains one malformed row: sheet row 42 holds `compression test`
in column C and nothing else, which appears to be a spreadsheet line break that
stranded the tail of the preceding record. It is why the file has 1155 data rows
where the paper describes 1154 alloys.

The `fixed` sheet is the published sheet with that row removed and nothing else
changed; the 1154 remaining records are identical to the originals field for field.

Both files are useful for testing. The single-sheet file opens directly; the
two-sheet file exercises sheet selection on import.
