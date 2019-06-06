package discovery

import (
	"context"
	"errors"
	"time"
)

type Cluster struct {
	ID               string    `firestore:"-"`
	CreatorIP        string    `firestore:"creator_ip"`
	CreatorUserAgent string    `firestore:"creator_user_agent"`
	CreatedAt        time.Time `firestore:"-"`
}

type Instance struct {
	ID            string         `json:"id" firestore:"id"`
	ClusterID     string         `json:"cluster_id" firestore:"-"`
	FlynnVersion  string         `json:"flynn_version,omitempty" firestore:"flynn_version"`
	SSHPublicKeys []SSHPublicKey `json:"ssh_public_keys,omitempty" firestore:"ssh_public_keys"`
	URL           string         `json:"url,omitempty" firestore:"url"`
	Name          string         `json:"name,omitempty" firestore:"name"`
	CreatorIP     string         `json:"-" firestore:"creator_ip"`
	CreatedAt     *time.Time     `json:"created_at,omitempty" firestore:"-"`
}

type SSHPublicKey struct {
	Type string `json:"type" firestore:"type"`
	Data []byte `json:"data" firestore:"data"`
}

var ErrExists = errors.New("object exists")

type StorageBackend interface {
	CreateCluster(context.Context, *Cluster) error
	CreateInstance(context.Context, *Instance) error
	GetClusterInstances(ctx context.Context, clusterID string) ([]*Instance, error)
}
