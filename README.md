# GoPCA - The Definitive Principal Component Analysis Application

Professional-grade PCA analysis made simple. Available as both a powerful command-line tool and an intuitive desktop application.

![GoPCA Desktop Application](docs/images/desktop-overview.png)
*[Screenshot: Full desktop application window showing the main interface with data loaded, PCA results displayed in the scores plot, and the various panels (data preview, settings, visualization tabs) visible. Should showcase the modern, professional UI with dark mode if available]*

## What is GoPCA?

GoPCA is **the** go-to application for Principal Component Analysis - a fundamental technique for understanding complex, multivariate data. Whether you're analyzing spectroscopic data, exploring gene expression patterns, or reducing dimensionality in machine learning pipelines, GoPCA provides the tools you need.

### Includes GoCSV Data Preparation Tool

GoPCA comes with **GoCSV**, a companion application for preparing your data. With its Excel-like interface, GoCSV makes it easy to clean, edit, and format your CSV files before PCA analysis - ensuring your data is analysis-ready.

### Three Powerful Tools

#### 🖥️ Desktop Application
Perfect for interactive data exploration, method development, and teaching.

![GoPCA Interactive Visualization](docs/images/interactive-viz.png)
*[Screenshot: Close-up of the visualization panel showing an interactive scores plot with colored groups, confidence ellipses, and hover tooltips displaying sample information. Include the plot controls (zoom, pan, export buttons) visible]*

#### 🚀 Command-Line Interface
Ideal for automation, batch processing, and integration into data pipelines.

```bash
# Analyze your data with a single command
gopca-cli analyze data.csv --components 3 --scale --output results/

# Get detailed component information
gopca-cli analyze spectra.csv --components 5 --export-loadings
```

#### 📝 GoCSV Data Editor
Clean and prepare your data with an intuitive spreadsheet-like interface.

- Edit cells directly like in Excel
- Add, remove, or reorder columns
- Handle missing values
- Export clean CSV files ready for PCA analysis

## Key Features

### 📊 Comprehensive Analysis
- **Multiple algorithms**: SVD (default) and NIPALS for handling missing data
- **Flexible preprocessing**: Mean centering, scaling, robust scaling, SNV, and more
- **Detailed outputs**: Scores, loadings, explained variance, and contribution plots

### 📈 Professional Visualizations

![GoPCA Visualization Gallery](docs/images/viz-gallery.png)
*[Screenshot: A 2x2 grid showing four different plot types: 1) Scores plot with groups and ellipses, 2) Loadings bar chart, 3) Scree plot showing explained variance, 4) Biplot combining scores and loadings. Each should be clearly labeled]*

- **Interactive scores plots** with zoom, pan, and selection
- **Loadings visualizations** to understand variable contributions  
- **Scree plots** for component selection
- **Biplots** showing samples and variables together
- **Export all plots** as publication-ready PNG images

### 🎯 Built for Real Work
- **Example datasets included**: Explore PCA immediately with wine and iris datasets
- **Handles real-world data**: Robust to missing values, mixed scales, and outliers
- **Data preparation included**: GoCSV helps clean messy data before analysis
- **Cross-platform**: Native performance on Windows, macOS, and Linux
- **Fast**: Optimized implementations handle large datasets efficiently

## Getting Started

### Desktop Application

1. **Download** the latest release for your platform
2. **Launch** GoPCA Desktop
3. **Load** your CSV data or try an example dataset
4. **Configure** preprocessing options
5. **Run** PCA and explore the results interactively

![GoPCA Workflow](docs/images/workflow.png)
*[Screenshot: A step-by-step visual showing the workflow - data loading screen with file browser, preprocessing options panel with checkboxes, and the resulting visualization after running PCA. Could be a single wide image with arrows between steps]*

### Data Preparation with GoCSV

1. **Launch** GoCSV from the GoPCA installation folder
2. **Open** your raw CSV file
3. **Clean** your data - remove empty rows, fix headers, handle missing values
4. **Save** the cleaned file
5. **Import** directly into GoPCA for analysis

### Command-Line Interface

```bash
# Install via download or package manager
wget https://github.com/bitjungle/gopca/releases/latest/gopca-cli
chmod +x gopca-cli

# Basic analysis
./gopca-cli analyze mydata.csv

# Advanced options
./gopca-cli analyze mydata.csv \
  --components 4 \
  --scale \
  --robust \
  --export-all \
  --output results/
```

## Use Cases

### 🧪 Chemometrics & Spectroscopy
Analyze NIR, FTIR, Raman, or other spectroscopic data to identify chemical patterns and outliers.

### 🧬 Bioinformatics
Explore gene expression, proteomics, or metabolomics data to find biological patterns and biomarkers.

### 🏭 Process Monitoring
Monitor industrial processes, detect anomalies, and understand multivariate relationships.

### 📊 Data Science
Reduce dimensionality before machine learning, explore feature relationships, or visualize high-dimensional data.

## Example Output

![GoPCA Results Example](docs/images/results-example.png)
*[Screenshot: Split view showing the CLI output on the left (text-based results table with explained variance, cumulative variance, etc.) and the GUI visualization on the right (a professional-looking scores plot with multiple groups). This demonstrates both interfaces analyzing the same dataset]*

## Documentation

- [Introduction to PCA](docs/intro_to_pca.md) - Learn the fundamentals
- [User Guide](docs/user_guide.md) - Detailed usage instructions
- [CLI Reference](docs/cli_reference.md) - Command-line options and examples

## System Requirements

- **Operating Systems**: Windows 10+, macOS 10.15+, Linux (Ubuntu 20.04+, Fedora 34+)
- **Memory**: 2GB RAM minimum, 8GB+ recommended for large datasets
- **Disk Space**: 100MB for application, additional space for data

## Support

- **Issues**: [GitHub Issues](https://github.com/bitjungle/gopca/issues)
- **Discussions**: [GitHub Discussions](https://github.com/bitjungle/gopca/discussions)
- **Email**: support@gopca.app

## License

GoPCA is open-source software licensed under the MIT License. See [LICENSE](LICENSE) for details.
