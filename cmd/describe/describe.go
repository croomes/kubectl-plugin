package describe

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/cluster"
	"github.com/croomes/kubectl-plugin/licence"
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
	OutputFormat() (output.Format, error)
	Namespace() (string, error)
}

// Client describes the functionality required by the CLI application
// to reasonably implement the "describe" verb commands.
type Client interface {
	Authenticate(ctx context.Context, username, password string) (apiclient.AuthSession, error)

	GetCluster(ctx context.Context) (*cluster.Resource, error)
	GetLicence(ctx context.Context) (*licence.Resource, error)
	GetUser(ctx context.Context, userID id.User) (*user.Resource, error)
	GetUserByName(ctx context.Context, username string) (*user.Resource, error)
	ListUsers(ctx context.Context) ([]*user.Resource, error)
	GetNode(ctx context.Context, uid id.Node) (*node.Resource, error)
	GetNodeByName(ctx context.Context, name string) (*node.Resource, error)
	ListNodes(ctx context.Context) ([]*node.Resource, error)
	GetListNodesByUID(ctx context.Context, uids ...id.Node) ([]*node.Resource, error)
	GetListNodesByName(ctx context.Context, names ...string) ([]*node.Resource, error)

	GetPolicyGroup(ctx context.Context, pgID id.PolicyGroup) (*policygroup.Resource, error)
	GetPolicyGroupByName(ctx context.Context, name string) (*policygroup.Resource, error)
	GetListPolicyGroupsByName(ctx context.Context, names ...string) ([]*policygroup.Resource, error)
	GetListPolicyGroupsByUID(ctx context.Context, gids ...id.PolicyGroup) ([]*policygroup.Resource, error)

	GetVolume(ctx context.Context, namespace id.Namespace, vid id.Volume) (*volume.Resource, error)
	GetVolumeByName(ctx context.Context, namespace id.Namespace, name string) (*volume.Resource, error)
	GetNamespaceVolumesByUID(ctx context.Context, namespaceID id.Namespace, volIDs ...id.Volume) ([]*volume.Resource, error)
	GetNamespaceVolumesByName(ctx context.Context, namespaceID id.Namespace, names ...string) ([]*volume.Resource, error)
	GetAllVolumes(ctx context.Context) ([]*volume.Resource, error)

	GetNamespace(ctx context.Context, namespaceID id.Namespace) (*namespace.Resource, error)
	GetNamespaceByName(ctx context.Context, name string) (*namespace.Resource, error)
	ListNamespaces(ctx context.Context) ([]*namespace.Resource, error)
	GetListNamespacesByName(ctx context.Context, names ...string) ([]*namespace.Resource, error)
	GetListNamespacesByUID(ctx context.Context, uids ...id.Namespace) ([]*namespace.Resource, error)
}

// Displayer defines the functionality required by the CLI application
// to display the results gathered by the "describe" verb commands.
type Displayer interface {
	DescribeCluster(ctx context.Context, w io.Writer, c *output.Cluster) error
	DescribeLicence(ctx context.Context, w io.Writer, l *output.Licence) error
	DescribeNamespace(ctx context.Context, w io.Writer, namespace *output.Namespace) error
	DescribeListNamespaces(ctx context.Context, w io.Writer, namespaces []*output.Namespace) error
	DescribeNode(ctx context.Context, w io.Writer, node *output.NodeDescription) error
	DescribeListNodes(ctx context.Context, w io.Writer, nodes []*output.NodeDescription) error
	DescribeVolume(ctx context.Context, w io.Writer, volume *output.Volume) error
	DescribeListVolumes(ctx context.Context, w io.Writer, volumes []*output.Volume) error
	DescribeUser(ctx context.Context, w io.Writer, user *output.User) error
	DescribeListUsers(ctx context.Context, w io.Writer, users []*output.User) error
	DescribePolicyGroup(ctx context.Context, w io.Writer, group *output.PolicyGroup) error
	DescribeListPolicyGroups(ctx context.Context, w io.Writer, groups []*output.PolicyGroup) error
}

// NewCommand configures the set of commands which are grouped by the "describe" verb.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	command := &cobra.Command{
		Use:   "describe",
		Short: "Fetch extended details for resources",
	}

	command.AddCommand(
		newNode(os.Stdout, client, config),
		newVolume(os.Stdout, client, config),
		newCluster(os.Stdout, client, config),
		newPolicyGroup(os.Stdout, client, config),
		newUser(os.Stdout, client, config),
		newNamespace(os.Stdout, client, config),
		newLicence(os.Stdout, client, config),
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
