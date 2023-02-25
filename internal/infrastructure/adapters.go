// Package adapters provides different layers to interact with the external world.
// Typically, software applications have the following behavior:
// 		1. Receive input to initiate an operation
// 		2. Interact with infrastructure services to complete a function or produce output.
// The entry points from which we receive information (i.e., requests) are called input ports.
// The gateways through which we integrate with external services are called interface adapters.

// Input Ports and interface adapters depend on frameworks/platforms and external services that are not part of the business or domain logic.
// For this reason, they belong to this separate layer named infrastructure. This layer is sometimes referred to as Ports & Adapters.
// Interface Adapters
//
//	The interface adapters are responsible for implementing domain and application services(interfaces) by integrating with specific frameworks/providers.
//	For example, we can use a SQL provider to implement a domain repository or integrate with an email/SMS provider to implement a Notification service.
//
// Input ports
//
//	The input ports provide the entry points of the application that receive input from the outside world.
//	For example, an input port could be an HTTP handler handling synchronous calls or a Kafka consumer handling asynchronous messages.
//
// The infrastructure layer interacts with the application layer only.
package infrastructure

import (
	"fry.org/cmo/cli/internal/application/logger"

	ilogger "fry.org/cmo/cli/internal/infrastructure/logger"

	"github.com/speijnik/go-errortree"
)

// An AdapterOption applies optional changes to the Kong application.
type AdapterOption interface {
	Apply(a *Adapters) error
}

// AdapterOptionFunc is function that adheres to the Option interface.
type AdapterOptionFunc func(a *Adapters) error

func (o AdapterOptionFunc) Apply(a *Adapters) error {
	return o(a)
}

// Adapters contains the exposed adapters of interface adapters
type Adapters struct {
	logger.Logger
}

// NewAdapters
func NewAdapters(opts ...AdapterOption) (Adapters, error) {
	var rcerror error

	a := Adapters{}

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&a); err != nil {
			rcerror = errortree.Add(rcerror, "infrastructure", err)
			return Adapters{}, rcerror
			// return Adapters{}, fmt.Errorf("NewAdapters: %w", err)
		}
	}

	return a, nil
}

// AdapterWithOptions
func AdapterWithOptions(a *Adapters, opts ...AdapterOption) error {
	var rcerror error

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(a); err != nil {
			rcerror = errortree.Add(rcerror, "infrastructure", err)
			return rcerror
			// return fmt.Errorf("AdapterWithOptions: %w", err)
		}
	}

	return nil
}

// WithLogger sets the logger .
func WithLogger(URI string) AdapterOption {
	return AdapterOptionFunc(func(a *Adapters) error {
		var err, rcerror error

		if a.Logger, err = ilogger.Parse(URI); err != nil {
			return errortree.Add(rcerror, "logger", err)
		}

		return nil
	})
}
