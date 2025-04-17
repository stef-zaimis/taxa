package quiz

import (
    "context"
    "fmt"
    "math/rand"
    "strings"
    "sync"
    "time"

    "github.com/jackc/pgx/v5/pgxpool"
    "github.com/stef-zaimis/taxa/internal/gbif"
)

// cache entry holds the two ID slices and a timestamp so you can
type descEntry struct {
    allIDs, mediaIDs []string
    ts               time.Time
}

var (
    descCache   = make(map[string]descEntry)
    descCacheMu sync.RWMutex
    // choose a TTL that makes sense for your data
    cacheTTL = 15 * time.Minute
)

func cacheKey(parentRank, parentName, targetRank string) string {
    return strings.ToLower(parentRank + ":" + parentName + ":" + targetRank)
}

// getDescendantsCached wraps fetchDescendantInfo with a simple in‑process cache.
func getDescendantsCached(pool *pgxpool.Pool, parentRank, parentName, targetRank string) (allIDs, mediaIDs []string, ancestorID string, err error) {
    key := cacheKey(parentRank, parentName, targetRank)

    // 1) fast read‑lock to see if we have a fresh entry
    descCacheMu.RLock()
    if e, ok := descCache[key]; ok && time.Since(e.ts) < cacheTTL {
        allIDs = e.allIDs
        mediaIDs = e.mediaIDs
        descCacheMu.RUnlock()
        // we still need to return ancestorID to downstream code,
        // so do a quick one‑row lookup & return it.
        err = pool.QueryRow(context.Background(), `
            SELECT taxon_id
              FROM taxon
             WHERE lower(taxon_rank) = $1
               AND lower(scientific_name) = $2
             LIMIT 1
        `, parentRank, parentName).Scan(&ancestorID)
        return
    }
    descCacheMu.RUnlock()

    // 2) miss or stale → fetch from DB
    allIDs, mediaIDs, ancestorID, err = fetchDescendantInfo(pool, parentRank, parentName, targetRank)
    if err != nil {
        return
    }

    // 3) write into cache
    descCacheMu.Lock()
    descCache[key] = descEntry{
        allIDs:   allIDs,
        mediaIDs: mediaIDs,
        ts:       time.Now(),
    }
    descCacheMu.Unlock()
    return
}

// GenerateQuestion is exactly the same as before, except we call
// getDescendantsCached instead of fetchDescendantInfo directly.
func GenerateQuestion(pool *pgxpool.Pool, parentRank, parentName, targetRank string, optionCount int) (Question, error) {
	fmt.Println("Starting generation")
    // 1) Fetch or load‐from‐cache every descendant ID under ancestor+rank
    descendantIDs, mediaIDs, _, err := getDescendantsCached(pool, parentRank, parentName, targetRank)
    if err != nil {
        return Question{}, fmt.Errorf("descendants lookup: %w", err)
    }
	fmt.Println("Found descendants")

    if len(mediaIDs) == 0 {
        return Question{}, fmt.Errorf("no taxa with media under %s/%s/%s", parentRank, parentName, targetRank)
    }

	cacheKey := cacheKey(parentRank, parentName, targetRank)

    var correctTaxon Taxon
    var imageURL string
	var correctID string

    // 2) Loop until we find a valid image or run out of candidates
    for len(mediaIDs) > 0 {
        idx := rand.Intn(len(mediaIDs))
        candidateID := mediaIDs[idx]

        // Fetch the candidate taxon row
        taxa, err := fetchTaxaByIDs(pool, []string{candidateID})
        if err != nil {
            return Question{}, fmt.Errorf("taxa fetch for candidate %s: %w", candidateID, err)
        }
        t := taxa[0]
        fmt.Printf("Trying taxon %s for image\n", t.ScientificName)

        // Attempt to get an image via GBIF
        gbifKey, img := gbif.GetImage(pool, t.ScientificName, t.Authorship, t.Rank)
        if gbifKey == "" || img == "" || strings.Contains(img, "localhost") {
            // No valid image: remove from our in-memory slice
            mediaIDs = append(mediaIDs[:idx], mediaIDs[idx+1:]...)

            // Also update the cache entry to reflect removal
            descCacheMu.Lock()
            if entry, ok := descCache[cacheKey]; ok {
                filtered := make([]string, 0, len(entry.mediaIDs)-1)
                for _, id := range entry.mediaIDs {
                    if id != candidateID {
                        filtered = append(filtered, id)
                    }
                }
                entry.mediaIDs = filtered
                descCache[cacheKey] = entry
            }
            descCacheMu.Unlock()
            continue
        }

        // Found a valid image
        t.GBIFKey = gbifKey
        correctTaxon = t
		correctID = t.TaxonID
        imageURL = img
        break
    }

    if correctTaxon.TaxonID == "" {
        return Question{}, fmt.Errorf("could not find any valid image among candidates")
    }

    others := make([]string, 0, len(descendantIDs)-1)
    for _, id := range descendantIDs {
        if id != correctID {
            others = append(others, id)
        }
    }
    if len(others) < optionCount-1 {
        return Question{}, fmt.Errorf("need %d distractors but only have %d", optionCount-1, len(others))
    }

    rand.Shuffle(len(others), func(i, j int) { others[i], others[j] = others[j], others[i] })
    distractorIDs := others[:optionCount-1]

    pickIDs := append(distractorIDs, correctID)
    taxaRows, err := fetchTaxaByIDs(pool, pickIDs)
    if err != nil {
        return Question{}, fmt.Errorf("taxa fetch: %w", err)
    }

    rand.Shuffle(len(taxaRows), func(i, j int) { taxaRows[i], taxaRows[j] = taxaRows[j], taxaRows[i] })
    correctIndex := 0
    for i, t := range taxaRows {
        if t.TaxonID == correctID {
            correctIndex = i
            break
        }
    }

	fmt.Println("Full quiz assembled")
    return Question{
        ImageURL:      imageURL,
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
