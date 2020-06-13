package main

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
    if _, err := db.Exec(
        "CREATE TABLE IF NOT EXISTS servers (id INT PRIMARY KEY, address STRING, ssl_grade STRING, country STRING, owner STRING)"); err != nil {
        log.Fatal(err)
    }
	// Insert two rows into the "accounts" table.
	if _, err := db.Exec(
		"INSERT INTO servers (id, address, ssl_grade, country, owner) VALUES (1, '192.68.1.1', 'B', 'MX', 'Example.com, Inc.')"); err != nil {
		log.Fatal(err)
	}

	// Print out the balances.
	rows, err := db.Query("SELECT * FROM servers")
	if err != nil {
		log.Fatal(err)
	}
	defer rows.Close()
	fmt.Println("Prueba de datos")
	for rows.Next() {
		var id int
		var address, ssl_grade, country, owner string

		if err := rows.Scan(&id, &address, &ssl_grade, &country, &owner); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("el registro es: %d %s %s %s %s\n", id, address, ssl_grade, country, owner) // tipos de dato %d=digito o numero, %s= string
	}
}
