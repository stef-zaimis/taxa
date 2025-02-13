package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"math/rand"
	"os"
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

	fmt.Printf("Querying for rank: '%s' and name: '%s'\n", rank, name)
	taxa, err := getRandomTaxa(conn, rank, name)
	if err != nil {
		log.Fatalf("Error fetching taxa: %v\n", err)
	}

	correct_answer_id := rand.Intn(len(taxa))
	correct_answer := taxa[correct_answer_id]

	fmt.Println("\nHere are 4 options under", name+":")
	for _, taxon := range taxa {
		fmt.Println("-", taxon)
	}

	fmt.Println("Guess the correct answer")
	answer, _ := reader.ReadString('\n')
	answer = strings.TrimSpace(answer)
	if strings.ToLower(correct_answer) == strings.ToLower(answer) {
		fmt.Println("Correct!")
	} else {
		fmt.Printf("Wrong! The right answer is %s\n", correct_answer)
	}
}

func getRandomTaxa(conn *pgx.Conn, rank, name string) ([]string, error) {
	query := `
		SELECT t2.scientific_name
		FROM taxon t1
		JOIN taxon t2 ON t1.taxon_id = t2.parent_id
		WHERE LOWER(t1.taxon_rank) = LOWER($1) AND LOWER(t1.scientific_name) = LOWER($2)
		ORDER BY RANDOM()
		LIMIT 4;
	`

	rows, err := conn.Query(context.Background(), query, rank, name)
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

	if len(taxa) == 0 {
		return nil, fmt.Errorf("no taxa found under %s (%s)", name, rank)
	}

	return taxa, nil
}
