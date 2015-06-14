package main

import (
	"errors"
	"time"
)

type Cluster struct {
	ID               string
	CreatorIP        string
	CreatorUserAgent string
	CreatedAt        time.Time
}

type Instance struct {
	ID            string         `json:"id"`
	ClusterID     string         `json:"cluster_id"`
	FlynnVersion  string         `json:"flynn_version,omitempty"`
	SSHPublicKeys []SSHPublicKey `json:"ssh_public_keys,omitempty"`
	URL           string         `json:"url,omitempty"`
	Name          string         `json:"name,omitempty"`
	CreatorIP     string         `json:"-"`
	CreatedAt     *time.Time     `json:"created_at,omitempty"`
}

type SSHPublicKey struct {
	Type string `json:"type"`
	Data []byte `json:"data"`
}

var ErrExists = errors.New("object exists")

type StorageBackend interface {
	CreateCluster(*Cluster) error
	CreateInstance(instance *Instance) error
	GetClusterInstances(clusterID string) ([]*Instance, error)
}
