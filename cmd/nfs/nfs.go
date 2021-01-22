package nfs

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/licence"
	"github.com/croomes/kubectl-plugin/namespace"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/output/jsonformat"
	"github.com/croomes/kubectl-plugin/output/textformat"
	"github.com/croomes/kubectl-plugin/output/yamlformat"
	"github.com/croomes/kubectl-plugin/pkg/id"
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
	GetLicence(ctx context.Context) (*licence.Resource, error)

	GetVolumeByName(ctx context.Context, namespace id.Namespace, name string) (*volume.Resource, error)
	GetVolume(ctx context.Context, namespaceID id.Namespace, uid id.Volume) (*volume.Resource, error)
	GetNamespaceByName(ctx context.Context, name string) (*namespace.Resource, error)

	AttachNFSVolume(ctx context.Context, namespaceID id.Namespace, volumeID id.Volume, params *apiclient.AttachNFSVolumeRequestParams) error
	UpdateNFSVolumeExports(ctx context.Context, namespaceID id.Namespace, volumeID id.Volume, exports []volume.NFSExportConfig, params *apiclient.UpdateNFSVolumeExportsRequestParams) error
	UpdateNFSVolumeMountEndpoint(ctx context.Context, namespaceID id.Namespace, volumeID id.Volume, endpoint string, params *apiclient.UpdateNFSVolumeMountEndpointRequestParams) error
}

// Displayer defines the functionality required by the CLI application to
// display the results returned by "update" verb operations.
type Displayer interface {
	UpdateNFSVolumeMountEndpoint(ctx context.Context, w io.Writer, volID id.Volume, endpoint string) error
	UpdateNFSVolumeExports(ctx context.Context, w io.Writer, volID id.Volume, exports []output.NFSExportConfig) error
	AttachVolume(ctx context.Context, w io.Writer) error
	AsyncRequest(ctx context.Context, w io.Writer) error
}

// NewCommand configures the set of commands which are grouped by the "nfs" verb.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	command := &cobra.Command{
		Use:   "nfs",
		Short: "Make changes and attach nfs volumes",
	}

	command.AddCommand(
		newAttach(os.Stdout, client, config),
		newSetEndpoint(os.Stdout, client, config),
		newSetExports(os.Stdout, client, config),
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
