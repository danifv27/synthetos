package actions

import (
	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"github.com/speijnik/go-errortree"
)

// PrintSecretRequest query params
type PrintSecretRequest struct {
	Mode      printer.PrinterMode
	ReceiveCh <-chan kms.Secret
	Decode    bool
}

type PrintSecretCommand interface {
	Handle(request PrintSecretRequest) error
}

// Implements PrintSecretCommand interface
type printSecretCommand struct {
	lgr     logger.Logger
	printer printer.Printer
}

// NewPrintSecretCommandHandler Handler Constructor
func NewPrintSecretCommandHandler(l logger.Logger, p printer.Printer) PrintSecretCommand {

	return printSecretCommand{
		lgr:     l,
		printer: p,
	}
}

func (h printSecretCommand) Handle(request PrintSecretRequest) error {
	var err, rcerror error

	if request.Mode != printer.PrinterModeNone {
		if err = h.printer.ListKmsSecrets(request.ReceiveCh, request.Mode); err != nil {
			return errortree.Add(rcerror, "Handle", err)
		}
	}

	return nil
}
