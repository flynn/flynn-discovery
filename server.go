package main

import (
	"context"
	"log"
	"net/http"
	"os"

	"cloud.google.com/go/firestore"
	"github.com/flynn/flynn-discovery/discovery"
	firestoreb "github.com/flynn/flynn-discovery/firestore"
	"github.com/flynn/flynn-discovery/postgres"
	"github.com/jackc/pgx"
)

func main() {
	var b discovery.StorageBackend
	if uri := os.Getenv("DATABASE_URL"); uri != "" {
		dbConfig, err := pgx.ParseURI(uri)
		if err != nil {
			log.Fatal(err)
		}
		db, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: dbConfig})
		if err != nil {
			log.Fatal(err)
		}
		b = postgres.NewBackend(db)
	} else if coll := os.Getenv("FIREBASE_COLLECTION"); coll != "" {
		client, err := firestore.NewClient(context.Background(), firestore.DetectProjectID)
		if err != nil {
			log.Fatal("error creating firestore client: %s", err)
		}
		b = firestoreb.NewBackend(client, coll)
	} else {
		log.Fatal("no backend configured")
	}

	log.Fatal(http.ListenAndServe(":"+os.Getenv("PORT"), discovery.NewServer(os.Getenv("URL"), b)))
}
