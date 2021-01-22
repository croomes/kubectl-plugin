package delete

import (
	"context"
	"errors"
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

type volumeCommand struct {
	config  ConfigProvider
	client  Client
	display Displayer

	namespace string

	// useCAS determines whether the command makes the delete request
	// constrained by the provided casVersion.
	useCAS     func() bool
	casVersion string

	useAsync         bool
	useOfflineDelete bool

	writer io.Writer
}

func (c *volumeCommand) runWithCtx(ctx context.Context, cmd *cobra.Command, args []string) error {
	useIDs, err := c.config.UseIDs()
	if err != nil {
		return err
	}

	params := &apiclient.DeleteVolumeRequestParams{
		OfflineDelete: c.useOfflineDelete,
	}

	if c.useCAS() {
		params.CASVersion = version.FromString(c.casVersion)
	}

	// If asynchrony is specified then source the timeout and set the
	// async timeout from it.
	if c.useAsync {
		timeout, err := c.config.CommandTimeout()
		if err != nil {
			return err
		}

		params.AsyncMax = timeout
	}

	namespaceID := id.Namespace(c.namespace)
	volumeID := id.Volume(args[0])

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

	err = c.client.DeleteVolume(
		ctx,
		namespaceID,
		volumeID,
		params,
	)
	if err != nil {
		return err
	}

	// Display the "request submitted" message if it was async, instead of
	// the deletion confirmation below.
	if c.useAsync {
		return c.display.AsyncRequest(ctx, c.writer)
	}

	volumeDisplay := output.NewVolumeDeletion(volumeID, namespaceID)

	return c.display.DeleteVolume(ctx, c.writer, volumeDisplay)
}

func newVolume(w io.Writer, client Client, config ConfigProvider) *cobra.Command {
	c := &volumeCommand{
		config: config,
		client: client,
		writer: w,
	}

	cobraCommand := &cobra.Command{
		Use:   "volume [volume name]",
		Short: "Delete a volume. By default the target volume must be online. If the volume is offline then the request must specify that an offline delete is desired.",
		Example: `
$ storageos delete volume my-test-volume my-unneeded-volume

$ storageos delete volume --namespace my-namespace my-old-volume
`,

		Args: argwrappers.WrapInvalidArgsError(func(_ *cobra.Command, args []string) error {
			if len(args) != 1 {
				return errors.New("must specify exactly one volume for deletion")
			}
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
	cobraCommand.Flags().BoolVar(&c.useOfflineDelete, "offline-delete", false, "request deletion of an offline volume. Volume data is not removed until the node reboots")

	return cobraCommand
}
