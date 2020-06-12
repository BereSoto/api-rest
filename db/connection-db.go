package connectiondb

import (
	"database/sql"
	"fmt"
	"log"

	_ "github.com/lib/pq"
)

func main() {
	// Connect to the "bank" database.
	db, err := sql.Open("postgres", "postgresql://test@localhost:26257/domains?sslmode=disable")
	if err != nil {
		log.Fatal("error connecting to the database: ", err)
	}
	defer db.Close()

	// Create the "accounts" table.

	// Insert two rows into the "accounts" table.
	if _, err := db.Exec(
		"INSERT INTO servers (id, address, ssl_grade, country, owner, local_date) VALUES (1, 192.68.1.1, ), (2, 250)"); err != nil {
		log.Fatal(err)
	}

	// Print out the balances.
	rows, err := db.Query("SELECT id, balance FROM accounts")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("Initial balances:")
	for rows.Next() {
		var id, balance int
		if err := rows.Scan(&id, &balance); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("%d %d\n", id, balance)
	}
}
