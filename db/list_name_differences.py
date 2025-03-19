import psycopg2
import pandas as pd
import re
from fuzzywuzzy import process

# Database connection details
DB_CONFIG = {
    "dbname": "taxa",
    "user": "postgres",
    "password": "toor",
    "host": "localhost",
    "port": "5432",
}

GBIF_CSV_PATH = "databases/gbif/occurrence_downloads/open_licensed_still_images.csv"
CHUNK_SIZE = 500000  # Process CSV in chunks
FUZZY_THRESHOLD = 90  # Minimum match score for fuzzy matching

# Function to strip authorship from scientific names
def strip_authorship(name):
    if not name:
        return None
    # Remove content within parentheses (common format for authorship)
    name = re.sub(r"\s*\([^)]*\)", "", name)
    # Remove author names (e.g., "Linnaeus, 1758")
    name = re.sub(r"\s+[A-Z][a-z]+,\s*\d{4}", "", name)
    name = re.sub(r"\s+[A-Z]+\s*\d{4}", "", name)  # Handles all caps authorships
    return name.strip()

# Load GBIF scientific names in chunks
gbif_names = set()
print("Processing GBIF CSV in chunks...")
for chunk in pd.read_csv(GBIF_CSV_PATH, sep="\t", usecols=["scientificName"], chunksize=CHUNK_SIZE):
    cleaned_names = chunk["scientificName"].dropna().apply(strip_authorship)
    gbif_names.update(cleaned_names.unique())

print(f"Loaded {len(gbif_names)} unique (authorship-stripped) scientific names from GBIF.")

# Fetch taxon table scientific names from PostgreSQL
print("Fetching taxon names from database...")
conn = psycopg2.connect(**DB_CONFIG)
cursor = conn.cursor()
cursor.execute("SELECT DISTINCT scientific_name FROM taxon;")
taxon_names = {strip_authorship(row[0]) for row in cursor.fetchall()}
cursor.close()
conn.close()

print(f"Loaded {len(taxon_names)} unique (authorship-stripped) scientific names from taxon database.")

# Find differences (after stripping authorship)
names_only_in_taxon = taxon_names - gbif_names
names_only_in_gbif = gbif_names - taxon_names

# Apply fuzzy matching for better comparison
def fuzzy_match_name(name, name_set):
    match, score = process.extractOne(name, name_set)
    return match if score >= FUZZY_THRESHOLD else None

# I removed fuzzy matching cause it was super slow, might use if absolutely necessary
#fuzzy_mismatches_gbif = {name for name in names_only_in_gbif if not fuzzy_match_name(name, taxon_names)}
#fuzzy_mismatches_taxon = {name for name in names_only_in_taxon if not fuzzy_match_name(name, gbif_names)}

# Write names that exist only in taxon database
with open("only_in_taxon.txt", "w") as f:
    f.write(f"Scientific names in taxon table but NOT in GBIF ({len(names_only_in_taxon)}):\n")
    f.writelines(f"{name}\n" for name in sorted(names_only_in_taxon))

# Write names that exist only in GBIF dataset
with open("only_in_gbif.txt", "w") as f:
    f.write(f"Scientific names in GBIF but NOT in taxon table ({len(names_only_in_gbif)}):\n")
    f.writelines(f"{name}\n" for name in sorted(names_only_in_gbif))

print(f"Comparison complete. Results saved to 'only_in_taxon.txt' and 'only_in_gbif.txt'.")

