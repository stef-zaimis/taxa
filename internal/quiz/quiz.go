package quiz

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stef-zaimis/taxa/internal/gbif"
)

// GenerateQuestion assembles a full quiz question with correct/incorrect taxa and an image.
func GenerateQuestion(pool *pgxpool.Pool, parentRank, parentName, targetRank string, optionCount int) (Question, error) {
	// Phase 1: fetch all media-enabled IDs once
	mediaIDs, ancestorID, err := fetchAllMediaIDs(pool, parentRank, parentName, targetRank)
	if err != nil {
		return Question{}, fmt.Errorf("no media taxa found: %w", err)
	}

	// Phase 2: retry in-memory picks
	maxAttempts := 10
	var correctTaxon Taxon
	var gbifKey, imageURL string
	for i := 0; i < maxAttempts; i++ {
		pickID := mediaIDs[rand.Intn(len(mediaIDs))]
		// fetch that single taxon
		t, err := fetchTaxonByID(pool, pickID)
		if err != nil {
			continue
		}
		// get image from GBIF
		key, img := gbif.GetImage(pool, t.ScientificName, t.Authorship, t.Rank)
		if img != "" && !strings.Contains(img, "localhost") {
			correctTaxon = t
			gbifKey = key
			imageURL = img
			break
		}
	}
	if correctTaxon.TaxonID == "" {
		return Question{}, fmt.Errorf("could not find media taxon after %d tries", maxAttempts)
	}
	correctTaxon.GBIFKey = gbifKey

	// Get distractors
	distractorCount := optionCount - 1
	incorrectTaxa, err := getRandomAdditionalTaxa(pool, parentRank, parentName, targetRank, correctTaxon.ScientificName, ancestorID, distractorCount)
	if err != nil {
		return Question{}, fmt.Errorf("failed to get distractors: %w", err)
	}

	// Shuffle options and find correct index
	options := append(incorrectTaxa, correctTaxon)
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })
	correctIndex := -1
	for i, t := range options {
		if t.ScientificName == correctTaxon.ScientificName {
			correctIndex = i
			break
		}
	}
	if correctIndex == -1 {
		return Question{}, fmt.Errorf("can't find correct taxon after shuffling")
	}

	// Return question
	return Question{
		ImageURL:     imageURL,
		Options:      options,
		CorrectIndex: correctIndex,
		CorrectAnswer: correctTaxon,
	}, nil
}

// fetchAllMediaIDs retrieves all descendant IDs with has_media = TRUE under the given ancestor+rank.
func fetchAllMediaIDs(pool *pgxpool.Pool, parentRank, parentName, targetRank string) ([]string, string, error) {
	ctx := context.Background()
	var ancestorID string
	if err := pool.QueryRow(ctx, `
		SELECT taxon_id
		  FROM taxon
		 WHERE lower(taxon_rank) = $1
		   AND lower(scientific_name) = $2
		 LIMIT 1
	`, parentRank, parentName).Scan(&ancestorID); err != nil {
		return nil, "", err
	}

	rows, err := pool.Query(ctx, `
		SELECT c.descendant_id
		  FROM taxon_closure c
		  JOIN taxon t ON t.taxon_id = c.descendant_id
		 WHERE c.ancestor_id       = $1
		   AND lower(t.taxon_rank) = lower($2)
		   AND t.has_media         = TRUE
	`, ancestorID, targetRank)
	if err != nil {
		return nil, "", err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, "", err
		}
		ids = append(ids, id)
	}
	if err := rows.Err(); err != nil {
		return nil, "", err
	}
	if len(ids) == 0 {
		return nil, "", fmt.Errorf("no taxa with media found")
	}
	return ids, ancestorID, nil
}

// fetchTaxonByID retrieves a single Taxon row by taxon_id.
func fetchTaxonByID(pool *pgxpool.Pool, id string) (Taxon, error) {
	ctx := context.Background()
	var t Taxon
	err := pool.QueryRow(ctx, `
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
	`, id).Scan(
		&t.TaxonID, &t.ScientificName, &t.Authorship, &t.Rank,
		&t.HasMedia, &t.Status, &t.Kingdom, &t.Phylum,
		&t.Class, &t.Order, &t.SuperFamily, &t.Family,
		&t.SubFamily, &t.Tribe,
	)
	return t, err
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
