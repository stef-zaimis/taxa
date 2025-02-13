-- Taxon table schema
postgres=# CREATE TABLE taxon (
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
