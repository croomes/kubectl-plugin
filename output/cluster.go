package output

import (
	"time"

	"github.com/croomes/kubectl-plugin/cluster"
	"github.com/croomes/kubectl-plugin/node"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

// Cluster defines a type that contains all the info needed to be outputted.
type Cluster struct {
	ID id.Cluster `json:"id" yaml:"id"`

	DisableTelemetry      bool `json:"disableTelemetry" yaml:"disableTelemetry"`
	DisableCrashReporting bool `json:"disableCrashReporting" yaml:"disableCrashReporting"`
	DisableVersionCheck   bool `json:"disableVersionCheck" yaml:"disableVersionCheck"`

	LogLevel  cluster.LogLevel  `json:"logLevel" yaml:"logLevel"`
	LogFormat cluster.LogFormat `json:"logFormat" yaml:"logFormat"`

	CreatedAt time.Time       `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt" yaml:"updatedAt"`
	Version   version.Version `json:"version" yaml:"version"`

	Nodes []*Node `json:"nodes" yaml:"nodes"`
}

// NewCluster returns a new Cluster object that contains all the info needed
// to be outputted.
func NewCluster(c *cluster.Resource, nodes []*node.Resource) *Cluster {
	return &Cluster{
		ID:                    c.ID,
		DisableTelemetry:      c.DisableTelemetry,
		DisableCrashReporting: c.DisableCrashReporting,
		DisableVersionCheck:   c.DisableVersionCheck,
		LogLevel:              c.LogLevel,
		LogFormat:             c.LogFormat,
		CreatedAt:             c.CreatedAt,
		UpdatedAt:             c.UpdatedAt,
		Version:               c.Version,
		Nodes:                 NewNodes(nodes),
	}
}
