package apiclient

import "github.com/croomes/kubectl-plugin/pkg/version"

// UpdateClusterRequestParams contains optional request parameters for a update
// cluster operation.
type UpdateClusterRequestParams struct {
	CASVersion version.Version
}
