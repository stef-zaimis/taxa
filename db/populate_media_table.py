import asyncio

import aiohttp
import psycopg2

DB_PARAMS = {
    "dbname": "col_dwca_db",
    "user": "postgres",
    "password": "toor",
    "host": "localhost",
    "port": "5432",
}


async def fetch_gbif_media(session, taxon_id):
    url = f"https://api.gbif.org/v1/occurrence/search?taxonKey={taxon_id}&mediaType=StillImage"
    async with session.get(url) as response:
        if response.status == 200:
            data = await response.json()
            for result in data.get("results", []):
                for media in result.get("media", []):
                    if media.get("identifier"):
                        return (
                            taxon_id,
                            True,
                            media["identifier"],
                            "GBIF",
                            media.get("license", "Unknown"),
                        )
    return (taxon_id, False, None, None, None)


async def process_taxa(taxon_ids):
    tasks = []
    async with aiohttp.ClientSession() as session:
        for taxon_id in taxon_ids:
            tasks.append(fetch_gbif_media(session, taxon_id))
        results = await asyncio.gather(*tasks)
    return results


def save_to_db(media_data):
    conn = psycopg2.connect(**DB_PARAMS)
    cursor = conn.cursor()
    query = """
    INSERT INTO media (taxon_id, has_media, media_url, source, license)
    VALUES (%s, %s, %s, %s, %s)
    ON CONFLICT (taxon_id) DO UPDATE
    SET has_media = EXCLUDED.has_media,
        media_url = EXCLUDED.media_url,
        source = EXCLUDED.source,
        license = EXCLUDED.license;
    """
    cursor.executemany(query, media_data)
    conn.commit()
    cursor.close()
    conn.close()


def main():
    batch_size = 10000
    offset = 0

    while True:
        conn = psycopg2.connect(*DB_PARAMS)
        cursor = conn.cursor()
        cursor.execute("SELECT gbif_id FROM ")
