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
	"fmt"

	"fry.org/cmo/cli/internal/application/logger"
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
}

// Commands operations that accept data to make a change or trigger an action
type Commands struct {
}

// Applications contains all exposed services of the application layer
type Applications struct {
	logger.Logger
	Queries  Queries
	Commands Commands
}

// NewApplications bootstraps Application Layer dependencies
func NewApplications(opts ...ApplicationOption) (Applications, error) {

	a := Applications{}

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&a); err != nil {
			return Applications{}, fmt.Errorf("NewApplications: %w", err)
		}
	}

	return a, nil
}

// WithOptions
func WithOptions(a *Applications, opts ...ApplicationOption) error {

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(a); err != nil {
			return fmt.Errorf("WithOptions: %w", err)
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
