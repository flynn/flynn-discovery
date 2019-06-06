package main

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"time"

	"cloud.google.com/go/firestore"
	firestoreapi "cloud.google.com/go/firestore/apiv1"
	"github.com/flynn/flynn-discovery/discovery"
	"github.com/jackc/pgx"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/option"
)

func main() {
	pgConfig, err := pgx.ParseURI(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatal(err)
	}
	pg, err := pgx.NewConnPool(pgx.ConnPoolConfig{ConnConfig: pgConfig})
	if err != nil {
		log.Fatal(err)
	}

	creds, err := google.CredentialsFromJSON(context.Background(), []byte(os.Getenv("GCP_AUTH")), firestoreapi.DefaultAuthScopes()...)
	if err != nil {
		log.Fatalf("error getting GCP credentials: %s", err)
	}

	client, err := firestore.NewClient(context.Background(), firestore.DetectProjectID, option.WithCredentials(creds))
	if err != nil {
		log.Fatalf("error creating firestore client: %s", err)
	}
	log.Fatal(migrate(pg, client, client.Collection("flynn-discovery-production/flynn-discovery/clusters")))
}

func migrate(pg *pgx.ConnPool, fc *firestore.Client, coll *firestore.CollectionRef) error {
	c := &discovery.Cluster{}
	for {
		batch := fc.Batch()
		rows, err := pg.Query("SELECT cluster_id, creator_ip, creator_user_agent, created_at FROM clusters WHERE created_at > $1 ORDER BY created_at ASC LIMIT 100", c.CreatedAt)
		if err != nil {
			return err
		}
		again := false
		for rows.Next() {
			again = true
			if err := scanCluster(rows, c); err != nil {
				rows.Close()
				return err
			}
			batch.Set(coll.Doc(c.ID), c)
		}
		if !again {
			break
		}
		if _, err := batch.Commit(context.Background()); err != nil {
			return err
		}
		fmt.Println("clusters:", c.CreatedAt.String())
	}

	inst := &discovery.Instance{CreatedAt: &time.Time{}}
	for {
		batch := fc.Batch()
		rows, err := pg.Query("SELECT cluster_id, instance_id, flynn_version, ssh_public_keys, url, name, creator_ip, created_at FROM instances WHERE created_at > $1 ORDER by created_at ASC LIMIT 100", inst.CreatedAt)
		if err != nil {
			return err
		}
		again := false
		for rows.Next() {
			again = true
			if err := scanInstance(rows, inst); err != nil {
				rows.Close()
				return err
			}
			docID := instanceID(inst.ClusterID, inst.URL)
			batch.Set(coll.Doc(inst.ClusterID).Collection("instances").Doc(docID), inst)
		}
		if !again {
			break
		}
		if _, err := batch.Commit(context.Background()); err != nil {
			return err
		}
		fmt.Println("instances:", inst.CreatedAt.String())
	}
	return nil
}

type pgxScanner interface {
	Scan(...interface{}) error
}

func scanCluster(row pgxScanner, c *discovery.Cluster) error {
	return row.Scan(&c.ID, &c.CreatorIP, &c.CreatorUserAgent, &c.CreatedAt)
}

func scanInstance(row pgxScanner, inst *discovery.Instance) error {
	if inst.CreatedAt == nil {
		inst.CreatedAt = &time.Time{}
	}
	var sshKeys string
	if err := row.Scan(&inst.ClusterID, &inst.ID, &inst.FlynnVersion, &sshKeys, &inst.URL, &inst.Name, &inst.CreatorIP, inst.CreatedAt); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(sshKeys), &inst.SSHPublicKeys); err != nil {
		return err
	}
	return nil
}

func instanceID(clusterID, instURL string) string {
	hash := sha256.Sum256([]byte("v1 " + clusterID + " " + instURL))
	return base64.URLEncoding.EncodeToString(hash[:])
}
