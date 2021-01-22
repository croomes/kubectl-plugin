package update

import (
	"context"
	"io"
	"time"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/namespace"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/output/jsonformat"
	"github.com/croomes/kubectl-plugin/output/textformat"
	"github.com/croomes/kubectl-plugin/output/yamlformat"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/labels"
	"github.com/croomes/kubectl-plugin/volume"
)

// ConfigProvider specifies the configuration
type ConfigProvider interface {
	Username() (string, error)
	Password() (string, error)

	CommandTimeout() (time.Duration, error)
	UseIDs() (bool, error)
	OutputFormat() (output.Format, error)
	Namespace() (string, error)
}

// Client defines the functionality required by the CLI application to
// reasonably implement the "update" verb commands.
type Client interface {
	Authenticate(ctx context.Context, username, password string) (apiclient.AuthSession, error)

	SetReplicas(ctx context.Context, nsID id.Namespace, volID id.Volume, numReplicas uint64, params *apiclient.SetReplicasRequestParams) error
	UpdateVolumeDescription(ctx context.Context, nsID id.Namespace, volID id.Volume, description string, params *apiclient.UpdateVolumeRequestParams) (*volume.Resource, error)
	UpdateVolumeLabels(ctx context.Context, nsID id.Namespace, volID id.Volume, labels labels.Set, params *apiclient.UpdateVolumeRequestParams) (*volume.Resource, error)
	ResizeVolume(ctx context.Context, nsID id.Namespace, volID id.Volume, sizeBytes uint64, params *apiclient.ResizeVolumeRequestParams) (*volume.Resource, error)

	GetVolumeByName(ctx context.Context, namespace id.Namespace, name string) (*volume.Resource, error)
	GetVolume(ctx context.Context, namespaceID id.Namespace, uid id.Volume) (*volume.Resource, error)
	GetNamespaceByName(ctx context.Context, name string) (*namespace.Resource, error)
}

// Displayer defines the functionality required by the CLI application to
// display the results returned by "update" verb operations.
type Displayer interface {
	SetReplicas(ctx context.Context, w io.Writer, new uint64) error
	ResizeVolume(ctx context.Context, w io.Writer, volUpdate output.VolumeUpdate) error
	UpdateVolumeDescription(ctx context.Context, w io.Writer, volUpdate output.VolumeUpdate) error
	UpdateVolumeLabels(ctx context.Context, w io.Writer, volUpdate output.VolumeUpdate) error
	AsyncRequest(ctx context.Context, w io.Writer) error
}

// NewCommand configures the set of commands which are grouped by the "update" verb.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	command := &cobra.Command{
		Use:   "update",
		Short: "Make changes to existing resources",
	}

	command.AddCommand(
		newVolumeUpdate(client, config),
	)

	return command
}

// SelectDisplayer returns the right command displayer specified in the
// config provider.
func SelectDisplayer(cp ConfigProvider) Displayer {
	out, err := cp.OutputFormat()
	if err != nil {
		return textformat.NewDisplayer(textformat.NewTimeFormatter())
	}

	switch out {
	case output.JSON:
		return jsonformat.NewDisplayer(jsonformat.DefaultEncodingIndent)
	case output.YAML:
		return yamlformat.NewDisplayer("")
	case output.Text:
		fallthrough
	default:
		return textformat.NewDisplayer(textformat.NewTimeFormatter())
	}
}
