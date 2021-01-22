package output

import (
	"reflect"
	"testing"
	"time"

	"github.com/kr/pretty"

	"github.com/croomes/kubectl-plugin/namespace"
	"github.com/croomes/kubectl-plugin/node"
	"github.com/croomes/kubectl-plugin/pkg/health"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/labels"
	"github.com/croomes/kubectl-plugin/pkg/version"
	"github.com/croomes/kubectl-plugin/volume"
)

func TestNewVolume(t *testing.T) {
	t.Parallel()

	labelsFromPairs := func(t *testing.T, pairs ...string) labels.Set {
		set, err := labels.NewSetFromPairs(pairs)
		if err != nil {
			t.Errorf("failed to set up test case: %v", err)
		}
		return set
	}

	tests := []struct {
		name string

		vol   *volume.Resource
		ns    *namespace.Resource
		nodes map[id.Node]*node.Resource

		wantOutputVol *Volume
	}{
		{
			name: "ok when master with nil sync progress",

			vol: &volume.Resource{
				ID:          "vol-id",
				Name:        "vol-name",
				Description: "vol-description",
				AttachedOn:  "attached-node",
				Nfs: volume.NFSConfig{
					Exports: []volume.NFSExportConfig{
						{
							ExportID:   1,
							Path:       "/",
							PseudoPath: "/",
							ACLs: []volume.NFSExportConfigACL{
								{
									Identity: volume.NFSExportConfigACLIdentity{
										IdentityType: "cidr",
										Matcher:      "10.0.0.0/8",
									},
									SquashConfig: volume.NFSExportConfigACLSquashConfig{
										GID:    0,
										UID:    0,
										Squash: "root",
									},
									AccessLevel: "rw",
								},
							},
						},
					},
					ServiceEndpoint: "10.0.0.1:/",
				},
				Namespace:  "namespace-id",
				Labels:     labelsFromPairs(t, "a=b", "b=c"),
				Filesystem: volume.FsTypeFromString("BLOCK"),
				SizeBytes:  42,

				Master: &volume.Deployment{
					ID:           "deploy-id",
					Node:         "node-id",
					Health:       health.MasterOnline,
					Promotable:   true,
					SyncProgress: nil, // explicitly nil
				},

				Replicas:  nil,
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Version:   version.FromString("some-version"),
			},
			ns: &namespace.Resource{
				Name: "namespace-name",
			},
			nodes: map[id.Node]*node.Resource{
				"attached-node": &node.Resource{
					Name: "attached-node-name",
				},
				"node-id": &node.Resource{
					Name: "node-name",
				},
			},

			wantOutputVol: &Volume{
				ID:             "vol-id",
				Name:           "vol-name",
				Description:    "vol-description",
				AttachedOn:     "attached-node",
				AttachedOnName: "attached-node-name",
				NFS: NFSConfig{
					Exports: []NFSExportConfig{
						{
							ExportID:   1,
							Path:       "/",
							PseudoPath: "/",
							ACLs: []NFSExportConfigACL{
								{
									Identity: NFSExportConfigACLIdentity{
										IdentityType: "cidr",
										Matcher:      "10.0.0.0/8",
									},
									SquashConfig: NFSExportConfigACLSquashConfig{
										GID:    0,
										UID:    0,
										Squash: "root",
									},
									AccessLevel: "rw",
								},
							},
						},
					},
					ServiceEndpoint: "10.0.0.1:/",
				},
				Namespace:     "namespace-id",
				NamespaceName: "namespace-name",
				Labels:        labelsFromPairs(t, "a=b", "b=c"),
				Filesystem:    volume.FsTypeFromString("BLOCK"),
				SizeBytes:     42,
				Master: &Deployment{
					ID:           "deploy-id",
					Node:         "node-id",
					NodeName:     "node-name",
					Health:       health.MasterOnline,
					Promotable:   true,
					SyncProgress: nil,
				},
				Replicas:  []*Deployment{},
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Version:   version.FromString("some-version"),
			},
		},
		{
			name: "ok when replicas both with sync progress and without",

			vol: &volume.Resource{
				ID:          "vol-id",
				Name:        "vol-name",
				Description: "vol-description",
				AttachedOn:  "attached-node",
				Nfs: volume.NFSConfig{
					Exports: []volume.NFSExportConfig{
						{
							ExportID:   1,
							Path:       "/",
							PseudoPath: "/",
							ACLs: []volume.NFSExportConfigACL{
								{
									Identity: volume.NFSExportConfigACLIdentity{
										IdentityType: "cidr",
										Matcher:      "10.0.0.0/8",
									},
									SquashConfig: volume.NFSExportConfigACLSquashConfig{
										GID:    0,
										UID:    0,
										Squash: "root",
									},
									AccessLevel: "rw",
								},
							},
						},
					},
					ServiceEndpoint: "10.0.0.1:/",
				},
				Namespace:  "namespace-id",
				Labels:     labelsFromPairs(t, "b=b", "a=c"),
				Filesystem: volume.FsTypeFromString("BLOCK"),
				SizeBytes:  42,

				Master: &volume.Deployment{
					ID:         "deploy-id",
					Node:       "node-id",
					Health:     health.MasterOnline,
					Promotable: true,
				},

				Replicas: []*volume.Deployment{
					{
						ID:           "repl-1",
						Node:         "node-1",
						Health:       health.ReplicaReady,
						Promotable:   true,
						SyncProgress: nil,
					},
					{
						ID:         "repl-2",
						Node:       "node-2",
						Health:     health.ReplicaSyncing,
						Promotable: false,
						SyncProgress: &volume.SyncProgress{
							BytesRemaining:            6,
							ThroughputBytes:           4,
							EstimatedSecondsRemaining: 2,
						},
					},
				},
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Version:   version.FromString("some-version"),
			},
			ns: &namespace.Resource{
				Name: "namespace-name",
			},
			nodes: map[id.Node]*node.Resource{
				"attached-node": &node.Resource{
					Name: "attached-node-name",
				},
				"node-id": &node.Resource{
					Name: "node-name",
				},
				"node-1": &node.Resource{
					Name: "node-1-name",
				},
				"node-2": &node.Resource{
					Name: "node-2-name",
				},
			},

			wantOutputVol: &Volume{
				ID:             "vol-id",
				Name:           "vol-name",
				Description:    "vol-description",
				AttachedOn:     "attached-node",
				AttachedOnName: "attached-node-name",
				NFS: NFSConfig{
					Exports: []NFSExportConfig{
						{
							ExportID:   1,
							Path:       "/",
							PseudoPath: "/",
							ACLs: []NFSExportConfigACL{
								{
									Identity: NFSExportConfigACLIdentity{
										IdentityType: "cidr",
										Matcher:      "10.0.0.0/8",
									},
									SquashConfig: NFSExportConfigACLSquashConfig{
										GID:    0,
										UID:    0,
										Squash: "root",
									},
									AccessLevel: "rw",
								},
							},
						},
					},
					ServiceEndpoint: "10.0.0.1:/",
				},
				Namespace:     "namespace-id",
				NamespaceName: "namespace-name",
				Labels:        labelsFromPairs(t, "b=b", "a=c"),
				Filesystem:    volume.FsTypeFromString("BLOCK"),
				SizeBytes:     42,
				Master: &Deployment{
					ID:         "deploy-id",
					Node:       "node-id",
					NodeName:   "node-name",
					Health:     health.MasterOnline,
					Promotable: true,
				},
				Replicas: []*Deployment{
					{
						ID:           "repl-1",
						Node:         "node-1",
						NodeName:     "node-1-name",
						Health:       health.ReplicaReady,
						Promotable:   true,
						SyncProgress: nil,
					},
					{
						ID:         "repl-2",
						Node:       "node-2",
						NodeName:   "node-2-name",
						Health:     health.ReplicaSyncing,
						Promotable: false,
						SyncProgress: &SyncProgress{
							BytesRemaining:            6,
							ThroughputBytes:           4,
							EstimatedSecondsRemaining: 2,
						},
					},
				},
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Version:   version.FromString("some-version"),
			},
		},
		{
			name: "missing attached on node information",

			vol: &volume.Resource{
				ID:          "vol-id",
				Name:        "vol-name",
				Description: "vol-description",
				AttachedOn:  "attached-node",
				Nfs: volume.NFSConfig{
					Exports: []volume.NFSExportConfig{
						{
							ExportID:   1,
							Path:       "/",
							PseudoPath: "/",
							ACLs: []volume.NFSExportConfigACL{
								{
									Identity: volume.NFSExportConfigACLIdentity{
										IdentityType: "cidr",
										Matcher:      "10.0.0.0/8",
									},
									SquashConfig: volume.NFSExportConfigACLSquashConfig{
										GID:    0,
										UID:    0,
										Squash: "root",
									},
									AccessLevel: "rw",
								},
							},
						},
					},
					ServiceEndpoint: "10.0.0.1:/",
				},
				Namespace:  "namespace-id",
				Labels:     labelsFromPairs(t, "b=b", "a=c"),
				Filesystem: volume.FsTypeFromString("BLOCK"),
				SizeBytes:  42,

				Master: &volume.Deployment{
					ID:         "deploy-id",
					Node:       "node-id",
					Health:     health.MasterOnline,
					Promotable: true,
				},

				Replicas:  []*volume.Deployment{},
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Version:   version.FromString("some-version"),
			},
			ns: &namespace.Resource{
				Name: "namespace-name",
			},
			nodes: map[id.Node]*node.Resource{
				"node-id": &node.Resource{
					Name: "node-name",
				},
			},

			wantOutputVol: &Volume{
				ID:             "vol-id",
				Name:           "vol-name",
				Description:    "vol-description",
				AttachedOn:     "attached-node",
				AttachedOnName: "unknown",
				NFS: NFSConfig{
					Exports: []NFSExportConfig{
						{
							ExportID:   1,
							Path:       "/",
							PseudoPath: "/",
							ACLs: []NFSExportConfigACL{
								{
									Identity: NFSExportConfigACLIdentity{
										IdentityType: "cidr",
										Matcher:      "10.0.0.0/8",
									},
									SquashConfig: NFSExportConfigACLSquashConfig{
										GID:    0,
										UID:    0,
										Squash: "root",
									},
									AccessLevel: "rw",
								},
							},
						},
					},
					ServiceEndpoint: "10.0.0.1:/",
				},
				Namespace:     "namespace-id",
				NamespaceName: "namespace-name",
				Labels:        labelsFromPairs(t, "b=b", "a=c"),
				Filesystem:    volume.FsTypeFromString("BLOCK"),
				SizeBytes:     42,
				Master: &Deployment{
					ID:           "deploy-id",
					Node:         "node-id",
					NodeName:     "node-name",
					Health:       health.MasterOnline,
					Promotable:   true,
					SyncProgress: nil,
				},
				Replicas:  []*Deployment{},
				CreatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				UpdatedAt: time.Date(2000, 1, 1, 0, 0, 0, 0, time.UTC),
				Version:   version.FromString("some-version"),
			},
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotOutputVol := NewVolume(tt.vol, tt.ns, tt.nodes)

			if !reflect.DeepEqual(gotOutputVol, tt.wantOutputVol) {
				pretty.Ldiff(t, gotOutputVol, tt.wantOutputVol)
				t.Errorf("got output vol %v, want %v", pretty.Sprint(gotOutputVol), pretty.Sprint(tt.wantOutputVol))
			}
		})
	}
}

func TestNewDeployment(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string

		inDeployment *volume.Deployment
		nodes        map[id.Node]*node.Resource

		wantOutputDeployment *Deployment
	}{
		{
			name: "ok",

			inDeployment: &volume.Deployment{
				ID:         "id",
				Node:       "node-id",
				Health:     "health",
				Promotable: true,
			},
			nodes: map[id.Node]*node.Resource{
				"node-id": &node.Resource{
					Name: "node-name",
				},
			},

			wantOutputDeployment: &Deployment{
				ID:         "id",
				Node:       "node-id",
				NodeName:   "node-name",
				Health:     "health",
				Promotable: true,
			},
		},
		{
			name: "missing",

			inDeployment: &volume.Deployment{
				ID:         "id",
				Node:       "node-id",
				Health:     "health",
				Promotable: true,
			},
			nodes: map[id.Node]*node.Resource{
				"some-other-node-id": &node.Resource{
					Name: "node-name",
				},
			},

			wantOutputDeployment: &Deployment{
				ID:         "id",
				Node:       "node-id",
				NodeName:   "unknown",
				Health:     "health",
				Promotable: true,
			},
		},
	}

	for _, tt := range tests {
		var tt = tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			gotOutputDeployment := newDeployment(tt.inDeployment, tt.nodes)

			if !reflect.DeepEqual(gotOutputDeployment, tt.wantOutputDeployment) {
				pretty.Ldiff(t, gotOutputDeployment, tt.wantOutputDeployment)
				t.Errorf("got output %v, want %v", gotOutputDeployment, tt.wantOutputDeployment)
			}
		})
	}
}
