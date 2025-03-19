package main

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

const dbURL = "postgres://postgres:toor@127.0.0.1:5432/taxa"
const gbifSearchAPI = "https://api.gbif.org/v1/species/search?datasetKey=d7dddbf4-2cf0-4f39-9b2a-bb099caae36c&q="
const gbifOccurrenceAPI = "https://api.gbif.org/v1/occurrence/search?mediaType=StillImage&taxonKey="

func main() {
	conn, err := pgx.Connect(context.Background(), dbURL)
	if err != nil {
		log.Fatalf("Can't connect to db: %v\n", err)
	}
	defer conn.Close(context.Background())

	reader := bufio.NewReader(os.Stdin)

	fmt.Print("Enter taxonomic rank (e.g. 'Kingdom'): ")
	rank, _ := reader.ReadString('\n')
	rank = strings.TrimSpace(strings.ToLower(rank))

	fmt.Print("Enter taxon name (e.g. 'Animalia'): ")
	name, _ := reader.ReadString('\n')
	name = strings.TrimSpace(name)

	fmt.Print("And what taxonomic level would you like to be tested on (e.g. 'Order')? ")
	choiceRank, _ := reader.ReadString('\n')
	choiceRank = strings.TrimSpace(strings.ToLower(choiceRank))

	fmt.Printf("Querying for rank: '%s' and name: '%s'\n", rank, name)

	// Fetch one taxon with media to be the correct answer
	correctTaxon, err := getTaxonWithMedia(conn, rank, name, choiceRank)
	if err != nil {
		log.Fatalf("No taxa with images found under '%s' (%s). Try another category.", name, rank)
	}

	// Fetch three additional taxa (without checking `has_media`)
	randomTaxa, err := getRandomAdditionalTaxa(conn, rank, name, choiceRank, correctTaxon)
	if err != nil {
		log.Fatalf("Error fetching additional taxa: %v\n", err)
	}

	// Create answer set
	taxa := append(randomTaxa, correctTaxon)
	rand.Shuffle(len(taxa), func(i, j int) { taxa[i], taxa[j] = taxa[j], taxa[i] })

	// Determine correct answer position
	correctAnswerID := -1
	for i, taxon := range taxa {
		if taxon == correctTaxon {
			correctAnswerID = i
			break
		}
	}

	// Get image for the correct answer
	gbifKey, imageURL := getGBIFImage(conn, correctTaxon)

	fmt.Printf("GBIF Key for %s: %s\n", correctTaxon, gbifKey)

	// Display image if found
	if imageURL != "" {
		fmt.Printf("\n Image for correct answer (%s): %s\n", correctTaxon, imageURL)
	} else {
		fmt.Println("No image found for the correct answer.")
	}

	// Display the options
	fmt.Println("\nHere are 4 options under", name+":")
	for i, taxon := range taxa {
		fmt.Printf("%d. %s\n", i+1, taxon)
	}

	// Display image URL for correct answer
	fmt.Printf("\nImage for correct answer (%s): %s\n", correctTaxon, imageURL)

	// Get user input
	fmt.Print("Guess the correct answer by inputting the number (e.g. '1'): ")
	userAnswer, _ := reader.ReadString('\n')
	userAnswer = strings.TrimSpace(userAnswer)
	userAnswerInt, _ := strconv.Atoi(userAnswer)

	if correctAnswerID == userAnswerInt-1 {
		fmt.Println("Correct!")
	} else {
		fmt.Printf("Wrong! The correct answer was %s\n", correctTaxon)
	}
}

// getTaxonWithMedia fetches a single taxon with has_media = TRUE
func getTaxonWithMedia(conn *pgx.Conn, parentRank, parentName, targetRank string) (string, error) {
	ctx := context.Background()

	query := `
		SELECT t.scientific_name
		FROM taxon_closure c
		JOIN taxon t ON t.taxon_id = c.descendant_id
		WHERE c.ancestor_id = (
			SELECT taxon_id FROM taxon WHERE lower(taxon_rank) = $1 AND lower(scientific_name) = $2 LIMIT 1
		)
		AND lower(t.taxon_rank) = $3
		AND t.has_media = TRUE
		LIMIT 1
	`

	var taxon string
	err := conn.QueryRow(ctx, query, parentRank, parentName, targetRank).Scan(&taxon)
	if err != nil {
		return "", fmt.Errorf("no taxa with images found")
	}

	return taxon, nil
}

// getRandomAdditionalTaxa fetches three random taxa (not checking has_media)
func getRandomAdditionalTaxa(conn *pgx.Conn, parentRank, parentName, targetRank, excludeTaxon string) ([]string, error) {
	ctx := context.Background()

	query := `
		SELECT t.scientific_name
		FROM taxon_closure c
		JOIN taxon t ON t.taxon_id = c.descendant_id
		WHERE c.ancestor_id = (
			SELECT taxon_id FROM taxon WHERE lower(taxon_rank) = $1 AND lower(scientific_name) = $2 LIMIT 1
		)
		AND lower(t.taxon_rank) = $3
		AND t.scientific_name != $4
		ORDER BY RANDOM()
		LIMIT 3
	`

	rows, err := conn.Query(ctx, query, parentRank, parentName, targetRank, excludeTaxon)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxa []string
	for rows.Next() {
		var taxon string
		if err := rows.Scan(&taxon); err != nil {
			return nil, err
		}
		taxa = append(taxa, taxon)
	}

	return taxa, nil
}

// getGBIFImage retrieves an image for a taxon and updates the database if needed
func getGBIFImage(conn *pgx.Conn, taxon string) (string, string) {
	ctx := context.Background()

	var gbifKey string
	query := "SELECT gbif_key FROM taxon WHERE scientific_name = $1 AND has_media = TRUE"

	err := conn.QueryRow(ctx, query, taxon).Scan(&gbifKey)
	if err != nil || gbifKey == "" {
		// Query GBIF API for taxon key
		gbifKey = fetchGBIFKeyFromAPI(taxon)
		if gbifKey == "" {
			fmt.Printf("No GBIF taxon key found for: %s\n", taxon)
			return "", ""
		}

		// Update the database with the newly found GBIF key
		updateQuery := "UPDATE taxon SET gbif_key = $1 WHERE scientific_name = $2"
		_, err := conn.Exec(ctx, updateQuery, gbifKey, taxon)
		if err != nil {
			fmt.Printf("Failed to update GBIF key for %s: %v\n", taxon, err)
		} else {
			fmt.Printf("Updated GBIF key for %s: %s\n", taxon, gbifKey)
		}
	}

	// Query the GBIF occurrence API for an image
	imageURL := fetchGBIFImageFromAPI(gbifKey)
	return gbifKey, imageURL
}

// fetchGBIFKeyFromAPI retrieves a GBIF taxon key using the scientific name
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

	body, _ := ioutil.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	if len(result.Results) > 0 {
		return strconv.Itoa(result.Results[0].Key)
	}

	return ""
}

// fetchGBIFImageFromAPI queries the occurrence API for an image
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

	if len(result.Results) > 0 && len(result.Results[0].Media) > 0 {
		return result.Results[0].Media[0].Identifier
	}

	fmt.Println("No images found for GBIF key:", gbifKey)
	return ""
}

