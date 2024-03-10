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

func MustInit(drop bool) *sql.DB {
	connStr := "user=popug dbname=ates password=pgdbpassword sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println(`Preparing tables...`)

	if drop {
		log.Println("!!! DROP all tables mode")
		_, err = db.Exec(sqlStatementsDrop)
	} else {
		_, err = db.Exec(sqlStatementsSetup)
	}

	if err != nil {
		log.Fatal(err)
	}
	log.Println(`... done`)

	return db
}
