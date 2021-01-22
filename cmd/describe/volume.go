package describe

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/cmd/argwrappers"
	"github.com/croomes/kubectl-plugin/cmd/clierr"
	"github.com/croomes/kubectl-plugin/cmd/flagutil"
	"github.com/croomes/kubectl-plugin/cmd/runwrappers"
	"github.com/croomes/kubectl-plugin/namespace"
	"github.com/croomes/kubectl-plugin/node"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/selectors"
	"github.com/croomes/kubectl-plugin/volume"
)

type volumeCommand struct {
	config  ConfigProvider
	client  Client
	display Displayer

	namespace string
	selectors []string

	writer io.Writer
}

func (c *volumeCommand) runWithCtx(ctx context.Context, cmd *cobra.Command, args []string) error {
	useIDs, err := c.config.UseIDs()
	if err != nil {
		return err
	}

	set, err := selectors.NewSetFromStrings(c.selectors...)
	if err != nil {
		return err
	}

	var ns *namespace.Resource

	if useIDs {
		ns, err = c.client.GetNamespace(ctx, id.Namespace(c.namespace))
		if err != nil {
			return err
		}
	} else {
		ns, err = c.client.GetNamespaceByName(ctx, c.namespace)
		if err != nil {
			return err
		}
	}

	switch len(args) {
	case 1:
		var vol *volume.Resource
		var err error

		if useIDs {
			vol, err = c.client.GetVolume(ctx, ns.ID, id.Volume(args[0]))
		} else {
			vol, err = c.client.GetVolumeByName(ctx, ns.ID, args[0])
		}
		if err != nil {
			return err
		}

		nodes, err := c.getNodeMapping(ctx)
		if err != nil {
			return err
		}

		return c.display.DescribeVolume(ctx, c.writer, output.NewVolume(vol, ns, nodes))

	default:
		volumes, err := c.listVolumes(ctx, ns.ID, args)
		if err != nil {
			return err
		}

		volumes = set.FilterVolumes(volumes)

		nodes, err := c.getNodeMapping(ctx)
		if err != nil {
			return err
		}

		outputVols := make([]*output.Volume, 0, len(volumes))

		for _, vol := range volumes {
			outputVols = append(outputVols, output.NewVolume(vol, ns, nodes))
		}

		return c.display.DescribeListVolumes(ctx, c.writer, outputVols)
	}
}

// getNodeMapping fetches the list of nodes from the API and builds a map from
// their ID to the full resource.
func (c *volumeCommand) getNodeMapping(ctx context.Context) (map[id.Node]*node.Resource, error) {
	nodeList, err := c.client.ListNodes(ctx)
	if err != nil {
		return nil, err
	}

	nodes := map[id.Node]*node.Resource{}
	for _, n := range nodeList {
		nodes[n.ID] = n
	}

	return nodes, nil
}

// listVolumes requests a list of volume resources using the configured API
// client, filtering using vols (if provided) as c's configuration dictates.
func (c *volumeCommand) listVolumes(ctx context.Context, ns id.Namespace, vols []string) ([]*volume.Resource, error) {
	useIDs, err := c.config.UseIDs()
	if err != nil {
		return nil, err
	}

	if !useIDs {
		return c.client.GetNamespaceVolumesByName(ctx, ns, vols...)
	}

	volIDs := []id.Volume{}
	for _, uid := range vols {
		volIDs = append(volIDs, id.Volume(uid))
	}

	return c.client.GetNamespaceVolumesByUID(ctx, ns, volIDs...)
}

func newVolume(w io.Writer, client Client, config ConfigProvider) *cobra.Command {
	c := &volumeCommand{
		config: config,
		client: client,
		writer: w,
	}

	cobraCommand := &cobra.Command{
		Aliases: []string{"volumes"},
		Use:     "volume [volume names...]",
		Short:   "Show detailed information for volumes",
		Example: `
$ storageos describe volumes

$ storageos describe volume --namespace my-namespace-name my-volume-name
`,

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
				runwrappers.EnsureTargetOrSelectors(&c.selectors),
				runwrappers.AuthenticateClient(c.config, c.client),
			)(c.runWithCtx)
			return run(context.Background(), cmd, args)
		},

		// If a legitimate error occurs as part of the VERB volume command
		// we don't need to barf the usage template.
		SilenceUsage: true,
	}

	flagutil.SupportSelectors(cobraCommand.Flags(), &c.selectors)

	return cobraCommand
}
