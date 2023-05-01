package actions

import (
	"context"
	"errors"

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

type ShowSummaryResult struct {
	items []provider.Summary
}

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

func summarize(u *unstructured.Unstructured) (provider.Summary, error) {
	var rcerror error

	s := provider.Summary{
		APIVersion: u.GetAPIVersion(),
		Kind:       u.GetKind(),
	}

	nameRaw := u.GetName()
	if nameRaw != "" {
		s.Name = nameRaw
		return s, nil
	}

	generateNameRaw := u.GetGenerateName()
	if generateNameRaw != "" {
		s.Name = generateNameRaw
		return s, nil
	}

	return provider.Summary{}, errortree.Add(rcerror, "summarize", errors.New("unable to find object name"))
}

func (h showSummaryQueryHandler) Handle(request ShowSummaryRequest) (ShowSummaryResult, error) {
	var err, rcerror error
	var resources []*unstructured.Unstructured

	ctx := context.Background()
	rc := ShowSummaryResult{
		items: make([]provider.Summary, 0),
	}
	if resources, err = h.provider.GetResources(ctx, request.Location, request.Selector); err != nil {
		return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
	}
	for _, r := range resources {
		if s, err := summarize(r); err != nil {
			return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
		} else {
			rc.items = append(rc.items, s)
		}
	}
	if request.Mode != printer.PrinterModeNone {
		if err = h.print.PrintResourceSummary(rc.items, request.Mode); err != nil {
			return ShowSummaryResult{}, errortree.Add(rcerror, "Handle", err)
		}
	}

	return rc, nil
}
