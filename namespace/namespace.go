package namespace

import (
	"time"

	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/labels"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

// Resource encapsulates a StorageOS namespace API resource as a data type.
type Resource struct {
	ID     id.Namespace `json:"id"`
	Name   string       `json:"name"`
	Labels labels.Set   `json:"labels"`

	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Version   version.Version `json:"version"`
}
