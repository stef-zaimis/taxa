package quiz

import (
	"context"
	"fmt"
	"math/rand"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/stef-zaimis/taxa/internal/gbif"
)

// Assemblae a full quiz question with correct/incorrect taxa and an image
func GenerateQuestion(pool *pgxpool.Pool, parentRank, parentName, targetRank string, optionCount int) (Question, error) {

	maxAttempts := 10
	var correctTaxon Taxon
	var ancestorID, gbifKey, imageURL string
	
	for i:=0; i<maxAttempts; i++ {
		// #1: Get correct taxon

		taxon, id, err := getTaxonWithMedia(pool, parentRank, parentName, targetRank)
		if err != nil {
			continue
		}

		key, img := gbif.GetImage(pool, taxon.ScientificName, taxon.Authorship, taxon.Rank)
		if img != "" && !strings.Contains(img, "localhost") {
			correctTaxon = taxon
			ancestorID = id
			gbifKey = key
			imageURL = img
			break
		}
	}
	correctTaxon.GBIFKey = gbifKey

	// #3: Get other options
	fmt.Println("Fetching other options")

	distractorCount := optionCount - 1
	incorrectTaxa, err := getRandomAdditionalTaxa(pool, parentRank, parentName, targetRank, correctTaxon.ScientificName, ancestorID, distractorCount)
	if err != nil {
		return Question{}, fmt.Errorf("failed to get distractors: %w", err)
	}

	// #4 Shuffle options
	options := append(incorrectTaxa, correctTaxon)
	rand.Shuffle(len(options), func(i, j int) { options[i], options[j] = options[j], options[i] })

	// #5: Find correct index
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

	// #6: Return question
	return Question {
		ImageURL: imageURL,
		Options: options,
		CorrectIndex: correctIndex,
		CorrectAnswer: correctTaxon,
	}, nil
}

// getTaxonWithMedia fetches a single taxon with has_media = TRUE
func getTaxonWithMedia(pool *pgxpool.Pool, parentRank, parentName, targetRank string) (Taxon, string, error) {
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
func getRandomAdditionalTaxa(pool *pgxpool.Pool, parentRank, parentName, targetRank, excludeTaxon, ancestorID string, distractorCount int) ([]Taxon, error) { 
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
