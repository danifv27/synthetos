// The application layer exposes all supported use cases of the application to the outside world.
// It consists of:

// 	Business logic/Use Cases
// 		Implementation of business requirements
// 		We can implement this with command/query separation. We cover this in our sample application.
// 	Application Services
// 		They provide isolated business logic/use cases functionality that is required. This functionality, is expressed by use cases.
// 		It can be an interface-only service if it is infrastructure-dependent

// The application layer code depends only on the domain layer.
package application

import (
	"fry.org/cmo/cli/internal/application/actions"
	"fry.org/cmo/cli/internal/application/healthchecker"
	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"fry.org/cmo/cli/internal/application/version"
	"github.com/speijnik/go-errortree"
)

// An Option applies optional changes to the Kong application.
type ApplicationOption interface {
	Apply(a *Applications) error
}

// AdapterOptionFunc is function that adheres to the Option interface.
type ApplicationOptionFunc func(a *Applications) error

func (o ApplicationOptionFunc) Apply(a *Applications) error {
	return o(a)
}

// Queries operations that request data
type Queries struct {
	ListGroups    actions.ListGroupsQueryHandler
	ListSecrets   actions.ListSecretsQuery
	DecryptSecret actions.DecryptSecretQuery
	ShowSummary   actions.ShowSummaryQueryHandler
	ListManifests actions.ListManifestsObjectsQuery
}

// Commands operations that accept data to make a change or trigger an action
type Commands struct {
	PrintVersion         actions.PrintVersionCommandHandler
	PrintResourceSummary actions.PrintResourceSummaryCommand
	PrintSecret          actions.PrintSecretCommand
}

// Applications contains all exposed services of the application layer
type Applications struct {
	logger.Logger
	healthchecker.Healthchecker
	Queries  Queries
	Commands Commands
}

// NewApplications bootstraps Application Layer dependencies
func NewApplications(opts ...ApplicationOption) (Applications, error) {
	var rcerror error

	a := Applications{}

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&a); err != nil {
			return Applications{}, errortree.Add(rcerror, "NewApplications", err)
		}
	}

	return a, nil
}

// WithOptions
func WithOptions(a *Applications, opts ...ApplicationOption) error {
	var rcerror error

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(a); err != nil {
			return errortree.Add(rcerror, "WithOptions", err)
		}
	}

	return nil
}

func WithLogger(l logger.Logger) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Logger = l

		return nil
	})
}

func WithHealthchecker(h healthchecker.Healthchecker) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Healthchecker = h

		return nil
	})
}

func WithPrintVersionCommand(v version.Version, p printer.Printer) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Commands.PrintVersion = actions.NewPrintVersionCommandHandler(v, p)

		return nil
	})
}

func WithPrintSecretCommand(l logger.Logger, p printer.Printer) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Commands.PrintSecret = actions.NewPrintSecretCommandHandler(l, p)

		return nil
	})
}

func WithPrintResourceSummaryCommand(l logger.Logger, p printer.Printer) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Commands.PrintResourceSummary = actions.NewPrintResourceSummaryCommandHandler(l, p)

		return nil
	})
}

func WithListGroupsQuery(l logger.Logger, p printer.Printer, k kms.KeyManager) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Queries.ListGroups = actions.NewListGroupsQueryHandler(l, p, k)

		return nil
	})
}

func WithListSecretsQuery(l logger.Logger, k kms.KeyManager) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Queries.ListSecrets = actions.NewListSecretsQueryHandler(l, k)

		return nil
	})
}

func WithShowSummaryQuery(l logger.Logger, pr provider.ResourceProvider) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Queries.ShowSummary = actions.NewShowSummaryQueryHandler(l, pr)

		return nil
	})
}

func WithDecryptSecretsQuery(l logger.Logger, k kms.KeyManager) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Queries.DecryptSecret = actions.NewDecryptSecretQueryHandler(l, k)

		return nil
	})
}
func WithListManifestsCommand(l logger.Logger, pr provider.ManifestProvider) ApplicationOption {

	return ApplicationOptionFunc(func(a *Applications) error {

		a.Queries.ListManifests = actions.NewListManifestsObjectsQueryHandler(l, pr)

		return nil
	})
}
