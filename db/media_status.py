import psycopg2
import pandas as pd
import re

# Database connection details
DB_CONFIG = {
    "dbname": "taxa",
    "user": "postgres",
    "password": "toor",
    "host": "localhost",
    "port": "5432",
}

GBIF_LOOKUP_PATH = "gbif_taxon_lookup.csv"

# Function to clean and standardize scientific names
def clean_scientific_name(name):
    """Process scientific names based on filtering rules."""
    if not name or not isinstance(name, str):
        return None  # Skip empty or invalid names

    name = name.strip()

    # **1 Ignore anything starting with 'BOLD:'**
    if name.startswith("BOLD:"):
        return None  # Skip

    # **2 Ignore hybrids (names containing '×')**
    if "×" in name:
        return None  # Skip

    # **3 Extract the first two words (Genus + Species)**
    words = name.split()
    if len(words) < 2:
        return None  # Skip names without genus & species
    
    cleaned_name = " ".join(words[:2])  # Keep only Genus + Species

    # **4 Remove special characters**
    cleaned_name = re.sub(r"[^a-zA-Z0-9\s]", "", cleaned_name)  # Keep only letters, numbers, spaces
    
    return cleaned_name.strip() if cleaned_name else None

# **Step 1: Load GBIF lookup table**
print("Loading GBIF lookup data...")
gbif_data = pd.read_csv(GBIF_LOOKUP_PATH, sep=r"\|\|", dtype=str, engine="python")

# Apply cleaning function and remove invalid entries
gbif_data["scientific_name"] = gbif_data["scientific_name"].apply(clean_scientific_name)
gbif_data = gbif_data.dropna().drop_duplicates()  # Remove empty & duplicate names

# Convert to dictionary for fast lookup
gbif_dict = dict(zip(gbif_data["scientific_name"], gbif_data["gbif_key"]))
print(f"Loaded {len(gbif_dict)} valid GBIF taxa.")

# **Step 2: Connect to the PostgreSQL database**
print("Fetching taxon data from database...")
conn = psycopg2.connect(**DB_CONFIG)
cursor = conn.cursor()

# **Step 3: Fetch scientific names from `taxon` table**
cursor.execute("SELECT taxon_id, scientific_name FROM taxon;")
taxa = cursor.fetchall()

update_count = 0
update_queries = []

# **Step 4: Process and update matching taxa**
for taxon_id, taxon_name in taxa:
    cleaned_name = clean_scientific_name(taxon_name)
    
    if cleaned_name in gbif_dict:
        gbif_key = gbif_dict[cleaned_name]
        update_queries.append((gbif_key, taxon_id))
        update_count += 1

# **Step 5: Perform batch updates for efficiency**
if update_queries:
    query = "UPDATE taxon SET has_media = TRUE, gbif_key = %s WHERE taxon_id = %s;"
    cursor.executemany(query, update_queries)

# Commit changes
conn.commit()
cursor.close()
conn.close()

print(f"Updated {update_count} taxa with GBIF media data.")

