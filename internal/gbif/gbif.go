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

	"github.com/jackc/pgx/v5"
)

const (
	gbifSearchAPI = "https://api.gbif.org/v1/species/search?datasetKey=d7dddbf4-2cf0-4f39-9b2a-bb099caae36c&q="
	gbifOccurrenceAPI = "https://api.gbif.org/v1/occurrence/search?mediaType=StillImage&license=CC0_1_0&license=CC_BY_4_0&taxonKey="
)

// Return a GBIF key and image URL for a given taxon
func GetImage(conn *pgx.Conn, taxon string, authorship string) (gbifKey string, imageURL string) {
	ctx := context.Background()

	var gbifKey string
	query := "SELECT gbif_key FROM taxon WHERE scientific_name = $1 AND has_media = TRUE"

	err := conn.QueryRow(ctx, query, taxon).Scan(&gbifKey)
	if err == nil && gbifKey != "" {
		imageURL := fetchGBIFImageFromAPI(gbifKey)
		return gbifKey, imageURL
	}

	// If there is an error (I'm assuming the error is related to gbifKey == "" maybe this should be specified later on)
	strippedName := strings.TrimSpace(strings.Replace(taxon, authorship, "", 1))
	gbifKey = fetchGBIFKeyFromAPI(strippedName)

	if gbifKey == "" {
		fmt.Printf("No GBIF taxon key found for: %s\n", taxon)
		return "", ""
	}

	// Add the retrieved key to the DB
	updateQuery := "UPDATE taxon SET gbif_key = $1 WHERE scientific_name = $2"
	_, err := conn.Exec(ctx, updateQuery, gbifKey, taxon)
	if err != nil {
		fmt.Printf("Failed to update GBIF key for %s: %v\n", taxon ,err)
	} else {
		fmt.Printf("Updated GBIF key for %s: %s\n", taxon, gbifKey)
	}

	imageURL := fetchGBIFImageFromAPI(gbifKey)
	return gbifKey, imageURL
}

// Retrieve a GBIF taxon key using the scientific name
func fetchGBIFKeyFromAPI(taxon string) string {
	resp, err := http.Get(gbifSearchAPI + taxon)
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

	if len(result.Results) > 0 {
		return strconv.Itoa(result.Results[0].Key)
	}

	return ""
}

// Query the occurrence API for an image
func fetchGBIFImageFromAPI(gbifKey string) string {
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

	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	var images []string
	for _, occurrence := range result.Results {
		for _, media := range occurrence.Media {
			images = append(images, media.Identifier)
		}
	}

	if len(images) == 0 {
		fmt.Println("No images found for GBIF key:", gbifKey)
		return ""
	}

	return images[rand.Intn(len(images))] 
}
