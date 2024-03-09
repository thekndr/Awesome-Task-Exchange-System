package main

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func initDB() {
	var err error

	connStr := "user=popug dbname=ates password=pgdbpassword sslmode=disable"
	db, err = sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println(`Preparing tables...`)
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS users (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    public_id UUID NOT NULL DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    role VARCHAR(50) NOT NULL DEFAULT 'worker'
);
`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(`... done`)
}
