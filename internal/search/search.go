package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type SearchResult struct {
	ScientificName string `json:"scientific_name"`
	Authorship string `json:"authorship"`
	Rank string `json:"rank"`
	TaxonID string `json:"taxon_id"`
	HasMedia bool `json:"has_media"`
}

// Perform fast case-insensitive substring search search (fuzzy if needed)
func SearchTaxa(pool *pgxpool.Pool, rawQuery string, limit int) ([]SearchResult, error) {
	ctx := context.Background()
	query := strings.ToLower(strings.TrimSpace(rawQuery))
	results := []SearchResult{}

	// Try fast substring match
	sqlPrefix := `
		SELECT scientific_name, scientific_name_authorship, taxon_rank, taxon_id, has_media
		FROM search_index
		WHERE lower(full_display_name) LIKE '%' || $1 || '%'
		ORDER BY has_media DESC, scientific_name
		LIMIT $2;
	`

	rows, err := pool.Query(ctx, sqlPrefix, query+"%", limit)
	if err != nil {
		return nil, fmt.Errorf("substring search error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ScientificName, &r.Authorship, &r.Rank, &r.TaxonID, &r.HasMedia); err == nil {
			results = append(results, r)
		}
	}

	// Return early if there are matches
	if len(results) > 0 {
		return results, nil
	}

	// Fuzzy fallback if substring returned nothing
	sqlFuzzy := `
		SELECT scientific_name, scientific_name_authorship, taxon_rank, taxon_id, has_media
		FROM search_index
		WHERE full_display_name % $1
		ORDER BY similarity(full_display_name, $1) DESC
		LIMIT $2;
	`

	rows, err = pool.Query(ctx, sqlFuzzy, query, limit)
	if err != nil {
		return nil, fmt.Errorf("fuzzy search error: %w", err)
	}
	defer rows.Close()

	for rows.Next() {
		var r SearchResult
		if err := rows.Scan(&r.ScientificName, &r.Authorship, &r.Rank, &r.TaxonID, &r.HasMedia); err == nil {
			results = append(results, r)
		}
	}

	return results, nil
}
