import asyncio

import aiohttp
import psycopg2

DB_PARAMS = {
    "dbname": "col_dwca",
    "user": "postgres",
    "password": "toor",
    "host": "localhost",
    "port": "5432",
}


async def fetch_gbif_taxon_key(session, taxon_name):
    url = f"https://api.gbif.org/v1/species/match?name={taxon_name}"
    async with session.get(url) as response:
        data = await response.json()
        return data.get("usageKey")
    return None


async def fetch_gbif_image_count(session, taxon_key):
    url = f"https://api.gbif.org/v1/occurrence/search?taxonKey={taxon_key}&mediaType=StillImage&limit=0"
    async with session.get(url) as response:
        if response.status == 200:
            data = await response.json()
            return data.get("count", 0)
    return 0

async def fetch_inaturalist_taxon_id(session, taxon_name):
    url = f"https://api.inaturalist.org/v1/taxa?q={taxon_name}"
    async with session.get(url) as response:
        if response.status == 200:
            data = await response.json()
            if data["results"]:
                return data["results"][0]["id"]
    return None

async def fetch_inaturalist_image_count(session, taxon_id):
    return

async def get_taxon_media_info(session, taxon_id, taxon_name):
    taxon_key = await fetch_gbif_taxon_key(session, taxon_name)
    if not taxon_key:
        return (taxon_id, "GBIF", None, False, 0)

    taxon_image_count = await fetch_gbif_image_count(session, taxon_key)
    if taxon_image_count == 0:
        return (taxon_id, "GBIF", taxon_key, False, 0)
    else:
        print(taxon_image_count)
    return (taxon_id, "GBIF", taxon_key, True, taxon_image_count)


async def process_taxa(taxon_ids):
    tasks = []
    async with aiohttp.ClientSession() as session:
        for taxon_id, taxon_name in taxon_ids:
            tasks.append(get_taxon_media_info(session, taxon_id, taxon_name))
        results = await asyncio.gather(*tasks)
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
    batch_size = 10000
    offset = 0

    while True:
        conn = psycopg2.connect(**DB_PARAMS)
        cursor = conn.cursor()
        cursor.execute(
            """
            SELECT taxon_id, scientific_name FROM taxon
            WHERE (taxon_id, 'GBIF') NOT IN (SELECT taxon_id, source FROM taxon_media_status)
            ORDER BY taxon_id OFFSET %s LIMIT %s
            """,
            (offset, batch_size),
        )

        taxon_data = cursor.fetchall()
        cursor.close()
        conn.close()

        if not taxon_data:
            print("All taxa processed.")
            break

        print(f"Processing {len(taxon_data)} taxa...")

        loop = asyncio.get_event_loop()
        media_data = loop.run_until_complete(process_taxa(taxon_data))

        if media_data:
            # save_to_db(media_data)
            print(f"Found {media_data}")
            print(f"Saved {len(media_data)} records to database.")

        offset += batch_size


if __name__ == "__main__":
    main()
