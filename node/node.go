package node

import (
	"time"

	"github.com/croomes/kubectl-plugin/pkg/capacity"
	"github.com/croomes/kubectl-plugin/pkg/health"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/labels"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

// Resource encapsulates a StorageOS node API resource as a data type.
type Resource struct {
	ID       id.Node          `json:"id"`
	Name     string           `json:"name"`
	Health   health.NodeState `json:"health"`
	Capacity capacity.Stats   `json:"capacity,omitempty"`

	IOAddr         string `json:"ioAddress"`
	SupervisorAddr string `json:"supervisorAddress"`
	GossipAddr     string `json:"gossipAddress"`
	ClusteringAddr string `json:"clusteringAddress"`

	Labels labels.Set `json:"labels"`

	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Version   version.Version `json:"version"`
}
