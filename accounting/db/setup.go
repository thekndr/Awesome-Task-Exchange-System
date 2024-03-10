package db

import (
	"database/sql"
	"log"

	_ "embed"
	_ "github.com/lib/pq"
)

var (
	//go:embed resources/accounting_setup.sql
	sqlStatementsSetup string

	//go:embed resources/accounting_drop.sql
	sqlStatementsDrop string
)

func MustInit(reset bool) *sql.DB {
	connStr := "user=popug dbname=ates password=pgdbpassword sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println(`Preparing tables...`)

	if reset {
		log.Println("!!! RESET all tables mode")
		log.Printf(`resetting tables...`)
		_, err = db.Exec(sqlStatementsDrop)
		if err != nil {
			log.Fatalf(`failed to reset tables: %w`, err)
		}
		log.Println(`... done`)
	}

	log.Printf(`establishing missing tables...`)
	_, err = db.Exec(sqlStatementsSetup)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(`... done`)

	return db
}
