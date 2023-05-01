package actions

import (
	"context"

	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"fry.org/cmo/cli/internal/application/provider"
	"github.com/speijnik/go-errortree"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ShowSummaryRequest query params
type ShowSummaryRequest struct {
	Mode     printer.PrinterMode
	Location string
	Selector string
}

type ShowSummaryResult struct{}

type ShowSummaryQueryHandler interface {
	Handle(request ShowSummaryRequest) (ShowSummaryResult, error)
}

// Implements ShowSummaryHandler interface
type showSummaryQueryHandler struct {
	lgr      logger.Logger
	print    printer.Printer
	provider provider.ResourceProvider
}

// NewShowSummaryQueryHandler Handler Constructor
func NewShowSummaryQueryHandler(l logger.Logger, p printer.Printer, pr provider.ResourceProvider) ShowSummaryQueryHandler {

	return showSummaryQueryHandler{
		lgr:      l,
		print:    p,
		provider: pr,
	}
}

func (h showSummaryQueryHandler) Handle(request ShowSummaryRequest) (ShowSummaryResult, error) {
	var err, rcerror error
	var rc ShowSummaryResult
	var uitems []*unstructured.Unstructured

	ctx := context.Background()

	if uitems, err = h.provider.GetResources(ctx, request.Location, request.Selector); err != nil {
		return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
	}
	if request.Mode != printer.PrinterModeNone {
		for _, u := range uitems {
			h.lgr.WithFields(logger.Fields{
				"item": u,
			}).Debug("Kustomization resource")
		}
		// 	if err = h.print.ListKmsGroups(rc.Groups, request.Mode); err != nil {
		// 		return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
		// 	}
	}

	return rc, nil
}
