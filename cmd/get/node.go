package get

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/cmd/flagutil"
	"github.com/croomes/kubectl-plugin/cmd/runwrappers"
	"github.com/croomes/kubectl-plugin/node"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/pkg/id"
	"github.com/croomes/kubectl-plugin/pkg/selectors"
)

type nodeCommand struct {
	config  ConfigProvider
	client  Client
	display Displayer

	selectors []string

	writer io.Writer
}

func (c *nodeCommand) runWithCtx(ctx context.Context, cmd *cobra.Command, args []string) error {
	switch len(args) {
	case 1:
		n, err := c.getNode(ctx, args[0])
		if err != nil {
			return err
		}

		return c.display.GetNode(ctx, c.writer, output.NewNode(n))
	default:
		set, err := selectors.NewSetFromStrings(c.selectors...)
		if err != nil {
			return err
		}

		nodes, err := c.listNodes(ctx, args)
		if err != nil {
			return err
		}

		nodes = set.FilterNodes(nodes)

		return c.display.GetListNodes(ctx, c.writer, output.NewNodes(nodes))
	}
}

// getNode retrieves a single node resource using the API client, determining
// whether to retrieve the node by name or ID based on the current command configuration.
func (c *nodeCommand) getNode(ctx context.Context, ref string) (*node.Resource, error) {
	useIDs, err := c.config.UseIDs()
	if err != nil {
		return nil, err
	}

	if !useIDs {
		return c.client.GetNodeByName(ctx, ref)
	}

	uid := id.Node(ref)
	return c.client.GetNode(ctx, uid)
}

// listNodes retrieves a list of node resources using the API client, determining
// whether to retrieve nodes by names by name or ID based on the current
// command configuration.
func (c *nodeCommand) listNodes(ctx context.Context, refs []string) ([]*node.Resource, error) {
	useIDs, err := c.config.UseIDs()
	if err != nil {
		return nil, err
	}

	if !useIDs {
		return c.client.GetListNodesByName(ctx, refs...)
	}

	uids := make([]id.Node, len(refs))
	for i, a := range refs {
		uids[i] = id.Node(a)
	}

	return c.client.GetListNodesByUID(ctx, uids...)
}

func newNode(w io.Writer, client Client, config ConfigProvider) *cobra.Command {
	c := &nodeCommand{
		config: config,
		client: client,
		writer: w,
	}
	cobraCommand := &cobra.Command{
		Aliases: []string{"nodes"},
		Use:     "node [node names...]",
		Short:   "Retrieve basic details of nodes in the cluster",
		Example: `
$ storageos get node my-node-name
`,
		PreRun: func(_ *cobra.Command, _ []string) {
			c.display = SelectDisplayer(c.config)
		},

		RunE: func(cmd *cobra.Command, args []string) error {
			run := runwrappers.Chain(
				runwrappers.RunWithTimeout(c.config),
				runwrappers.EnsureTargetOrSelectors(&c.selectors),
				runwrappers.AuthenticateClient(c.config, c.client),
			)(c.runWithCtx)
			return run(context.Background(), cmd, args)
		},

		// If a legitimate error occurs as part of the VERB node command
		// we don't need to barf the usage template.
		SilenceUsage: true,
	}

	flagutil.SupportSelectors(cobraCommand.Flags(), &c.selectors)

	return cobraCommand
}
