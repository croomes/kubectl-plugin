package update

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/cmd/argwrappers"
	"github.com/croomes/kubectl-plugin/cmd/clierr"
	"github.com/croomes/kubectl-plugin/cmd/flagutil"
	"github.com/croomes/kubectl-plugin/cmd/runwrappers"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/version"
)

type volumeDescriptionCommand struct {
	config  ConfigProvider
	client  Client
	display Displayer

	namespace   string
	volumeID    string
	description string

	// useCAS determines whether the command makes the set replicas request
	// constrained by the provided casVersion.
	useCAS     func() bool
	casVersion string
	useAsync   bool

	writer io.Writer
}

func (c *volumeDescriptionCommand) runWithCtx(ctx context.Context, cmd *cobra.Command, args []string) error {
	useIDs, err := c.config.UseIDs()
	if err != nil {
		return err
	}

	var nsID id.Namespace

	if useIDs {
		nsID = id.Namespace(c.namespace)
	} else {
		ns, err := c.client.GetNamespaceByName(ctx, c.namespace)
		if err != nil {
			return err
		}
		nsID = ns.ID
	}

	var volID id.Volume

	if useIDs {
		volID = id.Volume(c.volumeID)
	} else {
		vol, err := c.client.GetVolumeByName(ctx, nsID, c.volumeID)
		if err != nil {
			return err
		}
		volID = vol.ID
	}

	params := &apiclient.UpdateVolumeRequestParams{}

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

	updatedVol, err := c.client.UpdateVolumeDescription(ctx, nsID, volID, c.description, params)
	if err != nil {
		return err
	}

	if c.useAsync {
		return c.display.AsyncRequest(ctx, c.writer)
	}

	return c.display.UpdateVolumeDescription(ctx, c.writer, output.NewVolumeUpdate(updatedVol))
}

func newVolumeDescription(w io.Writer, client Client, config ConfigProvider) *cobra.Command {
	c := &volumeDescriptionCommand{
		config: config,
		client: client,
		writer: w,
	}

	cobraCommand := &cobra.Command{
		Use:   "description [volume name] [description]",
		Short: "Updates a volume's description",
		Example: `
$ storageos update volume description my-volume-name "Houston we have liftoff 🚀" --namespace my-namespace-name
`,

		Args: argwrappers.WrapInvalidArgsError(func(cmd *cobra.Command, args []string) error {

			if len(args) != 2 {
				return clierr.NewErrInvalidArgNum(args, 2, "storageos update volume description [volume] [description]")
			}

			c.volumeID = args[0]
			c.description = args[1]

			return nil
		}),

		PreRunE: argwrappers.WrapInvalidArgsError(func(_ *cobra.Command, args []string) error {
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
