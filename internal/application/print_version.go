package application

import (
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/version"
	"github.com/speijnik/go-errortree"
)

type PrintVersionRequest struct {
	Format string
}

type PrintVersionCommandHandler interface {
	Handle(command PrintVersionRequest) error
}

type printVersionCommandHandler struct {
	v version.Version
	p printer.Printer
}

// NewPrintVersionCommandHandler Constructor
func NewPrintVersionCommandHandler(version version.Version, printer printer.Printer) PrintVersionCommandHandler {

	return printVersionCommandHandler{
		v: version,
		p: printer,
	}
}

// Handle Handles the update request
func (h printVersionCommandHandler) Handle(command PrintVersionRequest) error {
	var err, rcerror error
	var mode printer.PrinterMode

	if command.Format == "json" {
		mode = printer.PrinterModeJSON
	} else {
		mode = printer.PrinterModeText
	}
	if err = h.p.PrintVersion(h.v, mode); err != nil {
		return errortree.Add(rcerror, "Handle", err)
	}

	return nil
}
