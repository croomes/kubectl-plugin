package volume

import (
	"strconv"
	"time"

	"github.com/croomes/kubectl-plugin/pkg/health"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/labels"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

const (
	// LabelNoCache is a StorageOS volume label which when enabled disables the
	// caching of volume data.
	LabelNoCache = "storageos.com/nocache"
	// LabelNoCompress is a StorageOS volume label which when enabled disables the
	// compression of volume data (both at rest and during transit).
	LabelNoCompress = "storageos.com/nocompress"
	// LabelReplicas is a StorageOS volume label which decides how many replicas
	// must be provisioned for that volume.
	LabelReplicas = "storageos.com/replicas"
	// LabelThrottle is a StorageOS volume label which when enabled deprioritises
	// the volume's traffic by reducing disk I/O rate.
	LabelThrottle = "storageos.com/throttle"
)

// FsType indicates the kind of filesystem which a volume has been given.
type FsType string

// String returns the name string for fs.
func (fs FsType) String() string {
	return string(fs)
}

// FsTypeFromString wraps name as an FsType. It doesn't perform validity
// checks.
func FsTypeFromString(name string) FsType {
	return FsType(name)
}

// AttachType The attachment type of a volume. "host" indicates that the volume
// is consumed by the node it is attached to.
type AttachType string

// List of AttachType
const (
	AttachTypeUnknown  AttachType = "unknown"
	AttachTypeDetached AttachType = "detached"
	AttachTypeNFS      AttachType = "nfs"
	AttachTypeHost     AttachType = "host"
)

// AttachTypeFromString wraps name as an AttachType. It doesn't perform validity
// checks.
func AttachTypeFromString(name string) AttachType {
	return AttachType(name)
}

// String returns the string representation of the current AttachType
func (a AttachType) String() string {
	return string(a)
}

// Resource encapsulates a StorageOS volume API resource as a data type.
type Resource struct {
	ID             id.Volume  `json:"id"`
	Name           string     `json:"name"`
	Description    string     `json:"description"`
	AttachedOn     id.Node    `json:"attachedOn"`
	AttachmentType AttachType `json:"attachmentType"`
	Nfs            NFSConfig  `json:"nfs"`

	Namespace  id.Namespace `json:"namespaceID"`
	Labels     labels.Set   `json:"labels"`
	Filesystem FsType       `json:"filesystem"`
	SizeBytes  uint64       `json:"sizeBytes"`

	Master   *Deployment   `json:"master"`
	Replicas []*Deployment `json:"replicas"`

	CreatedAt time.Time       `json:"createdAt"`
	UpdatedAt time.Time       `json:"updatedAt"`
	Version   version.Version `json:"version"`
}

// Deployment encapsulates a deployment instance for a
// volume as a data type.
type Deployment struct {
	ID           id.Deployment      `json:"id"`
	Node         id.Node            `json:"nodeID"`
	Health       health.VolumeState `json:"health"`
	Promotable   bool               `json:"promotable"`
	SyncProgress *SyncProgress      `json:"syncProgress,omitempty"`
}

// SyncProgress is a point-in-time snapshot of an ongoing sync operation.
type SyncProgress struct {
	BytesRemaining            uint64 `json:"bytesRemaining"`
	ThroughputBytes           uint64 `json:"throughputBytes"`
	EstimatedSecondsRemaining uint64 `json:"estimatedSecondsRemaining"`
}

// NFSConfig contains a config for NFS attaching containing and endpoint and a
// list of exports.
type NFSConfig struct {
	Exports         []NFSExportConfig `json:"exports"`
	ServiceEndpoint string            `json:"serviceEndpoint"`
}

// NFSExportConfig contains a single export configuration for NFS attaching.
type NFSExportConfig struct {
	ExportID   uint                 `json:"exportID"`
	Path       string               `json:"path"`
	PseudoPath string               `json:"pseudoPath"`
	ACLs       []NFSExportConfigACL `json:"acls"`
}

// NFSExportConfigACL contains a single ACL policy for NFS attaching export
// configuration.
type NFSExportConfigACL struct {
	Identity     NFSExportConfigACLIdentity     `json:"identity"`
	SquashConfig NFSExportConfigACLSquashConfig `json:"squashConfig"`
	AccessLevel  string                         `json:"accessLevel"`
}

// NFSExportConfigACLIdentity contains identity info for an ACL in a NFS export
// config.
type NFSExportConfigACLIdentity struct {
	IdentityType string `json:"identityType"`
	Matcher      string `json:"matcher"`
}

// NFSExportConfigACLSquashConfig contains squash info for an ACL in a NFS
// export config.
type NFSExportConfigACLSquashConfig struct {
	GID    int64  `json:"gid"`
	UID    int64  `json:"uid"`
	Squash string `json:"squash"`
}

// IsCachingDisabled returns if the volume resource is configured to disable
// caching of data.
func (r *Resource) IsCachingDisabled() (bool, error) {
	value, exists := r.Labels[LabelNoCache]
	if !exists {
		return false, nil
	}

	return strconv.ParseBool(value)
}

// IsCompressionDisabled returns if the volume resource is configured to disable
// compression of data at rest and during transit.
func (r *Resource) IsCompressionDisabled() (bool, error) {
	value, exists := r.Labels[LabelNoCompress]
	if !exists {
		return false, nil
	}

	return strconv.ParseBool(value)
}

// IsThrottleEnabled returns if the volume resource is configured to have its
// traffic deprioritised by reducing its disk I/O rate.
func (r *Resource) IsThrottleEnabled() (bool, error) {
	value, exists := r.Labels[LabelThrottle]
	if !exists {
		return false, nil
	}

	return strconv.ParseBool(value)
}
