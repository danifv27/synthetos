package actions

import (
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// PrintResourcesRequest query params
type PrintResourcesRequest struct {
	Mode      printer.PrinterMode
	ReceiveCh <-chan provider.ResourceList
}

type PrintResourcesCommand interface {
	Handle(request PrintResourcesRequest) error
}

// Implements PrintResourcesCommand interface
type printResourcesCommand struct {
	lgr     logger.Logger
	printer printer.Printer
}

// NewPrintResourcesCommandHandler Handler Constructor
func NewPrintResourcesCommandHandler(l logger.Logger, p printer.Printer) PrintResourcesCommand {

	return printResourcesCommand{
		lgr:     l,
		printer: p,
	}
}

func (h printResourcesCommand) Handle(request PrintResourcesRequest) error {
	var err, rcerror error

	if request.Mode != printer.PrinterModeNone {
		if err = h.printer.ListKbomResources(request.ReceiveCh, request.Mode); err != nil {
			return errortree.Add(rcerror, "Handle", err)
		}
	}

	return nil
}
