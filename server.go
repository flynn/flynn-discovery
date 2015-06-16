package main

import (
	"log"
	"net/http"
	"os"

	"github.com/flynn/flynn-discovery/Godeps/_workspace/src/github.com/jackc/pgx"
)

func main() {
	dbConfig, err := pgx.ParseURI(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	db, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: dbConfig})
	if err != nil {
		log.Fatal(err)
	}

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), NewServer(os.Getenv("URL"), NewPostgresBackend(db))))
}
