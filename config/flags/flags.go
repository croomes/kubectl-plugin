// Package flags exports an implementation of a configuration settings provider
// which operates using a set of flags.
package flags

import (
	"fmt"
	"time"

	"github.com/croomes/kubectl-plugin/config"
	"github.com/croomes/kubectl-plugin/output"
)

const (
	// AuthCacheDisabledFlag keys the long flag which determines whether the
	// CLI's auth cache must be disabled, if set.
	AuthCacheDisabledFlag = "no-auth-cache"
	// APIEndpointsFlag keys the long flag from which the list of API host
	// endpoints are sourced, if set.
	APIEndpointsFlag = "endpoints"
	// CacheDirFlag keys the long flag from which the directory to cache data
	// that can be re-used by later commands is stored in, if set.
	CacheDirFlag = "cache-dir"
	// CommandTimeoutFlag keys the long flag from which the timeout for API
	// operations is sourced, if set.
	CommandTimeoutFlag = "timeout"
	// UsernameFlag keys the long flag from which the username part of the
	// credentials used for authentication is sourced, if set.
	UsernameFlag = "username"
	// PasswordFlag keys the long flag from which the password part of the
	// credentials used for authentication is sourced, if set.
	PasswordFlag = "password"
	// UseIDsFlag keys the long flag from which the setting that decides
	// whether existing API resources are specified by unique identifier instead
	// of name is sourced.
	UseIDsFlag = "use-ids"
	// NamespaceFlag keys the long flag from which the namespace name or ID to
	// operate within is sourced for commands that required it.
	NamespaceFlag = "namespace"
	// ShortNamespaceFlag keys the short flag from which the namespace name or ID
	// to operate within is sourced for commands that required it.
	ShortNamespaceFlag = "n"
	// OutputFormatFlag keys the long flag from which the output format is
	// sourced for commands that requires it
	OutputFormatFlag = "output"
	// ShortOutputFormatFlag keys the short flag from which the output format is
	// sourced for commands that requires it
	ShortOutputFormatFlag = "o"
	// ConfigFileFlag keys the long flag from which the config file path is
	// sourced for commands that requires it
	ConfigFileFlag = "config"
	// ShortConfigFileFlag keys the short flag from which the config file path
	// is sourced for commands that requires it
	ShortConfigFileFlag = "c"
)

// FlagSet describes a set of typed flag set accessors/setters required by the
// Provider.
type FlagSet interface {
	Changed(name string) bool

	Bool(name string, value bool, usage string) *bool
	Duration(name string, value time.Duration, usage string) *time.Duration
	String(name string, value string, usage string) *string
	StringP(name string, shorthand string, value string, usage string) *string
	StringArray(name string, value []string, usage string) *[]string

	GetBool(name string) (bool, error)
	GetDuration(name string) (time.Duration, error)
	GetString(name string) (string, error)
	GetStringArray(name string) ([]string, error)
}

// Provider exports functionality to retrieve global configuration values from
// the global flag set if available. When a configuration value is not
// available from the flag set, the configured fallback is used.
type Provider struct {
	set      FlagSet
	fallback config.Provider
}

// AuthCacheDisabled sources the configuration setting which determines whether
// the CLI must disable use of the auth cache from flag's FlagSet. If the value
// stored has not changed then flag's fallback is used.
func (flag *Provider) AuthCacheDisabled() (bool, error) {
	disabled, err := flag.set.GetBool(AuthCacheDisabledFlag)
	if err != nil {
		return false, err
	}

	if !flag.set.Changed(AuthCacheDisabledFlag) {
		return flag.fallback.AuthCacheDisabled()
	}

	return disabled, nil
}

// APIEndpoints sources the list of comma-separated target API endpoints from
// flag's FlagSet. If the value stored has not changed then flag's fallback
// is used.
func (flag *Provider) APIEndpoints() ([]string, error) {
	hosts, err := flag.set.GetStringArray(APIEndpointsFlag)
	if err != nil {
		return nil, err
	}

	if len(hosts) == 0 || !flag.set.Changed(APIEndpointsFlag) {
		return flag.fallback.APIEndpoints()
	}

	return hosts, nil
}

// CacheDir sources the directory to use as the StorageOS CLI cache from flag's
// FlagSet. If the value stored has not changed then flag's fallback is used.
func (flag *Provider) CacheDir() (string, error) {
	cacheDir, err := flag.set.GetString(CacheDirFlag)
	if err != nil {
		return "", err
	}

	if cacheDir == "" || !flag.set.Changed(CacheDirFlag) {
		return flag.fallback.CacheDir()
	}

	return cacheDir, nil
}

// CommandTimeout sources the command timeout duration from flag's FlagSet.
// If the value stored has not changed then flag's fallback is used.
func (flag *Provider) CommandTimeout() (time.Duration, error) {
	timeout, err := flag.set.GetDuration(CommandTimeoutFlag)
	if err != nil {
		return 0, err
	}

	if timeout == 0 || !flag.set.Changed(CommandTimeoutFlag) {
		return flag.fallback.CommandTimeout()
	}

	return timeout, nil
}

// Username sources the StorageOS account username to authenticate with from
// flag's FlagSet. If the value stored has not changed then flag's fallback
// is used.
func (flag *Provider) Username() (string, error) {
	username, err := flag.set.GetString(UsernameFlag)
	if err != nil {
		return "", err
	}

	if username == "" || !flag.set.Changed(UsernameFlag) {
		return flag.fallback.Username()
	}

	return username, nil
}

// Password sources the StorageOS account password to authenticate with from
// flag's FlagSet. If the value stored has not changed then flag's fallback
// is used.
func (flag *Provider) Password() (string, error) {
	password, err := flag.set.GetString(PasswordFlag)
	if err != nil {
		return "", err
	}

	if password == "" || !flag.set.Changed(PasswordFlag) {
		return flag.fallback.Password()
	}

	return password, nil
}

// UseIDs sources the configuration setting to specify existing API resources
// by their unique identifier instead of name from flag's FlagSet. If the value
// stored has not changed then flag's fallback is used.
func (flag *Provider) UseIDs() (bool, error) {
	useIDs, err := flag.set.GetBool(UseIDsFlag)
	if err != nil {
		return false, err
	}

	if !flag.set.Changed(UseIDsFlag) {
		return flag.fallback.UseIDs()
	}

	return useIDs, nil
}

// Namespace sources the StorageOS namespace to operate within from flag's
// FlagSet, for operations that require a namespace. If the value stored has
// not changed then flag's fallback is used.
func (flag *Provider) Namespace() (string, error) {
	namespace, err := flag.set.GetString(NamespaceFlag)
	if err != nil {
		return "", err
	}

	if !flag.set.Changed(NamespaceFlag) {
		return flag.fallback.Namespace()
	}

	return namespace, nil
}

// OutputFormat sources the output format from flag's FlagSet. If the value
// stored has not changed, then flag's fallback is used.
func (flag *Provider) OutputFormat() (output.Format, error) {
	out, err := flag.set.GetString(OutputFormatFlag)
	if err != nil {
		return 0, err
	}

	if !flag.set.Changed(OutputFormatFlag) {
		return flag.fallback.OutputFormat()
	}

	outputType, err := output.FormatFromString(out)
	if err != nil {
		return output.Unknown, err
	}

	return outputType, nil
}

// ConfigFilePath sources the config file path from flag's FlagSet. If the value
// stored has not changed,, then flag's fallback is used.
func (flag *Provider) ConfigFilePath() (string, error) {
	path, err := flag.set.GetString(ConfigFileFlag)
	if err != nil {
		return "", err
	}

	if !flag.set.Changed(ConfigFileFlag) {
		return flag.fallback.ConfigFilePath()
	}

	return path, nil
}

// NewProvider initialises a new flag based configuration provider backed by
// flagset, falling back on the provided fallback if the value can
// not be sourced from flagset.
func NewProvider(flagset FlagSet, fallback config.Provider) *Provider {

	// Set up the flags for the config provider
	flagset.Bool(
		AuthCacheDisabledFlag,
		config.DefaultAuthCacheDisabled,
		"disable the CLI's caching of authentication sessions",
	)
	flagset.StringArray(
		APIEndpointsFlag,
		[]string{config.DefaultAPIEndpoint},
		"set the list of endpoints which are used when connecting to the StorageOS API",
	)
	flagset.String(
		CacheDirFlag,
		config.GetDefaultCacheDir(),
		"set the directory used by the StorageOS CLI to cache data that can be used for future commands",
	)
	flagset.Duration(
		CommandTimeoutFlag,
		config.DefaultCommandTimeout,
		"set the timeout duration to use for execution of the command",
	)
	flagset.String(
		UsernameFlag,
		config.DefaultUsername,
		"set the StorageOS account username to authenticate as",
	)
	flagset.String(
		PasswordFlag,
		config.DefaultPassword,
		"set the StorageOS account password to authenticate with",
	)
	flagset.Bool(
		UseIDsFlag,
		config.DefaultUseIDs,
		"specify existing StorageOS resources by their unique identifiers instead of by their names",
	)
	flagset.StringP(
		NamespaceFlag,
		ShortNamespaceFlag,
		config.DefaultNamespaceName,
		"specifies the namespace to operate within for commands that require one",
	)
	flagset.StringP(
		OutputFormatFlag,
		ShortOutputFormatFlag,
		config.DefaultOutput.String(),
		fmt.Sprintf("specifies the output format (one of %v)", output.ValidFormats),
	)

	flagset.StringP(
		ConfigFileFlag,
		ShortConfigFileFlag,
		config.GetDefaultConfigFile(),
		"specifies the config file path",
	)

	return &Provider{
		set:      flagset,
		fallback: fallback,
	}
}
