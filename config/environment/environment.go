// Package environment exports an implementation of a configuration settings
// provider which operates using the operating systems environment.
package environment

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/croomes/kubectl-plugin/config"
	"github.com/croomes/kubectl-plugin/output"
)

const (
	// AuthCacheDisabledVar keys the environment variable from which we source
	// the configuration setting which decides whether the auth cache is disabled.
	AuthCacheDisabledVar = "STORAGEOS_NO_AUTH_CACHE"
	// APIEndpointsVar keys the environment variable from which we source the
	// API host endpoints.
	APIEndpointsVar = "STORAGEOS_ENDPOINTS"
	// CacheDirVar keys the environment variable from which we source the
	// directory to use as a data cache.
	CacheDirVar = "STORAGEOS_CACHE_DIR"
	// CommandTimeoutVar keys the environment variable from which we source the
	// timeout for API operations.
	CommandTimeoutVar = "STORAGEOS_API_TIMEOUT"
	// UsernameVar keys the environment variable from which we source the
	// username of the StorageOS account to authenticate with.
	UsernameVar = "STORAGEOS_USERNAME"
	// PasswordVar keys the environment variable from which we source the
	// password of the StorageOS account to authenticate with.
	PasswordVar = "STORAGEOS_PASSWORD" // #nosec G101
	// PasswordCommandVar keys the environment variable from which we optionally
	// source the password of the StorageOS account to authenticate with through
	// command execution.
	PasswordCommandVar = "STORAGEOS_PASSWORD_COMMAND" // #nosec G101
	// UseIDsVar keys the environment variable from which we source the setting
	// which determines whether existing StorageOS API resources are specified
	// by their unique identifiers instead of names.
	UseIDsVar = "STORAGEOS_USE_IDS"
	// NamespaceVar keys the environment variable from which we source the
	// namespace name or unique identifier to operate within for commands that
	// require it.
	NamespaceVar = "STORAGEOS_NAMESPACE"
	// OutputFormatVar keys the environment variable from which we source the output
	// format to use when we print out the results.
	OutputFormatVar = "STORAGEOS_OUTPUT_FORMAT"
	// ConfigFilePathVar keys the environment variable from which we source the
	// config file path to use when we load configs from file.
	ConfigFilePathVar = "STORAGEOS_CONFIG"
)

// EnvConfigHelp holds the list of environment variable used to source
// configuration settings, along with a user facing help description.
var EnvConfigHelp = []struct {
	Name string
	Help string
}{
	{
		Name: AuthCacheDisabledVar,
		Help: "Disables the caching of authenticated sessions by the CLI",
	},
	{
		// TODO(CP-3924): Update this for multiple endpoints implementation
		Name: APIEndpointsVar,
		Help: "Sets the default StorageOS API endpoint for the CLI to connect to",
	},
	{
		Name: CacheDirVar,
		Help: "Sets the default directory for the CLI to cache re-usable data to",
	},
	{
		Name: CommandTimeoutVar,
		Help: "Specifies the default duration which the CLI will give a command to complete before aborting with a timeout",
	},
	{
		Name: UsernameVar,
		Help: "Sets the default username provided by the CLI for authentication",
	},
	{
		Name: PasswordVar,
		Help: "Sets the default password provided by the CLI for authentication",
	},
	{
		Name: PasswordCommandVar,
		Help: "If set the default password provided by the CLI for authentication is sourced from the output produced by executing the command",
	},
	{
		Name: UseIDsVar,
		Help: "When set to true, the CLI will use provided values as IDs instead of names for existing resources",
	},
	{
		Name: NamespaceVar,
		Help: "Specifies the default namespace for the CLI to operate in",
	},
	{
		Name: OutputFormatVar,
		Help: "Specifies the default format used by the CLI for output",
	},
	{
		Name: ConfigFilePathVar,
		Help: "Specifies the default path used by the CLI for the config file",
	},
}

// Provider exports functionality to retrieve global configuration values from
// environment variables if available. When a configuration value is not
// available from the environment, the configured fallback is used.
type Provider struct {
	fallback config.Provider
}

// AuthCacheDisabled sources the configuration setting which determines if the
// auth cache must be disabled from the environment if set. If not set then
// env's fallback is used.
func (env *Provider) AuthCacheDisabled() (bool, error) {
	disabledString := os.Getenv(AuthCacheDisabledVar)
	if disabledString == "" {
		return env.fallback.AuthCacheDisabled()
	}

	return strconv.ParseBool(disabledString)
}

// APIEndpoints sources the list of comma-separated target API endpoints from
// the environment if set. If not set in the environment then env's fallback
// is used.
func (env *Provider) APIEndpoints() ([]string, error) {
	hostString := os.Getenv(APIEndpointsVar)
	if hostString == "" {
		return env.fallback.APIEndpoints()
	}
	endpoints := strings.Split(hostString, ",")

	return endpoints, nil
}

// CacheDir sources the path to the directory for the CLI to use when caching
// data from the environment if set. If not set in the environment then env's fallback is used.
func (env *Provider) CacheDir() (string, error) {
	cacheString := os.Getenv(CacheDirVar)
	if cacheString == "" {
		return env.fallback.CacheDir()
	}

	return cacheString, nil
}

// CommandTimeout sources the command timeout duration from the environment
// if set. If not set in the environment then env's fallback is used.
func (env *Provider) CommandTimeout() (time.Duration, error) {
	timeoutString := os.Getenv(CommandTimeoutVar)
	if timeoutString == "" {
		return env.fallback.CommandTimeout()
	}

	return time.ParseDuration(timeoutString)
}

// Username sources the StorageOS account username to authenticate with from
// the environment if set. If not set in the environment then env's fallback
// is used.
func (env *Provider) Username() (string, error) {
	username := os.Getenv(UsernameVar)
	if username == "" {
		return env.fallback.Username()
	}

	return username, nil
}

// Password sources the StorageOS account password to authenticate with from
// the environment if set. If not set in the environment then env's fallback
// is used.
func (env *Provider) Password() (string, error) {
	passwordCommand := os.Getenv(PasswordCommandVar)
	if passwordCommand != "" {
		cmd := exec.Command(passwordCommand)
		b := &bytes.Buffer{}
		cmd.Stdout = b
		err := cmd.Run()
		if err != nil {
			switch v := err.(type) {
			case *exec.ExitError:
				return "", fmt.Errorf("password command exited with error code %d", v.ExitCode())
			default:
				return "", err
			}
		}
		return strings.TrimSpace(b.String()), nil
	}

	password := os.Getenv(PasswordVar)
	if password == "" {
		return env.fallback.Password()
	}

	return password, nil
}

// UseIDs sources the configuration setting to specify existing API resources
// by their unique identifier instead of name from the environment if set.
// If not set in the environment then env's fallback is used.
func (env *Provider) UseIDs() (bool, error) {
	useIDs := os.Getenv(UseIDsVar)
	if useIDs == "" {
		return env.fallback.UseIDs()
	}

	return strconv.ParseBool(useIDs)
}

// Namespace sources the StorageOS namespace to operate within from the
// environment if set. The value used must match up with the configuration
// setting for using IDs.
//
// If not set set in the environment then env's fallback is used.
func (env *Provider) Namespace() (string, error) {
	namespace := os.Getenv(NamespaceVar)
	if namespace == "" {
		return env.fallback.Namespace()
	}

	return namespace, nil
}

// OutputFormat returns the output format type taken from the environment, if set.
// If not set, the env's fallback is used.
func (env *Provider) OutputFormat() (output.Format, error) {
	out := os.Getenv(OutputFormatVar)
	if out == "" {
		return env.fallback.OutputFormat()
	}

	outputType, err := output.FormatFromString(out)
	if err != nil {
		return output.Unknown, err
	}

	return outputType, nil
}

// ConfigFilePath sources the config file path taken from the environment, if set.
// If not set, the env's fallback is used.
func (env *Provider) ConfigFilePath() (string, error) {
	path := os.Getenv(ConfigFilePathVar)
	if path == "" {
		return env.fallback.ConfigFilePath()
	}

	return path, nil
}

// NewProvider returns a configuration provider which sources
// its configuration setting values from the OS environment if
// available.
func NewProvider(fallback config.Provider) *Provider {
	return &Provider{
		fallback: fallback,
	}
}
