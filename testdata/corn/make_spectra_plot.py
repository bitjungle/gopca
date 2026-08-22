import pandas as pd
import matplotlib.pyplot as plt

# Load dataset
df = pd.read_csv("corn.csv")

# Drop index column if present
if df.columns[0].lower().startswith("unnamed"):
    df = df.drop(columns=df.columns[0])

# Identify spectral columns (numeric wavelength columns)
spectral_cols = [col for col in df.columns if col.isdigit()]

# Convert column names to numeric wavelengths
wavelengths = [int(col) for col in spectral_cols]

# Extract spectral data
X = df[spectral_cols].values

# Plot
plt.figure(figsize=(10, 6))

for spectrum in X:
    plt.plot(wavelengths, spectrum, alpha=0.3)

plt.xlabel("Wavelength (nm)")
plt.ylabel("Absorbance")
plt.title("NIR Spectra of Corn Samples")

plt.tight_layout()
# JPEG at dpi=150 keeps this figure to a few hundred KB. It is bundled into the
# desktop binary, so size matters; a 300 dpi PNG of the same plot was 1.3 MB.
# JPEG quality goes through pil_kwargs: the older savefig(quality=...) argument
# was deprecated in Matplotlib 3.3 and removed in 3.5.
plt.savefig("corn_spectra.jpg", dpi=150, pil_kwargs={"quality": 90})
plt.show()