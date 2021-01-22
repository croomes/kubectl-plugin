package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/blang/semver"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/apiclient/cache"
	"github.com/croomes/kubectl-plugin/apiclient/openapi"
	"github.com/croomes/kubectl-plugin/cmd/apply"
	"github.com/croomes/kubectl-plugin/cmd/attach"
	"github.com/croomes/kubectl-plugin/cmd/create"
	"github.com/croomes/kubectl-plugin/cmd/delete"
	"github.com/croomes/kubectl-plugin/cmd/describe"
	"github.com/croomes/kubectl-plugin/cmd/detach"
	"github.com/croomes/kubectl-plugin/cmd/get"
	"github.com/croomes/kubectl-plugin/cmd/nfs"
	"github.com/croomes/kubectl-plugin/cmd/update"
	"github.com/croomes/kubectl-plugin/config"
)

// UserAgentPrefix is used by the CLI application to identify itself to
// StorageOS.
var UserAgentPrefix string = "storageos-cli"

// InitCommand configures the CLI application's commands from the root down, using
// client as the method of communicating with the StorageOS API.
//
// The returned Command is configured with a flag set containing global configuration settings.
//
// Downstream errors are suppressed, so the caller is responsible for displaying messages.
func InitCommand(client *apiclient.Client, config config.Provider, globalFlags *pflag.FlagSet, version semver.Version) *cobra.Command {
	app := &cobra.Command{
		Use: "storageos <command>",
		Short: `Storage for Cloud Native Applications.

By using this product, you are agreeing to the terms of the the StorageOS Ltd. End
User Subscription Agreement (EUSA) found at: https://storageos.com/legal/#eusa

To be notified about stable releases and latest features, sign up at https://my.storageos.com.
`,

		PersistentPreRunE: func(_ *cobra.Command, _ []string) error {
			userAgent := strings.Join([]string{UserAgentPrefix, version.String()}, "/")

			disabledAuthCache, err := config.AuthCacheDisabled()
			if err != nil {
				return err
			}

			transport, err := func() (apiclient.Transport, error) {
				transport, err := openapi.NewOpenAPI(config, userAgent)
				if err != nil {
					return nil, err
				}

				if disabledAuthCache {
					return transport, nil
				}

				// Ensure that the cache dir exists
				cacheDir, err := config.CacheDir()
				if err != nil {
					return transport, nil
				}

				err = os.Mkdir(cacheDir, 0700)
				switch {
				case err == nil, os.IsExist(err):
					// This is ok
				default:
					return nil, err
				}

				// Only wrap with caching if desired
				return apiclient.NewAuthCachedTransport(
					transport,
					cache.NewSessionCache(config, time.Now, 5*time.Second),
				), nil
			}()
			if err != nil {
				return err
			}

			// Wrap the transport implementation in a re-authenticate layer.
			return client.ConfigureTransport(
				apiclient.NewTransportWithReauth(transport, config),
			)
		},

		SilenceErrors: true,
	}

	// Register the generic CLI commands that don't do any API interaction.
	versionCommand := &cobra.Command{
		Use:   "version",
		Short: "View version information for the StorageOS CLI",
		Run: func(_ *cobra.Command, _ []string) {
			fmt.Printf("StorageOS CLI version: %v\n", version.String())
		},
	}

	app.AddCommand(
		apply.NewCommand(client, config),
		update.NewCommand(client, config),
		create.NewCommand(client, config),
		get.NewCommand(client, config),
		describe.NewCommand(client, config),
		delete.NewCommand(client, config),
		attach.NewCommand(client, config),
		detach.NewCommand(client, config),
		nfs.NewCommand(client, config),
		versionCommand,
	)

	// Cobra subcommands which are not runnable and do not themselves have
	// subcommands are added as additional help topics.
	app.AddCommand(
		newConfigFileHelpTopic(config),
		newEnvConfigHelpTopic(),
		newExitCodeHelpTopic(),
	)

	app.PersistentFlags().AddFlagSet(globalFlags)

	return app
}
