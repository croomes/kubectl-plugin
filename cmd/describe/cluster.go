package describe

import (
	"context"
	"io"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/cmd/argwrappers"
	"github.com/croomes/kubectl-plugin/cmd/runwrappers"
	"github.com/croomes/kubectl-plugin/output"
)

type clusterCommand struct {
	config  ConfigProvider
	client  Client
	display Displayer

	writer io.Writer
}

func (c *clusterCommand) runWithCtx(ctx context.Context, cmd *cobra.Command, args []string) error {
	cluster, err := c.client.GetCluster(ctx)
	if err != nil {
		return err
	}

	nodes, err := c.client.ListNodes(ctx)
	if err != nil {
		return err
	}

	return c.display.DescribeCluster(ctx, c.writer, output.NewCluster(cluster, nodes))
}

func newCluster(w io.Writer, client Client, config ConfigProvider) *cobra.Command {
	c := &clusterCommand{
		config: config,
		client: client,
		writer: w,
	}

	cobraCommand := &cobra.Command{
		Aliases: []string{"nodes"},
		Use:     "cluster",
		Short:   "Retrieve detailed information for the current cluster",
		Example: `
$ storageos describe cluster
`,

		PreRunE: argwrappers.WrapInvalidArgsError(func(_ *cobra.Command, args []string) error {
			c.display = SelectDisplayer(c.config)
			return nil
		}),

		RunE: func(cmd *cobra.Command, args []string) error {
			run := runwrappers.Chain(
				runwrappers.RunWithTimeout(c.config),
				runwrappers.AuthenticateClient(c.config, c.client),
			)(c.runWithCtx)
			return run(context.Background(), cmd, args)
		},

		SilenceUsage: true,
	}

	return cobraCommand
}
