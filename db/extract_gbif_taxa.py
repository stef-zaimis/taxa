import pandas as pd

GBIF_CSV_PATH = "databases/gbif/occurrence_downloads/open_licensed_still_images.csv"
OUTPUT_LOOKUP_PATH = "gbif_taxon_lookup.csv"

CHUNK_SIZE = 2_000_000  # Larger chunks improve performance

# Use a set instead of a dictionary to collect unique scientific names
unique_names = set()

print("Processing GBIF CSV in chunks...")

for chunk in pd.read_csv(GBIF_CSV_PATH, sep="\t", usecols=["scientificName", "taxonKey"], dtype=str, chunksize=CHUNK_SIZE, iterator=True):
    chunk = chunk.dropna(subset=["scientificName", "taxonKey"])  # Remove rows with NaN
    
    # Add unique scientific names to set (very fast compared to dict)
    unique_names.update(zip(chunk["scientificName"].str.strip(), chunk["taxonKey"].str.strip()))

print(f"Loaded {len(unique_names)} unique taxa from GBIF.")

# Convert to dictionary after collecting all unique values
unique_taxa = dict(unique_names)

# Save to file using "||" as separator
print(f"Saving to {OUTPUT_LOOKUP_PATH}...")
with open(OUTPUT_LOOKUP_PATH, "w", encoding="utf-8") as f:
    f.write("scientific_name||gbif_key\n")  # Add header
    for name, key in unique_taxa.items():
        f.write(f"{name}||{key}\n")

print("Lookup table created successfully!")

