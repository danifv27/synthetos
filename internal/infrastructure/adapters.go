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
	"fry.org/cmo/cli/internal/application/exporters"
	"fry.org/cmo/cli/internal/application/healthchecker"
	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"fry.org/cmo/cli/internal/application/version"

	ihealthchecker "fry.org/cmo/cli/internal/infrastructure/endpoints/healthchecker"
	iexporters "fry.org/cmo/cli/internal/infrastructure/exporters"
	ikms "fry.org/cmo/cli/internal/infrastructure/kms"
	ilogger "fry.org/cmo/cli/internal/infrastructure/logger"
	itableprinter "fry.org/cmo/cli/internal/infrastructure/printer"
	iprovider "fry.org/cmo/cli/internal/infrastructure/provider"
	istorage "fry.org/cmo/cli/internal/infrastructure/storage"
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
	version.Version
	printer.Printer
	healthchecker.Healthchecker
	exporters.CucumberExporter
	kms.KeyManager
	provider.ResourceProvider
}

// NewAdapters
func NewAdapters(opts ...AdapterOption) (Adapters, error) {
	var rcerror error

	a := Adapters{}

	// Loop through each option
	for _, option := range opts {
		if err := option.Apply(&a); err != nil {
			return Adapters{}, errortree.Add(rcerror, "NewAdapters", err)
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
			return errortree.Add(rcerror, "AdapterWithOptions", err)
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
			return errortree.Add(rcerror, "WithLogger", err)
		}

		return nil
	})
}

func WithVersion(URI string) AdapterOption {

	return AdapterOptionFunc(func(a *Adapters) error {
		var err, rcerror error

		if a.Version, err = istorage.Parse(URI); err != nil {
			return errortree.Add(rcerror, "WithVersion", err)
		}

		return nil
	})
}

func WithTablePrinter() AdapterOption {

	return AdapterOptionFunc(func(a *Adapters) error {
		var err, rcerror error

		options := []itableprinter.PrinterOption{}

		if a.Printer, err = itableprinter.NewPrinter(options...); err != nil {
			return errortree.Add(rcerror, "WithTablePrinter", err)
		}

		return nil
	})
}

func WithHealthchecker(root string) AdapterOption {

	return AdapterOptionFunc(func(a *Adapters) error {

		a.Healthchecker = ihealthchecker.NewHealthchecker(root)

		return nil
	})
}

func WithCucumberExporter(opts ...iexporters.ExporterOption) AdapterOption {

	return AdapterOptionFunc(func(a *Adapters) error {
		var err error

		a.CucumberExporter, err = iexporters.NewCucumberExporter(opts...)

		return err
	})
}

func WithKeyManager(url string, l logger.Logger) AdapterOption {

	return AdapterOptionFunc(func(a *Adapters) error {
		var err error

		a.KeyManager, err = ikms.Parse(url, l)

		return err
	})
}

func WithResourceProvider(url string, l logger.Logger) AdapterOption {

	return AdapterOptionFunc(func(a *Adapters) error {
		var err error

		a.ResourceProvider, err = iprovider.Parse(url, l)

		return err
	})
}
