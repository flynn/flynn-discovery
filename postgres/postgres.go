package postgres

import (
	"encoding/json"
	"time"

	"github.com/flynn/flynn-discovery/discovery"
	"github.com/jackc/pgx"
)

func NewPostgresBackend(db *pgx.ConnPool) discovery.StorageBackend {
	return &PostgresBackend{db: db}
}

type PostgresBackend struct {
	db *pgx.ConnPool
}

func (b *PostgresBackend) CreateCluster(cluster *discovery.Cluster) error {
	return b.db.QueryRow("INSERT INTO clusters (creator_ip, creator_user_agent) VALUES ($1, $2) RETURNING cluster_id, created_at",
		cluster.CreatorIP, cluster.CreatorUserAgent).Scan(&cluster.ID, &cluster.CreatedAt)
}

func (b *PostgresBackend) CreateInstance(inst *discovery.Instance) error {
	if inst.SSHPublicKeys == nil {
		inst.SSHPublicKeys = []discovery.SSHPublicKey{}
	}
	// pgx doesn't like unmarshalling into **time.Time
	inst.CreatedAt = &time.Time{}
	sshKeys, _ := json.Marshal(inst.SSHPublicKeys)
	err := b.db.QueryRow("INSERT INTO instances (cluster_id, flynn_version, ssh_public_keys, url, name, creator_ip) VALUES ($1, $2, $3, $4, $5, $6) RETURNING instance_id, created_at",
		inst.ClusterID, inst.FlynnVersion, string(sshKeys), inst.URL, inst.Name, inst.CreatorIP).Scan(&inst.ID, inst.CreatedAt)
	if pgErr, ok := err.(pgx.PgError); ok && pgErr.Code == "23505" /*duplicate key violates unique constraint*/ && pgErr.ConstraintName == "instances_cluster_id_url_key" {
		row := b.db.QueryRow("SELECT instance_id, flynn_version, ssh_public_keys, url, name, creator_ip, created_at FROM instances WHERE cluster_id = $1 AND url = $2", inst.ClusterID, inst.URL)
		if err := scanInstance(row, inst); err != nil {
			return err
		}
		return discovery.ErrExists
	}
	return err
}

type pgxScanner interface {
	Scan(...interface{}) error
}

func scanInstance(row pgxScanner, inst *discovery.Instance) error {
	if inst.CreatedAt == nil {
		inst.CreatedAt = &time.Time{}
	}
	var sshKeys string
	if err := row.Scan(&inst.ID, &inst.FlynnVersion, &sshKeys, &inst.URL, &inst.Name, &inst.CreatorIP, inst.CreatedAt); err != nil {
		return err
	}
	if err := json.Unmarshal([]byte(sshKeys), &inst.SSHPublicKeys); err != nil {
		return err
	}
	return nil
}

func (b *PostgresBackend) GetClusterInstances(clusterID string) ([]*discovery.Instance, error) {
	rows, err := b.db.Query("SELECT instance_id, flynn_version, ssh_public_keys, url, name, creator_ip, created_at FROM instances WHERE cluster_id = $1", clusterID)
	if err != nil {
		return nil, err
	}
	var instances []*discovery.Instance
	for rows.Next() {
		inst := &discovery.Instance{ClusterID: clusterID}
		if err := scanInstance(rows, inst); err != nil {
			rows.Close()
			return nil, err
		}
		instances = append(instances, inst)
	}
	return instances, rows.Err()
}
