package output

import (
	"time"

	"github.com/croomes/kubectl-plugin/namespace"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/labels"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

// Namespace defines a type that contains all the info we need to output a
// namespace.
type Namespace struct {
	ID     id.Namespace `json:"id" yaml:"id"`
	Name   string       `json:"name" yaml:"name"`
	Labels labels.Set   `json:"labels" yaml:"labels"`

	CreatedAt time.Time       `json:"createdAt" yaml:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt" yaml:"updatedAt"`
	Version   version.Version `json:"version" yaml:"version"`
}

// NewNamespace returns a new Namespace object that contains all the info needed
// to be outputted.
func NewNamespace(n *namespace.Resource) *Namespace {
	return &Namespace{
		ID:        n.ID,
		Name:      n.Name,
		Labels:    n.Labels,
		CreatedAt: n.CreatedAt,
		UpdatedAt: n.UpdatedAt,
		Version:   n.Version,
	}
}

// NewNamespaces returns a list of Namespace objects that contains all the info
// needed to be outputted.
func NewNamespaces(ns []*namespace.Resource) []*Namespace {
	namespaces := make([]*Namespace, 0, len(ns))
	for _, n := range ns {
		namespaces = append(namespaces, NewNamespace(n))
	}
	return namespaces
}

// NamespaceDeletion defines a namespace deletion confirmation output
// representation.
type NamespaceDeletion struct {
	ID id.Namespace `json:"id" yaml:"id"`
}
