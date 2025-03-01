import requests
import psycopg2
import re
from fuzzywuzzy import process

DB PARAMS = {
        "dbname": "col_dwca",
        "user": "postgres",
        "password": "toor",
        "host": "localhost",
        "port": "5432",
}

GBIF_URL = "https://api.gbif.org/v1/occurrence/search?mediaType=StillImage&limit=100000"
INAT_URL = "https://api.inaturalist.org/v1/taxa?is_active=true&photos=true&per_page=100"

BATCH_SIZE = 50000

def strip_authorship(name, authorship):
    if not name:
        return None
    if authorship and authorship in name:
        return name.replace(authorship, "").strip()
    return re.sub(r"\s+\(,+\)$", "", name).strip()

def fetch_local_taxa():
    conn = psycopg2.connect(**DB_PARAMS)
    cursor = conn.cursor()

    query = "SELECT taxon_id, scientific_name, scientific_name_authorship, accepted_name_id FROM taxon"
    cursor.execute(query)

    local_taxa = {}
    for taxon_id, name, authorship, accepted_name in cursor.fetchall():
        stripped_name = strip_authorship(name, authorship)
        local_taxa[stripped_name] = taxon_id

        if accepted_name:
            local_taxa[strip_authorship(accepted_name, None)] = taxon_id

    cursor.close()
    conn.close()
    return local_taxa

def fetch_gbif_species_with_images():
    results = []
    offset = 0

    while True:
        url=f"{GBIF_URL}&offset={offset}"
        response = requests.get(url).json()

        if "results" in response:
            for record in response["results"]:
                if "speciesKey" in record and "scientificName" in record:
                    results.append((record["speciesKey"],record["scientificName"]))

        offset += 100000

        if "endOfRecords" in response and response["endOfRecords"]:
            break
    return results

def fuzzy_match_name(name, local_taxa, threshold=90):
    match, score = process.extractOne(name, local_taxa.keys())
    return local_taxa[match] if score >= threshold else None

def match_taxa(external_species, source, local_taxa):
    matched_data = []

    for external_id, external_name in external_species:
        stripped_name = strip_authorship(external_name, None)
        local_taxon_id = local_taxa.get(stripped_name)

        if not local_taxon_id:
            local_taxon_id = fuzzy_match_name(stripped_name, local_taxa)

        if local_taxon_id:
            matched_data.append((local_taxon_id, source, external_id, True))
        else:
            print(f"No match for {external_name} ({source})")
    
    return matched_data


def save_to_db(species_list, source):
    if not species_list:
        print(f"No {source} species with images found.")
        return
    conn = psycopg2.connect(**DB_PARAMS)
    cursor = conn.cursor()

    query = """
    INSERT INTO taxon_media_status (taxon_id, source, source_taxon_key, has_media)
    VALUES (%s, %s, %s, %s)
    ON CONFLICT (taxon_id, source) DO UPDATE
    SET has_media = EXCLUDED.has_media;
    """

    cursor.executemany(query, matched_species)
    conn.commit()
    cursor.close()
    conn.close()
    print(f"{len(matched_species)} {source} taxa with images saved to database.")

def main():
    print("Fetching data from db")
    local_taxa=fetch_local_taxa()

    print("Fetching GBIF species with images")
    gbif_species = fetch_gbif_species_with_images()

    print("Matching GBIF taxa")
    matched_gbif = match_taxa(gbif_species, "GBIF", local_taxa)

    print("Saving GBIF data to database")
    save_to_db(matched_gbif, "GBIF")

    print("Done")

if __name__ == "__main__":
    main()

