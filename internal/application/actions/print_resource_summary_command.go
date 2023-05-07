package actions

import (
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// PrintResourceSummaryRequest query params
type PrintResourceSummaryRequest struct {
	Mode printer.PrinterMode
	Ch   <-chan provider.Summary
}

type PrintResourceSummaryCommand interface {
	Handle(request PrintResourceSummaryRequest) error
}

// Implements PrintResourceSummaryCommand interface
type printResourceSummaryCommand struct {
	lgr     logger.Logger
	printer printer.Printer
}

// NewPrintResourceSummaryCommandHandler Handler Constructor
func NewPrintResourceSummaryCommandHandler(l logger.Logger, p printer.Printer) PrintResourceSummaryCommand {

	return printResourceSummaryCommand{
		lgr:     l,
		printer: p,
	}
}

func (h printResourceSummaryCommand) Handle(request PrintResourceSummaryRequest) error {
	var err, rcerror error

	if request.Mode != printer.PrinterModeNone {
		if err = h.printer.PrintResourceSummary(request.Ch, request.Mode); err != nil {
			return errortree.Add(rcerror, "Handle", err)
		}
	}

	return nil
}
