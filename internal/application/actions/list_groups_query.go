package actions

import (
	"context"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"fry.org/cmo/cli/internal/application/printer"
	"github.com/speijnik/go-errortree"
)

// ListGroupsRequest query params
type ListGroupsRequest struct {
	Mode printer.PrinterMode
}

type ListGroupsResult struct {
	Groups []kms.Group
}

type ListGroupsQueryHandler interface {
	Handle(request ListGroupsRequest) (ListGroupsResult, error)
}

// Implements ListGroupsHandler interface
type listGroupsQueryHandler struct {
	lgr   logger.Logger
	kmngr kms.KeyManager
	print printer.Printer
}

// NewListGroupsQueryHandler Handler Constructor
func NewListGroupsQueryHandler(l logger.Logger, p printer.Printer, k kms.KeyManager) ListGroupsQueryHandler {

	return listGroupsQueryHandler{
		lgr:   l,
		kmngr: k,
		print: p,
	}
}

func (h listGroupsQueryHandler) Handle(request ListGroupsRequest) (ListGroupsResult, error) {
	var err, rcerror error
	var rc ListGroupsResult

	ctx := context.Background()

	if rc.Groups, err = h.kmngr.ListGroups(ctx); err != nil {
		return ListGroupsResult{}, errortree.Add(rcerror, "Handle", err)
	}
	if request.Mode != printer.PrinterModeNone {
		if err = h.print.ListKmsGroups(rc.Groups, request.Mode); err != nil {
			return ListGroupsResult{}, errortree.Add(rcerror, "Handle", err)
		}
	}

	return rc, nil
}
