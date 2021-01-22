package apiclient

import (
	"context"
	"fmt"
	"time"

	"github.com/croomes/kubectl-plugin/node"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

// DeleteNodeRequestParams contains optional request parameters for a delete
// node operation.
type DeleteNodeRequestParams struct {
	CASVersion version.Version
	AsyncMax   time.Duration
}

// NodeNotFoundError indicates that the API could not find the StorageOS node
// specified.
type NodeNotFoundError struct {
	uid  id.Node
	name string
}

// Error returns an error message indicating that the node with a given
// ID or name was not found, as configured.
func (e NodeNotFoundError) Error() string {
	switch {
	case e.uid != "":
		return fmt.Sprintf("node with ID %v not found", e.uid)
	case e.name != "":
		return fmt.Sprintf("node with name %v not found", e.name)
	}

	return "node not found"
}

// NewNodeNotFoundError returns a NodeNotFoundError for the node with uid.
func NewNodeNotFoundError(uid id.Node) NodeNotFoundError {
	return NodeNotFoundError{
		uid: uid,
	}
}

// NewNodeNameNotFoundError returns a NodeNotFoundError for the node with name.
func NewNodeNameNotFoundError(name string) NodeNotFoundError {
	return NodeNotFoundError{
		name: name,
	}
}

// GetNodeByName requests basic information for the node resource which has
// name.
//
// The resource model for the API is build around using unique identifiers,
// so this operation is inherently more expensive than the corresponding
// GetNode() operation.
//
// Retrieving a node resource by name involves requesting a list of all nodes
// in the cluster from the StorageOS API and returning the first node where the
// name matches.
func (c *Client) GetNodeByName(ctx context.Context, name string) (*node.Resource, error) {
	nodes, err := c.Transport.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	for _, n := range nodes {
		if n.Name == name {
			return n, nil
		}
	}

	return nil, NewNodeNameNotFoundError(name)
}

// GetListNodesByUID requests a list containing basic information on each
// node resource in the cluster.
//
// The returned list is filtered using uids so that it contains only those
// resources which have a matching ID. Omitting uids will skip the filtering.
func (c *Client) GetListNodesByUID(ctx context.Context, uids ...id.Node) ([]*node.Resource, error) {
	nodes, err := c.Transport.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	return filterNodesForUIDs(nodes, uids...)
}

// GetListNodesByName requests a list containing basic information on each
// node resource in the cluster.
//
// The returned list is filtered using names so that it contains only those
// resources which have a matching name. Omitting names will skip the filtering.
func (c *Client) GetListNodesByName(ctx context.Context, names ...string) ([]*node.Resource, error) {
	nodes, err := c.Transport.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	return filterNodesForNames(nodes, names...)
}

// filterNodesForNames will return a subset of nodes containing resources
// which have one of the provided names. If names is not provided, nodes is
// returned as is.
//
// If there is no resource for a given name then an error is returned, thus
// this is a strict helper.
func filterNodesForNames(nodes []*node.Resource, names ...string) ([]*node.Resource, error) {
	// return everything if no filter names given
	if len(names) == 0 {
		return nodes, nil
	}

	retrieved := map[string]*node.Resource{}

	for _, n := range nodes {
		retrieved[n.Name] = n
	}

	filtered := make([]*node.Resource, 0, len(names))

	for _, name := range names {
		n, ok := retrieved[name]
		if !ok {
			return nil, NewNodeNameNotFoundError(name)
		}
		filtered = append(filtered, n)
	}

	return filtered, nil
}

// filterNodesForUIDs will return a subset of nodes containing resources
// which have one of the provided uids. If uids is not provided, nodes is
// returned as is.
//
// If there is no resource for a given uid then an error is returned, thus
// this is a strict helper.
func filterNodesForUIDs(nodes []*node.Resource, uids ...id.Node) ([]*node.Resource, error) {
	// return everything if no filter uids given
	if len(uids) == 0 {
		return nodes, nil
	}

	retrieved := map[id.Node]*node.Resource{}

	for _, n := range nodes {
		retrieved[n.ID] = n
	}

	filtered := make([]*node.Resource, 0, len(uids))

	for _, idVar := range uids {
		n, ok := retrieved[idVar]
		if !ok {
			return nil, NewNodeNotFoundError(idVar)
		}
		filtered = append(filtered, n)
	}

	return filtered, nil
}
