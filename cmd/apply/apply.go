package apply

import (
	"context"
	"io"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/licence"
	"github.com/croomes/kubectl-plugin/output"
	"github.com/croomes/kubectl-plugin/output/jsonformat"
	"github.com/croomes/kubectl-plugin/output/textformat"
	"github.com/croomes/kubectl-plugin/output/yamlformat"
)

// ConfigProvider specifies the configuration
type ConfigProvider interface {
	Username() (string, error)
	Password() (string, error)

	CommandTimeout() (time.Duration, error)
	OutputFormat() (output.Format, error)
}

// Client defines the functionality required by the CLI application to
// reasonably implement the "apply" verb commands.
type Client interface {
	Authenticate(ctx context.Context, username, password string) (apiclient.AuthSession, error)

	UpdateLicence(ctx context.Context, licenceKey []byte, params *apiclient.UpdateLicenceRequestParams) (*licence.Resource, error)
}

// Displayer defines the functionality required by the CLI application to
// display the results returned by "apply" verb operations.
type Displayer interface {
	UpdateLicence(ctx context.Context, w io.Writer, licence *output.Licence) error
}

// NewCommand configures the set of commands which are grouped by the "apply" verb.
func NewCommand(client Client, config ConfigProvider) *cobra.Command {
	command := &cobra.Command{
		Use:   "apply",
		Short: "Make changes to existing resources",
	}

	command.AddCommand(
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
