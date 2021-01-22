// Package runwrappers contains wrapper functions which implement shared
// functionality for command run functions.
package runwrappers

import (
	"context"
	"errors"

	"github.com/spf13/cobra"

	"github.com/croomes/kubectl-plugin/apiclient"
	"github.com/croomes/kubectl-plugin/config"
	"github.com/croomes/kubectl-plugin/pkg/cmdcontext"
)

var (
	// ErrTargetOrSelector is an error that indicates that a label selector
	// cannot be used when specifying resource names or unique identifiers.
	ErrTargetOrSelector = errors.New("a target name or unique identifier cannot be used with a label selector")
	// ErrMustSpecifyNamespaceID is an error indicating that the user has not
	// specified a namespace ID, which is required when using IDs.
	ErrMustSpecifyNamespaceID = errors.New("namespace ID must be specified when using resource IDs")
)

// CredentialsProvider is a type that source a set of authentication
// credentials to provide to the StorageOS API.
type CredentialsProvider interface {
	Username() (string, error)
	Password() (string, error)
}

// AuthAPIClient defines an API client that can be authenticated.
type AuthAPIClient interface {
	Authenticate(ctx context.Context, username, password string) (apiclient.AuthSession, error)
}

// NamespacedCommandConfigProvider abstracts a type which provides
// configuration settings for commands that are namespaced and can use IDs.
type NamespacedCommandConfigProvider interface {
	Namespace() (string, error)
	UseIDs() (bool, error)
}

// RunEWithContext is a function that extends a cobra.RunE function with a
// context parameter.
type RunEWithContext func(ctx context.Context, cmd *cobra.Command, args []string) error

// WrapRunEWithContext wraps next within another RunEWithContext.
type WrapRunEWithContext func(next RunEWithContext) RunEWithContext

// Chain chains wrappers from first to last, returning a wrap function that
// can be used to wrap an inner RunEWithContext.
func Chain(wrappers ...WrapRunEWithContext) WrapRunEWithContext {
	return func(next RunEWithContext) RunEWithContext {
		return func(ctx context.Context, cmd *cobra.Command, args []string) error {
			wrapped := next
			for i := len(wrappers) - 1; i >= 0; i-- {
				wrapped = wrappers[i](wrapped)
			}

			return wrapped(ctx, cmd, args)
		}
	}
}

// RunWithTimeout returns a wrapper function that uses provider to source a
// deadline for the context of the run function it is given to wrap.
func RunWithTimeout(provider cmdcontext.TimeoutProvider) WrapRunEWithContext {
	return func(next RunEWithContext) RunEWithContext {
		return func(ctx context.Context, cmd *cobra.Command, args []string) error {
			ctx, cancel := cmdcontext.WithTimeoutFromConfig(ctx, provider)
			defer cancel()

			return next(ctx, cmd, args)
		}
	}
}

// AuthenticateClient returns a wrapper function that uses provider to source
// the credentials which it authenticates client with before executing the run
// function it is given to wrap.
func AuthenticateClient(provider CredentialsProvider, client AuthAPIClient) WrapRunEWithContext {
	return func(next RunEWithContext) RunEWithContext {
		return func(ctx context.Context, cmd *cobra.Command, args []string) error {
			username, err := provider.Username()
			if err != nil {
				return err
			}

			password, err := provider.Password()
			if err != nil {
				return err
			}

			_, err = client.Authenticate(ctx, username, password)
			if err != nil {
				return err
			}

			return next(ctx, cmd, args)
		}
	}
}

// EnsureTargetOrSelectors returns a wrapper function that wraps next if
// both selectors and args have non-zero length.
func EnsureTargetOrSelectors(selectors *[]string) WrapRunEWithContext {
	return func(next RunEWithContext) RunEWithContext {
		return func(ctx context.Context, cmd *cobra.Command, args []string) error {
			if len(*selectors) > 0 && len(args) > 0 {
				return ErrTargetOrSelector
			}

			return next(ctx, cmd, args)
		}
	}
}

// EnsureNamespaceSetWhenUseIDs returns a wrapper function that ensures that
// a target namespace is configured when using IDs.
func EnsureNamespaceSetWhenUseIDs(provider NamespacedCommandConfigProvider) WrapRunEWithContext {
	return func(next RunEWithContext) RunEWithContext {
		return func(ctx context.Context, cmd *cobra.Command, args []string) error {
			useIDs, err := provider.UseIDs()
			if err != nil {
				return err
			}

			namespace, err := provider.Namespace()
			if err != nil {
				return err
			}

			if useIDs && namespace == config.DefaultNamespaceName {
				return ErrMustSpecifyNamespaceID
			}

			return next(ctx, cmd, args)
		}
	}
}
