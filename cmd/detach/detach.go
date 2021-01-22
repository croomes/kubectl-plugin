package detach

import (
	"context"
	"errors"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/cmd/argwrappers"
	"github.com/croomes/kubectl-plugin/cmd/clierr"
	"github.com/croomes/kubectl-plugin/cmd/flagutil"
	"github.com/croomes/kubectl-plugin/cmd/runwrappers"
	"github.com/croomes/kubectl-plugin/namespace"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/output/jsonformat"
	"github.com/croomes/kubectl-plugin/output/textformat"
	"github.com/croomes/kubectl-plugin/output/yamlformat"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/version"
	"github.com/croomes/kubectl-plugin/volume"
)

var (
	errArguments = errors.New("must specify exactly one volume to detach")
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

// Client describes the functionality required by the CLI application
// to reasonably implement the "detach" verb commands.
type Client interface {
	Authenticate(ctx context.Context, username, password string) (apiclient.AuthSession, error)

	GetNamespaceByName(ctx context.Context, name string) (*namespace.Resource, error)
	GetVolumeByName(ctx context.Context, namespaceID id.Namespace, name string) (*volume.Resource, error)
	DetachVolume(ctx context.Context, namespaceID id.Namespace, volumeID id.Volume, params *apiclient.DetachVolumeRequestParams) error
}

// Displayer describes the functionality required by the CLI application
// to display the resources produced by the "detach" verb commands.
type Displayer interface {
	DetachVolume(ctx context.Context, w io.Writer) error
	AsyncRequest(ctx context.Context, w io.Writer) error
}

type detachCommand struct {
	config  ConfigProvider
	client  Client
	display Displayer

	namespace string

	// useCAS determines whether the command makes the detach request
	// constrained by the provided casVersion.
	useCAS     func() bool
	casVersion string
	useAsync   bool

	writer io.Writer
}

func (c *detachCommand) runWithCtx(ctx context.Context, cmd *cobra.Command, args []string) error {
	useIDs, err := c.config.UseIDs()
	if err != nil {
		return err
	}

	namespaceID := id.Namespace(c.namespace)
	volumeID := id.Volume(args[0])

	// If not using IDs, get the appropriate IDs
	if !useIDs {
		ns, err := c.client.GetNamespaceByName(ctx, c.namespace)
		if err != nil {
			return err
		}
		namespaceID = ns.ID

		volName := args[0]
		vol, err := c.client.GetVolumeByName(ctx, namespaceID, volName)
		if err != nil {
			return err
		}
		volumeID = vol.ID
	}

	params := &apiclient.DetachVolumeRequestParams{}

	if c.useCAS() {
		params.CASVersion = version.FromString(c.casVersion)
	}

	// If asynchrony is specified then source the timeout and initialise the params.
	if c.useAsync {
		timeout, err := c.config.CommandTimeout()
		if err != nil {
			return err
		}

		params.AsyncMax = timeout
	}

	err = c.client.DetachVolume(
		ctx,
		namespaceID,
		volumeID,
		params,
	)
	if err != nil {
		return err
	}

	if c.useAsync {
		return c.display.AsyncRequest(ctx, c.writer)
	}

	return c.display.DetachVolume(ctx, c.writer)
}

// NewCommand configures the "detach" verb command.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	c := &detachCommand{
		config: config,
		client: client,
		writer: os.Stdout,
	}

	cobraCommand := &cobra.Command{
		Use:   "detach",
		Short: "Detach a volume from its current location",
		Example: `
$ storageos detach --namespace my-namespace-name my-volume
`,

		Args: argwrappers.WrapInvalidArgsError(func(cmd *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errArguments
			}
			return nil
		}),
		PreRunE: argwrappers.WrapInvalidArgsError(func(_ *cobra.Command, _ []string) error {
			c.display = SelectDisplayer(c.config)

			ns, err := c.config.Namespace()
			if err != nil {
				return err
			}

			if ns == "" {
				return clierr.ErrNoNamespaceSpecified
			}
			c.namespace = ns

			return nil
		}),

		RunE: func(cmd *cobra.Command, args []string) error {
			run := runwrappers.Chain(
				runwrappers.RunWithTimeout(c.config),
				runwrappers.EnsureNamespaceSetWhenUseIDs(c.config),
				runwrappers.AuthenticateClient(c.config, c.client),
			)(c.runWithCtx)

			return run(context.Background(), cmd, args)
		},

		SilenceUsage: true,
	}

	c.useCAS = flagutil.SupportCAS(cobraCommand.Flags(), &c.casVersion)
	flagutil.SupportAsync(cobraCommand.Flags(), &c.useAsync)

	return cobraCommand
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
		return jsonformat.NewDisplayer("")
	case output.YAML:
		return yamlformat.NewDisplayer("")
	case output.Text:
		fallthrough
	default:
		return textformat.NewDisplayer(textformat.NewTimeFormatter())
	}
}
