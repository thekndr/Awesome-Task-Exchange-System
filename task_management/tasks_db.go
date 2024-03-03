package task_management

import (
	"database/sql"
	"log"

	_ "github.com/lib/pq"
)

func MustInitDB() *sql.DB {
	connStr := "user=popug dbname=ates password=pgdbpassword sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		log.Fatal(err)
	}

	if err = db.Ping(); err != nil {
		log.Fatal(err)
	}

	log.Println(`Preparing tables...`)
	_, err = db.Exec(`
CREATE TABLE IF NOT EXISTS tasks (
    id SERIAL PRIMARY KEY,
    created_at TIMESTAMP NOT NULL DEFAULT now(),
    uuid UUID NOT NULL DEFAULT gen_random_uuid(),
    assignee_id UUID NOT NULL,
    description VARCHAR(255),
    status INT NOT NULL
);
`)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(`... done`)

	return db
}
