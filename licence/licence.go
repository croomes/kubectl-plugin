package licence

import (
	"fmt"
	"time"

	"github.com/alecthomas/units"

	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

// Resource describes a StorageOS product licence and the features included with
// it.
type Resource struct {
	ClusterID            id.Cluster      `json:"clusterID"`
	ExpiresAt            time.Time       `json:"expiresAt"`
	ClusterCapacityBytes uint64          `json:"clusterCapacityBytes"`
	UsedBytes            uint64          `json:"usedBytes"`
	Kind                 string          `json:"kind"`
	CustomerName         string          `json:"customerName"`
	Features             []string        `json:"features"`
	Version              version.Version `json:"version"`
}

func (l *Resource) String() string {
	return fmt.Sprintf(`Cluster ID: %v
Expires at: %v
Cluster capacity: %v
Used Bytes: %v
Kind: %v
Customer name: %v
Features: %v
Version: %v
`,
		l.ClusterID,
		l.ExpiresAt.Format(time.RFC3339),
		units.Base2Bytes(l.ClusterCapacityBytes).String(),
		units.Base2Bytes(l.UsedBytes).String(),
		l.Kind,
		l.CustomerName,
		l.Features,
		l.Version,
	)
}
