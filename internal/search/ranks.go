package search

import (
	"context"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

func SearchRanks(pool *pgxpool.Pool, rawQuery string, limit int) ([]string, error) {
	ctx := context.Background()
	query := strings.ToLower(strings.TrimSpace(rawQuery))

	sql := `
		SELECT rank
		FROM rank_index
		WHERE lower(rank) LIKE lower($1) || '%'
		ORDER BY rank ASC
		LIMIT $2
	`

	rows, err := pool.Query(ctx, sql, query, limit)
	if err != nil {
		return nil, fmt.Errorf("rank search failed: %w", err)
	}
	defer rows.Close()

	var results []string
	for rows.Next() {
		var rank string
		if err := rows.Scan(&rank); err == nil {
			results = append(results, rank)
		}
	}

	return results, nil
}
