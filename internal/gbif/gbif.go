package gbif

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"math/rand"
	"net/http"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

const (
	gbifSearchAPI     = "https://api.gbif.org/v1/species/search?datasetKey=d7dddbf4-2cf0-4f39-9b2a-bb099caae36c&status=ACCEPTED&extinct=false&q="
	gbifOccurrenceAPI = "https://api.gbif.org/v1/occurrence/search?mediaType=StillImage&license=CC0_1_0&license=CC_BY_4_0&taxonKey="
)

// Return a GBIF key and image URL for a given taxon
func GetImage(pool *pgxpool.Pool, taxon, authorship, rank string) (string, string) {
	ctx := context.Background()

	var gbifKey string
	query := "SELECT gbif_key FROM taxon WHERE lower(scientific_name) = lower($1) AND has_media = TRUE AND lower(taxon_rank) = lower($2)"

	err := pool.QueryRow(ctx, query, taxon, rank).Scan(&gbifKey)
	if err == nil && gbifKey != "" {
		imageURL := fetchGBIFImageFromAPI(pool, gbifKey, taxon, rank)
		return gbifKey, imageURL
	}

	// If there is an error (I'm assuming the error is related to gbifKey == "" maybe this should be specified later on)
	fmt.Printf("Querying for: %s\n", taxon)
	gbifKey = fetchGBIFKeyFromAPI(taxon, rank)

	if gbifKey == "" {
		pool.Exec(ctx, "UPDATE taxon SET has_media = FALSE WHERE lower(scientific_name) = lower($1) AND lower(taxon_rank) = lower($2)", taxon, rank)
		fmt.Printf("No GBIF taxon key found for: %s\n", taxon)
		return "", ""
	}

	// Add the retrieved key to the DB
	updateQuery := "UPDATE taxon SET gbif_key = $1 WHERE lower(scientific_name) = lower($2) AND lower(taxon_rank) = lower($3)"
	_, err = pool.Exec(ctx, updateQuery, gbifKey, taxon, rank)
	if err != nil {
		fmt.Printf("Failed to update GBIF key for %s: %v\n", taxon, err)
	} else {
		fmt.Printf("Updated GBIF key for %s: %s\n", taxon, gbifKey)
	}

	imageURL := fetchGBIFImageFromAPI(pool, gbifKey, taxon, rank)
	return gbifKey, imageURL
}

// Retrieve a GBIF taxon key using the scientific name
func fetchGBIFKeyFromAPI(taxon, rank string) string {
	url := fmt.Sprintf("%s%s&rank=%s", gbifSearchAPI, taxon, rank)
	resp, err := http.Get(url)
	if err != nil {
		fmt.Println("Error querying GBIF:", err)
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			Key int `json:"key"`
		} `json:"results"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)
	if err != nil {
		fmt.Println("Error unmarshalling GBIF response:", err)
		return ""
	}

	if len(result.Results) > 0 {
		return strconv.Itoa(result.Results[0].Key)
	}

	return ""
}

// Query the occurrence API for an image
func fetchGBIFImageFromAPI(pool *pgxpool.Pool, gbifKey, taxon, rank string) string {
	ctx := context.Background()
	resp, err := http.Get(gbifOccurrenceAPI + gbifKey)
	if err != nil {
		fmt.Println("Error querying GBIF occurrence API:", err)
		return ""
	}
	defer resp.Body.Close()

	var result struct {
		Results []struct {
			Media []struct {
				Identifier string `json:"identifier"`
			} `json:"media"`
		} `json:"results"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	var images []string
	for _, occurrence := range result.Results {
		for _, media := range occurrence.Media {
			if media.Identifier != "" && !strings.Contains(media.Identifier, "localhost") {
				images = append(images, media.Identifier)
			}
		}
	}

	if len(images) == 0 {
		pool.Exec(ctx, "UPDATE taxon SET has_media = FALSE WHERE lower(scientific_name) = lower($1) AND lower(taxon_rank) = lower($2)", taxon, rank)
		fmt.Println("No images found for GBIF key:", gbifKey)
		return ""
	}

	return images[rand.Intn(len(images))]
}
