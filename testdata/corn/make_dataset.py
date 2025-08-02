import pandas as pd

# Load spectral data
spectra_df = pd.read_csv('corn_m5spec.csv', sep=',')
#print(spectra_df)

# Properties: Moisture, Oil, Protein, Starch
propvals_df = pd.read_csv('corn_propvals.csv')
#print(propvals_df)

# Merge spectral data with properties
merged_df = pd.merge(propvals_df, spectra_df, left_index=True, right_index=True, how='inner')
# Append '#target' to the Moisture,Oil,Protein,Starch columns
for col in ['Moisture', 'Oil', 'Protein', 'Starch']:
    merged_df.rename(columns={col: f'{col}#target'}, inplace=True)
merged_df.to_csv('corn_target.csv', index=True)

# Make categorigal properties Low/Mid/High for Moisture, Oil, Protein, Starch
for prop in ['Moisture', 'Oil', 'Protein', 'Starch']:
    propvals_df[prop] = pd.cut(propvals_df[prop], bins=3, labels=['Low', 'Mid', 'High'])
#print(propvals_df)

# Merge spectral data with property categories
merged_category_data_df = pd.merge(propvals_df, merged_df, left_index=True, right_index=True, how='inner')

# Save the merged data to a new CSV file
merged_category_data_df.to_csv('corn.csv', index=True)
