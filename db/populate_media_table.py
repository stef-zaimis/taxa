import asyncio

import aiohttp
import psycopg2

import random

DB_PARAMS = {
    "dbname": "col_dwca",
    "user": "postgres",
    "password": "toor",
    "host": "localhost",
    "port": "5432",
}

MAX_CONCURRENT_REQUESTS = 10
RETRY_ATTEMPTS = 5
BACKOFF_TIME = (0.5, 5.0)
REQUEST_TIMEOUT = 15
BATCH_SIZE = 10000 

async def fetch_with_retries(session, url, retries=RETRY_ATTEMPTS):
    for attempt in range(retries):
        try:
            async with session.get(url, timeout=REQUEST_TIMEOUT) as response:
                if response.status == 200:
                    return await response.json()
                elif response.status in {429, 500, 502, 503, 504}:
                    print(f"API error, retrying {attempt +1}/{retries}")
                else:
                    print(f"HTTP {response.status} for {url}")
                    return None
        except (aiohttp.ClientError, asyncio.TimeoutError) as e:
            print(f"Request failed ({e}), attempt {attempt+1}/{retries}")
        
        await asyncio.sleep(random.uniform(*BACKOFF_TIME))
    return None

async def fetch_gbif_taxon_key(session, taxon_name):
    url = f"https://api.gbif.org/v1/species/match?name={taxon_name}"
    data = await fetch_with_retries(session, url)
    return data.get("usageKey") if data else None 

async def fetch_gbif_image_count(session, taxon_key):
    if not taxon_key:
        return 0
    url = f"https://api.gbif.org/v1/occurrence/search?taxonKey={taxon_key}&mediaType=StillImage&limit=0"
    data = await fetch_with_retries(session, url)
    return data.get("count", 0) if data else 0

async def fetch_inaturalist_taxon_id(session, taxon_name):
    url = f"https://api.inaturalist.org/v1/taxa?q={taxon_name}"
    data = await fetch_with_retries(session, url)
    return data["results"][0]["id"] if data and data.get("results") else None 

async def fetch_inaturalist_image_count(session, taxon_id):
    if not taxon_id:
        return 0
    url = f"https://api.inaturalist.org/v1/observations?taxon_id={taxon_id}&photos=true&per_page=0"
    data = await fetch_with_retries(session, url)
    return data.get("total_results", 0) if data else 0

async def get_taxon_media_info(session, taxon_id, taxon_name):
    media_entries = []

    gbif_taxon_key = await fetch_gbif_taxon_key(session, taxon_name)
    gbif_image_count = await fetch_gbif_image_count(session, gbif_taxon_key)

    #if gbif_image_count > 0:
        #print(f"GBIF - Taxon ID: {gbif_taxon_key}, Images: {gbif_image_count}")

    media_entries.append( (taxon_id, "GBIF", gbif_taxon_key, True, gbif_image_count) )
    
    inat_taxon_id = await fetch_inaturalist_taxon_id(session, taxon_name)
    inat_image_count = await fetch_inaturalist_image_count(session, inat_taxon_id)

    #if inat_image_count > 0:
        #print(f"INAT - Taxon ID: {inat_taxon_id}, Images: {inat_image_count}")

    media_entries.append( (taxon_id, "iNaturalist", inat_taxon_id, True, inat_image_count) )

    return media_entries

async def process_taxa(taxon_ids):
    results = []
    semaphore = asyncio.Semaphore(MAX_CONCURRENT_REQUESTS)

    async def worker(taxon_id, taxon_name):
        async with semaphore:
            return await get_taxon_media_info(session, taxon_id, taxon_name)

    async with aiohttp.ClientSession() as session:
        tasks = [worker(taxon_id, taxon_name) for taxon_id, taxon_name in taxon_ids]
        responses = await asyncio.gather(*tasks)

    for sublist in responses:
        results.extend(sublist)
    
    return results

def save_to_db(media_data):
    if not media_data:
        return

    conn = psycopg2.connect(**DB_PARAMS)
    cursor = conn.cursor()
    query = """
    INSERT INTO taxon_media_status (taxon_id, source, source_taxon_key, has_media, media_count)
    VALUES (%s, %s, %s, %s, %s)
    ON CONFLICT (taxon_id) DO UPDATE
    SET source_taxon_key = EXCLUDED.source_taxon_key,
        has_media = EXCLUDED.has_media,
        media_count = EXCLUDED.media_count;
    """

    cursor.executemany(query, media_data)
    conn.commit()
    cursor.close()
    conn.close()


def main():
    offset = 0

    while True:
        conn = psycopg2.connect(**DB_PARAMS)
        cursor = conn.cursor()
        cursor.execute(
            """
            SELECT taxon_id, scientific_name FROM taxon
            WHERE (taxon_id, 'GBIF') NOT IN (SELECT taxon_id, source FROM taxon_media_status)
                OR (taxon_id, 'iNaturalist') NOT IN (SELECT taxon_id, source FROM taxon_media_status)
            ORDER BY taxon_id OFFSET %s LIMIT %s
            """,
            (offset, BATCH_SIZE),
        )

        taxon_data = cursor.fetchall()
        cursor.close()
        conn.close()

        if not taxon_data:
            print("All taxa processed.")
            break

        print(f"Processing {len(taxon_data)} taxa...")

        media_data = asyncio.run(process_taxa(taxon_data))

        if media_data:
            # save_to_db(media_data)
            print(f"Saved {len(media_data)} records to database.")

        offset += BATCH_SIZE


if __name__ == "__main__":
    main()
