package openapi

import (
	"context"

	"github.com/antihax/optional"

	"code.storageos.net/storageos/openapi"
	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/licence"
)

// GetLicence requests the current cluster licence from the StorageOS API,
// translating it into a *licence.Resource.
func (o *OpenAPI) GetLicence(ctx context.Context) (*licence.Resource, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	model, resp, err := o.client.DefaultApi.GetLicence(ctx)
	if err != nil {
		return nil, mapOpenAPIError(err, resp)
	}

	return o.codec.decodeLicence(model)
}

// UpdateLicence sends a new version of the licence to apply to the current
// cluster. It returns the new licence resource if correctly applied.
func (o *OpenAPI) UpdateLicence(ctx context.Context, licence []byte, params *apiclient.UpdateLicenceRequestParams) (*licence.Resource, error) {
	o.mu.RLock()
	defer o.mu.RUnlock()

	// default
	req := openapi.UpdateLicence{
		Key: string(licence),
	}
	opts := &openapi.UpdateLicenceOpts{
		IgnoreVersion: optional.NewBool(true),
	}

	// check optional params
	if params != nil && params.CASVersion != "" {
		req.Version = params.CASVersion.String()
		opts.IgnoreVersion = optional.NewBool(false)
	}

	lic, resp, err := o.client.DefaultApi.UpdateLicence(ctx, req, opts)
	if err != nil {
		return nil, mapOpenAPIError(err, resp)
	}

	return o.codec.decodeLicence(lic)
}
