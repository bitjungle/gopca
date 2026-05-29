import pandas as pd
from scipy.io import arff

# === Load ARFF file ===
data, meta = arff.loadarff("EEG Eye State.arff")

# Convert to DataFrame
df = pd.DataFrame(data)

# === Decode byte strings (ARFF quirk) ===
for col in df.select_dtypes([object]):
    df[col] = df[col].str.decode("utf-8")

# === Rename target column ===
if "eyeDetection" in df.columns:
    df = df.rename(columns={"eyeDetection": "eye_state"})

# === Map eye state labels ===
df["eye_state"] = df["eye_state"].map({
    "0": "open",
    "1": "closed"
})

# === Create time column (seconds) ===
sampling_rate = 128  # Hz (samples per second)
time = [i / sampling_rate for i in range(len(df))]

# Insert as first column
df.insert(0, "time", time)

# === Move target column to the end ===
if "eye_state" in df.columns:
    cols = [c for c in df.columns if c != "eye_state"] + ["eye_state"]
    df = df[cols]

# === Save to CSV ===
df.to_csv("eeg_eye_state.csv", index=False)

print("Conversion complete: eeg_eye_state.csv")