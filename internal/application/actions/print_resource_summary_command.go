package actions

import (
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
)

// PrintResourceSummaryRequest query params
type PrintResourceSummaryRequest struct {
	Mode  printer.PrinterMode
	Items []provider.Summary
}

type PrintResourceSummaryResult struct{}

type PrintResourceSummaryCommand interface {
	Handle(request PrintResourceSummaryRequest) (PrintResourceSummaryResult, error)
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

func (h printResourceSummaryCommand) Handle(request PrintResourceSummaryRequest) (PrintResourceSummaryResult, error) {
	var err, rcerror error

	if request.Mode != printer.PrinterModeNone {
		if err = h.printer.PrintResourceSummary(request.Items, request.Mode); err != nil {
			return PrintResourceSummaryResult{}, errortree.Add(rcerror, "Handle", err)
		}
	}

	return PrintResourceSummaryResult{}, nil
}
