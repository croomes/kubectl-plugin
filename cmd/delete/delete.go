package delete

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/namespace"
	"github.com/croomes/kubectl-plugin/node"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/output/jsonformat"
	"github.com/croomes/kubectl-plugin/output/textformat"
	"github.com/croomes/kubectl-plugin/output/yamlformat"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/policygroup"
	"github.com/croomes/kubectl-plugin/user"
	"github.com/croomes/kubectl-plugin/volume"
)

// ConfigProvider specifies the configuration settings which commands require
// access to.
type ConfigProvider interface {
	Username() (string, error)
	Password() (string, error)

	CommandTimeout() (time.Duration, error)
	UseIDs() (bool, error)
	Namespace() (string, error)
	OutputFormat() (output.Format, error)
}

// Client defines the functionality required by the CLI application to
// reasonably implement the "delete" verb commands.
type Client interface {
	Authenticate(ctx context.Context, username, password string) (apiclient.AuthSession, error)

	GetUserByName(ctx context.Context, username string) (*user.Resource, error)
	GetNodeByName(ctx context.Context, name string) (*node.Resource, error)
	GetNamespaceByName(ctx context.Context, name string) (*namespace.Resource, error)
	GetVolumeByName(ctx context.Context, namespaceID id.Namespace, name string) (*volume.Resource, error)
	GetPolicyGroupByName(ctx context.Context, name string) (*policygroup.Resource, error)

	DeleteUser(ctx context.Context, uid id.User, params *apiclient.DeleteUserRequestParams) error
	DeleteNode(ctx context.Context, nodeID id.Node, params *apiclient.DeleteNodeRequestParams) error
	DeleteVolume(ctx context.Context, namespaceID id.Namespace, volumeID id.Volume, params *apiclient.DeleteVolumeRequestParams) error
	DeleteNamespace(ctx context.Context, uid id.Namespace, params *apiclient.DeleteNamespaceRequestParams) error
	DeletePolicyGroup(ctx context.Context, uid id.PolicyGroup, params *apiclient.DeletePolicyGroupRequestParams) error
}

// Displayer defines the functionality required by the CLI application to
// display the results gathered by the "delete" verb commands.
type Displayer interface {
	DeleteUser(ctx context.Context, w io.Writer, confirmation output.UserDeletion) error
	DeleteVolume(ctx context.Context, w io.Writer, confirmation output.VolumeDeletion) error
	AsyncRequest(ctx context.Context, w io.Writer) error
	DeleteNode(ctx context.Context, w io.Writer, confirmation output.NodeDeletion) error
	DeleteNamespace(ctx context.Context, w io.Writer, confirmation output.NamespaceDeletion) error
	DeletePolicyGroup(ctx context.Context, w io.Writer, confirmation output.PolicyGroupDeletion) error
}

// NewCommand configures the set of commands which are grouped by the "delete" verb.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	command := &cobra.Command{
		Use:   "delete",
		Short: "Delete resources in the cluster",
	}

	command.AddCommand(
		newVolume(os.Stdout, client, config),
		newNamespace(os.Stdout, client, config),
		newPolicyGroup(os.Stdout, client, config),
		newUser(os.Stdout, client, config),
		newNode(os.Stdout, client, config),
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
