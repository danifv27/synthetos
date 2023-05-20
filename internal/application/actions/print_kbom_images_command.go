package actions

import (
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// PrintImagesRequest query params
type PrintImagesRequest struct {
	Mode      printer.PrinterMode
	ReceiveCh <-chan provider.Image
}

type PrintImagesCommand interface {
	Handle(request PrintImagesRequest) error
}

// Implements PrintImagesCommand interface
type printImagesCommand struct {
	lgr     logger.Logger
	printer printer.Printer
}

// NewPrintImagesCommandHandler Handler Constructor
func NewPrintImagesCommandHandler(l logger.Logger, p printer.Printer) PrintImagesCommand {

	return printImagesCommand{
		lgr:     l,
		printer: p,
	}
}

func (h printImagesCommand) Handle(request PrintImagesRequest) error {
	var err, rcerror error

	if request.Mode != printer.PrinterModeNone {
		if err = h.printer.ListKbomImages(request.ReceiveCh, request.Mode); err != nil {
			return errortree.Add(rcerror, "Handle", err)
		}
	}

	return nil
}
