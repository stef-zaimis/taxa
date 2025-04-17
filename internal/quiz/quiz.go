package quiz

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stef-zaimis/taxa/internal/gbif"
)

// GenerateQuestion assembles a quiz question by
// 1) loading _all_ descendants in one pass,
// 2) splitting them into media vs non-media slices,
// 3) picking correct+N distractors in-memory,
// 4) fetching those rows in a single small query,
// 5) getting the GBIF image for the correct one.
func GenerateQuestion(pool *pgxpool.Pool, parentRank, parentName, targetRank string, optionCount int) (Question, error) {
	// 1) Fetch every descendant ID under ancestor+rank
	descendantIDs, mediaIDs, ancestorID, err := fetchDescendantInfo(pool, parentRank, parentName, targetRank)
	if err != nil {
		return Question{}, fmt.Errorf("fetchDescendantInfo: %w", err)
	}

	// Must have at least one media-enabled taxon
	if len(mediaIDs) == 0 {
		return Question{}, fmt.Errorf("no taxa with media found under %s/%s/%s", parentRank, parentName, targetRank)
	}

	// 2) Pick correct taxon ID at random from mediaIDs
	correctID := mediaIDs[rand.Intn(len(mediaIDs))]

	// 3) Build slice of candidate distractor IDs (exclude correct)
	others := make([]string, 0, len(descendantIDs)-1)
	for _, id := range descendantIDs {
		if id != correctID {
			others = append(others, id)
		}
	}
	if len(others) < optionCount-1 {
		return Question{}, fmt.Errorf("not enough distractors: have %d, need %d", len(others), optionCount-1)
	}

	// Shuffle and take first N distractors
	rand.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
	distractorIDs := others[:optionCount-1]

	// 4) Fetch all chosen rows (correct + distractors) in one small query
	pickIDs := append(distractorIDs, correctID)
	taxaRows, err := fetchTaxaByIDs(pool, pickIDs)
	if err != nil {
		return Question{}, fmt.Errorf("fetchTaxaByIDs: %w", err)
	}

	// 5) Find correctTaxon in the returned slice and get its GBIF image
	var correctTaxon Taxon
	for _, t := range taxaRows {
		if t.TaxonID == correctID {
			correctTaxon = t
			key, img := gbif.GetImage(t.ScientificName, t.Authorship, t.Rank)
			correctTaxon.GBIFKey = key
			if img == "" || strings.Contains(img, "localhost") {
				return Question{}, fmt.Errorf("no valid image for taxon %s", t.ScientificName)
			}
			break
		}
	}

	// Shuffle final options order
	rand.Shuffle(len(taxaRows), func(i, j int) { taxaRows[i], taxaRows[j] = taxaRows[j], taxaRows[i] })

	// Locate correctIndex
	correctIndex := -1
	for i, t := range taxaRows {
		if t.TaxonID == correctID {
			correctIndex = i
			break
		}
	}

	return Question{
		ImageURL:      "TODO_IMAGE_URL_NOT_USED_HERE",
		Options:       taxaRows,
		CorrectIndex:  correctIndex,
		CorrectAnswer: correctTaxon,
	}, nil
}

// fetchDescendantInfo loads ALL descendants under the given ancestor+rank
// returning three things in one pass: all IDs, media-enabled IDs, and the ancestorID.
func fetchDescendantInfo(pool *pgxpool.Pool, parentRank, parentName, targetRank string) (allIDs, mediaIDs []string, ancestorID string, err error) {
	ctx := context.Background()
	// 1) lookup ancestorID
	err = pool.QueryRow(ctx, `
		SELECT taxon_id
		  FROM taxon
		 WHERE lower(taxon_rank) = $1
		   AND lower(scientific_name) = $2
		 LIMIT 1`, parentRank, parentName).Scan(&ancestorID)
	if err != nil {
		return
	}

	// 2) one big pass: pull every descendant_id + has_media flag
	rows, err := pool.Query(ctx, `
		SELECT c.descendant_id, t.has_media
		  FROM taxon_closure c
		  JOIN taxon t ON t.taxon_id = c.descendant_id
		 WHERE c.ancestor_id       = $1
		   AND lower(t.taxon_rank) = lower($2)
		`, ancestorID, targetRank)
	if err != nil {
		return
	}
	defer rows.Close()

	for rows.Next() {
		var id string
		var hasMedia bool
		if err = rows.Scan(&id, &hasMedia); err != nil {
			return
		}
		allIDs = append(allIDs, id)
		if hasMedia {
			mediaIDs = append(mediaIDs, id)
		}
	}
	err = rows.Err()
	return
}

// fetchTaxaByIDs gets the full Taxon rows for a small slice of IDs in one go.
func fetchTaxaByIDs(pool *pgxpool.Pool, ids []string) ([]Taxon, error) {
	ctx := context.Background()
	rows, err := pool.Query(ctx, `
		SELECT taxon_id,
		       scientific_name,
		       scientific_name_authorship,
		       taxon_rank,
		       has_media,
		       taxonomic_status,
		       kingdom,
		       phylum,
		       class_name,
		       order_name,
		       superfamily,
		       family,
		       subfamily,
		       tribe
		  FROM taxon
		 WHERE taxon_id = ANY($1)
	`, ids)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []Taxon
	for rows.Next() {
		var t Taxon
		if err := rows.Scan(
			&t.TaxonID,
			&t.ScientificName,
			&t.Authorship,
			&t.Rank,
			&t.HasMedia,
			&t.Status,
			&t.Kingdom,
			&t.Phylum,
			&t.Class,
			&t.Order,
			&t.SuperFamily,
			&t.Family,
			&t.SubFamily,
			&t.Tribe,
		); err != nil {
			return nil, err
		}
		result = append(result, t)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}



// LEGACY THINGS, KEEPING THEM JUST IN CASE
// getTaxonWithMedia picks one random taxon WITH has_media = TRUE 
// under the given ancestor + rank, **without** COUNT or OFFSET.
func getTaxonWithMedia(pool *pgxpool.Pool, parentRank, parentName, targetRank string) (Taxon, string, error) {
  ctx := context.Background()

  // 1) find the ancestorID
  var ancestorID string
  err := pool.QueryRow(ctx, `
    SELECT taxon_id
      FROM taxon
     WHERE lower(taxon_rank) = $1
       AND lower(scientific_name) = $2
     LIMIT 1
  `, parentRank, parentName).Scan(&ancestorID)
  if err != nil {
    return Taxon{}, "", fmt.Errorf("ancestor lookup: %w", err)
  }

  // 2) pull **all** descendant IDs that have_media = TRUE
  rows, err := pool.Query(ctx, `
    SELECT c.descendant_id
      FROM taxon_closure c
      JOIN taxon t 
        ON t.taxon_id = c.descendant_id
     WHERE c.ancestor_id      = $1
       AND lower(t.taxon_rank) = lower($2)
       AND t.has_media         = TRUE
  `, ancestorID, targetRank)
  if err != nil {
    return Taxon{}, "", fmt.Errorf("fetch media IDs: %w", err)
  }
  defer rows.Close()

  var ids []string
  for rows.Next() {
    var id string
    if err := rows.Scan(&id); err != nil {
      return Taxon{}, "", fmt.Errorf("scan id: %w", err)
    }
    ids = append(ids, id)
  }
  if err := rows.Err(); err != nil {
    return Taxon{}, "", fmt.Errorf("rows err: %w", err)
  }
  if len(ids) == 0 {
    return Taxon{}, "", fmt.Errorf("no taxa with media found")
  }

  // 3) pick one at random
  chosenID := ids[rand.Intn(len(ids))]

  // 4) fetch that taxon’s full record
  var t Taxon
  err = pool.QueryRow(ctx, `
    SELECT taxon_id,
           scientific_name,
           scientific_name_authorship,
           taxon_rank,
           has_media,
           taxonomic_status,
           kingdom,
           phylum,
           class_name,
           order_name,
           superfamily,
           family,
           subfamily,
           tribe
      FROM taxon
     WHERE taxon_id = $1
  `, chosenID).Scan(
    &t.TaxonID, &t.ScientificName, &t.Authorship, &t.Rank, &t.HasMedia,
    &t.Status, &t.Kingdom, &t.Phylum, &t.Class, &t.Order,
    &t.SuperFamily, &t.Family, &t.SubFamily, &t.Tribe,
  )
  if err != nil {
    return Taxon{}, "", fmt.Errorf("fetch chosen taxon: %w", err)
  }

  return t, ancestorID, nil
}

// getTaxonWithMedia fetches a single taxon with has_media = TRUE
func getTaxonWithMediaOffsetTrick(pool *pgxpool.Pool, parentRank, parentName, targetRank string) (Taxon, string, error) {
	ctx := context.Background()

	ancestorQuery := `
		SELECT taxon_id
		FROM taxon
		WHERE lower(taxon_rank) = $1 AND lower(scientific_name) = $2
		LIMIT 1
	`

	var ancestorID string
	err := pool.QueryRow(ctx, ancestorQuery, parentRank, parentName).Scan(&ancestorID)
	if err != nil {
		return Taxon{}, "", fmt.Errorf("failed to find ancestorID: %v", err)
	}

	countQuery := `
		SELECT COUNT(*)
		FROM taxon_closure c
		JOIN taxon t ON t.taxon_id = c.descendant_id
		WHERE c.ancestor_id = $1
		AND lower(t.taxon_rank) = lower($2)
		AND t.has_media = TRUE
	`

	var count int

	err = pool.QueryRow(ctx, countQuery, ancestorID, targetRank).Scan(&count)
	if err != nil || count == 0 {
		return Taxon{}, "", fmt.Errorf("no taxa with images found")
	}

	fmt.Printf("Count is %d \n", count)

	offset := rand.Intn(count)

	query := `
		SELECT t.taxon_id, t.scientific_name, t.scientific_name_authorship, t.taxon_rank, t.has_media, t.taxonomic_status, t.kingdom, t.phylum, t.class_name, t.order_name, t.superfamily, t.family, t.subfamily, t.tribe
		FROM taxon_closure c
		JOIN taxon t ON t.taxon_id = c.descendant_id
		WHERE c.ancestor_id = $1
		AND lower(t.taxon_rank) = lower($2)
		AND t.has_media = TRUE
		OFFSET $3
		LIMIT 1
	`

	var t Taxon
	err = pool.QueryRow(ctx, query, ancestorID, targetRank, offset).Scan(&t.TaxonID, &t.ScientificName, &t.Authorship, &t.Rank, &t.HasMedia, &t.Status, &t.Kingdom, &t.Phylum, &t.Class, &t.Order, &t.SuperFamily, &t.Family, &t.SubFamily, &t.Tribe)
	if err != nil {
		return Taxon{}, "", fmt.Errorf("failed to fetch taxon at offset %d: %v", offset, err)
	}

	return t, ancestorID, nil
}

// getRandomAdditionalTaxa fetches three random taxa (not checking has_media)
func getRandomAdditionalTaxaOffsetTrick(pool *pgxpool.Pool, parentRank, parentName, targetRank, excludeTaxon, ancestorID string, distractorCount int) ([]Taxon, error) { 
	ctx := context.Background()

	countQuery := `
		SELECT COUNT(*)
		FROM taxon_closure c
		JOIN taxon t ON t.taxon_id = c.descendant_id
		WHERE c.ancestor_id = $1 
		AND lower(t.taxon_rank) = lower($2)
		AND t.scientific_name != $3
	`

	var availableCount int
	err := pool.QueryRow(ctx, countQuery, ancestorID, targetRank, excludeTaxon).Scan(&availableCount)
	if err != nil || availableCount < distractorCount {
		return nil, fmt.Errorf("not enouhg taxa to choose from")
	}

	result := make([]Taxon, 0, distractorCount)
	usedOffsets := make(map[int]struct{})

	for len(result) < distractorCount {
		if len(usedOffsets) >= availableCount {
			return nil, fmt.Errorf("ran out of unique offsets to try")
		}

		offset := rand.Intn(availableCount)
		if _, tried := usedOffsets[offset]; tried {
			continue
		}
		usedOffsets[offset] = struct{}{}

		query := `
			SELECT t.taxon_id, t.scientific_name, t.scientific_name_authorship, t.taxon_rank, t.has_media, t.taxonomic_status, t.kingdom, t.phylum, t.class_name, t.order_name, t.superfamily, t.family, t.subfamily, t.tribe
			FROM taxon_closure c
			JOIN taxon t ON t.taxon_id = c.descendant_id
			WHERE c.ancestor_id = $1
			AND lower(t.taxon_rank) = lower($2)
			AND t.scientific_name != $3
			OFFSET $4
			LIMIT 1
		`

		var t Taxon
		err := pool.QueryRow(ctx, query, ancestorID, targetRank, excludeTaxon, offset).Scan(&t.TaxonID, &t.ScientificName, &t.Authorship, &t.Rank, &t.HasMedia, &t.Status, &t.Kingdom, &t.Phylum, &t.Class, &t.Order, &t.SuperFamily, &t.Family, &t.SubFamily, &t.Tribe)
		if err == nil {
			result = append(result, t)
		}
	}

	return result, nil
}
