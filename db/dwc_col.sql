-- Taxon table schema COL
CREATE TABLE taxon (
    taxon_id VARCHAR(50) PRIMARY KEY,
    parent_id VARCHAR(50),
    accepted_name_id VARCHAR(50),
    original_name_id VARCHAR(50),
    scientific_name_id VARCHAR(50),
    dataset_id VARCHAR(50),
    taxonomic_status VARCHAR(50),
    taxon_rank VARCHAR(50),
    scientific_name TEXT,
    scientific_name_authorship TEXT,
    notho VARCHAR(50),
    generic_name VARCHAR(100),
    infrageneric_epithet VARCHAR(100),
    specific_epithet VARCHAR(100),
    infraspecific_epithet VARCHAR(100),
    cultivar_epithet VARCHAR(100),
    name_according_to TEXT,
    name_published_in TEXT,
    nomenclatural_code VARCHAR(50),
    nomenclatural_status VARCHAR(50),
    kingdom VARCHAR(50),
    phylum VARCHAR(50),
    class_name VARCHAR(50),
    order_name VARCHAR(50),
    superfamily VARCHAR(50),
    family VARCHAR(50),
    subfamily VARCHAR(50),
    tribe VARCHAR(50),
    taxon_remarks TEXT,
    taxon_references TEXT,
    has_media BOOLEAN NOT NULL DEFAULT FALSE,
    gbif_key VARCHAR(50)
);

-- Copying from a Taxon.tsv file of the dwca version of COL, 
-- YOU FIRST NEED TO DO THE FOLLOWING PRE-PROCESSING TO THE DATA IN A TERMINAL:
-- tail -n +2 Taxon.tsv > Taxon_no_header.tsv //removes the header with column names 
-- ALSO MAKE SURE THAT THE Taxon.tsv IS CORRECT, FOR SOME REASON DOWNLOADING THE LATEST LINK IN COL'S MAIN PAGE JUST DOESN'T CONTAIN HALF THE INFO, GO TO THE DEDICATED DOWNLOAD PAGE

\COPY taxon( -- Use \ to execute as the current user (postgres can't see /tmp)
    taxon_id, parent_id, accepted_name_id, original_name_id, scientific_name_id, dataset_id,
    taxonomic_status, taxon_rank, scientific_name, scientific_name_authorship,
    notho, generic_name, infrageneric_epithet, specific_epithet, infraspecific_epithet, cultivar_epithet,
    name_according_to, name_published_in, nomenclatural_code, nomenclatural_status,
    kingdom, phylum, class_name, order_name, superfamily, family, subfamily, tribe,
    taxon_remarks, taxon_references
)
FROM '/tmp/Taxon_no_header.tsv' -- Use a .tsv file with its header removed (there were issues with mismatched quotes so I had to use format text to avoid postgresql trying to parse quotes, especially since they are not used to delimit or for syntax in .tsv files) and copy to tmp to avoid permission issues with postgres
WITH (
  FORMAT text,
  DELIMITER E'\t'
);

-- Remove authorship from scientific name (let's keep it clean for better searches, we'll add authorship later when displaying to the user)
UPDATE taxon
SET scientific_name = trim(replace(scientific_name, scientific_name_authorship, ''))
WHERE scientific_name_authorship IS NOT NULL
  AND scientific_name ILIKE '%' || scientific_name_authorship || '%';

-----------------------------------------------------------------------------------------------------------------------------------------
-- Closure table for ancestors/descendants
CREATE TABLE taxon_closure (
  ancestor_id  VARCHAR(50) NOT NULL,
  descendant_id VARCHAR(50) NOT NULL,
  depth VARCHAR(50) NOT NULL,
  PRIMARY KEY (ancestor_id, descendant_id)
);

-- Indexes for this
CREATE INDEX ON taxon_closure (ancestor_id);
CREATE INDEX ON taxon_closure (descendant_id);
CREATE INDEX ON taxon_closure (depth);

-- Script to populate it
TRUNCATE taxon_closure; -- Clear table

-- Step 1: Insert self relationships (each taxon is its own ancestor with depth 0)
INSERT INTO taxon_closure (ancestor_id, descendant_id, depth)
SELECT taxon_id, taxon_id, 0
FROM taxon;

-- Step 2: Recursively insert ancestor relationships (depth > 0)
WITH RECURSIVE taxon_paths AS (
  -- Base case: direct parent-child relationships
  SELECT 
    taxon_id AS descendant_id,
    parent_id AS ancestor_id,
    1 AS depth
  FROM taxon
  WHERE parent_id IS NOT NULL

  UNION ALL

  -- Recursive step: for each already-found relationship,
  -- get the parent of the current ancestor.
  SELECT
    tp.descendant_id,
    t.parent_id AS ancestor_id,
    tp.depth + 1 AS depth
  FROM taxon_paths tp
  JOIN taxon t ON t.taxon_id = tp.ancestor_id
  WHERE t.parent_id IS NOT NULL
)
INSERT INTO taxon_closure (ancestor_id, descendant_id, depth)
SELECT ancestor_id, descendant_id, depth
FROM taxon_paths;


-----------------------------------------------------------------------------------------------------------------------------------------
-- Populating the has_media and gbif_key fields
-- After running the python scripts, to update ancestors:

UPDATE taxon 
SET has_media = TRUE
WHERE taxon_id IN (
    SELECT ancestor_id
    FROM taxon_closure
    WHERE descendant_id IN (
        SELECT taxon_id FROM taxon WHERE has_media = TRUE
    )
);

-----------------------------------------------------------------------------------------------------------------------------------------
-- To enable fuzzy matching (pg_trm extension)
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Search table to be used in the quicksearch functionality
CREATE TABLE search_index (
	id SERIAL PRIMARY KEY,
	scientific_name TEXT NOT NULL,
	scientific_name_authorship TEXT,
	taxon_rank TEXT NOT NULL,
	search_text TEXT GENERATED ALWAYS AS (
		lower(scientific_name || ' ' || COALESCE(scientific_name_authorship, ''))
	) STORED,
	taxon_id VARCHAR(50) NOT NULL,
	gbif_key VARCHAR(50),
	has_media BOOLEAN NOT NULL DEFAULT FALSE,
	rank_priority INTEGER NOT NULL DEFAULT 999
);

-- Populating it
INSERT INTO search_index (
  scientific_name,
  scientific_name_authorship,
  taxon_rank,
  taxon_id,
  gbif_key,
  has_media,
  rank_priority
)
SELECT
  scientific_name,
  scientific_name_authorship,
  taxon_rank,
  taxon_id,
  gbif_key,
  has_media,
  CASE lower(taxon_rank)
	WHEN 'domain' THEN 1
    WHEN 'realm' THEN 2
    WHEN 'kingdom' THEN 3
    WHEN 'subkingdom' THEN 4
    WHEN 'phylum' THEN 5
    WHEN 'division' THEN 5
    WHEN 'subphylum' THEN 6
    WHEN 'subdivision' THEN 6
    WHEN 'parvphylum' THEN 7
    WHEN 'infraphylum' THEN 8
    WHEN 'superclass' THEN 9
    WHEN 'class' THEN 10
    WHEN 'megaclass' THEN 10
    WHEN 'subterclass' THEN 10
    WHEN 'subclass' THEN 11
    WHEN 'infraclass' THEN 12
    WHEN 'cohort' THEN 12
    WHEN 'subcohort' THEN 13
    WHEN 'superlegion' THEN 14
    WHEN 'legion' THEN 15
    WHEN 'sublegion' THEN 16
    WHEN 'magnorder' THEN 17
    WHEN 'grandorder' THEN 18
    WHEN 'superorder' THEN 19
    WHEN 'order' THEN 21
    WHEN 'nanorder' THEN 22
    WHEN 'parvorder' THEN 22
    WHEN 'suborder' THEN 23
    WHEN 'infraorder' THEN 24
    WHEN 'section' THEN 24
    WHEN 'subsection' THEN 25
    WHEN 'superfamily' THEN 26
    WHEN 'epifamily' THEN 27
    WHEN 'family' THEN 28
    WHEN 'subfamily' THEN 29
    WHEN 'infrafamily' THEN 30
    WHEN 'supertribe' THEN 31
    WHEN 'tribe' THEN 32
    WHEN 'subtribe' THEN 32
    WHEN 'infratribe' THEN 33
    WHEN 'genus' THEN 35
    WHEN 'infrageneric name' THEN 36
    WHEN 'subgenus' THEN 36
    WHEN 'series' THEN 39
    WHEN 'subseries' THEN 40
    WHEN 'species' THEN 41
    WHEN 'species aggregate' THEN 42
    WHEN 'complex' THEN 42
    WHEN 'subspecies' THEN 43
    WHEN 'variety' THEN 44
    WHEN 'subvariety' THEN 45
    WHEN 'form' THEN 47
    WHEN 'subform' THEN 48
    WHEN 'forma specialis' THEN 49
    WHEN 'aberration' THEN 50
    WHEN 'lusus' THEN 50
    WHEN 'mutatio' THEN 50
    WHEN 'proles' THEN 50
    WHEN 'morph' THEN 50
    WHEN 'infraspecific name' THEN 50
    WHEN 'infrasubspecific name' THEN 51
    WHEN 'grex' THEN 52
    WHEN 'strain' THEN 52
    WHEN 'lineage' THEN 999
    WHEN 'clade' THEN 999
    WHEN 'group' THEN 999
    WHEN 'unranked' THEN 999
    WHEN 'other' THEN 999
    ELSE 999
  END
FROM taxon;

-- Rank search table
CREATE TABLE rank_index AS 
SELECT DISTINCT taxon_rank AS rank
FROM taxon
WHERE taxon_rank IS NOT NULL;
-----------------------------------------------------------------------------------------------------------------------------------------
--------------------------------------------------------- Indices:
CREATE INDEX idx_taxon_rank_media ON taxon (lower(taxon_rank), has_media);
CREATE INDEX idx_closure_query_fast ON taxon_closure (ancestor_id) INCLUDE (descendant_id);
CREATE INDEX idx_taxon_rank_name ON taxon (taxon_id, scientific_name, taxon_rank);
CREATE INDEX idx_taxon_rank_name_lower ON taxon (lower(taxon_rank), lower(scientific_name));

-- Taxon lookup
CREATE INDEX idx_taxon_rank_name_lower ON taxon (lower(taxon_rank), lower(scientific_name));
CREATE INDEX idx_taxon_rank_media_composite ON taxon (lower(taxon_rank), scientific_name) WHERE has_media = TRUE;
CREATE INDEX idx_taxon_name_media ON taxon (scientific_name) WHERE has_media = TRUE;
CREATE INDEX idx_taxon_gbif_update ON taxon (scientific_name);
CREATE INDEX idx_taxon_id ON taxon (taxon_id);

-- Closure table
CREATE INDEX idx_taxon_closure_ancestor ON taxon_closure (ancestor_id);
CREATE INDEX idx_taxon_closure_ancestor_desc ON taxon_closure (ancestor_id, descendant_id);

-- taxon Search table
CREATE INDEX idx_search_text_prefix ON search_index (lower(search_text));
CREATE INDEX idx_search_text_trgm ON search_index USING gin (search_text gin_trgm_ops); -- Make sure to have the extension set up
CREATE INDEX idx_search_rank_media ON search_index (rank_priority, has_media);

CREATE INDEX IF NOT EXISTS idx_search_index_lower_pattern
  ON search_index (lower(search_text) text_pattern_ops);

CREATE INDEX IF NOT EXISTS idx_search_index_lower_trgm
  ON search_index USING gin (lower(search_text) gin_trgm_ops);

-- rank search table
CREATE INDEX idx_rank_search ON rank_index (rank);

-- Correct taxon selection
-- support fast filtering by ancestor_id
CREATE INDEX IF NOT EXISTS idx_taxon_closure_ancestor
  ON taxon_closure (ancestor_id);

-- fast case‑insensitive rank lookup
CREATE INDEX IF NOT EXISTS idx_taxon_rank_lower
  ON taxon (lower(taxon_rank));

-- fast has_media filter
CREATE INDEX IF NOT EXISTS idx_taxon_has_media
  ON taxon (has_media)
 WHERE has_media = TRUE;

-- Random taxa selection
-- so lookups by ancestor_id are index-only
CREATE INDEX IF NOT EXISTS idx_taxon_closure_ancestor
  ON taxon_closure (ancestor_id);

CREATE INDEX IF NOT EXISTS idx_taxon_rank_lower
  ON taxon (lower(taxon_rank));

-- has_media filter will be very fast
CREATE INDEX IF NOT EXISTS idx_taxon_has_media
  ON taxon (has_media)
  WHERE has_media = TRUE;

----------------------------------------------------MEDIA TABLES -> LIKELY USELESS, BUT STILL KEEPING THEM IN CASE----------------------
-- Old media table
CREATE TABLE taxon_media_status (
    id SERIAL PRIMARY KEY,
    taxon_id VARCHAR(255) REFERENCES taxon(taxon_id),
    source VARCHAR(50) NOT NULL,
    source_taxon_key VARCHAR(50),
    has_media BOOLEAN NOT NULL DEFAULT FALSE,
    media_count INT DEFAULT 0,
    UNIQUE (taxon_id, source)
);

-- Add a "has_media" column to the taxon table for quick queries
ALTER TABLE taxon ADD COLUMN has_media BOOLEAN NOT NULL DEFAULT FALSE;

-- Update the has_media column in the taxon table every time the media table is updated:
UPDATE taxon
SET has_media = EXISTS (
	SELECT 1 FROM taxon_media_status
	WHERE taxon_media_status.taxon_id = taxon.taxon_id AND has_media = TRUE

-----------------------------------------------------------------------------------------------------------------------------------------------------------
-- Full media table from GBIF
-- THIS IS USELESS, THE SIMPLE CSV DOESNT CONTAIN LINKS TO IMAGES, SO WE WONT ACTUALLY MAKE THIS TABLE AT ALL
CREATE TABLE taxon_media (
    gbifID BIGINT PRIMARY KEY,
    datasetKey VARCHAR(50),
    occurrenceID TEXT UNIQUE,
    kingdom VARCHAR(50),
    phylum VARCHAR(50), 
	class VARCHAR(50),   
	order VARCHAR(50),  
	family VARCHAR(50),
	genus VARCHAR(50),  
	species VARCHAR(50),
	infraspecificEpithet    
	taxonRank       
	scientificName  
	verbatimScientificName  
	verbatimScientificNameAuthorship        
	countryCode    
	locality 
	stateProvince   
	occurrenceStatus        
	individualCount 
	publishingOrgKey        
	decimalLatitude 
	decimalLongitude        
	coordinateUncertaintyInMeters   
	coordinatePrecision     
	elevation       
	elevationAccuracy       
	depth  
	depthAccuracy    
	eventDate       
	day     
	month   
	year    
	taxonKey        
	speciesKey      
	basisOfRecord   
	institutionCode
	collectionCode  
	catalogNumber   
	recordNumber    
	identifiedBy    
	dateIdentified  
	license 
	rightsHolder    
	recorded
	By      
	typeStatus      
	establishmentMeans      
	lastInterpreted 
	mediaType       
	issue
);

-- Copying from a TaxonMedia.tsv file of the csv (simple) version of the GBIF occurrence download
-- YOU FIRST NEED TO DO THE FOLLOWING PRE-PROCESSING TO THE DATA IN A TERMINAL:
-- tail -n +2 TaxonMedia.tsv > TaxonMedia_no_header.tsv //removes the header with column names 

\COPY taxon( -- Use \ to execute as the current user (postgres can't see /tmp)
    taxon_id, parent_id, accepted_name_id, original_name_id, scientific_name_id, dataset_id,
    taxonomic_status, taxon_rank, scientific_name, scientific_name_authorship,
    notho, generic_name, infrageneric_epithet, specific_epithet, infraspecific_epithet, cultivar_epithet,
    name_according_to, name_published_in, nomenclatural_code, nomenclatural_status,
    kingdom, phylum, class_name, order_name, superfamily, family, subfamily, tribe,
    taxon_remarks, taxon_references
)
FROM '/tmp/TaxonMedia_no_header.tsv' -- Use a .tsv file with its header removed (there were issues with mismatched quotes so I had to use format text to avoid postgresql trying to parse quotes, especially since they are not used to delimit or for syntax in .tsv files) and copy to tmp to avoid permission issues with postgres
WITH (
  FORMAT text,
  DELIMITER E'\t'
);

