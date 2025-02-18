package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"github.com/jackc/pgx/v5"
)

const dbURL = "postgres://postgres:toor@127.0.0.1:5432/col_dwca_db"

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
	taxa, err := getRandomTaxa(conn, rank, name, choiceRank)
	if err != nil {
		log.Fatalf("Error fetching taxa: %v\n", err)
	}

	correctAnswerID := rand.Intn(len(taxa))
	fmt.Println("The right answer is ", correctAnswerID)
	correctAnswer := taxa[correctAnswerID]

	fmt.Println("\nHere are 4 options under", name+":")
	id := 1
	for _, taxon := range taxa {
		fmt.Println(id, taxon)
		id++
	}

	fmt.Print("Guess the correct answer by inputting the number (e.g. '1'): ")
	userAnswer, _ := reader.ReadString('\n')
	userAnswer = strings.TrimSpace(userAnswer)
	userAnswerInt, _ := strconv.Atoi(userAnswer)
	if correctAnswerID == userAnswerInt-1 {
		fmt.Println("Correct!")
	} else {
		fmt.Printf("Wrong! The right answer is %s\n", correctAnswer)
	}
}

func getRandomTaxa(conn *pgx.Conn, parentRank, parentName, targetRank string) ([]string, error) {
	ctx := context.Background()

	findAncestor := `
		SELECT taxon_id
		FROM taxon
		WHERE lower(taxon_rank) = $1
		AND lower(scientific_name) = $2
		LIMIT 1
	`

	var ancestorID string
	err := conn.QueryRow(ctx, findAncestor, parentRank, parentName).Scan(&ancestorID)
	if err != nil {
		return nil, fmt.Errorf("Couldn't find parent (%s %s): %w", parentRank, parentName, err)
	}

	findDescendants := `
			SELECT t.scientific_name
			FROM taxon_closure c
			JOIN taxon t ON t.taxon_id = c.descendant_id
			WHERE c.ancestor_id = $1
			AND lower(t.taxon_rank) = $2
		`
	rows, err := conn.Query(ctx, findDescendants, ancestorID, targetRank)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var taxa []string
	for rows.Next() {
		var taxonName string
		if err := rows.Scan(&taxonName); err != nil {
			return nil, err
		}
		taxa = append(taxa, taxonName)
	}
	rows.Close()

	if len(taxa) == 0 {
		return nil, fmt.Errorf("no taxa found under %s (%s)", parentName, parentRank)
	}

	rand.Shuffle(len(taxa), func(i, j int) {
		taxa[i], taxa[j] = taxa[j], taxa[i]
	})

	if len(taxa) > 4 {
		taxa = taxa[:4]
	}

	return taxa, nil
}
