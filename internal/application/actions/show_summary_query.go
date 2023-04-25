package actions

import (
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
)

// ShowSummaryRequest query params
type ShowSummaryRequest struct {
	Mode printer.PrinterMode
}

type ShowSummaryResult struct{}

type ShowSummaryQueryHandler interface {
	Handle(request ShowSummaryRequest) (ShowSummaryResult, error)
}

// Implements ShowSummaryHandler interface
type showSummaryQueryHandler struct {
	lgr   logger.Logger
	print printer.Printer
}

// NewShowSummaryQueryHandler Handler Constructor
func NewShowSummaryQueryHandler(l logger.Logger, p printer.Printer) ShowSummaryQueryHandler {

	return showSummaryQueryHandler{
		lgr:   l,
		print: p,
	}
}

func (h showSummaryQueryHandler) Handle(request ShowSummaryRequest) (ShowSummaryResult, error) {
	// var err, rcerror error
	var rc ShowSummaryResult

	// ctx := context.Background()

	// if rc.Groups, err = h.kmngr.ShowSummary(ctx); err != nil {
	// 	return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
	// }
	// if request.Mode != printer.PrinterModeNone {
	// 	if err = h.print.ListKmsGroups(rc.Groups, request.Mode); err != nil {
	// 		return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
	// 	}
	// }

	return rc, nil
}
