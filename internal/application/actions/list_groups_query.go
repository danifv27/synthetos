package actions

import (
	"context"
	"errors"

	"fry.org/cmo/cli/internal/application/kms"
	"fry.org/cmo/cli/internal/application/logger"
	"github.com/speijnik/go-errortree"
)

// ListGroupsRequest query params
type ListGroupsRequest struct {
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
}

// NewListGroupsQueryHandler Handler Constructor
func NewListGroupsQueryHandler(l logger.Logger, k kms.KeyManager) ListGroupsQueryHandler {

	return listGroupsQueryHandler{
		lgr:   l,
		kmngr: k,
	}
}

func (h listGroupsQueryHandler) Handle(request ListGroupsRequest) (ListGroupsResult, error) {
	var rcerror error

	ctx := context.Background()
	if err := h.kmngr.List(ctx); err != nil {
		return ListGroupsResult{}, errortree.Add(rcerror, "Handle", err)
	}

	return ListGroupsResult{}, errortree.Add(rcerror, "Handle", errors.New("method not implemented"))
}
