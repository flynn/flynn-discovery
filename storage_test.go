package main

import (
	"context"
	"fmt"
	"os"
	"reflect"
	"testing"

	"cloud.google.com/go/firestore"
	"github.com/flynn/flynn-discovery/discovery"
	firestoreb "github.com/flynn/flynn-discovery/firestore"
	"github.com/flynn/flynn/pkg/random"
)

func TestFirestoreBackend(t *testing.T) {
	projectID := os.Getenv("GCP_PROJECT")
	if projectID == "" {
		t.Skip()
	}

	client, err := firestore.NewClient(context.Background(), projectID)
	if err != nil {
		t.Fatalf("error creating firestore client: %s", err)
	}

	collection := fmt.Sprintf("flynn-discovery-test-%s", random.UUID())
	testStorageBackend(t, firestoreb.NewBackend(client, collection))
}

func testStorageBackend(t *testing.T, b discovery.StorageBackend) {
	cluster := &discovery.Cluster{CreatorIP: "1.1.1.1", CreatorUserAgent: "foo/1.0"}
	if err := b.CreateCluster(context.Background(), cluster); err != nil {
		t.Fatalf("error creating cluster: %s", err)
	}
	if cluster.ID == "" {
		t.Errorf("cluster wasn't assigned ID")
	}
	if cluster.CreatedAt.IsZero() {
		t.Errorf("cluster wasn't assigned timestamp")
	}

	inst1 := discovery.Instance{
		ClusterID:     cluster.ID,
		FlynnVersion:  "20190101.1",
		SSHPublicKeys: []discovery.SSHPublicKey{{Type: "ssh-rsa", Data: []byte("asdf")}},
		URL:           "http://1.1.1.1:1113",
		Name:          "asdf",
		CreatorIP:     "1.1.1.1",
	}
	if err := b.CreateInstance(context.Background(), &inst1); err != nil {
		t.Fatalf("error creating first instance: %s", err)
	}
	if inst1.ID == "" {
		t.Errorf("first instance wasn't assigned ID")
	}
	if inst1.CreatedAt == nil || inst1.CreatedAt.IsZero() {
		t.Errorf("first instance wasn't assigned timestamp")
	}

	inst2 := inst1
	inst2.ID = ""
	inst2.CreatedAt = nil
	// create same instance again
	if err := b.CreateInstance(context.Background(), &inst2); err != discovery.ErrExists {
		t.Errorf("wrong error when recreating instance with duplicate URL, expected %#v, got %#v", discovery.ErrExists, err)
	}
	if inst2.ID != inst1.ID {
		t.Errorf("unexpected ID for existing instance: expected %q, got %q", inst1.ID, inst2.ID)
	}
	if inst2.CreatedAt == nil || !inst2.CreatedAt.Equal(*inst1.CreatedAt) {
		t.Errorf("unexpected timestamp for existing instance: expected %s, got %s", inst1.CreatedAt, inst2.CreatedAt)
	}

	inst3 := inst1
	inst3.ID = ""
	inst3.CreatedAt = nil
	inst3.URL = "https://2.2.2.2:1113"
	if err := b.CreateInstance(context.Background(), &inst3); err != nil {
		t.Fatalf("error creating third instance: %s", err)
	}
	if inst3.ID == "" {
		t.Errorf("third instance wasn't assigned ID")
	}
	if inst3.CreatedAt == nil || inst3.CreatedAt.IsZero() {
		t.Errorf("third instance wasn't assigned timestamp")
	}

	insts, err := b.GetClusterInstances(context.Background(), cluster.ID)
	if err != nil {
		t.Fatalf("error getting cluster instance list: %s", err)
	}

	if len(insts) != 2 {
		t.Fatalf("unexpected instance list, wanted 2, got %d: %#v", len(insts), insts)
	}

	want := []*discovery.Instance{&inst1, &inst3}
	if insts[0].ID == want[1].ID {
		want[0], want[1] = want[1], want[0]
	}
	if !reflect.DeepEqual(want, insts) {
		t.Fatalf("unexpected instance list:\n\twant: %#v\n\tgot:  %#v", want, insts)
	}

	insts, err = b.GetClusterInstances(context.Background(), random.UUID())
	if err != nil {
		t.Fatalf("error getting nonexistant cluster instance list: %s", err)
	}
	if len(insts) != 0 {
		t.Fatalf("unexpected instance list, wanted 0, got %d: %#v", len(insts), insts)
	}
}
