package firestore

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"strings"

	"cloud.google.com/go/firestore"
	"github.com/flynn/flynn-discovery/discovery"
	"github.com/flynn/flynn/pkg/random"
)

func NewBackend(c *firestore.Client, baseCollection string) discovery.StorageBackend {
	return &backend{
		client: c,
		c:      c.Collection(baseCollection + "/flynn-discovery/clusters"),
	}
}

type backend struct {
	client *firestore.Client
	c      *firestore.CollectionRef
}

func (b *backend) CreateCluster(ctx context.Context, cluster *discovery.Cluster) error {
	cluster.ID = random.UUID()
	wr, err := b.c.Doc(cluster.ID).Create(ctx, cluster)
	if err != nil {
		return err
	}
	cluster.CreatedAt = wr.UpdateTime
	return nil
}

func (b *backend) CreateInstance(ctx context.Context, inst *discovery.Instance) error {
	docID := instanceID(inst.ClusterID, inst.URL)
	inst.ID = random.UUID()
	dr := b.c.Doc(inst.ClusterID).Collection("instances").Doc(docID)
	wr, err := dr.Create(ctx, inst)
	if err != nil && strings.Contains(err.Error(), "already exists") {
		doc, err := dr.Get(ctx)
		if err != nil {
			return err
		}
		if err := doc.DataTo(inst); err != nil {
			return err
		}
		inst.CreatedAt = &doc.CreateTime
		return discovery.ErrExists
	} else if err != nil {
		return err
	}
	inst.CreatedAt = &wr.UpdateTime
	return nil
}

func (b *backend) GetClusterInstances(ctx context.Context, clusterID string) ([]*discovery.Instance, error) {
	refs, err := b.c.Doc(clusterID).Collection("instances").DocumentRefs(ctx).GetAll()
	if err != nil {
		return nil, err
	}
	docs, err := b.client.GetAll(ctx, refs)
	if err != nil {
		return nil, err
	}
	res := make([]*discovery.Instance, len(docs))
	for i, d := range docs {
		inst := &discovery.Instance{}
		if err := d.DataTo(inst); err != nil {
			return nil, err
		}
		inst.ClusterID = clusterID
		inst.CreatedAt = &d.CreateTime
		res[i] = inst
	}

	return res, nil
}

func instanceID(clusterID, instURL string) string {
	hash := sha256.Sum256([]byte("v1 " + clusterID + " " + instURL))
	return base64.URLEncoding.EncodeToString(hash[:])
}
