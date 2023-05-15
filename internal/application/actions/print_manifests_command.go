package actions

import (
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// PrintManifestsRequest query params
type PrintManifestsRequest struct {
	ReceiveCh <-chan provider.Manifest
}

type PrintManifestsCommand interface {
	Handle(request PrintManifestsRequest) error
}

// Implements PrintManifestCommand interface
type printManifestsCommand struct {
	lgr     logger.Logger
	printer printer.Printer
}

// NewPrintManifestsCommandHandler Handler Constructor
func NewPrintManifestsCommandHandler(l logger.Logger, p printer.Printer) PrintManifestsCommand {

	return printManifestsCommand{
		lgr:     l,
		printer: p,
	}
}

func (h printManifestsCommand) Handle(request PrintManifestsRequest) error {
	var err, rcerror error

	if err = h.printer.ListManifests(request.ReceiveCh); err != nil {
		return errortree.Add(rcerror, "Handle", err)
	}

	return nil
}
