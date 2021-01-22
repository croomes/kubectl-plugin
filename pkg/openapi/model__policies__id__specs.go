/*
 * StorageOS API
 *
 * No description provided (generated by Openapi Generator https://github.com/openapitools/openapi-generator)
 *
 * API version: 2.3.0-alpha
 * Contact: info@storageos.com
 * Generated by: OpenAPI Generator (https://openapi-generator.tech)
 */

package openapi

// PoliciesIdSpecs struct for PoliciesIdSpecs
type PoliciesIdSpecs struct {
	// A unique identifier for a namespace. The format of this type is undefined and may change but the defined properties will not change..
	NamespaceID string `json:"namespaceID,omitempty"`
	// The resource type this policy grants access to.
	ResourceType string `json:"resourceType,omitempty"`
	// If true, disallows requests that attempt to mutate the resource.
	ReadOnly bool `json:"readOnly,omitempty"`
}
