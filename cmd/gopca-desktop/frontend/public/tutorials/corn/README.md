## Corn (corn_*.csv) near-infrared (NIR) data

This dataset consists of 80 samples of corn measured on 3 different NIR spectrometers. The wavelength range is 1100-2498nm at 2 nm intervals (700 channels). The moisture, oil, protein and starch values for each of the samples is also included. A number of NBS glass standards were also measured on each instrument. The data was originally made at Cargill. 

* [Data source](https://eigenvector.com/data/Corn/)
* Research article: Fatemi, Singh & Kamruzzaman (2022), *Identification of informative spectral ranges for predicting major chemical constituents in corn using NIR spectroscopy*, Food Chemistry 383:132442. https://doi.org/10.1016/j.foodchem.2022.132442
* License: Unknown

## Regenerating the data

The full Eigenvector dataset holds the same 80 samples measured on three instruments.
The `corn.csv` bundled with GoPCA uses the **m5** instrument only. It is built by
[`make_dataset.py`](make_dataset.py) from two files in this directory,
`corn_m5spec.csv` (the spectra) and `corn_propvals.csv` (the laboratory values):

```bash
cd testdata/corn
python make_dataset.py            # writes corn.csv
```

The script also derives the categorical `Low`/`Mid`/`High` columns from the
laboratory values, alongside the continuous `#target` columns.

The spectra figure used by the tutorial is regenerated with
[`make_spectra_plot.py`](make_spectra_plot.py).