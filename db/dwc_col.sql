-- Taxon table schema
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
    scientific_name_authorship VARCHAR(255),
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
    taxon_references TEXT
);

-- Copying from a Taxon.tsv file
COPY taxon(
    taxon_id, parent_id, accepted_name_id, original_name_id, scientific_name_id, dataset_id,
    taxonomic_status, taxon_rank, scientific_name, scientific_name_authorship,
    notho, generic_name, infrageneric_epithet, specific_epithet, infraspecific_epithet, cultivar_epithet,
    name_according_to, name_published_in, nomenclatural_code, nomenclatural_status,
    kingdom, phylum, class_name, order_name, superfamily, family, subfamily, tribe,
    taxon_remarks, taxon_references
)
FROM '/home/stefanos/Desktop/Uni/thesis/2025col-darwin/Taxon_noheader.tsv' -- Use a .tsv file with its header removed (there were issues with mismatched quotes so I had to use format text to avoid postgresql trying to parse quotes, especially since they are not used to delimit or for syntax in .tsv files)
WITH (
  FORMAT text,
  DELIMITER E'\t'
);



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
